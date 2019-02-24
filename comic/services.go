package comic

import (
	"fmt"
	"github.com/aimeelaplant/comiccruncher/imaging"
	"github.com/aimeelaplant/comiccruncher/internal/log"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"strconv"
	"time"
)

const (
	// disableSourcesSql is the SQL for disabling character sources.
	disableSourcesSQL = `
	UPDATE character_sources
	SET is_disabled = TRUE	
	WHERE character_id = ?
		AND is_disabled = false -- no need to reset ones already disabled
		AND vendor_name ILIKE ANY(ARRAY[%s]);`
	// mainSourcesSql is the SQL for setting main sources.
	mainSourcesSQL = `
	UPDATE character_sources
	SET is_main = TRUE	
	WHERE character_id = ?
		AND is_main = FALSE -- ignore ones already already set
		AND is_disabled = FALSE -- ignore disabled ones
		AND vendor_name NOT ILIKE ALL(ARRAY[%s])`
	// altSourcesSql is the sql for setting alternate sources.
	altSourcesSQL = `
	UPDATE character_sources
	SET is_main = FALSE
	WHERE character_id = ?
		AND is_main = TRUE -- ignore ones already already set
		AND is_disabled = FALSE -- ignore disabled ones
		AND vendor_name ILIKE ANY(ARRAY[%s])`
)

// PublisherServicer is the service interface for publishers.
type PublisherServicer interface {
	// Publisher gets a publisher by its slug.
	Publisher(slug PublisherSlug) (*Publisher, error)
}

// IssueServicer is the service interface for issues.
type IssueServicer interface {
	// Issues gets issues by their IDs.
	Issues(ids []IssueID, limit, offset int) ([]*Issue, error)
	// IssuesByVendor gets issues by their vendor IDs and vendor types.
	IssuesByVendor(vendorIds []string, vendorType VendorType, limit, offset int) ([]*Issue, error)
	// Creates an issue.
	Create(issue *Issue) error
	// CreateP ceates an issue from parameters.
	CreateP(
		vendorID, vendorPublisher, vendorSeriesName, vendorSeriesNumber string,
		pubDate, saleDate time.Time,
		isVariant, isMonthUncertain, isReprint bool,
		format Format) error
}

// CharacterServicer is the service interface for characters.
// TODO: This interface is huge and not idiomatic Go...fix later.
type CharacterServicer interface {
	// Creates a character
	Create(character *Character) error
	// Character gets a character by its slug.
	Character(slug CharacterSlug) (*Character, error)
	// Updates a character.
	Update(character *Character) error
	// UpdateAll updates all characters
	UpdateAll(characters []*Character) error
	// CharactersWithSources gets all the enabled characters who have sources.
	CharactersWithSources(slug []CharacterSlug, limit, offset int) ([]*Character, error)
	// Characters gets all enabled characters by their slugs.
	Characters(slugs []CharacterSlug, limit, offset int) ([]*Character, error)
	// CharacterByVendor gets all the characters by the vendor. If `includeIsDisabled` is true, it will include disabled characters.
	CharacterByVendor(vendorID string, vendorType VendorType, includeIsDisabled bool) (*Character, error)
	// CharactersByPublisher list characters alphabetically. If `filterSources` is true, it will only list characters with sources.
	CharactersByPublisher(slugs []PublisherSlug, filterSources bool, limit, offset int) ([]*Character, error)
	// CreateSource creates a character source if it doesn't exist. If it exists, it returns the found source and an error.
	// And also modifies `source` to get the found values.
	CreateSource(source *CharacterSource) error
	// UpdateSource updates a character source
	UpdateSource(source *CharacterSource) error
	// MustNormalizeSources so that main vs alternate sources are categorized correctly and disables any unnecessary sources.
	// panics if there's an error.
	MustNormalizeSources(*Character)
	// Source gets a unique source by its character ID and vendor url.
	Source(id CharacterID, vendorURL string) (*CharacterSource, error)
	// Sources gets all the sources for a  character.
	Sources(id CharacterID, vendorType VendorType, isMain *bool) ([]*CharacterSource, error)
	// TotalSources gets the total sources for a character.
	TotalSources(id CharacterID) (int64, error)
	// CreateIssueP creates an issue for a character with the parameters.
	CreateIssueP(
		characterID CharacterID,
		issueID IssueID,
		appearanceType AppearanceType,
		importance *Importance) (*CharacterIssue, error)
	// CreateIssue creates an issue for a character.
	CreateIssue(issue *CharacterIssue) error
	// CreateIssues creates multiple issues for a character. // TODO: Autogenerated IDs not returned in struct!!
	CreateIssues(issues []*CharacterIssue) error
	// Issue gets a character issue by its character ID and issue ID
	Issue(characterID CharacterID, issueID IssueID) (*CharacterIssue, error)
	// RemoveIssues removes all the issues w/ the associated character ID.
	RemoveIssues(ids ...CharacterID) (int, error)
	// CreateSyncLogP creates a sync log for a character with the parameters.
	CreateSyncLogP(
		id CharacterID,
		status CharacterSyncLogStatus,
		syncType CharacterSyncLogType,
		syncedAt *time.Time) (*CharacterSyncLog, error)
	// CreateSyncLog creates a sync log.
	CreateSyncLog(syncLog *CharacterSyncLog) error
	// UpdateSyncLog updates a sync log
	UpdateSyncLog(syncLog *CharacterSyncLog) error
}

// RankedServicer is the interface for getting ranked and popular characters.
type RankedServicer interface {
	AllPopular(cr PopularCriteria) ([]*RankedCharacter, error)
	DCPopular(cr PopularCriteria) ([]*RankedCharacter, error)
	MarvelPopular(cr PopularCriteria) ([]*RankedCharacter, error)
	MarvelTrending(limit, offset int) ([]*RankedCharacter, error)
	DCTrending(limit, offset int) ([]*RankedCharacter, error)
}

// ExpandedServicer is the interface for getting a character with expanded details.
type ExpandedServicer interface {
	Character(slug CharacterSlug) (*ExpandedCharacter, error)
}

// CharacterThumbServicer is the interface for creating and getting thumbnails for a character.
type CharacterThumbServicer interface {
	Upload(c *Character) (*CharacterThumbnails, error)
}

// ExpandedService gets an expanded character.
type ExpandedService struct {
	cr  CharacterRepository
	ar  AppearancesByYearsRepository
	r   RedisClient
	slr CharacterSyncLogRepository
	ctr CharacterThumbRepository
}

// RankedService is the service for getting ranked and popular characters.
type RankedService struct {
	popRepo PopularRepository
}

// CharacterThumbService is the service for creating and uploading thumbnails for characters.
type CharacterThumbService struct {
	r RedisClient
	tu imaging.ThumbnailUploader
}

// Upload uploads thumbnails for the given character struct.
func (cts *CharacterThumbService) Upload(c *Character) (*CharacterThumbnails, error) {
	vendorKey := c.VendorImage
	imageKey := c.Image
	slug := c.Slug
	thumbs := &CharacterThumbnails{
		Slug: slug,
		Image: &ThumbnailSizes{},
		VendorImage: &ThumbnailSizes{},
	}
	opts := imaging.NewDefaultThumbnailOptions(
		"images/characters/",
		imaging.NewThumbnailSize(100, 100),
		imaging.NewThumbnailSize(300, 300),
		imaging.NewThumbnailSize(600, 0),
	)
	if vendorKey != "" {
		sizes1, err := cts.makeThumbs(vendorKey, opts)
		if err != nil {
			return thumbs, err
		}
		thumbs.VendorImage = sizes1
	}
	if imageKey != "" {
		sizes2, err := cts.makeThumbs(imageKey, opts)
		if err != nil {
			return thumbs, err
		}
		thumbs.Image = sizes2
	}
	vendorImg := thumbs.VendorImage
	img := thumbs.Image
	// the last `-` delimeter is for the regular image.
	redisVal := fmt.Sprintf(
		"small:%s;medium:%s;large:%s-small:%s;medium:%s;large:%s",
		vendorImg.Small,
		vendorImg.Medium,
		vendorImg.Large,
		img.Small,
		img.Medium,
		img.Large)
	return thumbs, cts.r.Set(redisThumbnailKey(slug), redisVal, 0).Err()
}

func (cts *CharacterThumbService) makeThumbs(key string, opts *imaging.ThumbnailOptions) (*ThumbnailSizes, error) {
	results, err := cts.tu.Generate(key, opts)
	if err != nil {
		return nil, err
	}
	sizes := &ThumbnailSizes{}
	for _, result := range results {
		width := int(result.Dimensions.Width)
		pathname := result.Pathname
		switch width {
		case 100:
			sizes.Small = pathname
			break
		case 300:
			sizes.Medium = pathname
			break
		case 600:
			sizes.Large = pathname
			break
		}
	}
	return sizes, nil
}

// Character gets an expanded character.
func (s *ExpandedService) Character(slug CharacterSlug) (*ExpandedCharacter, error) {
	c, err := s.cr.FindBySlug(slug, false)
	if c == nil || err != nil {
		return nil, err
	}
	sl, err := s.slr.LastSyncs(c.ID)
	if err != nil {
		return nil, err
	}
	res, err := s.r.HGetAll(fmt.Sprintf("%s:stats", slug.Value())).Result()
	if err != nil {
		return nil, err
	}
	ec := &ExpandedCharacter{}
	if len(res) > 0 {
		atCount, err := parseUint(res["all_time_issue_count"])
		if err != nil {
			return nil, err
		}
		atRank, err := parseUint(res["all_time_issue_count_rank"])
		if err != nil {
			return nil, err
		}
		atAvg, err := strconv.ParseFloat(res["all_time_average_per_year"], 64)
		if err != nil {
			return nil, err
		}
		atAvgRank, err := parseUint(res["all_time_average_per_year_rank"])
		if err != nil {
			return nil, err
		}
		allTime := NewCharacterStats(AllTimeStats, atRank, atCount, atAvgRank, atAvg)
		miCount, err := parseUint(res["main_issue_count"])
		if err != nil {
			return nil, err
		}
		miRank, err := parseUint(res["main_issue_count_rank"])
		if err != nil {
			return nil, err
		}
		miAvgRank, err := parseUint(res["main_average_per_year_rank"])
		if err != nil {
			return nil, err
		}
		miAvg, err := strconv.ParseFloat(res["main_average_per_year"], 64)
		if err != nil {
			return nil, err
		}
		mainStats := NewCharacterStats(MainStats, miRank, miCount, miAvgRank, miAvg)
		stats := make([]CharacterStats, 2)
		stats[0] = allTime
		stats[1] = mainStats
		ec.Stats = stats
	}
	apps, err := s.ar.List(slug)
	if err != nil {
		return nil, err
	}
	thumbs, err := s.ctr.Thumbnails(slug)
	if err != nil {
		return nil, err
	}
	ec.Appearances = apps
	ec.Character = c
	ec.LastSyncs = sl
	ec.Thumbnails = thumbs
	return ec, nil
}

// AllPopular gets the most popular characters per year ordered by either issue count or
// average appearances per year.
func (s *RankedService) AllPopular(cr PopularCriteria) ([]*RankedCharacter, error) {
	return s.popRepo.All(cr)
}

// DCPopular gets DC's most popular characters per year.
func (s *RankedService) DCPopular(cr PopularCriteria) ([]*RankedCharacter, error) {
	return s.popRepo.DC(cr)
}

// MarvelPopular gets Marvel's most popular characters per year ordered by either issue count o
// or average appearances per year.
func (s *RankedService) MarvelPopular(cr PopularCriteria) ([]*RankedCharacter, error) {
	return s.popRepo.Marvel(cr)
}

// MarvelTrending gets the trending characters for marvel.
func (s *RankedService) MarvelTrending(limit, offset int) ([]*RankedCharacter, error) {
	return s.popRepo.MarvelTrending(limit, offset)
}

// DCTrending gets the trending characters for marvel.
func (s *RankedService) DCTrending(limit, offset int) ([]*RankedCharacter, error) {
	return s.popRepo.DCTrending(limit, offset)
}

// PublisherService is the service for publishers.
type PublisherService struct {
	repository PublisherRepository
}

// IssueService is the service for issues.
type IssueService struct {
	repository IssueRepository
}

// CharacterService is the service for characters.
type CharacterService struct {
	tx                    Transactional
	repository            CharacterRepository
	issueRepository       CharacterIssueRepository
	sourceRepository      CharacterSourceRepository
	syncLogRepository     CharacterSyncLogRepository
	appearancesRepository AppearancesByYearsRepository
}

// Publisher gets a publisher by its slug.
func (s *PublisherService) Publisher(slug PublisherSlug) (*Publisher, error) {
	return s.repository.FindBySlug(slug)
}

// Issues gets all the issues by their IDs. A `limit` of `0` means no limit.
func (s *IssueService) Issues(ids []IssueID, limit, offset int) ([]*Issue, error) {
	return s.repository.FindAll(IssueCriteria{
		Ids:    ids,
		Limit:  limit,
		Offset: offset,
	})
}

// IssuesByVendor gets all the issues by the vendor IDs and vendor type.
// A limit of `0` means no limit.
func (s *IssueService) IssuesByVendor(ids []string, vendorType VendorType, limit, offset int) ([]*Issue, error) {
	return s.repository.FindAll(IssueCriteria{
		VendorIds:  ids,
		VendorType: vendorType,
		Limit:      limit,
		Offset:     offset,
	})
}

// Create creates an issue.
func (s *IssueService) Create(i *Issue) error {
	return s.repository.Create(i)
}

// CreateP Creates an issue from the parameters.
func (s *IssueService) CreateP(vendorID, vendorPublisher, vendorSeriesName, vendorSeriesNumber string, pubDate, saleDate time.Time, isVariant, isMonthUncertain, isReprint bool, format Format) error {
	return s.repository.Create(NewIssue(
		vendorID,
		vendorPublisher,
		vendorSeriesNumber,
		vendorSeriesNumber,
		pubDate,
		saleDate,
		isVariant,
		isMonthUncertain,
		isReprint,
		format,
	))
}

// Create creates a new character
func (s *CharacterService) Create(c *Character) error {
	return s.repository.Create(c)
}

// Character gets a non-disabled character by its slug.
func (s *CharacterService) Character(slug CharacterSlug) (*Character, error) {
	return s.repository.FindBySlug(slug, false)
}

// Update updates a character.
func (s *CharacterService) Update(c *Character) error {
	return s.repository.Update(c)
}

// UpdateAll updates all characters
func (s *CharacterService) UpdateAll(characters []*Character) error {
	return s.repository.UpdateAll(characters)
}

// CharactersWithSources gets non-disabled characters who have sources.
// A `limit` of `0` means unlimited.
func (s *CharacterService) CharactersWithSources(slugs []CharacterSlug, limit, offset int) ([]*Character, error) {
	return s.repository.FindAll(CharacterCriteria{
		Slugs:             slugs,
		IncludeIsDisabled: false,
		FilterSources:     true,
		Limit:             limit,
		Offset:            offset,
	})
}

// Characters gets all non-disabled characters by their slugs.
// A `limit` of `0` means unlimited.
func (s *CharacterService) Characters(slugs []CharacterSlug, limit, offset int) ([]*Character, error) {
	return s.repository.FindAll(CharacterCriteria{
		Slugs:             slugs,
		Limit:             limit,
		Offset:            offset,
		IncludeIsDisabled: false,
	})
}

// CreateSource creates a source for a character, if it doesn't exist.
// If it exists, an ErrAlreadyExists gets returned as an error.
// A little janky right now.
func (s *CharacterService) CreateSource(source *CharacterSource) error {
	err := s.sourceRepository.Create(source)
	return err
}

// UpdateSource updates an existing source
func (s *CharacterService) UpdateSource(source *CharacterSource) error {
	return s.sourceRepository.Update(source)
}

// Source gets a unique character source by its character ID and vendor url
func (s *CharacterService) Source(id CharacterID, vendorURL string) (*CharacterSource, error) {
	sources, err := s.sourceRepository.FindAll(CharacterSourceCriteria{
		CharacterIDs:      []CharacterID{id},
		IncludeIsDisabled: true,
		VendorUrls:        []string{vendorURL},
		Limit:             1,
	})
	if err != nil {
		return nil, err
	}
	if len(sources) == 0 {
		return nil, nil
	}
	return sources[0], nil
}

// Sources lists all non-disabled character sources from the given parameters.
// If `isMain` is `nil`, it will list both types of sources. If `isMain` is true, it will list main sources.
// If `isMain` is `false`, it will list alternate sources.
func (s *CharacterService) Sources(id CharacterID, vendorType VendorType, isMain *bool) ([]*CharacterSource, error) {
	return s.sourceRepository.FindAll(CharacterSourceCriteria{
		CharacterIDs:      []CharacterID{id},
		IncludeIsDisabled: false, // Don't include sources that are disabled!
		VendorType:        vendorType,
		IsMain:            isMain,
	})
}

// MustNormalizeSources normalizes sources for main and alternate sources and disables any unneeded sources.
func (s *CharacterService) MustNormalizeSources(c *Character) {
	id := c.ID.Value()
	var altUniverses []universeDefinition
	var disabledUniverses []universeDefinition
	if c.Publisher.Slug == "marvel" {
		altUniverses = marvelAltUniverses
		disabledUniverses = marvelDisabledUniverses
	} else if c.Publisher.Slug == "dc" {
		altUniverses = dcAltUniverses
		disabledUniverses = dcDisabledUniverses
	} else {
		panic(fmt.Sprintf("unknown publisher: %s", c.Publisher.Slug.Value()))
	}
	// todo: better to run all this in a transaction.
	// disable clones, impostors, etc.
	if !ignoreIDsForDisabled[id] {
		must(s.sourceRepository.Raw(fmt.Sprintf(disableSourcesSQL, pgSearchString(disabledUniverses)), id))
	}
	// set the main universes from alt universes.
	must(s.sourceRepository.Raw(fmt.Sprintf(mainSourcesSQL, pgSearchString(altUniverses)), id))
	// now set the alternate sources from alternate sources.
	// b/c if we add any more sources after running the above query, we
	// won't be able to set is_main = false for any of them. sooo stupid and i'm sure there's a better way to do this but whatever.
	must(s.sourceRepository.Raw(fmt.Sprintf(altSourcesSQL, pgSearchString(altUniverses)), id))
	// Now make sure earth-616 is set as main. (Some sources have 616 .. some don't. :( )
	if c.Publisher.Slug == "marvel" {
		must(s.sourceRepository.Raw("UPDATE character_sources SET is_main = TRUE WHERE vendor_name ILIKE '%earth-616)%' AND character_id = ?", id))
	}
}

// TotalSources gets the total number of sources for a character
func (s *CharacterService) TotalSources(id CharacterID) (int64, error) {
	return s.repository.Total(CharacterCriteria{FilterSources: true, IDs: []CharacterID{id}})
}

// CreateIssueP creates an issue from the parameters.
func (s *CharacterService) CreateIssueP(characterID CharacterID, issueID IssueID, appearanceType AppearanceType, importance *Importance) (*CharacterIssue, error) {
	issue := &CharacterIssue{
		CharacterID:    characterID,
		IssueID:        issueID,
		AppearanceType: appearanceType,
		Importance:     importance,
	}
	err := s.issueRepository.Create(issue)
	return issue, err
}

// CreateIssue creates an issue.
func (s *CharacterService) CreateIssue(issue *CharacterIssue) error {
	return s.issueRepository.Create(issue)
}

// CreateIssues creates multiple issues in a bulk query. TODO: Generated ID doesn't get set!
func (s *CharacterService) CreateIssues(issues []*CharacterIssue) error {
	return s.issueRepository.InsertFast(issues)
}

// Issue gets a character issue by its character ID and issue ID.
func (s *CharacterService) Issue(characterID CharacterID, issueID IssueID) (*CharacterIssue, error) {
	return s.issueRepository.FindOneBy(characterID, issueID)
}

// CreateSyncLogP creates a sync log with the parameters.
func (s *CharacterService) CreateSyncLogP(id CharacterID, status CharacterSyncLogStatus, syncType CharacterSyncLogType, syncedAt *time.Time) (*CharacterSyncLog, error) {
	syncLog := &CharacterSyncLog{
		CharacterID: id,
		SyncStatus:  status,
		SyncType:    syncType,
		SyncedAt:    syncedAt,
	}
	err := s.syncLogRepository.Create(syncLog)
	return syncLog, err
}

// CreateSyncLog creates a sync log for a character.
func (s *CharacterService) CreateSyncLog(syncLog *CharacterSyncLog) error {
	return s.syncLogRepository.Create(syncLog)
}

// UpdateSyncLog updates a sync log for a character.
func (s *CharacterService) UpdateSyncLog(syncLog *CharacterSyncLog) error {
	return s.syncLogRepository.Update(syncLog)
}

// CharacterByVendor gets a character from the specified vendor and whether the character is disabled or not.
func (s *CharacterService) CharacterByVendor(vendorID string, vendorType VendorType, includeIsDisabled bool) (*Character, error) {
	characters, err := s.repository.FindAll(CharacterCriteria{
		VendorIds:         []string{vendorID},
		VendorTypes:       []VendorType{vendorType},
		Limit:             1,
		IncludeIsDisabled: includeIsDisabled,
	})
	if err != nil {
		return nil, err
	}
	if len(characters) > 0 {
		return characters[0], nil
	}
	return nil, nil
}

// CharactersByPublisher lists enabled characters by their publisher.
func (s *CharacterService) CharactersByPublisher(slugs []PublisherSlug, filterSources bool, limit, offset int) ([]*Character, error) {
	return s.repository.FindAll(CharacterCriteria{
		FilterSources:     filterSources,
		PublisherSlugs:    slugs,
		IncludeIsDisabled: false,
		Limit:             limit,
		Offset:            offset,
	})
}

// RemoveIssues deletes all associated issues for the given character IDs.
func (s *CharacterService) RemoveIssues(ids ...CharacterID) (int, error) {
	results := 0
	res := s.tx.RunInTransaction(func(tx *pg.Tx) error {
		for _, id := range ids {
			res, err := s.issueRepository.RemoveAllByCharacterID(id)
			if err != nil {
				log.COMIC().Error("error removing character issues for character. transaction will be rolled back.", zap.Uint("ID", id.Value()))
				return err
			}
			results += res
		}
		return nil
	})
	return results, res
}

// NewPublisherService creates a new publisher service
func NewPublisherService(container *PGRepositoryContainer) *PublisherService {
	return &PublisherService{
		repository: container.PublisherRepository(),
	}
}

// NewCharacterService creates a new character service but with the appearances by years coming from postgres.
func NewCharacterService(container *PGRepositoryContainer) *CharacterService {
	return &CharacterService{
		tx:                    container.DB(),
		repository:            container.CharacterRepository(),
		issueRepository:       container.CharacterIssueRepository(),
		sourceRepository:      container.CharacterSourceRepository(),
		syncLogRepository:     container.CharacterSyncLogRepository(),
		appearancesRepository: container.AppearancesByYearsRepository(),
	}
}

// NewIssueService creates a new issue service from the repository container.
func NewIssueService(container *PGRepositoryContainer) *IssueService {
	return &IssueService{
		repository: container.IssueRepository(),
	}
}

// NewRankedService creates a new service for ranked characters.
func NewRankedService(repository PopularRepository) *RankedService {
	return &RankedService{
		popRepo: repository,
	}
}

// NewExpandedService creates a new service for getting expanded details for a character
func NewExpandedService(cr CharacterRepository, ar AppearancesByYearsRepository, rc RedisClient, slr CharacterSyncLogRepository, ctr CharacterThumbRepository) *ExpandedService {
	return &ExpandedService{
		cr:  cr,
		ar:  ar,
		r:   rc,
		slr: slr,
		ctr: ctr,
	}
}

// NewCharacterThumbnailService creates a new thumbnail service.
func NewCharacterThumbnailService(r RedisClient, tu imaging.ThumbnailUploader) *CharacterThumbService {
	return &CharacterThumbService{
		r: r,
		tu: tu,
	}
}
