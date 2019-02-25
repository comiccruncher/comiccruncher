package comic_test

import (
	"github.com/aimeelaplant/comiccruncher/internal/pgo"
	"github.com/go-pg/pg"
)

// The test database instance.
var testInstance *pg.DB

func init() {
	// The test database instance.
	testInstance = pgo.MustInstanceTest()
}
