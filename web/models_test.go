package web

import (
	"encoding/json"
	"github.com/comiccruncher/comiccruncher/comic"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewCharacter(t *testing.T) {
	c := &comic.Character{
		Name:        "test",
		VendorImage: "test.jpg",
		Slug:        "test",
	}
	th := &comic.CharacterThumbnails{
		Slug: "test",
	}
	ch := NewCharacter(c, th)
	assert.NotNil(t, ch)
}

func TestCharacterMarshalJSON(t *testing.T) {
	c := &comic.Character{
		Name:        "test",
		VendorImage: "test.jpg",
		Slug:        "test-test",
	}
	th := &comic.CharacterThumbnails{
		Slug: "test-test",
		VendorImage: &comic.ThumbnailSizes{
			Small:  "test.jpg",
			Medium: "test.jpg",
			Large:  "test.jpg",
		},
	}
	ch := NewCharacter(c, th)
	bts, err := ch.MarshalJSON()
	assert.Nil(t, err)

	expected := `{"publisher":{"name":"","slug":""},"name":"test","other_name":"","description":"","vendor_url":"","vendor_description":"","slug":"test-test","image":"","vendor_image":"https://d2jsu6fyd1g4ln.cloudfront.net/test.jpg","thumbnails":{"slug":"test-test","image":null,"vendor_image":{"small":"https://d2jsu6fyd1g4ln.cloudfront.net/test.jpg","medium":"https://d2jsu6fyd1g4ln.cloudfront.net/test.jpg","large":"https://d2jsu6fyd1g4ln.cloudfront.net/test.jpg"}}}`
	assert.Equal(t, expected, string(bts))

	m := make(map[string]interface{})
	err = json.Unmarshal(bts, &m)

	assert.Nil(t, err)

	assertions := map[string]string{
		"name":               "test",
		"other_name":         "",
		"description":        "",
		"vendor_description": "",
		"vendor_url":         "",
		"slug":               "test-test",
		"image":              "",
		"vendor_image":       "https://d2jsu6fyd1g4ln.cloudfront.net/test.jpg",
	}

	for key, val := range assertions {
		assertKeyHasValue(t, key, val, m)
	}
	_, ok := m["publisher"]
	assert.True(t, ok)

	thumbs, ok := m["thumbnails"]
	assert.True(t, ok)
	ths := thumbs.(map[string]interface{})
	assertKeyHasValue(t, "slug", "test-test", ths)
}

func assertKeyHasValue(t *testing.T, key, value string, m map[string]interface{}) {
	val, ok := m[key]
	assert.True(t, ok)
	assert.Equal(t, value, val)
}
