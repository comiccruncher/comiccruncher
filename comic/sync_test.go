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
		{Year: 2017, Main: 11, Alternate: 0},
		{Year: 2018, Main: 0, Alternate: 10},
	}
	mains := comic.AppearancesByYears{CharacterSlug: "test", Aggregates: aggs}
	r := mock_comic.NewMockAppearancesByYearsRepository(ctrl)
	r.EXPECT().List(gomock.Any()).Return(mains, nil)

	w := mock_comic.NewMockAppearancesByYearsWriter(ctrl)
	w.EXPECT().Set(mains).Return(nil)

	s := comic.NewAppearancesSyncerRW(r, w)
	total, err := s.Sync(comic.CharacterSlug("test"))
	assert.Nil(t, err)
	assert.Equal(t, 21, total)
}

func TestAppearancesSyncerSyncReaderError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	aggs := []comic.YearlyAggregate{
		{Year: 2017, Main: 11, Alternate: 10},
		{Year: 2018, Main: 10, Alternate: 0},
	}
	mains := comic.AppearancesByYears{CharacterSlug: "test", Aggregates: aggs}
	r := mock_comic.NewMockAppearancesByYearsRepository(ctrl)
	r.EXPECT().List(gomock.Any()).Return(mains, errors.New("bad error"))

	w := mock_comic.NewMockAppearancesByYearsWriter(ctrl)
	s := comic.NewAppearancesSyncerRW(r, w)
	_, err := s.Sync(comic.CharacterSlug("test"))
	assert.Error(t, err)
}

func TestAppearancesSyncerSyncWriterError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	aggs := []comic.YearlyAggregate{
		{Year: 2017, Main: 11, Alternate: 10},
		{Year: 2018, Main: 10, Alternate: 20},
	}
	mains := comic.AppearancesByYears{CharacterSlug: "test", Aggregates: aggs}

	r := mock_comic.NewMockAppearancesByYearsRepository(ctrl)
	r.EXPECT().List(gomock.Any()).Return(mains, nil)
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

func TestRedisCharacterStatsSyncerSyncAll(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	rds := mock_comic.NewMockRedisClient(ctrl)
	rds.EXPECT().HMSet(gomock.Any(), gomock.Any()).Return(&redis.StatusCmd{})
	rds.EXPECT().HMSet(gomock.Any(), gomock.Any()).Return(&redis.StatusCmd{})
	cr := mock_comic.NewMockCharacterRepository(ctrl)
	cr.EXPECT().FindBySlug(gomock.Any(), false).Return(&comic.Character{
		Slug: "a",
		Publisher: comic.Publisher{
			Slug: "marvel",
		},
	}, nil)
	cr.EXPECT().FindBySlug(gomock.Any(), false).Return(&comic.Character{
		Slug: "b",
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
	pr.EXPECT().FindOneByMarvel(gomock.Any()).Return(&comic.RankedCharacter{
		Stats: comic.CharacterStats{
			IssueCountRank: 10,
			IssueCount:     1,
		},
	}, nil)
	pr.EXPECT().FindOneByAll(gomock.Any()).Return(&comic.RankedCharacter{
		Stats: comic.CharacterStats{
			IssueCountRank: 10,
			IssueCount:     20,
		},
	}, nil)

	syncer := comic.NewCharacterStatsSyncer(rds, cr, pr)

	c := []*comic.Character{
		{Slug: "a"},
		{Slug: "b"},
	}
	slugs := []string{"a", "b"}
	res := syncer.SyncAll(c)
	for i := 0; i < len(c); i++ {
		result := <-res
		assert.Nil(t, result.Error)
		assert.Contains(t, slugs, c[i].Slug.Value())
	}
}

func TestRedisCharacterStatsSyncerSyncAllError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	rds := mock_comic.NewMockRedisClient(ctrl)
	rds.EXPECT().HMSet(gomock.Any(), gomock.Any()).Times(0)
	cr := mock_comic.NewMockCharacterRepository(ctrl)
	cr.EXPECT().FindBySlug(gomock.Any(), false).Return(&comic.Character{
		Slug: "a",
		Publisher: comic.Publisher{
			Slug: "marvel",
		},
	}, nil)
	cr.EXPECT().FindBySlug(gomock.Any(), false).Return(&comic.Character{
		Slug: "b",
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
	}, errors.New("some error"))
	pr.EXPECT().FindOneByMarvel(gomock.Any()).Return(&comic.RankedCharacter{
		Stats: comic.CharacterStats{
			IssueCountRank: 10,
			IssueCount:     1,
		},
	}, nil)
	pr.EXPECT().FindOneByAll(gomock.Any()).Return(&comic.RankedCharacter{
		Stats: comic.CharacterStats{
			IssueCountRank: 10,
			IssueCount:     20,
		},
	}, errors.New("some error"))

	syncer := comic.NewCharacterStatsSyncer(rds, cr, pr)

	c := []*comic.Character{
		{Slug: "a"},
		{Slug: "b"},
	}
	slugs := []string{"a", "b"}
	res := syncer.SyncAll(c)
	for i := 0; i < len(c); i++ {
		result := <-res
		assert.Error(t, result.Error)
		assert.Contains(t, slugs, c[i].Slug.Value())
	}
}
