package comic

import (
	"errors"
	"fmt"
	"github.com/aimeelaplant/comiccruncher/internal/log"
	"github.com/go-pg/pg"
	"github.com/go-redis/redis"
	"github.com/gosimple/slug"
	"go.uber.org/zap"
	"math"
	"strings"
	"sync"
	"time"
)

var (
	// AllView is the materialized view for all characters with both main and alternate appearances.
	AllView MaterializedView = "mv_ranked_characters"
	// MainView is the materialized view for all characters with main appearances.
	MainView MaterializedView = "mv_ranked_characters_main"
	// AltView is the materialized view for all characters with alternate appearances.
	AltView MaterializedView = "mv_ranked_characters_alternate"
	// DcMainView is the materialized view for DC characters with main appearances.
	DcMainView MaterializedView = "mv_ranked_characters_dc_main"
	// MarvelMainView is the materialized view for Marvel characters with main appearances.
	MarvelMainView MaterializedView = "mv_ranked_characters_marvel_main"
	// MarvelTrendingView is the materialized view for trending Marvel characters for main appearances only.
	MarvelTrendingView MaterializedView = "mv_trending_characters_marvel"
	// DCTrendingView is the materialized view for trending DC characters for main appearances only.
	DCTrendingView MaterializedView = "mv_trending_characters_dc"
	// Sooo many. In hindsight I should have used something like MongoDB. ¯\_(ツ)_/¯
)

// MaterializedView is the name of a table with a materialized view to cache expensive query results.
type MaterializedView string

// Value returns the string value.
func (v MaterializedView) Value() string {
	return string(v)
}

// PublisherRepository is the repository interface for publishers.
type PublisherRepository interface {
	FindBySlug(slug PublisherSlug) (*Publisher, error)
}

// IssueRepository is the repository interface for issues.
type IssueRepository interface {
	Create(issue *Issue) error
	CreateAll(issues []*Issue) error
	Update(issue *Issue) error
	FindByVendorID(vendorID string) (*Issue, error)
	FindAll(c IssueCriteria) ([]*Issue, error)
}

// CharacterRepository is the repository interface for characters.
type CharacterRepository interface {
	Create(c *Character) error
	Update(c *Character) error
	FindBySlug(slug CharacterSlug, includeIsDisabled bool) (*Character, error)
	FindAll(cr CharacterCriteria) ([]*Character, error)
	UpdateAll(characters []*Character) error
	Remove(id CharacterID) error
	Total(cr CharacterCriteria) (int64, error)
}

// CharacterSourceRepository is the repository interface for character sources.
type CharacterSourceRepository interface {
	Create(s *CharacterSource) error
	FindAll(criteria CharacterSourceCriteria) ([]*CharacterSource, error)
	Remove(id CharacterSourceID) error
	// Raw runs a raw query on the character sources.
	Raw(query string, params ...interface{}) error
	Update(s *CharacterSource) error
}

// CharacterSyncLogRepository is the repository interface for character sync logs.
type CharacterSyncLogRepository interface {
	Create(s *CharacterSyncLog) error
	FindAllByCharacterID(characterID CharacterID) ([]*CharacterSyncLog, error)
	Update(s *CharacterSyncLog) error
	FindByID(id CharacterSyncLogID) (*CharacterSyncLog, error)
	LastSyncs(id CharacterID) ([]*LastSync, error)
}

// CharacterIssueRepository is the repository interface for character issues.
type CharacterIssueRepository interface {
	CreateAll(cis []*CharacterIssue) error
	Create(ci *CharacterIssue) error
	FindOneBy(characterID CharacterID, issueID IssueID) (*CharacterIssue, error)
	InsertFast(issues []*CharacterIssue) error
}

// AppearancesByYearsRepository is the repository interface for getting a characters appearances per year.
type AppearancesByYearsRepository interface {
	List(slugs CharacterSlug) (AppearancesByYears, error)
}

// AppearancesByYearsMapRepository is the repository for listing a character's appearances by years in a map.
type AppearancesByYearsMapRepository interface {
	ListMap(slugs ...CharacterSlug) (map[CharacterSlug][]AppearancesByYears, error)
}

// StatsRepository is the repository interface for general stats about the db.
type StatsRepository interface {
	Stats() (Stats, error)
}

// PopularRepository is the repository interface for popular character rankings.
type PopularRepository interface {
	All(cr PopularCriteria) ([]*RankedCharacter, error)
	DC(cr PopularCriteria) ([]*RankedCharacter, error)
	Marvel(cr PopularCriteria) ([]*RankedCharacter, error)
	FindOneByDC(id CharacterID) (*RankedCharacter, error)
	FindOneByMarvel(id CharacterID) (*RankedCharacter, error)
	FindOneByAll(id CharacterID) (*RankedCharacter, error)
	MarvelTrending(limit, offset int) ([]*RankedCharacter, error)
	DCTrending(limit, offset int) ([]*RankedCharacter, error)
}

// PopularRefresher concurrently refreshes the materialized views.
type PopularRefresher interface {
	Refresh(view MaterializedView) error
	RefreshAll() error
}

// CharacterThumbRepository is the repository for getting character thumbnails.
type CharacterThumbRepository interface {
	AllThumbnails(slugs ...CharacterSlug) (map[CharacterSlug]*CharacterThumbnails, error)
	Thumbnails(slug CharacterSlug) (*CharacterThumbnails, error)
}

// RedisCharacterThumbRepository is for a redis repository for getting character thumbnails.
type RedisCharacterThumbRepository struct {
	r RedisClient
}

// Thumbnails gets the thumbnails for a character.
func (ctr *RedisCharacterThumbRepository) Thumbnails(slug CharacterSlug) (*CharacterThumbnails, error) {
	result, err := ctr.r.Get(redisThumbnailKey(slug)).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	thumbs := &CharacterThumbnails{}
	err = parseRedisThumbnails(result, thumbs)
	return thumbs, err
}

// AllThumbnails efficiently gets the thumbnails for many characters.
func (ctr *RedisCharacterThumbRepository) AllThumbnails(slugs ...CharacterSlug) (map[CharacterSlug]*CharacterThumbnails, error) {
	slcLen := len(slugs)
	allKeys := make([]string, slcLen)
	for i, s := range slugs {
		allKeys[i] = redisThumbnailKey(s)
	}
	all, err := ctr.r.MGet(allKeys...).Result()
	if err != nil {
		return nil, err
	}
	allThumbs := make(map[CharacterSlug]*CharacterThumbnails, slcLen)
	for i, s := range slugs {
		thumb := &CharacterThumbnails{Slug: s,}
		allThumbs[s] = thumb
		if all != nil {
			val := all[i]
			if val != nil {
				parseRedisThumbnails(val.(string), thumb)
			}
		}
	}

	return allThumbs, nil
}

// PGPopularRepository is the postgres implementation for the popular character repository.
type PGPopularRepository struct {
	db  *pg.DB
	ctr CharacterThumbRepository
}

// PGAppearancesByYearsRepository is the postgres implementation for the appearances per year repository.
type PGAppearancesByYearsRepository struct {
	db *pg.DB
}

// PGCharacterRepository is the postgres implementation for the character repository.
type PGCharacterRepository struct {
	db *pg.DB
}

// PGPublisherRepository is the postgres implementation for the publisher repository.
type PGPublisherRepository struct {
	db *pg.DB
}

// PGIssueRepository is the postgres implementation for the issue repository.
type PGIssueRepository struct {
	db *pg.DB
}

// PGCharacterSourceRepository is the postgres implementation for the character source repository.
type PGCharacterSourceRepository struct {
	db *pg.DB
}

// PGCharacterIssueRepository is the postgres implementation for the character issue repository.
type PGCharacterIssueRepository struct {
	db *pg.DB
}

// PGCharacterSyncLogRepository is the postgres implementation for the character sync log repository.
type PGCharacterSyncLogRepository struct {
	db *pg.DB
}

// PGStatsRepository is the postgres implementation for the stats repository.
type PGStatsRepository struct {
	db *pg.DB
}

// RedisAppearancesByYearsRepository is the Redis implementation for appearances per year repository.
type RedisAppearancesByYearsRepository struct {
	redisClient RedisClient
	deserializer YearlyAggregateDeserializer
	serializer YearlyAggregateSerializer
}

// FindBySlug gets a publisher by its slug.
func (r *PGPublisherRepository) FindBySlug(slug PublisherSlug) (*Publisher, error) {
	publisher := &Publisher{}
	if err := r.db.Model(publisher).Where("publisher.slug = ?", slug).Select(); err != nil {
		if err == pg.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return publisher, nil
}

// Create creates a character.
func (r *PGCharacterRepository) Create(c *Character) error {
	c.Name = strings.TrimSpace(c.Name)
	c.Slug = CharacterSlug(slug.Make(c.Name))
	count, err := r.db.
		Model(&Character{}).
		Where("slug = ?", c.Slug).
		WhereOr("name = ?", c.Name).
		Count()
	if err != nil {
		return err
	}
	if count != 0 {
		// slug must be unique, so just take the nanosecond to increment the slug.
		// @TODO: use a trigger.
		c.Slug = CharacterSlug(fmt.Sprintf("%s-%d", c.Slug, time.Now().Nanosecond()))
	}
	if _, err = r.db.Model(c).Insert(); err != nil {
		return err
	}

	// Now load the publisher. Ugh, there's probably a better way to do this...
	return r.db.Model(c).Column("character.*", "Publisher").Where("character.id = ?", c.ID).Select()
}

// Update updates a character.
func (r *PGCharacterRepository) Update(c *Character) error {
	return r.db.Update(c)
}

// FindBySlug finds a character by its slug. `includeIsDisabled` means to also include disabled characters
// in the find.
func (r *PGCharacterRepository) FindBySlug(slug CharacterSlug, includeIsDisabled bool) (*Character, error) {
	if result, err := r.FindAll(CharacterCriteria{
		Slugs:             []CharacterSlug{slug},
		IncludeIsDisabled: includeIsDisabled,
		Limit:             1}); err != nil {
		return nil, err
	} else if len(result) != 0 {
		return result[0], nil
	}
	return nil, nil
}

// Remove removes a character by its ID.
func (r *PGCharacterRepository) Remove(id CharacterID) error {
	if _, err := r.db.Model(&Character{}).Where("id = ?", id).Delete(); err != nil {
		return err
	}
	return nil
}

// Total gets total number of characters based on the criteria.
func (r *PGCharacterRepository) Total(cr CharacterCriteria) (int64, error) {
	query := r.db.Model(&Character{})

	if len(cr.PublisherSlugs) > 0 {
		query.
			Join("JOIN publishers p ON character.publisher_id = p.id").
			Where("p.slug IN (?)", pg.In(cr.PublisherSlugs))
	}

	if len(cr.PublisherIDs) > 0 {
		query.Where("publisher_id IN (?)", pg.In(cr.PublisherIDs))
	}

	if len(cr.VendorIds) > 0 {
		query.Where("character.vendor_id IN (?)", pg.In(cr.VendorIds))
	}

	if len(cr.IDs) > 0 {
		query.Where("character.id IN (?)", pg.In(cr.IDs))
	}

	if len(cr.Slugs) > 0 {
		query.Where("character.slug IN (?)", pg.In(cr.Slugs))
	}

	if !cr.IncludeIsDisabled {
		query.Where("character.is_disabled = ?", false)
	}

	if cr.FilterSources {
		query.Where("EXISTS (SELECT 1 FROM character_sources cs WHERE cs.character_id = character.id)")
	}

	if cr.FilterIssues {
		query.Where("EXISTS (SELECT 1 FROM character_issues ci WHERE ci.character_id = character.id)")
	}

	count, err := query.Count()
	return int64(count), err
}

// FindAll finds characters by the criteria.
func (r *PGCharacterRepository) FindAll(cr CharacterCriteria) ([]*Character, error) {
	var characters []*Character
	query := r.db.Model(&characters).Column("character.*", "Publisher")

	if len(cr.PublisherSlugs) > 0 {
		query.
			Join("JOIN publishers p").
			JoinOn("character.publisher_id = p.id").
			Where("p.slug IN (?)", pg.In(cr.PublisherSlugs))
	}

	if len(cr.PublisherIDs) > 0 {
		query.Where("character.publisher_id IN (?)", pg.In(cr.PublisherIDs))
	}

	if len(cr.VendorIds) > 0 {
		query.Where("character.vendor_id IN (?)", pg.In(cr.VendorIds))
	}

	if len(cr.IDs) > 0 {
		query.Where("character.id IN (?)", pg.In(cr.IDs))
	}

	if len(cr.Slugs) > 0 {
		query.Where("character.slug IN (?)", pg.In(cr.Slugs))
	}

	if !cr.IncludeIsDisabled {
		query.Where("character.is_disabled = ?", false)
	}

	if cr.FilterSources {
		query.Where("EXISTS (SELECT 1 FROM character_sources cs WHERE cs.character_id = character.id)")
	}

	if cr.FilterIssues {
		query.Where("EXISTS (SELECT 1 FROM character_issues ci WHERE ci.character_id = character.id)")
	}

	if cr.Limit > 0 {
		query.Limit(cr.Limit)
	}

	if cr.Offset > 0 {
		query.Offset(cr.Offset)
	}

	if err := query.Order("character.slug").Select(); err != nil {
		return nil, err
	}

	return characters, nil
}

// UpdateAll updates all the characters in the slice.
func (r *PGCharacterRepository) UpdateAll(characters []*Character) error {
	if len(characters) > 0 {
		_, err := r.db.Model(&characters).Update()
		return err
	}
	return nil
}

// CreateAll creates the issues in the slice.
func (r *PGCharacterIssueRepository) CreateAll(issues []*CharacterIssue) error {
	// pg-go gives error if you pass an empty slice.
	// interface should handle empty slice accordingly.
	if len(issues) > 0 {
		if _, err := r.db.Model(&issues).OnConflict("DO NOTHING").Insert(); err != nil {
			return err
		}
	}
	return nil
}

// InsertFast creates all the issues in the db ...
// but NOTE it does not generate the autoincremented ID's into the models of the slice. :(
// TODO: Find out why ORM can't do this?!?!
func (r *PGCharacterIssueRepository) InsertFast(issues []*CharacterIssue) error {
	if len(issues) > 0 {
		query := `INSERT INTO character_issues (character_id, issue_id, appearance_type, importance, created_at, updated_at)
			VALUES %s ON CONFLICT DO NOTHING`
		values := ""
		for i, c := range issues {
			var importance string
			if c.Importance == nil {
				importance = "NULL"
			} else {
				importance = string(*c.Importance)
			}
			values += fmt.Sprintf("(%d, %d, '%08b', %s, now(), now())", c.CharacterID, c.IssueID, c.AppearanceType, importance)
			if i != len(issues)-1 {
				values += ", "
			}
		}
		if _, err := r.db.Query(&issues, fmt.Sprintf(query, values)); err != nil {
			return err
		}
	}
	return nil
}

// Create creates a character issue.
func (r *PGCharacterIssueRepository) Create(ci *CharacterIssue) error {
	if _, err := r.db.Model(ci).Returning("*").Insert(); err != nil {
		return err
	}
	return nil
}

// FindOneBy finds a character issue by the params.
func (r *PGCharacterIssueRepository) FindOneBy(characterID CharacterID, issueID IssueID) (*CharacterIssue, error) {
	characterIssue := &CharacterIssue{}
	if err := r.db.Model(characterIssue).Where("character_id = ?", characterID).Where("issue_id = ?", issueID).Select(); err != nil {
		if err == pg.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return characterIssue, nil
}

// Create creates a character source.
func (r *PGCharacterSourceRepository) Create(s *CharacterSource) error {
	_, err := r.db.Model(s).Insert(s)
	return err
}

// FindAll finds all the character sources for the criteria.
func (r *PGCharacterSourceRepository) FindAll(cr CharacterSourceCriteria) ([]*CharacterSource, error) {
	var characterSources []*CharacterSource

	query := r.db.Model(&characterSources)

	if len(cr.VendorUrls) > 0 {
		query.Where("vendor_url IN (?)", pg.In(cr.VendorUrls))
	}

	if len(cr.CharacterIDs) > 0 {
		query.Where("character_id IN (?)", pg.In(cr.CharacterIDs))
	}

	if cr.IsMain != nil {
		query.Where("is_main = ?", cr.IsMain)
	}

	query.Where("vendor_type = ?", cr.VendorType)

	if !cr.IncludeIsDisabled {
		query.Where("is_disabled = False")
	}

	if cr.Limit > 0 {
		query.Limit(cr.Limit)
	}

	if cr.Offset > 0 {
		query.Offset(cr.Offset)
	}

	if err := query.Select(); err != nil {
		return nil, err
	}

	return characterSources, nil
}

// Remove removes a character source by its ID.
func (r *PGCharacterSourceRepository) Remove(id CharacterSourceID) error {
	_, err := r.db.Model(&CharacterSource{}).Where("id = ?", id).Delete()
	return err
}

// Raw performs a raw query on the character source. Not ideal but fine for now.
func (r *PGCharacterSourceRepository) Raw(query string, params ...interface{}) error {
	if _, err := r.db.Exec(query, params...); err != nil {
		return err
	}
	return nil
}

// Update updates a character source...
func (r *PGCharacterSourceRepository) Update(s *CharacterSource) error {
	return r.db.Update(s)
}

// Create creates a new character sync log.
func (r *PGCharacterSyncLogRepository) Create(s *CharacterSyncLog) error {
	_, err := r.db.Model(s).Insert(s)
	return err
}

// FindAllByCharacterID gets all the sync logs by the character ID.
func (r *PGCharacterSyncLogRepository) FindAllByCharacterID(id CharacterID) ([]*CharacterSyncLog, error) {
	var syncLogs []*CharacterSyncLog
	if err := r.db.
		Model(&syncLogs).
		Where("character_id = ?", id).
		Select(); err != nil {
		if err == pg.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return syncLogs, nil
}

// Update updates a sync log.
func (r *PGCharacterSyncLogRepository) Update(l *CharacterSyncLog) error {
	return r.db.Update(l)
}

// FindByID finds a character sync log by the id.
func (r *PGCharacterSyncLogRepository) FindByID(id CharacterSyncLogID) (*CharacterSyncLog, error) {
	syncLog := &CharacterSyncLog{}
	if err := r.db.Model(syncLog).Where("id = ?", id).Select(); err != nil {
		if err == pg.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return syncLog, nil
}

// LastSyncs gets the last successful sync logs for a character.
func (r *PGCharacterSyncLogRepository) LastSyncs(id CharacterID) ([]*LastSync, error) {
	var ls []*LastSync
	sql := `SELECT
		character_id,
		synced_at,
		(message)::int as num_issues
		FROM character_sync_logs
		WHERE character_id = ?
			AND message IS NOT NULL
			AND synced_at IS NOT NULL
			AND sync_status = ?
		ORDER BY synced_at DESC LIMIT 3`
	_, err := r.db.Query(&ls, sql, id, Success)
	return ls, err
}

// Create creates an issue.
func (r *PGIssueRepository) Create(issue *Issue) error {
	_, err := r.db.Model(issue).Returning("*").Insert(issue)
	return err
}

// CreateAll creates all the issue in the slice.
func (r *PGIssueRepository) CreateAll(issues []*Issue) error {
	// pg-go returns an error if you bulk-insert an empty slice.
	if len(issues) > 0 {
		return r.db.Insert(&issues)
	}
	return nil
}

// Update updates an issue.
func (r *PGIssueRepository) Update(issue *Issue) error {
	return r.db.Update(issue)
}

// FindByVendorID finds the issues with the specified vendor IDs.
func (r *PGIssueRepository) FindByVendorID(vendorID string) (*Issue, error) {
	if issues, err := r.FindAll(IssueCriteria{VendorIds: []string{vendorID}, Limit: 1}); err != nil {
		return nil, err
	} else if len(issues) != 0 {
		return issues[0], nil
	} else {
		return nil, nil
	}
}

// FindAll finds all the issues from the criteria.
func (r *PGIssueRepository) FindAll(cr IssueCriteria) ([]*Issue, error) {
	var issues []*Issue

	query := r.db.Model(&issues)

	if len(cr.Ids) > 0 {
		query.Where("id IN (?)", pg.In(cr.Ids))
	}

	query.Where("vendor_type = ?", cr.VendorType)

	if len(cr.VendorIds) > 0 {
		query.Where("vendor_id IN (?)", pg.In(cr.VendorIds))
	}

	if len(cr.Formats) > 0 {
		query.Where("format IN (?)", pg.In(cr.Formats))
	}

	if cr.Limit > 0 {
		query.Limit(cr.Limit)
	}

	if cr.Offset > 0 {
		query.Offset(cr.Offset)
	}

	if err := query.Select(); err != nil {
		return nil, err
	}

	return issues, nil
}

// Stats gets stats for the comic repository.
func (r *PGStatsRepository) Stats() (Stats, error) {
	stats := Stats{}
	_, err := r.db.QueryOne(&stats, `
		SELECT date_part('year', min(i.sale_date)) AS min_year, 
		date_part('year', max(i.sale_date)) AS max_year,
		count(ci.id) as total_appearances,
       	(SELECT count(*) FROM characters c
       		WHERE EXISTS (SELECT 1 FROM character_sources cs WHERE cs.character_id = c.id)
			AND EXISTS (SELECT 1 FROM character_issues ci WHERE ci.character_id = C.id)) as total_characters,
       	(SELECT count(*) FROM issues) AS total_issues
		FROM character_issues ci
		INNER JOIN issues i ON ci.issue_id = i.id`)
	return stats, err
}

func (r *PGAppearancesByYearsRepository) createQuery(slug CharacterSlug, t AppearanceType) string {
	return fmt.Sprintf(`
	SELECT years.year as year, count(issues.id) AS count, '%s' as category
	FROM generate_series(
       (SELECT date_part('year', min(i.sale_date)) FROM issues i
        INNER JOIN character_issues ci ON ci.issue_id = i.id
        INNER JOIN characters c on c.id = ci.character_id
        WHERE c.slug = ?0) :: INT,
        date_part('year', CURRENT_DATE) :: INT
       ) AS years(year)
       LEFT JOIN (
		SELECT i.sale_date AS sale_date, i.id AS id
		FROM issues i
		INNER JOIN character_issues ci ON i.id = ci.issue_id
		INNER JOIN characters c on c.id = ci.character_id
		WHERE c.slug = ?0 AND ci.appearance_type & B'%08b' > 0::BIT(8)
       ) issues ON years.year = date_part('year', issues.sale_date)
      GROUP BY years.year`, t.String(), t)
}

// List gets a slice of a character's main and alternate appearances. This isn't very efficient for multiple characters
// so you should use the Redis repo instead.
func (r *PGAppearancesByYearsRepository) List(s CharacterSlug) (AppearancesByYears, error) {
	mainQ := r.createQuery(s, Main)
	altQ := r.createQuery(s, Alternate)
	q := mainQ + " UNION " + altQ + " ORDER BY year, category"
	vals := make([]struct{
		Year     int
		Count    int
		Category string
	}, 0)
	_, err := r.db.Query(&vals, q, s)
	if err != nil {
		return AppearancesByYears{}, err
	}
	length := int(math.Round(float64(len(vals) / 2)))
	yas := make([]YearlyAggregate, length)
 	t := 0
	for i := 0; i < len(vals); i+=2 {
		alt := vals[i]
		main := vals[i+1]
		yas[t] = YearlyAggregate{
			Year:      alt.Year,
			Alternate: alt.Count,
			Main:      main.Count,
		}
		t++
	}
	return NewAppearancesByYears(s, yas), nil
}

// List returns a slice of appearances per year for the given characters' slugs main and alternate appearances.
func (r *RedisAppearancesByYearsRepository) List(s CharacterSlug) (AppearancesByYears, error) {
	key := getAppearanceKey(s)
	all, err := r.redisClient.Get(key).Result()
	if err != nil  && err != redis.Nil {
		return AppearancesByYears{}, err
	}
	if all != "" {
		return NewAppearancesByYears(s, r.deserializer.Deserialize(all)), nil
	}
	return NewAppearancesByYears(s, nil), nil
}

// ListMap returns a map of appearances per year for the given characters' slugs main and alternate appearances.
func (r *RedisAppearancesByYearsRepository) ListMap(slugs ...CharacterSlug) (map[CharacterSlug]AppearancesByYears, error) {
	slcLen := len(slugs)
	allKeys := make([]string, slcLen)
	for i, s := range slugs {
		allKeys[i] = getAppearanceKey(s)
	}
	all, err := r.redisClient.MGet(allKeys...).Result()
	if err != nil {
		return nil, err
	}
	allApps := make(map[CharacterSlug]AppearancesByYears, slcLen)
	for i, s := range slugs {
		val := all[i]
		if val != nil {
			allApps[s] = NewAppearancesByYears(s, r.deserializer.Deserialize(val.(string)))
		} else {
			allApps[s] = NewAppearancesByYears(s, nil)
		}
	}
	return allApps, nil
}

// Sets the character's appearances in Redis.
func (r *RedisAppearancesByYearsRepository) Set(character AppearancesByYears) error {
	if character.CharacterSlug.Value() == "" {
		return errors.New("wtf. got blank character slug")
	}
	return r.redisClient.Set(
		getAppearanceKey(character.CharacterSlug),
		r.serializer.Serialize(character.Aggregates),
		0).
		Err()
}

func getAppearanceKey(s CharacterSlug) string {
	return s.Value() + ":appearances"
}

// All returns all the popular characters for DC and Marvel.
func (r *PGPopularRepository) All(cr PopularCriteria) ([]*RankedCharacter, error) {
	if cr.AppearanceType == Main {
		return r.query(MainView, cr)
	}
	if cr.AppearanceType == Alternate {
		return r.query(AltView, cr)
	}
	return r.query(AllView, cr)
}

// DC gets the popular characters for DC characters only. The rank will be adjusted for DC.
func (r *PGPopularRepository) DC(cr PopularCriteria) ([]*RankedCharacter, error) {
	return r.query(DcMainView, cr)
}

// Marvel gets the popular characters for Marvel characters only. The rank will be adjusted for Marvel.
func (r *PGPopularRepository) Marvel(cr PopularCriteria) ([]*RankedCharacter, error) {
	return r.query(MarvelMainView, cr)
}

// MarvelTrending gets the trending characters for Marvel.
func (r *PGPopularRepository) MarvelTrending(limit, offset int) ([]*RankedCharacter, error) {
	return r.query(MarvelTrendingView, PopularCriteria{
		AppearanceType: Main,
		SortBy:         MostIssues,
		Limit:          limit,
		Offset:         offset,
	})
}

// DCTrending gets the trending characters for DC.
func (r *PGPopularRepository) DCTrending(limit, offset int) ([]*RankedCharacter, error) {
	return r.query(DCTrendingView, PopularCriteria{
		AppearanceType: Main,
		SortBy:         MostIssues,
		Limit:          limit,
		Offset:         offset,
	})
}

func (r *PGPopularRepository) findOneBy(id CharacterID, view MaterializedView) (*RankedCharacter, error) {
	sql := fmt.Sprintf(`SELECT
		average_per_year_rank as stats__average_rank,
		average_per_year as stats__average,
		issue_count as stats__issue_count,
		issue_count_rank as stats__issue_count_rank,
		id,
		publisher_id,
		name,
		other_name,
		description,
		image,
		slug,
		vendor_image,
		vendor_url,
		vendor_description,
		publisher__id,
		publisher__slug,
		publisher__name
		FROM %s
		WHERE id = ?
		`, view)
	c := &RankedCharacter{}
	_, err := r.db.QueryOne(c, sql, id)
	return c, err
}

// FindOneByDC finds a ranked character for DC main appearances.
func (r *PGPopularRepository) FindOneByDC(id CharacterID) (*RankedCharacter, error) {
	return r.findOneBy(id, DcMainView)
}

// FindOneByMarvel finds a ranked character for Marvel main appearances.
func (r *PGPopularRepository) FindOneByMarvel(id CharacterID) (*RankedCharacter, error) {
	return r.findOneBy(id, MarvelMainView)
}

// FindOneByAll finds a ranked character for all-time types of appearances.
func (r *PGPopularRepository) FindOneByAll(id CharacterID) (*RankedCharacter, error) {
	return r.findOneBy(id, AllView)
}

// Refresh refreshes the specified the materialized view. Note this can take several seconds!
func (r *PGPopularRepository) Refresh(view MaterializedView) error {
	_, err := r.db.Exec("REFRESH MATERIALIZED VIEW CONCURRENTLY " + view.Value())
	return err
}

// RefreshAll refreshes all the materialized views in a transaction. Note this can take a while, so refreshing is done concurrently
// for all tables!
func (r *PGPopularRepository) RefreshAll() error {
	allViews := []MaterializedView{
		AllView,
		MainView,
		AltView,
		DcMainView,
		MarvelMainView,
		MarvelTrendingView,
		MarvelMainView,
	}
	var wg sync.WaitGroup
	wg.Add(len(allViews))
	errCh := make(chan error, len(allViews))
	defer close(errCh)
	for idx := range allViews {
		go func(idx int, wg *sync.WaitGroup) {
			defer wg.Done()
			err := r.Refresh(allViews[idx])
			if err != nil {
				errCh <- err
				log.COMIC().Error("error refreshing", zap.String("view", allViews[idx].Value()), zap.Error(err))
			} else {
				log.COMIC().Info("done refreshing", zap.String("view", allViews[idx].Value()))
			}
		}(idx, &wg)
	}
	wg.Wait() // done goroutines
	select {
	case err, ok := <-errCh:
		if ok {
			return err
		}
	default:
		return nil
	}

	return nil
}

// Generates the SQL for the materialized view table.
func (r *PGPopularRepository) sql(table MaterializedView, sort PopularSortCriteria) string {
	cat := "main"
	if table == AllView {
		cat = "all_time"
	}
	if table == AltView {
		cat = "alternate"
	}
	return fmt.Sprintf(`SELECT 
			average_per_year_rank as stats__average_rank, 
			average_per_year as stats__average,
			issue_count as stats__issue_count, 
			issue_count_rank as stats__issue_count_rank, 
			'%s' as stats__category,
			id,
			publisher_id, 
			name, 
			other_name,
			description,
			image,
			slug,
			vendor_image,
			vendor_url,
			vendor_description,
			publisher__id,
			publisher__slug,
			publisher__name
		FROM %s
		ORDER BY %s ASC
		LIMIT ?0 OFFSET ?1`, cat, table.Value(), string(sort))
}

// queries the database for the table and criteria.
func (r *PGPopularRepository) query(table MaterializedView, cr PopularCriteria) ([]*RankedCharacter, error) {
	var characters []*RankedCharacter
	_, err := r.db.Query(&characters, r.sql(table, cr.SortBy), cr.Limit, cr.Offset)
	if err != nil {
		return nil, err
	}
	slugs := make([]CharacterSlug, len(characters))
	for i, c := range characters {
		slugs[i] = c.Slug
	}
	thumbs, err := r.ctr.AllThumbnails(slugs...)
	if err != nil {
		return characters, err
	}
	for _, c := range characters {
		thumb := thumbs[c.Slug]
		if thumb != nil && (thumb.VendorImage != nil || thumb.Image != nil) {
			c.Thumbnails = thumbs[c.Slug]
		}
	}
	return characters, err
}

// NewPGAppearancesPerYearRepository creates the new appearances by year repository for postgres.
func NewPGAppearancesPerYearRepository(db *pg.DB) *PGAppearancesByYearsRepository {
	return &PGAppearancesByYearsRepository{
		db: db,
	}
}

// NewRedisAppearancesPerYearRepository creates the redis yearly appearances repository.
func NewRedisAppearancesPerYearRepository(client RedisClient) *RedisAppearancesByYearsRepository {
	return &RedisAppearancesByYearsRepository{redisClient: client, deserializer: &RedisYearlyAggregateDeserializer{}, serializer: &RedisYearlyAggregateSerializer{}}
}

// NewPGStatsRepository creates a new stats repository for the postgres implementation.
func NewPGStatsRepository(db *pg.DB) *PGStatsRepository {
	return &PGStatsRepository{db: db}
}

// NewPGCharacterIssueRepository creates the new character issue repository for the postgres implementation.
func NewPGCharacterIssueRepository(db *pg.DB) *PGCharacterIssueRepository {
	return &PGCharacterIssueRepository{db: db}
}

// NewPGCharacterSourceRepository creates the new character source repository for the postgres implementation.
func NewPGCharacterSourceRepository(db *pg.DB) *PGCharacterSourceRepository {
	return &PGCharacterSourceRepository{
		db: db,
	}
}

// NewPGPublisherRepository creates a new publisher repository for the postgres implementation.
func NewPGPublisherRepository(db *pg.DB) *PGPublisherRepository {
	return &PGPublisherRepository{db: db}
}

// NewPGIssueRepository creates a new issue repository for the postgres implementation.
func NewPGIssueRepository(db *pg.DB) *PGIssueRepository {
	return &PGIssueRepository{db: db}
}

// NewPGCharacterRepository creates the new character repository.
func NewPGCharacterRepository(db *pg.DB) *PGCharacterRepository {
	return &PGCharacterRepository{db: db}
}

// NewPGCharacterSyncLogRepository creates the new character sync log repository.
func NewPGCharacterSyncLogRepository(db *pg.DB) *PGCharacterSyncLogRepository {
	return &PGCharacterSyncLogRepository{db: db}
}

// NewPGPopularRepository creates the new popular characters repository for postgres
// and the redis cache for appearances.
func NewPGPopularRepository(db *pg.DB, ctr CharacterThumbRepository) *PGPopularRepository {
	return &PGPopularRepository{
		db: db,
		ctr: ctr,
	}
}

// NewPopularRefresher creates a new refresher for refreshing the materialized views.
func NewPopularRefresher(db *pg.DB) *PGPopularRepository {
	return &PGPopularRepository{
		db: db,
	}
}

// NewRedisCharacterThumbRepository creates a new redis character thumb repository.
func NewRedisCharacterThumbRepository(r RedisClient) *RedisCharacterThumbRepository {
	return &RedisCharacterThumbRepository{r: r}
}
