package web

import (
	"encoding/json"
	"fmt"
	"github.com/aimeelaplant/comiccruncher/comic"
	"os"
)

var cdnURL = os.Getenv("CC_CDN_URL")

// Character is the character struct a character
type Character struct {
	comic.Character                            // just extend the character from the comic package
	Appearances     []comic.AppearancesByYears `json:"appearances"`
}

// MarshalJSON overrides the marshaling of JSON with presentation for CDN urls.
func (c *Character) MarshalJSON() ([]byte, error) {
	type Alias Character
	return json.Marshal(&struct {
		Image       string `json:"image"`
		VendorImage string `json:"vendor_image"`
		*Alias
	}{
		Alias:       (*Alias)(c),
		Image:       fmt.Sprintf("%s/%s", cdnURL, c.Image),
		VendorImage: fmt.Sprintf("%s/%s", cdnURL, c.VendorImage),
	})
}

// NewCharacter creates a new character from params.
func NewCharacter(character comic.Character, appearances []comic.AppearancesByYears) Character {
	c := Character{
		Character:   character,
		Appearances: appearances,
	}

	// TODO: stupid hack. figure out why MarshalJSON() override isn't being called in Echo context. Bug in Echo???
	if c.VendorImage != "" {
		c.VendorImage = cdnURL + "/" + c.VendorImage
	}
	if c.Image != "" {
		c.Image = cdnURL + "/" + c.Image
	}
	return c
}
