package comic_test

import (
	"errors"
	"github.com/aimeelaplant/comiccruncher/comic"
	"github.com/aimeelaplant/comiccruncher/internal/mocks/comic"
	"github.com/go-redis/redis"
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

func TestRedisCharacterStatsSyncerSyncMarvel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	rds := mock_comic.NewMockRedisClient(ctrl)
	rds.EXPECT().HMSet(gomock.Any(), gomock.Any()).Return(&redis.StatusCmd{})
	cr := mock_comic.NewMockCharacterRepository(ctrl)
	cr.EXPECT().FindBySlug(gomock.Any(), false).Return(&comic.Character{
		Slug: "emma-frost",
		Publisher: comic.Publisher{
			Slug: "marvel",
		},
	}, nil)
	pr := mock_comic.NewMockPopularRepository(ctrl)
	pr.EXPECT().FindOneByDC(gomock.Any()).Times(0)
	pr.EXPECT().FindOneByMarvel(gomock.Any()).Return(&comic.RankedCharacter{
		Stats: comic.CharacterStats{
			IssueCountRank: 1,
			IssueCount:     100,
		},
	}, nil)
	pr.EXPECT().FindOneByAll(gomock.Any()).Return(&comic.RankedCharacter{
		Stats: comic.CharacterStats{
			IssueCountRank: 2,
			IssueCount:     200,
		},
	}, nil)

	syncer := comic.NewCharacterStatsSyncer(rds, cr, pr)
	err := syncer.Sync(comic.CharacterSlug("emma-frost"))
	assert.Nil(t, err)
}

func TestRedisCharacterStatsSyncerSyncDC(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	rds := mock_comic.NewMockRedisClient(ctrl)
	rds.EXPECT().HMSet(gomock.Any(), gomock.Any()).Return(&redis.StatusCmd{})
	cr := mock_comic.NewMockCharacterRepository(ctrl)
	cr.EXPECT().FindBySlug(gomock.Any(), false).Return(&comic.Character{
		Slug: "emma-frost",
		Publisher: comic.Publisher{
			Slug: "dc",
		},
	}, nil)
	pr := mock_comic.NewMockPopularRepository(ctrl)
	pr.EXPECT().FindOneByMarvel(gomock.Any()).Times(0)
	pr.EXPECT().FindOneByDC(gomock.Any()).Return(&comic.RankedCharacter{
		Stats: comic.CharacterStats{
			IssueCountRank: 1,
			IssueCount:     100,
		},
	}, nil)
	pr.EXPECT().FindOneByAll(gomock.Any()).Return(&comic.RankedCharacter{
		Stats: comic.CharacterStats{
			IssueCountRank: 2,
			IssueCount:     200,
		},
	}, nil)

	syncer := comic.NewCharacterStatsSyncer(rds, cr, pr)
	err := syncer.Sync(comic.CharacterSlug("emma-frost"))
	assert.Nil(t, err)
}

func TestRedisCharacterStatsSyncerNoSyncNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	rds := mock_comic.NewMockRedisClient(ctrl)
	rds.EXPECT().HMSet(nil, nil).Times(0)
	pr := mock_comic.NewMockPopularRepository(ctrl)
	cr := mock_comic.NewMockCharacterRepository(ctrl)
	cr.EXPECT().FindBySlug(gomock.Any(), false).Return(nil, nil)
	syncer := comic.NewCharacterStatsSyncer(rds, cr, pr)
	err := syncer.Sync(comic.CharacterSlug("emma-frost"))
	assert.Error(t, err)
}
