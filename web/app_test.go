package web_test

import (
	"context"
	"github.com/comiccruncher/comiccruncher/internal/mocks/comic"
	"github.com/comiccruncher/comiccruncher/internal/mocks/search"
	"github.com/comiccruncher/comiccruncher/web"
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

func TestAppShutdown(t *testing.T) {
	ctrl := gomock.NewController(t)
	es := mock_comic.NewMockExpandedServicer(ctrl)
	srchr := mock_search.NewMockSearcher(ctrl)
	sr := mock_comic.NewMockStatsRepository(ctrl)
	rs := mock_comic.NewMockRankedServicer(ctrl)
	ctr := mock_comic.NewMockCharacterThumbRepository(ctrl)
	a := web.NewApp(es, srchr, sr, rs, ctr)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	assert.Nil(t, a.Shutdown(ctx))
}
