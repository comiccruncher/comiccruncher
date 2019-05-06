package search_test

import (
	"fmt"
	"github.com/comiccruncher/comiccruncher/comic"
	"github.com/comiccruncher/comiccruncher/internal/pgo"
	"github.com/comiccruncher/comiccruncher/search"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var db = pgo.MustInstanceTest()

func TestMain(m *testing.M) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recovered from panic and cleaning up test data")
			tearDownTestData()
			panic(r)
		}
	}()
	tearDownTestData()
	setUpTestData()
	code := m.Run()
	tearDownTestData()
	os.Exit(code)
}

func setUpTestData() {
	p := &comic.Publisher{Name: "Marvel", Slug: "marvel"}
	_, err := db.Model(p).SelectOrInsert()
	if err != nil {
		panic(err)
	}
	_, err = db.Model(&comic.Character{
		Name:        "Emma Frost",
		Slug:        "emma-frost",
		PublisherID: p.ID,
	}).Insert()
	if err != nil {
		panic(err)
	}
}

func tearDownTestData() {
	db.Exec("DELETE FROM characters")
	db.Exec("DELETE FROM publishers")
}

func TestFindAllCharactersByName(t *testing.T) {
	var searchService = search.NewSearchService(db)

	characters, err := searchService.Characters("emm frost", 5, 0)
	assert.Nil(t, err)
	assert.Len(t, characters, 1)
	assert.Equal(t, "Marvel", characters[0].Publisher.Name)
	characters, err = searchService.Characters("emsfdsfsfs'", 5, 0)
	assert.Nil(t, err)
	assert.Len(t, characters, 0)
}
