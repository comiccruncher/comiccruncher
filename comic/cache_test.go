package comic_test

import (
	"errors"
	"github.com/aimeelaplant/comiccruncher/comic"
	"github.com/aimeelaplant/comiccruncher/internal/mocks/comic"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAppearancesSyncerSync(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	aggs := []comic.YearlyAggregate{
		{Year: 2017, Count: 11},
		{Year: 2018, Count: 10},
	}
	mains := comic.AppearancesByYears{CharacterSlug: "test", Category: comic.Main, Aggregates: aggs}
	alts := comic.AppearancesByYears{CharacterSlug: "test", Category: comic.Alternate, Aggregates: aggs}
	both := comic.AppearancesByYears{CharacterSlug: "test", Category: comic.Main | comic.Alternate, Aggregates: aggs}

	r := mock_comic.NewMockAppearancesByYearsRepository(ctrl)
	r.EXPECT().Main(gomock.Any()).Return(mains, nil)
	r.EXPECT().Alternate(gomock.Any()).Return(alts, nil)
	r.EXPECT().Both(gomock.Any()).Return(both, nil)
	w := mock_comic.NewMockAppearancesByYearsWriter(ctrl)
	w.EXPECT().Set(mains).Return(nil)
	w.EXPECT().Set(alts).Return(nil)

	s := comic.NewAppearancesSyncerRW(r, w)
	total, err := s.Sync(comic.CharacterSlug("test"))
	assert.Nil(t, err)
	assert.Equal(t, 21, total)
}

func TestAppearancesSyncerSyncReaderError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	aggs := []comic.YearlyAggregate{
		{Year: 2017, Count: 11},
		{Year: 2018, Count: 10},
	}
	mains := comic.AppearancesByYears{CharacterSlug: "test", Category: comic.Main, Aggregates: aggs}
	r := mock_comic.NewMockAppearancesByYearsRepository(ctrl)
	r.EXPECT().Main(gomock.Any()).Return(mains, errors.New("bad error"))

	w := mock_comic.NewMockAppearancesByYearsWriter(ctrl)
	s := comic.NewAppearancesSyncerRW(r, w)
	_, err := s.Sync(comic.CharacterSlug("test"))
	assert.Error(t, err)
}

func TestAppearancesSyncerSyncWriterError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	aggs := []comic.YearlyAggregate{
		{Year: 2017, Count: 11},
		{Year: 2018, Count: 10},
	}
	mains := comic.AppearancesByYears{CharacterSlug: "test", Category: comic.Main, Aggregates: aggs}

	r := mock_comic.NewMockAppearancesByYearsRepository(ctrl)
	r.EXPECT().Main(gomock.Any()).Return(mains, nil)
	w := mock_comic.NewMockAppearancesByYearsWriter(ctrl)
	w.EXPECT().Set(mains).Return(errors.New("some error"))

	s := comic.NewAppearancesSyncerRW(r, w)
	_, err := s.Sync(comic.CharacterSlug("test"))
	assert.Error(t, err)
}
