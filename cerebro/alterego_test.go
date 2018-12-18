package cerebro_test

import (
	"github.com/aimeelaplant/comiccruncher/cerebro"
	"github.com/aimeelaplant/comiccruncher/comic"
	"github.com/aimeelaplant/comiccruncher/internal/mocks/cerebro"
	"github.com/aimeelaplant/comiccruncher/internal/mocks/comic"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestAlterEgoIdentifier_Name_For_DC(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		file, err := os.Open("./testdata/aquaman.html")
		if err != nil {
			panic(err)
		}
		bytes, err := ioutil.ReadAll(file)
		if err != nil {
			panic(err)
		}
		w.Write(bytes)
	}))
	defer ts.Close()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSvc := mock_comic.NewMockCharacterServicer(ctrl)

	mockRepo := mock_comic.NewMockCharacterSourceRepository(ctrl)
	mockRepo.EXPECT().FindAll(gomock.Any()).Times(0)

	identifier := cerebro.NewAlterEgoIdentifier(ts.Client(), mockSvc)
	character := comic.Character{
		Publisher: comic.Publisher{
			Slug: "dc",
		},
		VendorURL: ts.URL,
	}

	realName, err := identifier.Name(character)

	assert.Nil(t, err)
	assert.Equal(t, "Arthur Curry", realName)
}

func TestNewAlterEgoIdentifier(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	h := mock_cerebro.NewMockHTTPClient(ctrl)
	svc := mock_comic.NewMockCharacterServicer(ctrl)
	i := cerebro.NewAlterEgoIdentifier(h, svc)
	assert.NotNil(t, i)
}

func TestNewAlterEgoImporter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	h := mock_cerebro.NewMockHTTPClient(ctrl)
	svc := mock_comic.NewMockCharacterServicer(ctrl)
	i := cerebro.NewAlterEgoImporter(h, svc)
	assert.NotNil(t, i)
}
