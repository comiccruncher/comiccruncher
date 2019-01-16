package comic

import (
	"errors"
	"fmt"
	"github.com/aimeelaplant/comiccruncher/internal/log"
	"github.com/go-redis/redis"
	"go.uber.org/zap"
)

// Syncer is the interface for syncing yearly appearances from persistence to a cache.
type Syncer interface {
	// Syncs appearances from postgres to redis. Returns the number of issues synced and an error if any.
	Sync(slug CharacterSlug) (int, error)
}

// AppearancesSyncer to sync yearly appearances from Postgres to Redis.
type AppearancesSyncer struct {
	reader AppearancesByYearsRepository
	writer AppearancesByYearsWriter
}

// Sync gets all the character's appearances from the database and syncs them to Redis.
// returns the total number of issues synced and an error if any.
func (s *AppearancesSyncer) Sync(slug CharacterSlug) (int, error) {
	mainAppsPerYear, err := s.reader.Main(slug)
	if err != nil {
		return 0, err
	}
	if mainAppsPerYear.Aggregates != nil {
		log.COMIC().Info("main appearances to send to redis", zap.Int("total", mainAppsPerYear.Total()))
		err = s.writer.Set(mainAppsPerYear)
		if err != nil {
			return 0, err
		}
	}
	altAppsPerYear, err := s.reader.Alternate(slug)
	if err != nil {
		return 0, err
	}
	if altAppsPerYear.Aggregates != nil {
		log.COMIC().Info("alt appearances to send to redis", zap.Int("total", altAppsPerYear.Total()))
		err = s.writer.Set(altAppsPerYear)
		if err != nil {
			return 0, err
		}
	}
	all, err := s.reader.Both(slug)
	total := all.Total()
	log.COMIC().Info(
		"done syncing postgres appearances to redis!",
		zap.String("character", string(slug)),
		zap.String("appearances", fmt.Sprintf("%v", all.Aggregates)),
		zap.Int("total", total),
		zap.Error(err))
	return total, nil
}

// CharacterStatsSyncer is the interface for syncing characters.
type CharacterStatsSyncer interface {
	Sync(slug CharacterSlug) error
	SyncAll(characters []*Character) <- chan CharacterSyncResult
}

// RedisHmSetter is a redis client for setting hash-sets.
type RedisHmSetter interface {
	HMSet(key string, fields map[string]interface{}) *redis.StatusCmd
}

// RedisCharacterStatsSyncer is for syncing characters to redis.
type RedisCharacterStatsSyncer struct {
	r  RedisClient
	cr CharacterRepository
	pr PopularRepository
}

// Sync syncs the character's ranking stats to Redis.
func (s *RedisCharacterStatsSyncer) Sync(slug CharacterSlug) error {
	c, err := s.cr.FindBySlug(slug, false)
	if err != nil {
		return err
	}
	if c == nil {
		return errors.New("character doesn't exist or is disabled")
	}
	if c.Publisher.Slug == "marvel" {
		rcm, err := s.pr.FindOneByMarvel(c.ID)
		if err != nil {
			return err
		}
		rc, err := s.pr.FindOneByAll(c.ID)
		if err != nil {
			return err
		}
		if err := s.set(c, rc, rcm); err != nil {
			return err
		}
	}
	if c.Publisher.Slug == "dc" {
		rcd, err := s.pr.FindOneByDC(c.ID)
		if err != nil {
			return err
		}
		rc, err := s.pr.FindOneByAll(c.ID)
		if err != nil {
			return err
		}
		if err := s.set(c, rc, rcd); err != nil {
			return err
		}
	}
	return nil
}

// CharacterSyncResult is the result set for a synced character to redis and an error if any.
type CharacterSyncResult struct {
	Slug CharacterSlug
	Error error
}

func (s *RedisCharacterStatsSyncer) syncConcurrent(slugs <-chan CharacterSlug, results chan<- CharacterSyncResult) {
	for slug := range slugs {
		results <- CharacterSyncResult{Slug: slug, Error: s.Sync(slug)}
	}
}

// SyncAll syncs multiple characters to redis in goroutines.
func (s *RedisCharacterStatsSyncer) SyncAll(characters []*Character) <-chan CharacterSyncResult {
	slugLen := len(characters)
	slugCh := make(chan CharacterSlug, slugLen)
	defer close(slugCh)
	resultCh := make(chan CharacterSyncResult, slugLen)
	jobLimit := 50
	// make sure we aren't firing off goroutines greater than the job limit.
	if slugLen < jobLimit {
		jobLimit = slugLen
	}
	for i := 0; i < jobLimit; i++ {
		go s.syncConcurrent(slugCh, resultCh)
	}
	// send work over
	for _, chrctr := range characters {
		slugCh <- chrctr.Slug
	}
	// Return the results so caller can collect them.
	return resultCh
}

func (s *RedisCharacterStatsSyncer) set(c *Character, allTime *RankedCharacter, main *RankedCharacter) error {
	at := allTime.Stats
	ma := main.Stats
	m := make(map[string]interface{}, 8)
	m["all_time_issue_count_rank"] = at.IssueCountRank
	m["all_time_issue_count"] = at.IssueCount
	m["all_time_average_per_year"] = at.Average
	m["all_time_average_per_year_rank"] = at.AverageRank
	m["main_issue_count_rank"] = ma.IssueCountRank
	m["main_issue_count"] = ma.IssueCount
	m["main_average_per_year"] = ma.Average
	m["main_average_per_year_rank"] = ma.AverageRank
	key := fmt.Sprintf("%s:stats", c.Slug)
	return s.r.HMSet(key, m).Err()
}

// NewAppearancesSyncer returns a new appearances syncer
func NewAppearancesSyncer(r *PGRepositoryContainer, w *RedisAppearancesByYearsRepository) *AppearancesSyncer {
	return &AppearancesSyncer{
		reader: r.AppearancesByYearsRepository(),
		writer: w,
	}
}

// NewAppearancesSyncerRW returns a new appearances syncer with the reader and writer for the cache.
func NewAppearancesSyncerRW(r AppearancesByYearsRepository, w AppearancesByYearsWriter) *AppearancesSyncer {
	return &AppearancesSyncer{
		reader: r,
		writer: w,
	}
}

// NewCharacterStatsSyncer returns a new character stats syncer with dependencies.
func NewCharacterStatsSyncer(r RedisClient, cr CharacterRepository, pr PopularRepository) *RedisCharacterStatsSyncer {
	return &RedisCharacterStatsSyncer{
		r:  r,
		cr: cr,
		pr: pr,
	}
}
