package marvel_test

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"github.com/aimeelaplant/comiccruncher/marvel"
)

func TestAPI_Characters(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		file, err := os.Open("./testdata/characters_result.json")
		if err != nil {
			panic(err)
		}
		bytes, err := ioutil.ReadAll(file)
		if err != nil {
			panic(err)
		}
		w.Write(bytes)
	}))

	marvelApi := marvel.NewMarvelAPI(ts.Client())
	marvelApi.CharacterEndpoint = ts.URL

	result, apiError, err := marvelApi.Characters(&marvel.Criteria{})
	assert.Nil(t, err)
	assert.Nil(t, apiError)
	assert.Len(t, result.Results, 50)
	assert.Equal(t, result.Results[0].Comics.Available, 12)
}

func TestAPI_FetchCharactersFails(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		file, err := os.Open("./testdata/error_result.json")
		if err != nil {
			panic(err)
		}
		bytes, err := ioutil.ReadAll(file)
		if err != nil {
			panic(err)
		}
		w.Write(bytes)
	}))

	marvelApi := marvel.NewMarvelAPI(ts.Client())
	marvelApi.CharacterEndpoint = ts.URL

	result, apiError, err := marvelApi.Characters(&marvel.Criteria{})
	assert.Nil(t, apiError)
	assert.Nil(t, err)
	assert.Equal(t, "You may not request more than 100 items.", result.Status)
	assert.Equal(t, 409, result.Code)
}
