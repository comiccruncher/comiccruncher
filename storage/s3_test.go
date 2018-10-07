package storage_test

import (
	"github.com/aimeelaplant/comiccruncher/storage"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestCrc32TimeNamingStrategy(t *testing.T) {
	assert.True(t, strings.HasSuffix(storage.Crc32TimeNamingStrategy()("test.txt"), ".txt"))
}
