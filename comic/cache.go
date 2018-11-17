package comic

import (
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
