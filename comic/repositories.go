package comic

import (
	"fmt"
	"github.com/go-pg/pg"
	"github.com/go-redis/redis"
	"github.com/gosimple/slug"
	"strconv"
	"strings"
	"time"
)

const appearancesPerYearsKey = "yearly"

// The repository interface for publishers.
type PublisherRepository interface {
	FindBySlug(slug PublisherSlug) (*Publisher, error)
}

// The repository interface for issues.
type IssueRepository interface {
	Create(issue *Issue) error
	CreateAll(issues []*Issue) error
	FindByVendorId(vendorId string) (*Issue, error)
	FindAll(c IssueCriteria) ([]*Issue, error)
}

// The repository interface for characters.
type CharacterRepository interface {
	Create(c *Character) error
	Update(c *Character) error
	FindBySlug(slug CharacterSlug, includeIsDisabled bool) (*Character, error)
	FindAll(cr CharacterCriteria) ([]*Character, error)
	UpdateAll(characters []*Character) error
	Remove(id CharacterID) error
	Total(cr CharacterCriteria) (int64, error)
}

// The repository interface for character sources.
type CharacterSourceRepository interface {
	Create(s *CharacterSource) error
	FindAll(criteria CharacterSourceCriteria) ([]*CharacterSource, error)
	Remove(id uint) error
	// Run a raw query on the character sources.
	Raw(query string, params ...interface{}) error
	Update(s *CharacterSource) error
}

// The repository interface for character sync logs.
type CharacterSyncLogRepository interface {
	Create(s *CharacterSyncLog) error
	FindAllByCharacterId(characterID CharacterID) ([]*CharacterSyncLog, error)
	Update(s *CharacterSyncLog) error
	FindById(id CharacterSyncLogID) (*CharacterSyncLog, error)
}

// The repository interface for character issues.
type CharacterIssueRepository interface {
	CreateAll([]*CharacterIssue) error
	Create(ci *CharacterIssue) error
	FindOneBy(characterID CharacterID, issueID IssueID) (*CharacterIssue, error)
	InsertFast(issues []*CharacterIssue) error
}

// The repository interface for getting a characters appearances per year.
type AppearancesByYearsRepository interface {
	//	Set(character *AppearancesByYears) error
	Both(slug CharacterSlug) (AppearancesByYears, error)
	Main(slug CharacterSlug) (AppearancesByYears, error)
	Alternate(slug CharacterSlug) (AppearancesByYears, error)
	List(slug CharacterSlug) ([]AppearancesByYears, error)
}

// The repository interface for general stats about the db.
type StatsRepository interface {
	Stats() (Stats, error)
}

// The postgres implementation for the appearances per year repository.
type PGAppearancesByYearsRepository struct {
	db *pg.DB
}

// The postgres implementation for the character repository.
type PGCharacterRepository struct {
	db *pg.DB
}

// The postgres implementation for the publisher repository.
type PGPublisherRepository struct {
	db *pg.DB
}

// The postgres implementation for the issue repository.
type PGIssueRepository struct {
	db *pg.DB
}

// The postgres implementation for the character source repository.
type PGCharacterSourceRepository struct {
	db *pg.DB
}

// The postgres implementation for the character issue repository.
type PGCharacterIssueRepository struct {
	db *pg.DB
}

// The postgres implementation for the character sync log repository.
type PGCharacterSyncLogRepository struct {
	db *pg.DB
}

// The postgres implementation for the stats repository.
type PGStatsRepository struct {
	db *pg.DB
}

// The container for all the postgres repositories.
type PGRepositoryContainer struct {
	publisherRepository          PublisherRepository
	issueRepository              IssueRepository
	characterRepository          CharacterRepository
	characterIssueRepository     CharacterIssueRepository
	characterSourceRepository    CharacterSourceRepository
	characterSyncLogRepository   CharacterSyncLogRepository
	appearancesByYearsRepository *PGAppearancesByYearsRepository
	statsRepository StatsRepository
}

// PGYearlyAppearancesRepository
// The Redis implementation for appearances per year repository.
type RedisAppearancesByYearsRepository struct {
	redisClient *redis.Client
}

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

func (r *PGCharacterRepository) Update(c *Character) error {
	return r.db.Update(c)
}

func (r *PGCharacterRepository) FindBySlug(slug CharacterSlug, includeIsDisabled bool) (*Character, error) {
	if result, err := r.FindAll(CharacterCriteria{
		Slugs: []CharacterSlug{slug},
		IncludeIsDisabled: includeIsDisabled,
		Limit: 1}); err != nil {
		return nil, err
	} else if len(result) != 0 {
		return result[0], nil
	}
	return nil, nil
}

func (r *PGCharacterRepository) Remove(id CharacterID) error {
	if _, err := r.db.Model(&Character{}).Where("id = ?", id).Delete(); err != nil {
		return err
	}
	return nil
}

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

func (r *PGCharacterRepository) UpdateAll(characters []*Character) error {
	if len(characters) > 0 {
		_, err := r.db.Model(&characters).Update()
		return err
	}
	return nil
}

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

func (r *PGCharacterIssueRepository) Create(ci *CharacterIssue) error {
	if _, err := r.db.Model(ci).Returning("*").Insert(); err != nil {
		return err
	}
	return nil
}

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

func (r *PGCharacterSourceRepository) Create(s *CharacterSource) error {
	_, err := r.db.Model(s).Insert(s)
	return err
}

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

func (r *PGCharacterSourceRepository) Remove(id uint) error {
	_, err := r.db.Model(&CharacterSource{}).Where("id = ?", id).Delete()
	return err
}

func (r *PGCharacterSourceRepository) Raw(query string, params ...interface{}) error {
	if _, err := r.db.Exec(query, params...); err != nil {
		return err
	}
	return nil
}

func (r *PGCharacterSourceRepository) Update(s *CharacterSource) error {
	return r.db.Update(s)
}

func (r *PGCharacterSyncLogRepository) Create(s *CharacterSyncLog) error {
	_, err := r.db.Model(s).Insert(s)
	return err
}

func (r *PGCharacterSyncLogRepository) FindAllByCharacterId(characterID CharacterID) ([]*CharacterSyncLog, error) {
	var syncLogs []*CharacterSyncLog
	if err := r.db.
		Model(&syncLogs).
		Where("character_id = ?", characterID).
		Select(); err != nil {
		if err == pg.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return syncLogs, nil
}

func (r *PGCharacterSyncLogRepository) Update(l *CharacterSyncLog) error {
	return r.db.Update(l)
}

func (r *PGCharacterSyncLogRepository) FindById(id CharacterSyncLogID) (*CharacterSyncLog, error) {
	syncLog := &CharacterSyncLog{}
	if err := r.db.Model(syncLog).Where("id = ?", id).Select(); err != nil {
		if err == pg.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return syncLog, nil
}


func (r *PGIssueRepository) Create(issue *Issue) error {
	_, err := r.db.Model(issue).Returning("*").Insert(issue)
	return err
}

func (r *PGIssueRepository) CreateAll(issues []*Issue) error {
	// pg-go returns an error if you bulk-insert an empty slice.
	if len(issues) > 0 {
		return r.db.Insert(&issues)
	}
	return nil
}

func (r *PGIssueRepository) FindByVendorId(vendorId string) (*Issue, error) {
	if issues, err := r.FindAll(IssueCriteria{VendorIds: []string{vendorId}, Limit: 1}); err != nil {
		return nil, err
	} else if len(issues) != 0 {
		return issues[0], nil
	} else {
		return nil, nil
	}
}

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

// Gets all of a character's appearances per year, which includes its main and alternate counterparts.
func (r *PGAppearancesByYearsRepository) Both(slug CharacterSlug) (AppearancesByYears, error) {
	var appearancesPerYears []YearlyAggregate
	_, err := r.db.Query(
		&appearancesPerYears,
		`SELECT years.year AS year, count(issues.id) AS count
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
		WHERE c.slug = ?0
	) issues ON years.year = date_part('year', issues.sale_date)
	GROUP BY years.year
	ORDER BY years.year`, slug)

	if err != nil {
		return AppearancesByYears{}, err
	}
	if len(appearancesPerYears) > 0 {
		return AppearancesByYears{CharacterSlug: slug, Aggregates: appearancesPerYears, Category: Main | Alternate}, nil
	}
	return AppearancesByYears{}, nil
}

// Gets a character's main appearances per year. No alternate realities.
func (r *PGAppearancesByYearsRepository) Main(slug CharacterSlug) (AppearancesByYears, error) {
	var appearances []YearlyAggregate
	_, err := r.db.Query(
		&appearances,
		fmt.Sprintf(`
		SELECT years.year as year, count(issues.id) AS count
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
		GROUP BY years.year
		ORDER BY years.year`, Main), slug)
	if err != nil {
		return AppearancesByYears{}, err
	}
	if len(appearances) > 0 {
		return AppearancesByYears{CharacterSlug: slug, Aggregates: appearances, Category: Main}, nil
	}
	return AppearancesByYears{}, nil
}

// Gets a character's alternate appearances per year. Yes to alternate realities.
func (r *PGAppearancesByYearsRepository) Alternate(slug CharacterSlug) (AppearancesByYears, error) {
	var appearances []YearlyAggregate
	_, err := r.db.Query(&appearances, fmt.Sprintf(`
	SELECT years.year AS year, count(issues.id) AS count
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
	GROUP BY years.year
	ORDER BY years.year`, Alternate), slug)
	if err != nil {
		return AppearancesByYears{}, err
	}
	if len(appearances) > 0 {
		return AppearancesByYears{CharacterSlug: slug, Aggregates: appearances, Category: Alternate}, nil
	}
	return AppearancesByYears{}, nil
}

// Gets a slice of a character's main and alternate appearances.
func (r *PGAppearancesByYearsRepository) List(slug CharacterSlug) ([]AppearancesByYears, error) {
	main, err := r.Main(slug)
	if err != nil {
		return nil, err
	}
	alt, err := r.Alternate(slug)
	if err != nil {
		return nil, err
	}
	list := make([]AppearancesByYears, 0)
	if main.Aggregates != nil {
		list = append(list, main)
	}
	if alt.Aggregates != nil {
		list = append(list, alt)
	}
	return list, nil
}

func (r *RedisAppearancesByYearsRepository) byType(slug CharacterSlug, t AppearanceType) (AppearancesByYears, error) {
	sortedResult, err := r.redisClient.Sort(appearancesPerYearsZKey(slug, t), &redis.Sort{}).Result()
	if err != nil {
		return AppearancesByYears{}, err
	}
	if len(sortedResult) == 0 {
		return AppearancesByYears{}, nil
	}
	c := AppearancesByYears{}
	// Get the years sorted by member. TODO: Ugh, probably a better way to do this. Still faster than SQL!
	for _, year := range sortedResult {
		// Get each year's score.
		zScore, err := r.redisClient.ZScore(appearancesPerYearsZKey(slug, t), year).Result()
		if err != nil {
			return AppearancesByYears{}, err
		}
		yearInt, err := strconv.Atoi(year)
		if err != nil {
			return AppearancesByYears{}, err
		}
		c.AddAppearance(YearlyAggregate{Year: yearInt, Count: int(zScore)})
	}
	c.Category = t
	c.CharacterSlug = slug
	return c, nil
}

func (r *RedisAppearancesByYearsRepository) Both(slug CharacterSlug) (AppearancesByYears, error) {
	return r.byType(slug, Main|Alternate)
}

func (r *RedisAppearancesByYearsRepository) Main(slug CharacterSlug) (AppearancesByYears, error) {
	return r.byType(slug, Main)
}

func (r *RedisAppearancesByYearsRepository) Alternate(slug CharacterSlug) (AppearancesByYears, error) {
	return r.byType(slug, Alternate)
}

func (r *RedisAppearancesByYearsRepository) List(slug CharacterSlug) ([]AppearancesByYears, error) {
	main, err := r.Main(slug)
	if err != nil {
		return nil, err
	}
	alt, err := r.Alternate(slug)
	if err != nil {
		return nil, err
	}
	both := make([]AppearancesByYears, 0)
	if main.Aggregates != nil {
		both = append(both, main)
	}
	if alt.Aggregates != nil {
		both = append(both, alt)
	}
	return both, nil
}

// Set's the character's info like this: HMSET KEY name "character.Name"
// Sets the character's appearances like this: ZADDNX KEY:yearly 1 "1979" 2 "1980"
func (r *RedisAppearancesByYearsRepository) Set(character AppearancesByYears) error {
	key := character.CharacterSlug
	var zScores []redis.Z
	for _, appearance := range character.Aggregates {
		z := redis.Z{
			Score:  float64(appearance.Count),
			Member: appearance.Year,
		}
		zScores = append(zScores, z)
	}
	if len(zScores) > 0 {
		zResult := r.redisClient.ZAdd(appearancesPerYearsZKey(key, character.Category), zScores...)
		return zResult.Err()
	}
	return nil
}

// Gets the publisher repository.
func (c *PGRepositoryContainer) PublisherRepository() PublisherRepository {
	return c.publisherRepository
}

// Gets the issue repository.
func (c *PGRepositoryContainer) IssueRepository() IssueRepository {
	return c.issueRepository
}

// Gets the character repository.
func (c *PGRepositoryContainer) CharacterRepository() CharacterRepository {
	return c.characterRepository
}

// Gets the character issue repository.
func (c *PGRepositoryContainer) CharacterIssueRepository() CharacterIssueRepository {
	return c.characterIssueRepository
}

// Gets the character source repository.
func (c *PGRepositoryContainer) CharacterSourceRepository() CharacterSourceRepository {
	return c.characterSourceRepository
}

// Gets the character sync log repository.
func (c *PGRepositoryContainer) CharacterSyncLogRepository() CharacterSyncLogRepository {
	return c.characterSyncLogRepository
}

func (c *PGRepositoryContainer) AppearancesByYearsRepository() *PGAppearancesByYearsRepository {
	return c.appearancesByYearsRepository
}

func (c *PGRepositoryContainer) StatsRepository() StatsRepository {
	return c.statsRepository
}

func appearancesPerYearsZKey(key CharacterSlug, cat AppearanceType) string {
	return fmt.Sprintf("%s:%s:%d", key, appearancesPerYearsKey, cat)
}

// Returns a pointer to an instance of the gorm appearance repository.
func NewPGAppearancesPerYearRepository(db *pg.DB) *PGAppearancesByYearsRepository {
	return &PGAppearancesByYearsRepository{
		db: db,
	}
}

// Returns a pointer to an instance of the redis yearly appearances repository.
func NewRedisAppearancesPerYearRepository(client *redis.Client) *RedisAppearancesByYearsRepository {
	return &RedisAppearancesByYearsRepository{redisClient: client}
}

func NewPGStatsRepository(db *pg.DB) StatsRepository {
	return &PGStatsRepository{db: db}
}

func NewPGCharacterIssueRepository(db *pg.DB) CharacterIssueRepository {
	return &PGCharacterIssueRepository{db: db}
}

func NewPGCharacterSourceRepository(db *pg.DB) CharacterSourceRepository {
	return &PGCharacterSourceRepository{
		db: db,
	}
}

func NewPGPublisherRepository(db *pg.DB) PublisherRepository {
	return &PGPublisherRepository{db: db}
}

func NewPGIssueRepository(db *pg.DB) IssueRepository {
	return &PGIssueRepository{db: db}
}

func NewPGCharacterRepository(db *pg.DB) CharacterRepository {
	return &PGCharacterRepository{db: db}
}

func NewPGCharacterSyncLogRepository(db *pg.DB) CharacterSyncLogRepository {
	return &PGCharacterSyncLogRepository{db: db}
}

func NewPGRepositoryContainer(db *pg.DB) *PGRepositoryContainer {
	return &PGRepositoryContainer{
		publisherRepository:          NewPGPublisherRepository(db),
		issueRepository:              NewPGIssueRepository(db),
		characterRepository:          NewPGCharacterRepository(db),
		characterSyncLogRepository:   NewPGCharacterSyncLogRepository(db),
		characterSourceRepository:    NewPGCharacterSourceRepository(db),
		characterIssueRepository:     NewPGCharacterIssueRepository(db),
		appearancesByYearsRepository: NewPGAppearancesPerYearRepository(db),
		statsRepository:			  NewPGStatsRepository(db),
	}
}

