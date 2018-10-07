package comic_test

import (
	"github.com/aimeelaplant/comiccruncher/internal/pgo"
	"github.com/aimeelaplant/comiccruncher/comic"
	"github.com/go-pg/pg"
)

// The test database instance.
var testInstance *pg.DB

// The repository container with the injected db instance.
var testContainer *comic.PGRepositoryContainer

func init() {
	// The test database instance.
	testInstance = pgo.MustInstanceTest()
	// The repository container with the injected db instance.
	testContainer = comic.NewPGRepositoryContainer(testInstance)
}
