package cerebro_test

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"github.com/aimeelaplant/comiccruncher/cerebro"
)

func TestParsePublisherName(t *testing.T) {
	testData := map[string]string{
		"Marvel": "Storm (Marvel) () (blah)",
		"":       "Storm",
	}

	for k, v := range testData {
		assert.Equal(t, k, cerebro.ParsePublisherName(v))
	}
}

func TestParseCharacterName(t *testing.T) {
	testData := map[string]string{
		"Storm":      "Storm (Marvel) () (blah)",
		"Something)": "Something) (Another)",
	}

	for k, v := range testData {
		assert.Equal(t, k, cerebro.ParseCharacterName(v))
	}
}

func TestSearchableName(t *testing.T) {
	name := "Test. Ing (ABC)"
	parensIndex := strings.Index(name, "(")

	assert.Equal(t, "ABC", cerebro.SearchableName(name, parensIndex))
	assert.Equal(t, "Test Ing ABC", cerebro.SearchableName(name, -1))
}
