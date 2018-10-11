package web

import (
	"encoding/json"
	"github.com/aimeelaplant/comiccruncher/comic"
	"github.com/stretchr/testify/assert"
	"testing"
)

func assertMapHasKey(t *testing.T, key string, m map[string]*json.RawMessage) {
	_, ok := m[key]
	assert.True(t, ok)
}

func TestCharacter_MarshalJSON(t *testing.T) {
	character := Character{
		Character: comic.Character{
			Name: "Aimee",
			Publisher: comic.Publisher{
				Name: "Aimee",
				Slug: "aimee",
			},
			Image:       "images/image1.jpg",
			VendorImage: "images/image.jpg",
			Slug:        "aimee",
		},
		Appearances: []comic.AppearancesByYears{
			{CharacterSlug: "aimee",
				Category: comic.Main,
				Aggregates: []comic.YearlyAggregate{
					{Year: 2017, Count: 100},
					{Year: 2018, Count: 100},
				},
			},
		},
	}
	j, err := json.MarshalIndent(&character, "", "  ")
	assert.Nil(t, err)

	var objMap map[string]*json.RawMessage
	err = json.Unmarshal(j, &objMap)

	assert.Nil(t, err)
	assertMapHasKey(t, "name", objMap)
	assertMapHasKey(t, "vendor_url", objMap)
	assertMapHasKey(t, "vendor_image", objMap)
	assert.NotEqual(t, "images/image.jpg", objMap["vendor_image"])
	assertMapHasKey(t, "vendor_description", objMap)
	assertMapHasKey(t, "image", objMap)
	assert.NotEqual(t, "images/image1.jpg", objMap["image"])
	assertMapHasKey(t, "publisher", objMap)
	assertMapHasKey(t, "description", objMap)
	assertMapHasKey(t, "slug", objMap)
	// TODO: assertMapHasKey(t, "appearances", objMap)
}
