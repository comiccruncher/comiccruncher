package pgo

import (
	"errors"
	"fmt"
	"github.com/aimeelaplant/comiccruncher/internal/log"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"os"
	"sync"
	"time"
)

var once sync.Once
var instance *pg.DB

// Configuration is for configuring the connection to postgres.
type Configuration struct {
	Host     string
	Port     string
	Database string
	User     string
	Password string
	LogMode  bool
}

// NewConfiguration creates a new database configuration from the parameters.
func NewConfiguration(host, port, database, user, password string, logMode bool) *Configuration {
	return &Configuration{
		Host:     host,
		Port:     port,
		Database: database,
		User:     user,
		Password: password,
		LogMode:  logMode,
	}
}

// NewTestConfiguration is a test configuration from environment variables.
func NewTestConfiguration() *Configuration {
	return NewConfiguration(
		os.Getenv("CC_POSTGRES_TEST_HOST"),
		os.Getenv("CC_POSTGRES_TEST_PORT"),
		os.Getenv("CC_POSTGRES_TEST_DB"),
		os.Getenv("CC_POSTGRES_TEST_USER"),
		os.Getenv("CC_POSTGRES_TEST_PASSWORD"),
		false, // Makes it easier to read test results. TODO: print out db logs at postgres layer.
	)
}

// NewDevelopmentConfiguration is a development configuration from environment variables.
func NewDevelopmentConfiguration() *Configuration {
	return NewConfiguration(
		os.Getenv("CC_POSTGRES_DEV_HOST"),
		os.Getenv("CC_POSTGRES_DEV_PORT"),
		os.Getenv("CC_POSTGRES_DEV_DB"),
		os.Getenv("CC_POSTGRES_DEV_USER"),
		os.Getenv("CC_POSTGRES_DEV_PASSWORD"),
		false, // Can look at Docker logs for postgres container instead.
	)
}

// NewProductionConfiguration is a production configuration from environment variables.
func NewProductionConfiguration() *Configuration {
	return NewConfiguration(
		os.Getenv("CC_POSTGRES_HOST"),
		os.Getenv("CC_POSTGRES_PORT"),
		os.Getenv("CC_POSTGRES_DB"),
		os.Getenv("CC_POSTGRES_USER"),
		os.Getenv("CC_POSTGRES_PASSWORD"),
		false,
	)
}

// DB creates a new db connection from the configuration struct.
func DB(config *Configuration) (*pg.DB, error) {
	db := pg.Connect(
		&pg.Options{
			User:     config.User,
			Password: config.Password,
			Addr:     fmt.Sprintf("%s:%s", config.Host, config.Port),
			Database: config.Database,
		},
	)
	var queryError error
	if config.LogMode {
		db.OnQueryProcessed(
			func(event *pg.QueryProcessedEvent) {
				query, err := event.FormattedQuery()
				if err != nil {
					queryError = err
				}
				log.Db().Info(query, zap.String("time", time.Since(event.StartTime).String()))
			},
		)
	}
	return db, queryError
}

// InstanceTest returns a new instance of test database with configured env vars.
func InstanceTest() (*pg.DB, error) {
	return DB(NewTestConfiguration())
}

// MustInstanceTest returns a new instance of the test database with configured env vars and panics if there is an error.
func MustInstanceTest() *pg.DB {
	db, err := InstanceTest()
	if err != nil {
		panic(err)
	}
	return db
}

// Instance returns a singleton instance to the database.
func Instance() (*pg.DB, error) {
	var err error
	once.Do(func() {
		if os.Getenv("CC_ENVIRONMENT") == "development" {
			instance, err = DB(NewDevelopmentConfiguration())
		} else if os.Getenv("CC_ENVIRONMENT") == "production" {
			instance, err = DB(NewProductionConfiguration())
		}
	})
	if instance == nil {
		return instance, errors.New("unknown environment")
	}
	return instance, err
}

// MustInstance returns a singleton instance to the database and panics if there's an error.
func MustInstance() *pg.DB {
	instance, err := Instance()
	if err != nil {
		log.Db().Error("error", zap.Error(err))
		panic(err)
	}
	return instance
}
