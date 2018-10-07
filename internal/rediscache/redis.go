package rediscache

import (
	"fmt"
	"github.com/go-redis/redis"
	"os"
	"strings"
	"sync"
)

var onceRedis sync.Once
var redisClient *redis.Client

type Configuration struct {
	Host     string
	Port     string
	Password string
}

func (c *Configuration) Address() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

func NewConfigurationFromEnv() *Configuration {
	switch env := strings.ToLower(os.Getenv("CC_ENVIRONMENT")); env {
	default:
		return &Configuration{
			Host:     os.Getenv("CC_REDIS_HOST"),
			Port:     os.Getenv("CC_REDIS_PORT"),
			Password: os.Getenv("CC_REDIS_PASSWORD")}
	}
}

// Creates a redis client from the configuration struct.
func Client(config *Configuration) *redis.Client {
	opts := redis.Options{
		Addr:     config.Address(),
		Password: config.Password,
	}
	client := redis.NewClient(&opts)
	return client
}

// Returns a singleton instance with a configuration from environment variables.
func Instance() *redis.Client {
	onceRedis.Do(func() {
		redisClient = Client(NewConfigurationFromEnv())
	})
	return redisClient
}
