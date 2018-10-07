package dc

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestApi_FetchCharacters(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		file, err := os.Open("./testdata/api.json")
		if err != nil {
			panic(err)
		}
		bytes, err := ioutil.ReadAll(file)
		if err != nil {
			panic(err)
		}
		w.Write(bytes)
	}))

	dcApi := Api{
		httpClient:        ts.Client(),
		CharacterEndpoint: ts.URL,
	}

	result, err := dcApi.FetchCharacters(1)
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Results, 25)
	assert.Equal(t, result.TotalResults, 153)
}
