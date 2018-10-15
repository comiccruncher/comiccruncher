package comic_test

import (
	"github.com/aimeelaplant/comiccruncher/comic"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCharacterService_NormalizeSources(t *testing.T) {
	svc := comic.NewCharacterService(testContainer)
	err := svc.NormalizeSources(1)
	assert.Nil(t, err)
}
