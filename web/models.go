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
	if c.Character.Image != "" {
		c.Character.Image = "https://d2jsu6fyd1g4ln.cloudfront.net" + "/" + c.Character.Image
	}
	if c.Character.VendorImage != "" {
		c.Character.VendorImage = "https://d2jsu6fyd1g4ln.cloudfront.net" + "/" + c.Character.VendorImage
	}
	cdnUrlForThumbnails(c.CharacterThumbnails)
	if c.CharacterThumbnails.Image == nil && c.CharacterThumbnails.VendorImage == nil {
		c.CharacterThumbnails = nil
	}
	type Alias Character
	return json.Marshal(&struct {
		*Alias
		Image string `json:"image"`
		VendorImage string `json:"vendor_image"`
		Thumbnails *comic.CharacterThumbnails `json:"thumbnails"`
	}{
		Alias:       (*Alias)(c),
		Image: c.Character.Image,
		VendorImage: c.Character.VendorImage,
		Thumbnails: c.CharacterThumbnails,
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
