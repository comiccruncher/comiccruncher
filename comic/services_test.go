package comic_test

import (
	"testing"
	"github.com/aimeelaplant/comiccruncher/comic"
	"github.com/stretchr/testify/assert"
)

func TestCharacterService_NormalizeSources(t *testing.T) {
	svc := comic.NewCharacterService(testContainer)
	err := svc.NormalizeSources(1)
	assert.Nil(t, err)
}
