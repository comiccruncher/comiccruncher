package web

import (
	"encoding/json"
	"github.com/aimeelaplant/comiccruncher/comic"
	"os"
)

var cdnURL = os.Getenv("CC_CDN_URL")

// Character is a character with thumbnails attached.
type Character struct {
	*comic.Character
	*comic.CharacterThumbnails
}

// MarshalJSON overrides JSON marshaling for presentation.
func (c *Character) MarshalJSON() ([]byte, error) {
	ch := c.Character
	if ch.Image != "" {
		ch.Image = cdnURL + "/" + ch.Image
	}
	if ch.VendorImage != "" {
		ch.VendorImage = cdnURL + "/" + ch.VendorImage
	}
	thumbs := c.CharacterThumbnails
	cdnUrlForThumbnails(thumbs)
	if thumbs.Image == nil && thumbs.VendorImage == nil {
		thumbs = nil
	}
	type Alias Character
	return json.Marshal(&struct {
		*Alias
		Slug string `json:"slug"`
		Image string `json:"image"`
		VendorImage string `json:"vendor_image"`
		Thumbnails *comic.CharacterThumbnails `json:"thumbnails"`
	}{
		Alias:       (*Alias)(c),
		Slug: ch.Slug.Value(),
		Image: ch.Image,
		VendorImage: ch.VendorImage,
		Thumbnails: thumbs,
	})
}

// NewCharacter creates a new character for presentation.
func NewCharacter(c *comic.Character, th *comic.CharacterThumbnails) *Character {
	return &Character{
		Character: c,
		CharacterThumbnails: th,
	}
}

func cdnUrlForThumbnails(thumbs *comic.CharacterThumbnails) {
	if thumbs != nil {
		if thumbs.VendorImage != nil {
			cdnUrlForSizes(thumbs.VendorImage)
		}
		if thumbs.Image != nil {
			cdnUrlForSizes(thumbs.Image)
		}
	}
}

func cdnUrlForSizes(sizes *comic.ThumbnailSizes) {
	if sizes != nil {
		sizes.Small = cdnURL + "/" + sizes.Small
		sizes.Medium = cdnURL + "/" + sizes.Medium
		sizes.Large = cdnURL + "/" + sizes.Large
	}
}
