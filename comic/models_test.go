package comic_test

import (
	"github.com/aimeelaplant/comiccruncher/comic"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewCharacterSlugs(t *testing.T) {
	slugs := []string{"a", "b", "c"}
	assert.Len(t, comic.NewCharacterSlugs(slugs...), 3)

	slugs1 := []string{}
	assert.Len(t, comic.NewCharacterSlugs(slugs1...), 0)
}

func TestAppearanceType_HasAll(t *testing.T) {
	c := comic.AppearanceType(comic.Main)
	assert.True(t, c.HasAll(comic.Main))
	assert.False(t, c.HasAll(comic.Alternate))
	assert.False(t, c.HasAll(comic.Main|comic.Alternate))

	c = comic.AppearanceType(comic.Main | comic.Alternate)
	assert.True(t, c.HasAll(comic.Main|comic.Alternate))
	assert.True(t, c.HasAll(comic.Alternate))
	assert.True(t, c.HasAll(comic.Main))
}

func TestAppearanceType_HasAny(t *testing.T) {
	c := comic.AppearanceType(comic.Main)
	assert.True(t, c.HasAny(comic.Main|comic.Alternate))
	assert.True(t, c.HasAny(comic.Main))
	assert.False(t, c.HasAny(comic.Alternate))

	c = comic.AppearanceType(comic.Main | comic.Alternate)
	assert.True(t, c.HasAny(comic.Main|comic.Alternate))
	assert.True(t, c.HasAny(comic.Main))
	assert.True(t, c.HasAny(comic.Alternate))
}

func TestAppearanceType_MarshalJSON(t *testing.T) {
	main := comic.AppearanceType(comic.Main)
	bytes, err := main.MarshalJSON()
	assert.Nil(t, err)
	assert.Equal(t, "\"main\"", string(bytes))

	alternate := comic.AppearanceType(comic.Alternate)
	bytes, err = alternate.MarshalJSON()
	assert.Nil(t, err)
	assert.Equal(t, "\"alternate\"", string(bytes))

	all := comic.AppearanceType(comic.Main | comic.Alternate)
	bytes, err = all.MarshalJSON()
	assert.Nil(t, err)
	assert.Equal(t, "\"all\"", string(bytes))
}
