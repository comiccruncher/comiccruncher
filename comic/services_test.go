package comic_test

import (
	"github.com/aimeelaplant/comiccruncher/comic"
	"github.com/stretchr/testify/assert"
	"testing"
)

// db test to test query.
func TestCharacterService_MustNormalizeSources(t *testing.T) {
	svc := comic.NewCharacterService(testContainer)
	c, err := svc.Character("emma-frost")
	assert.Nil(t, err)
	svc.MustNormalizeSources(c)
}
