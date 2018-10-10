package log

import (
	"github.com/gosimple/slug"
	"go.uber.org/zap"
	"os"
	"strings"
	"sync"
)

var once sync.Once
var loggers map[string]*zap.Logger
var rw sync.RWMutex

// Name  is the name type for a logger.
type Name string

// Value gets the string value.
func (n Name) Value() string {
	return string(n)
}

const (
	// DB is a logger name for db-related things.
	DB                      Name = "db"
	// Cerebro is the logger name for cerebro-related.
	Cerebro                 Name = "cerebro"
	// DCCharacterImporter is the logger name for importing dc stuff.
	DCCharacterImporter     Name = "dccharacterimporter"
	// MarvelCharacterImporter is the logge rname for importing marvel stuff.
	MarvelCharacterImporter Name = "marvelcharacterimporter"
	// Queue is the logger for queuing stuff.
	Queue                   Name = "charactersyncqueue"
	// Web is the logger for web-related stuff.
	Web                     Name = "web"
	// Migrations is the logger for migration-related stuff.
	Migrations              Name = "migrations"
	// Comic is the logger for comic package stuff.
	Comic                   Name = "comic"
	// Messaging is the logger for messaging stuff.
	Messaging               Name = "messaging"
)

// Creates a new logger with a configuration based on the environment.
func loggerfromEnv(name string) *zap.Logger {
	env := strings.ToLower(os.Getenv("CC_ENVIRONMENT"))
	if env == "production" {
		cfg := zap.NewProductionConfig()
		cfg.Encoding = "console"
		cfg.EncoderConfig = zap.NewDevelopmentEncoderConfig()
		logger, err := cfg.Build()
		logger = logger.Named(name)
		if err != nil {
			panic(err)
		}
		return logger
	} else {
		logger, err := zap.NewDevelopment()
		logger = logger.Named(slug.Make(name))
		if err != nil {
			panic(err)
		}
		return logger
	}
}

// Logger safely gets a logger from a name (concurrent-safe).
func Logger(name Name) *zap.Logger {
	once.Do(func() {
		loggers = make(map[string]*zap.Logger)
	})
	defer rw.Unlock()
	rw.Lock()
	logger, ok := loggers[name.Value()]
	if !ok {
		l := loggerfromEnv(name.Value())
		loggers[name.Value()] = l
		return loggers[name.Value()]
	}
	return logger
}

// Db is a method for getting the DB logger.
func Db() *zap.Logger {
	return Logger(DB)
}

// CEREBRO is a method for getting the Cerebro logger.
func CEREBRO() *zap.Logger {
	return Logger(Cerebro)
}

// WEB is a method for getting the Web logger.
func WEB() *zap.Logger {
	return Logger(Web)
}

// MIGRATIONS is a method for getting the Migrations logger.
func MIGRATIONS() *zap.Logger {
	return Logger(Migrations)
}

// DCIMPORTER is a method for getting the DC character importer logger.
func DCIMPORTER() *zap.Logger {
	return Logger(DCCharacterImporter)
}

// MARVELIMPORTER is a method for getting the Marvel character importer logger.
func MARVELIMPORTER() *zap.Logger {
	return Logger(MarvelCharacterImporter)
}

// QUEUE is a method for getting the queue logger.
func QUEUE() *zap.Logger {
	return Logger(Queue)
}

// COMIC is a method for getting the comic logger.
func COMIC() *zap.Logger {
	return Logger(Comic)
}

// MESSAGING is a method for getting the messaging logger.
func MESSAGING() *zap.Logger {
	return Logger(Messaging)
}
