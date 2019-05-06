package rediscache_test

import (
	"github.com/comiccruncher/comiccruncher/internal/rediscache"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestClient(t *testing.T) {
	c := &rediscache.Configuration{
		Host:     "host",
		Port:     "port",
		Password: "password",
	}
	r := rediscache.Client(c)
	assert.NotNil(t, r)
}
