package comic_test

import (
	"fmt"
	"github.com/aimeelaplant/comiccruncher/comic"
	"github.com/aimeelaplant/comiccruncher/imaging"
	"github.com/aimeelaplant/comiccruncher/internal/mocks/comic"
	"github.com/aimeelaplant/comiccruncher/internal/mocks/imaging"
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
	slug := comic.CharacterSlug("emma-frost")
	ctr := mock_comic.NewMockCharacterThumbRepository(ctrl)
	ctr.EXPECT().Thumbnails(gomock.Any()).Return(&comic.CharacterThumbnails{
		Slug: slug,
		Image: &comic.ThumbnailSizes{
			Small: "a",
			Medium: "b",
			Large: "c",
		},
		VendorImage: &comic.ThumbnailSizes{
			Small: "d",
			Medium: "e",
			Large: "f",
		},
	}, nil)
	svc := comic.NewExpandedService(cr, ar, rc, slr, ctr)
	ec, err := svc.Character(slug)
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
	assert.NotNil(t, ec.Thumbnails)
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
	ctr := mock_comic.NewMockCharacterThumbRepository(ctrl)
	svc := comic.NewExpandedService(cr, ar, rc, slr, ctr)
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
	ctr := mock_comic.NewMockCharacterThumbRepository(ctrl)
	ctr.EXPECT().Thumbnails(gomock.Any()).Return(nil, nil)

	svc := comic.NewExpandedService(cr, ar, rc, slr, ctr)
	ec, err := svc.Character(comic.CharacterSlug("emma-frost"))
	assert.Nil(t, err)
	assert.Len(t, ec.Appearances, 0)
	assert.Len(t, ec.Stats, 0)
}

func TestRankedServiceDCTrending(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r := mock_comic.NewMockPopularRepository(ctrl)
	r.EXPECT().DCTrending(25, 0).Times(1).Return([]*comic.RankedCharacter{}, nil)
	svc := comic.NewRankedService(r)
	results, err := svc.DCTrending(25, 0)
	assert.Nil(t, err)
	assert.Len(t, results, 0)
}

func TestRankedServiceMarvelTrending(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r := mock_comic.NewMockPopularRepository(ctrl)
	r.EXPECT().MarvelTrending(25, 0).Times(1).Return([]*comic.RankedCharacter{}, nil)
	svc := comic.NewRankedService(r)
	results, err := svc.MarvelTrending(25, 0)
	assert.Nil(t, err)
	assert.Len(t, results, 0)
}

func TestCharacterThumbServiceUpload(t *testing.T) {
	c := &comic.Character{
		VendorImage: "myvendorimg.jpg",
		Image: "myimage.jpg",
		Slug: "test",
	}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	size100 := imaging.NewThumbnailSize(100, 100)
	size300 := imaging.NewThumbnailSize(300, 300)
	size600 := imaging.NewThumbnailSize(600, 0)
	r := mock_comic.NewMockRedisClient(ctrl)
	r.EXPECT().Set(gomock.Any(), gomock.Any(), time.Duration(0)).Return(redis.NewStatusCmd()).Times(1)
	th := mock_imaging.NewMockThumbnailUploader(ctrl)
	th.EXPECT().Generate(c.VendorImage, gomock.Any()).Return([]*imaging.ThumbnailResult{
		{Pathname: "a.jpg", Dimensions: size100},
		{Pathname: "b.jpg", Dimensions: size300},
		{Pathname: "c.jpg", Dimensions: size600},
	}, nil)
	th.EXPECT().Generate(c.Image, gomock.Any()).Return([]*imaging.ThumbnailResult{
		{Pathname: "d.jpg", Dimensions: size100},
		{Pathname: "e.jpg", Dimensions: size300},
		{Pathname: "f.jpg", Dimensions: size600},
	}, nil)
	svc := comic.NewCharacterThumbnailService(r, th)
	thmbs, err := svc.Upload(c)
	assert.Nil(t, err)
	assert.NotNil(t, thmbs)
	vndrImg := thmbs.VendorImage
	img := thmbs.Image
	assert.NotNil(t, vndrImg)
	assert.NotNil(t, img)
	assert.Equal(t, "a.jpg", vndrImg.Small)
	assert.NotEmpty(t, "b.jpg", vndrImg.Medium)
	assert.NotEmpty(t, "c.jpg", vndrImg.Large)
	assert.NotEmpty(t, "d.jpg", img.Small)
	assert.NotEmpty(t, "e.jpg", img.Medium)
	assert.NotEmpty(t, "f.jpg", img.Large)
}
