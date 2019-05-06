package listutil_test

import (
	"github.com/comiccruncher/comiccruncher/internal/listutil"
	"github.com/stretchr/testify/assert"
	"testing"
)

func in(slice []string, s string) bool {
	for i := range slice {
		if slice[i] == s {
			return true
		}
	}
	return false
}

func TestStringKeys(t *testing.T) {
	m := make(map[string]string)
	m["a"] = "a"
	m["b"] = "b"
	m["c"] = "c"
	stringKeys := listutil.StringKeys(m)
	assert.True(t, in(stringKeys, "a"))
	assert.True(t, in(stringKeys, "b"))
	assert.True(t, in(stringKeys, "c"))
}

func TestStringInSlice(t *testing.T) {
	strs := []string{"a", "b", "c"}
	assert.True(t, listutil.StringInSlice(strs, "a"))
	assert.True(t, listutil.StringInSlice(strs, "b"))
	assert.True(t, listutil.StringInSlice(strs, "c"))
	assert.False(t, listutil.StringInSlice(strs, "d"))
}
