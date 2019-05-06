package pgo_test

import (
	"github.com/comiccruncher/comiccruncher/internal/pgo"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewConfiguration(t *testing.T) {
	c := pgo.NewConfiguration("a", "b", "c", "d", "e", true)
	assert.NotNil(t, c)
	assert.Equal(t, "a", c.Host)
	assert.Equal(t, "b", c.Port)
	assert.Equal(t, "c", c.Database)
	assert.Equal(t, "d", c.User)
	assert.Equal(t, "e", c.Password)
	assert.True(t, c.LogMode)
}
