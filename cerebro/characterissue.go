package cerebro

import (
	"errors"
	"github.com/aimeelaplant/comiccruncher/comic"
	"github.com/aimeelaplant/comiccruncher/internal/listutil"
	"github.com/aimeelaplant/comiccruncher/internal/log"
	"github.com/aimeelaplant/externalissuesource"
	"github.com/avast/retry-go"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"
)

// Concurrency limit for fetching issues from an external source.
const jobLimit = 10

// An error returned from the http client. Unfortunately it has no variable associated with it.
const errClientTimeoutString = "Client.Timeout exceeded"

var (
	// Maps formats from the external source to our own formats.
	externalToLocalFormatMap = map[externalissuesource.Format]comic.Format{
		externalissuesource.Unknown:      comic.FormatUnknown,
		externalissuesource.Standard:     comic.FormatStandard,
		externalissuesource.TPB:          comic.FormatTPB,
		externalissuesource.Manga:        comic.FormatManga,
		externalissuesource.HC:           comic.FormatHC,
		externalissuesource.OGN:          comic.FormatOGN,
		externalissuesource.Web:          comic.FormatWeb,
		externalissuesource.Magazine:     comic.FormatMagazine,
		externalissuesource.DigitalMedia: comic.FormatDigitalMedia,
		externalissuesource.MiniComic:    comic.FormatMiniComic,
		externalissuesource.Flipbook:     comic.FormatFlipbook,
		externalissuesource.Anthology:    comic.FormatAnthology,
		externalissuesource.Prestige:     comic.FormatPrestige,
	}
	// The types of formats that count as an appearance for a character.
	countsAsAppearance = map[comic.Format]bool{
		comic.FormatStandard:     true,
		comic.FormatOGN:          true,
		comic.FormatMiniComic:    true,
		comic.FormatAnthology:    true,
		comic.FormatFlipbook:     true,
		comic.FormatWeb:          true,
		comic.FormatDigitalMedia: true,
		comic.FormatManga:        true,
		comic.FormatPrestige:     true,
	}
)

// CharacterIssueImporter is the importer for getting a character's issues from a character source.
type CharacterIssueImporter struct {
	appearanceSyncer comic.Syncer
	characterSvc     comic.CharacterServicer
	issueSvc         comic.IssueServicer
	externalSource   externalissuesource.ExternalSource
	logger           *zap.Logger
}

// ImportWithSyncLog does A LOT. It's for importing a character's issues with an existing sync log attached.
// Imports a character's issues from their character sources and polls an external source for issue information
// and then persists the character's appearances to the db and Redis.
// A channel is opened listening for a SIGINT if the caller quits the process.
// In that case, the character sync log is set to failed and the process quits cleanly.
func (i *CharacterIssueImporter) ImportWithSyncLog(character comic.Character, syncLog *comic.CharacterSyncLog) error {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, os.Kill)
	go func() error {
		for sig := range sigCh {
			if sig == os.Interrupt || sig == os.Kill {
				close(sigCh)
				i.updateSyncLog(syncLog, comic.Fail, sigCh)
				// TODO: ok, so it doesn't clean it _that_ gracefully ... would have to close the channels that get opened
				// below for that to happen.
				i.logger.Info("cleaned up gracefully and failed the sync log.", zap.Uint("sync log ID", syncLog.ID.Value()))
				os.Exit(0)
			}
		}
		return nil
	}()
	// Set to in progress.
	i.updateSyncLog(syncLog, comic.InProgress, nil)
	i.logger.Info("started import", zap.String("character", character.Slug.Value()))
	if character.IsDisabled {
		i.updateSyncLog(syncLog, comic.Fail, sigCh)
		// it's not a pressing error so just log and return nil.
		i.logger.Warn("the character is disabled. won't sync appearances", zap.String("character", character.Slug.Value()))
		return nil
	}
	sources, err := i.characterSvc.Sources(character.ID, comic.VendorTypeCb, nil)
	if err != nil || len(sources) == 0 {
		i.updateSyncLog(syncLog, comic.Fail, nil)
		if len(sources) == 0 {
			i.logger.Warn("cannot import issues for character. no sources for character", zap.String("character", character.Slug.Value()))
		}
		return err
	}

	// Links to issues we have to actually go and fetch if we don't have them.
	linksToFetch := make([]string, 0)
	// A map containing the vendor ID of an issue and the link to the issue.
	vendorIdsMap := make(map[string]string)
	// A map containing vendor IDs for each source that is marked as a main source for a character.
	mainSourcesMap := make(map[string]bool)
	// A map containing vendor ID's marked for alternate appearances.
	altSourcesMap := make(map[string]bool)
	for _, source := range sources {
		page, err := i.requestCharacterPage(source.VendorURL)
		if err != nil {
			i.logger.Warn("error getting character page. skipping", zap.Error(err), zap.String("character", character.Slug.Value()), zap.String("source", source.VendorURL))
			// Skip.
			continue
		}
		i.logger.Info(
			"got issue links from source",
			zap.Int("issue links", len(page.IssueLinks)),
			zap.String("source", source.VendorURL),
			zap.String("vendor name", source.VendorName),
			zap.String("character", character.Slug.Value()))
		for _, l := range page.IssueLinks {
			idIndex := strings.Index(l, "=")
			if idIndex != -1 {
				vendorID := l[idIndex+1:]
				vendorIdsMap[vendorID] = l
				// If it's a main source, then put it in the `mainSourcesMap` so we can reference it later as a main
				// issue for a character. Note the vendor id can in both a main source or alternate source.
				if source.IsMain {
					mainSourcesMap[vendorID] = true
				} else {
					altSourcesMap[vendorID] = true
				}
			}
		}
	}
	// Find all the issues that we could have in the database.
	localIssues, err := i.issueSvc.IssuesByVendor(listutil.StringKeys(vendorIdsMap), comic.VendorTypeCb, 0, 0)
	if err != nil {
		i.updateSyncLog(syncLog, comic.Fail, sigCh)
		return err
	}
	// a map of issues we have with the associated external vendor id.
	localIssueVendorIds := make(map[string]bool)
	// character issues we don't have and need to create
	characterIssues := make([]*comic.CharacterIssue, 0)
	// Determines the appearance type from an issue.
	appearanceType := func(issue *comic.Issue) comic.AppearanceType {
		if mainSourcesMap[issue.VendorID] && altSourcesMap[issue.VendorID] {
			return comic.Main | comic.Alternate
		} else if mainSourcesMap[issue.VendorID] {
			return comic.Main
		} else {
			return comic.Alternate
		}
	}

	for _, localIssue := range localIssues {
		// add the vendor id to the map
		localIssueVendorIds[localIssue.VendorID] = true
		// create a character issue if it counts as an appearance
		if isAppearance(localIssue, character.Publisher.Slug) {
			characterIssues = append(characterIssues, comic.NewCharacterIssue(character.ID, localIssue.ID, appearanceType(localIssue)))
		}
	}
	// Insert ignore what we possibly don't have.
	i.logger.Info("inserting what we possibly don't have", zap.Int("character issues", len(characterIssues)))
	if err := i.characterSvc.CreateIssues(characterIssues); err != nil {
		i.updateSyncLog(syncLog, comic.Fail, sigCh)
		return err
	}
	// release the resource. no longer need the slice
	characterIssues = nil
	for k, v := range vendorIdsMap {
		// if the vendor id isn't in the map of local issues, then we want to put the link in the `links` slice.
		// it means we don't have the issue.
		if !localIssueVendorIds[k] {
			linksToFetch = append(linksToFetch, v)
		}
	}

	// release the resource. no longer needed.
	localIssueVendorIds = nil
	i.logger.Info("Total issues we have", zap.Int("count", len(localIssues)), zap.String("character", character.Slug.Value()))
	i.logger.Info("Total issues we don't have and need to fetch", zap.Int("count", len(linksToFetch)), zap.String("character", character.Slug.Value()))

	if len(linksToFetch) > 0 {
		linkCh := make(chan string, len(linksToFetch))
		issueCh := make(chan *comic.Issue, len(linksToFetch))
		for w := 0; w < jobLimit; w++ {
			go i.requestIssues(w, linkCh, issueCh)
		}
		// Send the work over.
		for _, l := range linksToFetch {
			linkCh <- l
		}
		// Close the channel for no more sends.
		close(linkCh)

		// Collect the results of the work.
		for idx := 0; idx < len(linksToFetch); idx++ {
			ish := <-issueCh
			// Skip the issue if we get a blank one or the year is is less than one.
			if ish.VendorID == "" || ish.SaleDate.Year() <= 1 {
				i.logger.Warn("received blank issue. skipping.", zap.String("link", linksToFetch[idx]))
				continue
			}
			i.logger.Info("received issue", zap.String("issue.VendorId", ish.VendorID))
			// Ideally creating an issue and character issue shouldn't be in a for loop.
			// Maybe put chunks in a pool and do bulk creation.
			// There's too much info that gets lost if a caller quits the process and _ALL_ the issues don't get saved.
			// So we want to save incrementally here and not do an all-or-nothing transaction for _ALL_ issues that get fetched.
			if err := i.issueSvc.Create(ish); err != nil {
				close(issueCh)
				return err
			}
			if isAppearance(ish, character.Publisher.Slug) {
				if _, err := i.characterSvc.CreateIssueP(character.ID, ish.ID, appearanceType(ish), nil); err != nil {
					close(issueCh)
					return err
				}
			}
		}
		close(issueCh)
		i.logger.Info("issues to attempt to sync!", zap.Int("total", len(linksToFetch)), zap.String("character", character.Slug.Value()))
	}

	// Now send the new character issues over to redis.
	total, err := i.appearanceSyncer.Sync(character.Slug)
	if err != nil {
		i.updateSyncLog(syncLog, comic.Fail, sigCh)
		i.logger.Error("could not sync appearances", zap.Error(err))
		return err
	}
	// get the total appearances synced and use it as the message.
	syncLog.Message = strconv.Itoa(total)
	i.updateSyncLog(syncLog, comic.Success, sigCh)
	return nil
}

// ImportAll imports characters from the specified slugs and creates the sync log for each character and sets it to PENDING,
// then sequentially imports the issues for the character.
func (i *CharacterIssueImporter) ImportAll(slugs []comic.CharacterSlug) error {
	characters, err := i.characterSvc.CharactersWithSources(slugs, 0, 0)
	if err != nil {
		return err
	}
	// A channel for sync logs that are done being created.
	syncLogCh := make(chan *comic.CharacterSyncLog, len(characters))
	defer close(syncLogCh)
	for idx := range characters {
		go func(idx int) {
			character := characters[idx]
			// create the sync log
			syncLog := comic.NewSyncLogPending(character.ID, comic.YearlyAppearances)
			if err := i.characterSvc.CreateSyncLog(syncLog); err != nil {
				i.logger.Error("error creating sync log", zap.String("character", character.Slug.Value()), zap.Error(err))
				// send a blank one
				syncLogCh <- &comic.CharacterSyncLog{}
			} else {
				// attach the character to the sync log (a lil hacky)
				syncLog.Character = character
				// send the sync log over
				syncLogCh <- syncLog
			}
		}(idx)
	}
	// Read results from the channel.
	for idx := 0; idx < len(characters); idx++ {
		syncLog, ok := <-syncLogCh
		if ok && syncLog.Character != nil {
			character := syncLog.Character
			// start the import. one-by-one -- no concurrency here. maybe in the future if the external source can handle it. :)
			if err := i.ImportWithSyncLog(*character, syncLog); err != nil {
				i.logger.Error("error importing character issues", zap.String("character", character.Slug.Value()), zap.Error(err))
				if syncLog.SyncStatus != comic.Fail {
					// we need to fail it!
					i.logger.Info("failing sync log", zap.String("character", character.Slug.Value()))
					i.updateSyncLog(syncLog, comic.Fail, nil)
				}
			}
		}
	}
	return nil
}

// Persists the sync log with the new status and closes the signal channel if the new status is fail or success.
func (i *CharacterIssueImporter) updateSyncLog(cLog *comic.CharacterSyncLog, newStatus comic.CharacterSyncLogStatus, sigCh chan<- os.Signal) {
	if sigCh != nil && (newStatus == comic.Fail || newStatus == comic.Success) {
		// TODO: https://go101.org/article/channel-closing.html
		defer close(sigCh)
	}
	if newStatus == comic.Success {
		now := time.Now()
		cLog.SyncedAt = &now
	}
	cLog.SyncStatus = newStatus
	err := i.characterSvc.UpdateSyncLog(cLog)
	if err != nil {
		i.logger.Error("error updating sync log", zap.Error(err))
	}
}

// requestIssues requests issue information from an external source link (the caller sends string to the `links`) and then converts the
// external issue to our own model and sends it over to the `issues` chan (this method sends strings to the `issues` chan).
func (i *CharacterIssueImporter) requestIssues(workerID int, links <-chan string, issues chan<- *comic.Issue) {
	i.logger.Info("started worker", zap.Int("workerId", workerID))
	for l := range links {
		externalIssueCh := make(chan *externalissuesource.Issue, 1)
		retry.Do(func() error {
			i.logger.Info("fetching issue", zap.String("url", l))
			externalIssue, err := i.externalSource.Issue(l)
			if err != nil {
				if isConnectionError(err) {
					i.logger.Info("got connection issue. retrying", zap.String("url", l), zap.Error(err))
					return err
				}
				i.logger.Error("received error from external source", zap.Int("workerId", workerID), zap.String("link", l), zap.Error(err))
				// Send a blank issue
				issues <- &comic.Issue{}
				// close the channel. won't send anymore.
				close(externalIssueCh)
				// return nil to exit the retry.
				return nil
			}
			// send over the issue.
			externalIssueCh <- externalIssue
			return nil
		}, retryDelay)
		// read from it if the value was sent.
		if externalIssue, ok := <-externalIssueCh; ok {
			issueFormat := comic.FormatOther
			for k, v := range externalToLocalFormatMap {
				if k == externalIssue.Format {
					issueFormat = v
					break
				}
			}
			issues <- comic.NewIssue(
				externalIssue.Id, // the vendor ID
				externalIssue.Vendor,
				externalIssue.Series,
				externalIssue.Number,
				externalIssue.PublicationDate,
				externalIssue.OnSaleDate,
				externalIssue.IsVariant,
				externalIssue.MonthUncertain,
				issueFormat)
			i.logger.Info("finished job and worker", zap.String("url", l), zap.Int("workerId", workerID))
		}
	}
}

// requestCharacterPage requests a character source page and retries if there's a connection failure.
func (i *CharacterIssueImporter) requestCharacterPage(source string) (externalissuesource.CharacterPage, error) {
	pageChan := make(chan *externalissuesource.CharacterPage, 1)
	retry.Do(func() error {
		i.logger.Info("getting character page", zap.String("source", source))
		cPage, err := i.externalSource.CharacterPage(source)
		if err != nil {
			if isConnectionError(err) {
				i.logger.Info("got connection issue. retrying.", zap.String("source", source))
				return err
			}
			// Close the channel.
			close(pageChan)
			i.logger.Error("error from page", zap.String("source", source), zap.Error(err))
			// exit the retry
			return nil
		}
		// Send the page over
		pageChan <- cPage
		return nil
	}, retryDelay)
	if page, ok := <-pageChan; ok {
		return *page, nil
	}
	// return empty page
	return externalissuesource.CharacterPage{}, errors.New("couldn't get the page")
}

// isAppearance checks that the issue should count as an issue appearance for the character.
func isAppearance(issue *comic.Issue, slug comic.PublisherSlug) bool {
	if !issue.IsVariant && // it's not a variant
		!issue.IsReprint && // it's not a reprint.
		countsAsAppearance[issue.Format] && // the format counts as an appearance
		// Checks that the external issue's publisher matches up with the publisher of the character.
		// check their slugs since it is lower cased and just "marvel" or "dc".
		strings.Contains(strings.ToLower(issue.VendorPublisher), slug.Value()) &&
		// the issue actually has a sale date.
		issue.SaleDate.Year() > 1 {
		return true
	}
	return false
}

// NewCharacterIssueImporter creates a new character issue importer.
func NewCharacterIssueImporter(
	container *comic.PGRepositoryContainer,
	appearancesSyncer comic.Syncer,
	externalSource externalissuesource.ExternalSource) *CharacterIssueImporter {
	return &CharacterIssueImporter{
		characterSvc:     comic.NewCharacterService(container),
		issueSvc:         comic.NewIssueService(container),
		externalSource:   externalSource,
		appearanceSyncer: appearancesSyncer,
		logger:           log.CEREBRO(),
	}
}
