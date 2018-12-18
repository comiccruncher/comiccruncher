package comic

import (
	"fmt"
	"github.com/go-redis/redis"
	"time"
)

// RedisClient is the interface for interacting with Redis.
type RedisClient interface {
	Get(key string) *redis.StringCmd
	MGet(keys ...string) *redis.SliceCmd
	Set(key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	HMSet(key string, fields map[string]interface{}) *redis.StatusCmd
	HGetAll(key string) *redis.StringStringMapCmd
}

// Returns the redis key for appearances per year for a character and appearance type.
func redisKey(key CharacterSlug, cat AppearanceType) string {
	return fmt.Sprintf("%s:%s:%d", key, appearancesPerYearsKey, cat)
}

// redisThumbnailKey returns the key for character profile thumbnails.
func redisThumbnailKey(s CharacterSlug) string {
	return s.Value() + ":profile:thumbnails"
}
