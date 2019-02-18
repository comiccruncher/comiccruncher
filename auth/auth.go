package auth

import (
	"fmt"
	"github.com/go-redis/redis"
	"time"
)

// Redis is the interface for interacting with the Redis client.
type Redis interface {
	HMSet(key string, fields map[string]interface{}) *redis.StatusCmd
	Exists(keys ...string) *redis.IntCmd
}

// Token is the struct for details about an issued JWT token.
type Token struct  {
	Payload 	string
	UUID 		string
	CreatedAt   time.Time
}

// TokenRepository is the interface for token repos.
type TokenRepository interface {
	Create(t *Token) error
	Exists(uuid string) (bool, error)
}

// RedisTokenRepository is the implementation of the token repository.
type RedisTokenRepository struct {
	client Redis
}

// Create creates a new token repository.
func (r *RedisTokenRepository) Create(t *Token) error {
	m := make(map[string]interface{}, 3)
	m["CreatedAt"] = t.CreatedAt.String()
	// m["Payload"] = t.Payload
	return r.client.HMSet(fmt.Sprintf("token:%s", t.UUID), m).Err()
}

// Exists checks if the uuid exists.
func (r *RedisTokenRepository) Exists(uuid string) (bool, error) {
	cmd := r.client.Exists(fmt.Sprintf("token:%s", uuid))
	val, err := cmd.Result()
	if err != nil {
		return false, err
	}
	if val == 1 {
		return true, nil
	}
	return false, nil
}

// NewToken creates a new token struct
func NewToken(payload string, UUID string) *Token {
	return &Token{Payload: payload, UUID: UUID}
}

// NewRedisTokenRepository creates a new Redis token repository.
func NewRedisTokenRepository(r Redis) *RedisTokenRepository {
	return &RedisTokenRepository{client: r}
}
