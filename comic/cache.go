package comic

import (
	"fmt"
	"github.com/aimeelaplant/comiccruncher/internal/log"
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

// NewAppearancesSyncer returns a new appearances syncer
func NewAppearancesSyncer(r *PGRepositoryContainer, w *RedisAppearancesByYearsRepository) Syncer {
	return &AppearancesSyncer{
		reader: r.AppearancesByYearsRepository(),
		writer: w,
	}
}


func NewAppearancesSyncerRW(r AppearancesByYearsRepository, w AppearancesByYearsWriter) Syncer {
	return &AppearancesSyncer{
		reader: r,
		writer: w,
	}
}
