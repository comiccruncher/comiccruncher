package cerebro

import (
	"errors"
	"fmt"
	"github.com/aimeelaplant/comiccruncher/comic"
	"github.com/aimeelaplant/comiccruncher/internal/log"
	"github.com/aimeelaplant/comiccruncher/internal/stringutil"
	"github.com/aimeelaplant/externalissuesource"
	"github.com/avast/retry-go"
	"github.com/go-pg/pg"
	"github.com/go-redis/redis"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"
)

// Concurrency limit for fetching issues from an external source.
const jobLimit = 10

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
	// allowedPublishers is the list of publishers allowed to count as an appearance for a character.
	allowedPublishers = []string{
		"Archie",
		"Atlas Comics",
		"Dark Horse",
		"DC Comics",
		"Dynamite Entertainment",
		"Marvel",
		"IDW Publishing",
		"Image Comics",
		"Malibu",
		"Timely Comics",
	}
)

// ExternalVendorID is a vendor ID for a third-party vendor.
type ExternalVendorID string

// ExternalVendorURL is a vendor URL from a third-party.
type ExternalVendorURL string

// String gets the string value of the external URL.
func (u ExternalVendorURL) String() string {
	return string(u)
}

// CharacterVendorExtractor parses information about a vendor for a character.
type CharacterVendorExtractor interface {
	// Extract extracts character source information into CharacterVendorInfo.
	Extract(sources []*comic.CharacterSource) (CharacterVendorInfo, error)
}

// CharacterIssueImporter is the importer for getting a character's issues from a character source.
type CharacterIssueImporter struct {
	appearancesWriter comic.AppearancesByYearsWriter
	appearanceSyncer  comic.Syncer
	characterSvc      comic.CharacterServicer
	issueSvc          comic.IssueServicer
	externalSource    externalissuesource.ExternalSource
	extractor         CharacterVendorExtractor
	refresher         comic.PopularRefresher
	statsSyncer       comic.CharacterStatsSyncer
	logger            *zap.Logger
}

// CharacterVendorInfo contains information about a character's vendor IDs and
// the main and alternate sources associated to the character's sources.
type CharacterVendorInfo struct {
	// VendorIDs contains a map of VendorIDs and their corresponding URLs.
	VendorIDs map[ExternalVendorID]ExternalVendorURL
	// MainSources contains all the VendorIDs that are main sources.
	MainSources map[ExternalVendorID]bool
	// AltSources contains all the VendorIDs that are alternate sources.
	AltSources map[ExternalVendorID]bool
}

// CharacterCBExtractor parses a character's sources and into CharacterVendorInfo.
type CharacterCBExtractor struct {
	src    externalissuesource.ExternalSource
	logger *zap.Logger
}

// requestCharacterPage requests a character source page and retries if there's a connection failure.
func (p *CharacterCBExtractor) requestCharacterPage(source string) (externalissuesource.CharacterPage, error) {
	pageChan := make(chan *externalissuesource.CharacterPage, 1)
	retry.Do(func() error {
		p.logger.Info("getting character page", zap.String("source", source))
		cPage, err := p.src.CharacterPage(source)
		if err != nil {
			if isConnectionError(err) {
				p.logger.Info("got connection issue. retrying.", zap.String("source", source))
				return err
			}
			// Close the channel.
			close(pageChan)
			p.logger.Error("error from page", zap.String("source", source), zap.Error(err))
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

// Extract parses the vendor information from a character's many character sources.
func (p *CharacterCBExtractor) Extract(sources []*comic.CharacterSource) (CharacterVendorInfo, error) {
	ei := CharacterVendorInfo{}
	if len(sources) == 0 {
		return ei, fmt.Errorf("0 sources returned. no sources to import")
	}
	// A map containing the vendor ID of an issue and the link to the issue.
	vendorIDs := make(map[ExternalVendorID]ExternalVendorURL)
	// A map containing vendor IDs for each source that is marked as a main source for a character.
	mainSources := make(map[ExternalVendorID]bool)
	// A map containing vendor ID's marked for alternate appearances.
	altSources := make(map[ExternalVendorID]bool)
	for _, s := range sources {
		page, err := p.requestCharacterPage(s.VendorURL)
		if err != nil {
			p.logger.Warn("error getting character page. skipping", zap.Error(err), zap.String("source", s.VendorURL))
			// Skip.
			continue
		}
		p.logger.Info(
			"got issue links from source",
			zap.Int("issue links", len(page.IssueLinks)),
			zap.String("source", s.VendorURL),
			zap.String("vendor name", s.VendorName))
		for _, l := range page.IssueLinks {
			idIndex := strings.Index(l, "=")
			if idIndex == -1 {
				continue
			}
			vendorID := ExternalVendorID(l[idIndex+1:])
			vendorIDs[vendorID] = ExternalVendorURL(l)
			// If it's a main source, then put it in the `mainSourcesMap` so we can reference it later as a main
			// issue for a character. Note the vendor id can in both a main source or alternate source.
			if s.IsMain {
				mainSources[vendorID] = true
			} else {
				altSources[vendorID] = true
			}
		}
	}
	ei.AltSources = altSources
	ei.MainSources = mainSources
	ei.VendorIDs = vendorIDs
	return ei, nil
}

// AppearanceType determines the type of appearance from the issue's vendor ID.
// So if the issue's vendor ID is in the externalInfo's mainSources map, it's a main source, etc.
func (vi CharacterVendorInfo) AppearanceType(issue *comic.Issue) comic.AppearanceType {
	externalToLocalVendorID := ExternalVendorID(issue.VendorID)
	if vi.MainSources[externalToLocalVendorID] && vi.AltSources[externalToLocalVendorID] {
		return comic.Main | comic.Alternate
	} else if vi.MainSources[externalToLocalVendorID] {
		return comic.Main
	}
	return comic.Alternate
}

// vendorIDStrings gets all the vendor IDs from the `vendorIDs` attribute as a string slice.
func (vi CharacterVendorInfo) vendorIDStrings() []string {
	vendorIDs := make([]string, len(vi.VendorIDs))
	count := 0
	for vendorID := range vi.VendorIDs {
		vendorIDs[count] = string(vendorID)
		count++
	}
	return vendorIDs
}

// listenOnSigInt creates an os.Signal chan and fails the sync logs if the caller interrupts the process.
func (i *CharacterIssueImporter) listenOnSigInt(syncLogs []*comic.CharacterSyncLog) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, os.Kill)
	for sig := range sigCh {
		if sig == os.Interrupt || sig == os.Kill {
			for idx := range syncLogs {
				syncLog := syncLogs[idx]
				syncLog.Message = "user quit process"
				i.updateSyncLog(syncLog, comic.Fail)
				i.logger.Info("failed sync log", zap.String("character", syncLog.Character.Slug.Value()))
			}
			i.logger.Info("quitting.")
			os.Exit(-1)
		}
	}
}

// Gets all the links to the issues we do not have in the database.
func (i *CharacterIssueImporter) nonExistingURLs(vi CharacterVendorInfo, c comic.Character) ([]ExternalVendorURL, error) {
	// Find all the issues that we could have in the database.
	localIssues, err := i.issueSvc.IssuesByVendor(vi.vendorIDStrings(), comic.VendorTypeCb, 0, 0)
	if err != nil {
		return nil, err
	}
	// a map of issues we have with the associated external vendor id.
	localIssueVendorIDs := make(map[string]bool)
	// character issues we don't have and need to create
	characterIssues := make([]*comic.CharacterIssue, 0)
	for _, localIssue := range localIssues {
		// add the vendor id to the map
		localIssueVendorIDs[localIssue.VendorID] = true
		// create a character issue if it counts as an appearance
		if isAppearance(localIssue) {
			characterIssues = append(characterIssues, comic.NewCharacterIssue(c.ID, localIssue.ID, vi.AppearanceType(localIssue)))
		}
	}
	// Insert ignore what we possibly don't have.
	i.logger.Info("inserting what we possibly don't have", zap.Int("character issues", len(characterIssues)))
	if err := i.characterSvc.CreateIssues(characterIssues); err != nil {
		return nil, err
	}
	linksToFetch := make([]ExternalVendorURL, 0)
	for k, v := range vi.VendorIDs {
		// if the vendor id isn't in the map of local issues, then we want to put the link in the `links` slice.
		// it means we don't have the issue.
		if !localIssueVendorIDs[string(k)] {
			linksToFetch = append(linksToFetch, v)
		}
	}
	i.logger.Info("Total issues we have", zap.Int("count", len(localIssues)), zap.String("character", c.Slug.Value()))
	i.logger.Info("Total issues we don't have and need to fetch", zap.Int("count", len(linksToFetch)), zap.String("character", c.Slug.Value()))
	return linksToFetch, nil
}

// importIssues imports a character's issues from their character sources and polls an external source for issue information
// and then persists the character's appearances to the db and Redis.
func (i *CharacterIssueImporter) importIssues(character comic.Character, doReset bool) (int, error) {
	if doReset {
		res, err := i.characterSvc.RemoveIssues(character.ID)
		if err != nil {
			return 0, err
		}
		log.CEREBRO().Info("removed issues for character", zap.String("slug", character.Slug.Value()), zap.Int("total", res))
	}
	// Set to in progress.
	i.logger.Info("started import", zap.String("character", character.Slug.Value()))
	if character.IsDisabled {
		return 0, errors.New("won't sync appearances for disabled character")
	}
	sources, err := i.characterSvc.Sources(character.ID, comic.VendorTypeCb, nil)
	if err != nil {
		return 0, err
	}
	vi, err := i.extractor.Extract(sources)
	if err != nil {
		return 0, err
	}
	linksToFetch, err := i.nonExistingURLs(vi, character)
	if err != nil {
		return 0, err
	}
	linkCh := make(chan ExternalVendorURL, len(linksToFetch))
	defer close(linkCh)
	issueCh := make(chan *comic.Issue, len(linksToFetch))
	defer close(issueCh)
	for w := 0; w < jobLimit; w++ {
		go i.requestIssues(w, linkCh, issueCh)
	}
	// Send the work over.
	for _, l := range linksToFetch {
		linkCh <- l
	}
	// Collect the results of the work.
	for idx := 0; idx < len(linksToFetch); idx++ {
		ish := <-issueCh
		// Skip the issue if we get a blank one or the year is is less than one.
		if ish.VendorID == "" || ish.SaleDate.Year() <= 1 {
			i.logger.Warn("received blank issue. skipping.", zap.String("link", linksToFetch[idx].String()))
			continue
		}
		i.logger.Info("received issue", zap.String("issue.VendorId", ish.VendorID))
		// Ideally creating an issue and character issue shouldn't be in a for loop.
		// Maybe put chunks in a pool and do bulk creation.
		// There's too much info that gets lost if a caller quits the process and _ALL_ the issues don't get saved.
		// So we want to save incrementally here and not do an all-or-nothing transaction for _ALL_ issues that get fetched.
		if err := i.issueSvc.Create(ish); err != nil {
			return 0, err
		}
		if isAppearance(ish) {
			if _, err := i.characterSvc.CreateIssueP(character.ID, ish.ID, vi.AppearanceType(ish), nil); err != nil {
				return 0, err
			}
		}
	}
	i.logger.Info("issues to attempt to sync!", zap.Int("total", len(linksToFetch)), zap.String("character", character.Slug.Value()))
	if doReset {
		// delete all the issues first.
		res, err := i.appearancesWriter.Delete(character.Slug)
		if err != nil {
			log.CEREBRO().Error("error removing appearances from redis", zap.String("character", character.Slug.Value()), zap.Error(err))
		} else {
			log.CEREBRO().Info("deleted appearances from redis", zap.String("character", character.Slug.Value()), zap.Int64("total", res))
		}
	}
	// Now send the new character issues over to redis.
	total, err := i.appearanceSyncer.Sync(character.Slug)
	if err != nil {
		return 0, err
	}
	return total, nil
}

// ImportWithSyncLog does A LOT. It's for importing a character's issues with an existing sync log attached.
// Imports a character's issues from their character sources and polls an external source for issue information
// and then persists the character's appearances to the db and Redis.
// A channel is opened listening for a SIGINT if the caller quits the process.
// In that case, the character sync log is set to failed and the process quits cleanly.
func (i *CharacterIssueImporter) ImportWithSyncLog(character comic.Character, syncLog *comic.CharacterSyncLog, doReset bool) error {
	// Set to in progress.
	i.updateSyncLog(syncLog, comic.InProgress)
	total, err := i.importIssues(character, doReset)
	if err != nil {
		i.updateSyncLog(syncLog, comic.Fail)
		return err
	}
	syncLog.Message = strconv.Itoa(total)
	i.updateSyncLog(syncLog, comic.Success)
	return nil
}

// ImportAll imports characters from the specified slugs and creates the sync log for each character and sets it to PENDING,
// then sequentially imports the issues for the character.
// Fatals if failed to create a sync log or character cannot be fetched.
// If doReset is set to true, it will delete all associated character issues first and re-import new ones.
func (i *CharacterIssueImporter) ImportAll(slugs []comic.CharacterSlug, doReset bool) error {
	characters, err := i.characterSvc.CharactersWithSources(slugs, 0, 0)
	if err != nil {
		i.logger.Fatal("cannot get characters", zap.Error(err))
	}
	// A channel for sync logs that are done being created.
	syncLogCh := make(chan *comic.CharacterSyncLog, len(characters))
	defer close(syncLogCh)
	syncLogs := make([]*comic.CharacterSyncLog, len(characters))
	for idx := range characters {
		character := characters[idx]
		// create the sync log
		syncLog := comic.NewSyncLogPending(character.ID, comic.YearlyAppearances)
		if err := i.characterSvc.CreateSyncLog(syncLog); err != nil {
			i.logger.Fatal("error creating sync log", zap.String("character", character.Slug.Value()), zap.Error(err))
		}
		// TODO: This is hacky.
		syncLog.Character = character
		syncLogs[idx] = syncLog
		// send the sync log over
		syncLogCh <- syncLog
	}
	// listen for sig int. TODO: make better.
	go i.listenOnSigInt(syncLogs)
	// Read results from the channel.
	for idx := 0; idx < len(characters); idx++ {
		syncLog, ok := <-syncLogCh
		if ok && syncLog.Character != nil {
			character := syncLog.Character
			// start the import. one-by-one -- no concurrency here. maybe in the future if the external source can handle it. :)
			if err := i.ImportWithSyncLog(*character, syncLog, doReset); err != nil {
				i.logger.Error("error importing character issues", zap.String("character", character.Slug.Value()), zap.Error(err))
			}
		}
	}
	// Now refresh all the views.
	if err := i.refresher.RefreshAll(); err != nil {
		return err
	}
	// Now sync them all to redis
	results := i.statsSyncer.SyncAll(characters)
	for idx := 0; idx < len(characters); idx++ {
		res := <-results
		err = res.Error
		slug := res.Slug.Value()
		if res.Error != nil {
			i.logger.Error("error syncing character to redis", zap.Error(err), zap.String("character", slug))
		} else {
			i.logger.Info("synced character to redis", zap.String("character", slug))
		}
	}
	return nil
}

// Persists the sync log with the new status and closes the signal channel if the new status is fail or success.
func (i *CharacterIssueImporter) updateSyncLog(cLog *comic.CharacterSyncLog, newStatus comic.CharacterSyncLogStatus) {
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
func (i *CharacterIssueImporter) requestIssues(workerID int, links <-chan ExternalVendorURL, issues chan<- *comic.Issue) {
	for l := range links {
		externalIssueCh := make(chan *externalissuesource.Issue, 1)
		err := retryURL(func() (string, error) {
			externalIssue, err := i.externalSource.Issue(l.String())
			if err != nil {
				return l.String(), err
			}
			externalIssueCh <- externalIssue
			return l.String(), err
		})
		if err != nil {
			i.logger.Error("received error from external source", zap.Int("workerId", workerID), zap.String("link", l.String()), zap.Error(err))
			// Send a blank issue
			issues <- &comic.Issue{}
			continue
		}
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
				externalIssue.IsReprint,
				issueFormat)
		}
	}
}

// isAppearance checks that the issue should count as an issue appearance for the character.
func isAppearance(issue *comic.Issue) bool {
	if !issue.IsVariant && // it's not a variant
		!issue.IsReprint && // it's not a reprint.
		countsAsAppearance[issue.Format] && // the format counts as an appearance
		// the issue actually has a sale date.
		issue.SaleDate.Year() > 1 &&
		// Checks that the external issue's publisher matches up with allowed publishers
		// There can be multiple publishers such as Timely Comics that are actually Marvel
		// or crossovers with different publishers.
		stringutil.AnyFunc(issue.VendorPublisher, allowedPublishers, strings.Contains) {
		return true
	}
	return false
}

// NewCharacterIssueImporter creates a new character issue importer.
func NewCharacterIssueImporter(
	db *pg.DB,
	redis *redis.Client) *CharacterIssueImporter {
	as := comic.NewAppearancesSyncer(db, redis)
	cr := comic.NewPGCharacterRepository(db)
	ctr := comic.NewRedisCharacterThumbRepository(redis)
	pr := comic.NewPGPopularRepository(db, ctr)
	ss := comic.NewCharacterStatsSyncer(redis, cr, pr)
	aw := comic.NewRedisAppearancesPerYearRepository(redis)
	p := comic.NewPGPopularRepository(db, ctr)
	externalSource := externalissuesource.NewCbExternalSource(externalissuesource.NewHttpClient(), &externalissuesource.CbExternalSourceConfig{})
	return &CharacterIssueImporter{
		appearancesWriter: aw,
		characterSvc:      comic.NewCharacterServiceFactory(db),
		issueSvc:          comic.NewIssueServiceFactory(db),
		externalSource:    externalSource,
		appearanceSyncer:  as,
		logger:            log.CEREBRO(),
		extractor:         NewCharacterCBExtractor(externalSource),
		refresher:         p,
		statsSyncer:       ss,
	}
}

// NewCharacterCBExtractor creates a new character CB vendor extractor from the params.
func NewCharacterCBExtractor(externalSource externalissuesource.ExternalSource) *CharacterCBExtractor {
	return &CharacterCBExtractor{
		src:    externalSource,
		logger: log.CEREBRO(),
	}
}
