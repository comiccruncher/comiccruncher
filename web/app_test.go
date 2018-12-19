package web_test

import (
	"github.com/aimeelaplant/comiccruncher/internal/mocks/comic"
	"github.com/aimeelaplant/comiccruncher/internal/mocks/search"
	"github.com/aimeelaplant/comiccruncher/web"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewApp(t *testing.T) {
	ctrl := gomock.NewController(t)
	es := mock_comic.NewMockExpandedServicer(ctrl)
	srchr := mock_search.NewMockSearcher(ctrl)
	sr := mock_comic.NewMockStatsRepository(ctrl)
	rs := mock_comic.NewMockRankedServicer(ctrl)
	ctr := mock_comic.NewMockCharacterThumbRepository(ctrl)
	a := web.NewApp(es, srchr, sr, rs, ctr)
	assert.NotNil(t, a)
}

func TestAppRun(t *testing.T) {
	ctrl := gomock.NewController(t)
	es := mock_comic.NewMockExpandedServicer(ctrl)
	srchr := mock_search.NewMockSearcher(ctrl)
	sr := mock_comic.NewMockStatsRepository(ctrl)
	rs := mock_comic.NewMockRankedServicer(ctrl)
	ctr := mock_comic.NewMockCharacterThumbRepository(ctrl)
	a := web.NewApp(es, srchr, sr, rs, ctr)
	go func() {
		err := a.Run("0")
		assert.Nil(t, err)
	}()
	time.Sleep(200 * time.Millisecond)
}

func TestAppClose(t *testing.T) {
	ctrl := gomock.NewController(t)
	es := mock_comic.NewMockExpandedServicer(ctrl)
	srchr := mock_search.NewMockSearcher(ctrl)
	sr := mock_comic.NewMockStatsRepository(ctrl)
	rs := mock_comic.NewMockRankedServicer(ctrl)
	ctr := mock_comic.NewMockCharacterThumbRepository(ctrl)
	a := web.NewApp(es, srchr, sr, rs, ctr)
	assert.Nil(t, a.Close())
}
