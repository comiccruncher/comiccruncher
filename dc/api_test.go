package dc_test

import (
	"github.com/comiccruncher/comiccruncher/dc"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestAPI_Characters(t *testing.T) {
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
	dcAPI := dc.NewDcAPI(ts.Client())
	dcAPI.CharacterEndpoint = ts.URL

	result, err := dcAPI.Characters(1)
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Results, 25)
	assert.Equal(t, result.TotalResults, 153)
}
