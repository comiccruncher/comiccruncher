package comic_test

import (
	"fmt"
	"github.com/aimeelaplant/comiccruncher/comic"
	"github.com/aimeelaplant/comiccruncher/internal/mocks/comic"
	"github.com/go-redis/redis"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

// db test to test query.
func TestCharacterServiceMustNormalizeSources(t *testing.T) {
	svc := comic.NewCharacterService(testContainer)
	c, err := svc.Character("emma-frost")
	assert.Nil(t, err)
	svc.MustNormalizeSources(c)
}

func TestExpandedServiceCharacter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ch := &comic.Character{
		ID: 1,
		Publisher: comic.Publisher{
			Slug: "marvel",
			Name: "Marvel",
		},
		PublisherID: 1,
		Slug:        "emma-frost",
	}

	cr := mock_comic.NewMockCharacterRepository(ctrl)
	cr.EXPECT().FindBySlug(gomock.Any(), false).Times(1).Return(ch, nil)
	ar := mock_comic.NewMockAppearancesByYearsRepository(ctrl)
	ar.EXPECT().List(gomock.Any()).Return([]comic.AppearancesByYears{
		{
			Category: comic.Main,
			Aggregates: []comic.YearlyAggregate{
				{Count: 10, Year: 1979},
			},
		},
		{
			Category: comic.Main | comic.Alternate,
			Aggregates: []comic.YearlyAggregate{
				{Count: 10, Year: 1979},
			},
		},
	}, nil)
	rc := mock_comic.NewMockRedisClient(ctrl)
	val := make(map[string]string)
	val["all_time_issue_count"] = "100"
	val["all_time_issue_count_rank"] = "1"
	val["all_time_average_per_year"] = "20.23"
	val["all_time_average_per_year_rank"] = "2"
	val["main_issue_count"] = "300"
	val["main_issue_count_rank"] = "4"
	val["main_average_per_year_rank"] = "5"
	val["main_average_per_year"] = "60.23"
	cmd := redis.NewStringStringMapResult(val, nil)
	rc.EXPECT().HGetAll(fmt.Sprintf("%s:stats", ch.Slug)).Times(1).Return(cmd)

	tm := time.Now()
	sl := []*comic.LastSync{
		{
			CharacterID: 1,
			SyncedAt:    tm,
			NumIssues:   10,
		},
	}
	slr := mock_comic.NewMockCharacterSyncLogRepository(ctrl)
	slr.EXPECT().LastSyncs(gomock.Any()).Times(1).Return(sl, nil)

	svc := comic.NewExpandedService(cr, ar, rc, slr)
	ec, err := svc.Character(comic.CharacterSlug("emma-frost"))
	at := ec.Stats[0]
	m := ec.Stats[1]
	assert.Nil(t, err)
	assert.Len(t, ec.Appearances, 2)
	assert.Len(t, ec.Stats, 2)
	assert.Equal(t, at.Category, comic.AllTimeStats)
	assert.Equal(t, at.IssueCount, uint(100))
	assert.Equal(t, at.IssueCountRank, uint(1))
	assert.Equal(t, at.Average, float64(20.23))
	assert.Equal(t, at.AverageRank, uint(2))
	assert.Equal(t, m.Category, comic.MainStats)
	assert.Equal(t, m.IssueCount, uint(300))
	assert.Equal(t, m.IssueCountRank, uint(4))
	assert.Equal(t, m.Average, float64(60.23))
	assert.Equal(t, m.AverageRank, uint(5))
	assert.NotNil(t, ec.LastSyncs)
	assert.Len(t, ec.LastSyncs, 1)
	assert.Equal(t, ec.LastSyncs[0].NumIssues, 10)
	assert.Equal(t, ec.LastSyncs[0].SyncedAt, tm)
	assert.Equal(t, ec.LastSyncs[0].CharacterID, comic.CharacterID(1))
}

func TestExpandedServiceCharacterNoResult(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cr := mock_comic.NewMockCharacterRepository(ctrl)
	cr.EXPECT().FindBySlug(gomock.Any(), false).Times(1).Return(nil, nil)
	ar := mock_comic.NewMockAppearancesByYearsRepository(ctrl)
	ar.EXPECT().List(gomock.Any()).Times(0)
	rc := mock_comic.NewMockRedisClient(ctrl)
	val := make(map[string]string, 0)
	cmd := redis.NewStringStringMapResult(val, nil)
	rc.EXPECT().HGetAll(gomock.Any()).Times(0).Return(cmd)
	slr := mock_comic.NewMockCharacterSyncLogRepository(ctrl)
	slr.EXPECT().LastSyncs(gomock.Any()).Times(0)

	svc := comic.NewExpandedService(cr, ar, rc, slr)
	ec, err := svc.Character(comic.CharacterSlug("emma-frost"))
	assert.Nil(t, err)
	assert.Nil(t, ec)
}

func TestExpandedServiceCharacterNoRedisResult(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ch := &comic.Character{
		ID: 1,
		Publisher: comic.Publisher{
			Slug: "marvel",
			Name: "Marvel",
		},
		PublisherID: 1,
		Slug:        "emma-frost",
	}

	cr := mock_comic.NewMockCharacterRepository(ctrl)
	cr.EXPECT().FindBySlug(gomock.Any(), false).Times(1).Return(ch, nil)
	ar := mock_comic.NewMockAppearancesByYearsRepository(ctrl)
	ar.EXPECT().List(gomock.Any()).Return([]comic.AppearancesByYears{}, nil)
	rc := mock_comic.NewMockRedisClient(ctrl)
	val := make(map[string]string, 0)
	cmd := redis.NewStringStringMapResult(val, nil)
	rc.EXPECT().HGetAll(fmt.Sprintf("%s:stats", ch.Slug)).Times(1).Return(cmd)
	slr := mock_comic.NewMockCharacterSyncLogRepository(ctrl)
	slr.EXPECT().LastSyncs(gomock.Any()).Times(1).Return([]*comic.LastSync{}, nil)

	svc := comic.NewExpandedService(cr, ar, rc, slr)
	ec, err := svc.Character(comic.CharacterSlug("emma-frost"))
	assert.Nil(t, err)
	assert.Len(t, ec.Appearances, 0)
	assert.Len(t, ec.Stats, 0)
}
