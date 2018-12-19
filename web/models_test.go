package web

import (
	"github.com/aimeelaplant/comiccruncher/comic"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewCharacter(t *testing.T) {
	c := &comic.Character{
		Name: "test",
		VendorImage: "test.jpg",
		Slug: "test",
	}
	th := &comic.CharacterThumbnails{
		Slug: "test",
	}
	ch := NewCharacter(c, th)
	assert.NotNil(t, ch)
}

func TestCharacterMarshalJSON(t *testing.T) {
	c := &comic.Character{
		Name: "test",
		VendorImage: "test.jpg",
		Slug: "test",
	}
	th := &comic.CharacterThumbnails{
		Slug: "test",
	}
	ch := NewCharacter(c, th)

	bts, err := ch.MarshalJSON()

	expected := `{"publisher":{"name":"","slug":""},"name":"test","other_name":"","description":"","vendor_url":"","vendor_description":"","image":"","vendor_image":"https://d2jsu6fyd1g4ln.cloudfront.net/test.jpg","thumbnails":null}`
	assert.Nil(t, err)
	assert.Equal(t, expected, string(bts))
}

func TestCharacterMarshalJSONThumbs(t *testing.T) {
	c := &comic.Character{
		Name: "test",
		VendorImage: "test.jpg",
		Slug: "test",
	}
	th := &comic.CharacterThumbnails{
		Slug: "test",
		VendorImage: &comic.ThumbnailSizes{
			Small: "test.jpg",
			Medium: "test.jpg",
			Large: "test.jpg",
		},
	}
	ch := NewCharacter(c, th)
	bts, err := ch.MarshalJSON()
	expected := `{"publisher":{"name":"","slug":""},"name":"test","other_name":"","description":"","vendor_url":"","vendor_description":"","image":"","vendor_image":"https://d2jsu6fyd1g4ln.cloudfront.net/test.jpg","thumbnails":{"slug":"test","image":null,"vendor_image":{"small":"https://d2jsu6fyd1g4ln.cloudfront.net/test.jpg","medium":"https://d2jsu6fyd1g4ln.cloudfront.net/test.jpg","large":"https://d2jsu6fyd1g4ln.cloudfront.net/test.jpg"}}}`

	assert.Nil(t, err)
	assert.Equal(t, expected, string(bts))
}
