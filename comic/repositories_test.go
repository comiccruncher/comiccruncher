package comic_test

import (
	"errors"
	"fmt"
	"github.com/aimeelaplant/comiccruncher/comic"
	"github.com/aimeelaplant/comiccruncher/internal/mocks/comic"
	"github.com/go-pg/pg/orm"
	"github.com/go-redis/redis"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func must(_ orm.Result, err error) {
	if err != nil {
		panic(err)
	}
}

func TestMain(m *testing.M) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recovered from panic and cleaning up test data")
			tearDownData()
			panic(r)
		}
	}()
	tearDownData()
	setupTestData()
	code := m.Run()
	tearDownData()
	os.Exit(code)
}

func tearDownData() {
	db := testInstance
	must(db.Exec("DELETE FROM character_sync_logs"))
	must(db.Exec("DELETE FROM character_sources"))
	must(db.Exec("DELETE FROM character_issues"))
	must(db.Exec("DELETE FROM issues"))
	must(db.Exec("DELETE FROM characters"))
	must(db.Exec("DELETE FROM publishers"))
}

func setupTestData() {
	db := testInstance
	publisher := &comic.Publisher{Name: "Marvel", Slug: "marvel"}
	must(db.Model(publisher).Insert())
	must(db.Model(&comic.Publisher{Name: "DC", Slug: "dc"}).Insert())
	character := &comic.Character{
		Name:        "Emma Frost",
		Slug:        "emma-frost",
		VendorID:    "1",
		PublisherID: publisher.ID,
	}
	must(db.Model(character).Insert())
	source := &comic.CharacterSource{
		VendorName:  "Emma Frost (Marvel Universe)",
		CharacterID: comic.CharacterID(character.ID),
		VendorURL:   "https://example.com",
		VendorType:  comic.VendorTypeCb,
	}
	must(db.Model(source).Insert())
	syncLog := &comic.CharacterSyncLog{
		CharacterID: character.ID,
		SyncType:    comic.YearlyAppearances,
		SyncStatus:  comic.InProgress,
	}
	must(db.Model(syncLog).Insert())
	character2 := &comic.Character{
		Name:        "Emma Frost",
		Slug:        "emma-frost-2",
		VendorID:    "2",
		PublisherID: publisher.ID,
	}
	must(db.Model(character2).Insert())
	date := time.Date(1979, time.November, 1, 0, 0, 0, 0, time.UTC)
	issue := &comic.Issue{
		PublicationDate:    date,
		SaleDate:           date,
		Format:             comic.FormatStandard,
		VendorPublisher:    "Marvel",
		VendorSeriesName:   "Uncanny X-Men",
		VendorSeriesNumber: "129",
		VendorID:           "123",
	}
	issue2 := &comic.Issue{
		PublicationDate:    date.AddDate(1, 0, 0),
		SaleDate:           date.AddDate(1, 0, 0),
		Format:             comic.FormatStandard,
		VendorPublisher:    "Marvel",
		VendorSeriesName:   "Uncanny X-Men",
		VendorSeriesNumber: "130",
		VendorID:           "124",
	}
	issue3 := &comic.Issue{
		PublicationDate:    issue2.PublicationDate,
		SaleDate:           issue2.PublicationDate,
		Format:             comic.FormatStandard,
		VendorPublisher:    "Marvel",
		VendorSeriesName:   "Uncanny X-Men",
		VendorSeriesNumber: "131",
		VendorID:           "125",
	}
	if err := db.Insert(issue, issue2, issue3); err != nil {
		panic(err)
	}
	if err := db.Insert(
		&comic.CharacterIssue{CharacterID: comic.CharacterID(character2.ID), IssueID: comic.IssueID(issue.ID), AppearanceType: comic.Main},
		&comic.CharacterIssue{CharacterID: comic.CharacterID(character2.ID), IssueID: comic.IssueID(issue2.ID), AppearanceType: comic.Main | comic.Alternate},
		&comic.CharacterIssue{CharacterID: comic.CharacterID(character2.ID), IssueID: comic.IssueID(issue3.ID), AppearanceType: comic.Alternate}); err != nil {
		panic(err)
	}
}

func TestPGPublisherRepositoryFindBySlug(t *testing.T) {
	r := comic.NewPGPublisherRepository(testInstance)
	marvel, err := r.FindBySlug("marvel")
	assert.Nil(t, err)
	assert.Equal(t, "marvel", string(marvel.Slug))
}

func TestNewPGPublisherRepositoryFindBySlugReturnsNil(t *testing.T) {
	r := comic.NewPGPublisherRepository(testInstance)
	bogus, err := r.FindBySlug("bogus")
	assert.Nil(t, err)
	assert.Nil(t, bogus)
}

func TestPGCharacterRepositoryFindBySlugReturnsNil(t *testing.T) {
	r := comic.NewPGCharacterRepository(testInstance)
	bogus, err := r.FindBySlug("bogus", true)
	assert.Nil(t, bogus)
	assert.Nil(t, err)
}

func TestCharacterRepositoryFindBySlug(t *testing.T) {
	r := comic.NewPGCharacterRepository(testInstance)
	c, err := r.FindBySlug("emma-frost", false)
	assert.Nil(t, err)
	assert.NotNil(t, c)
	assert.Equal(t, comic.CharacterSlug("emma-frost"), c.Slug)

	assert.True(t, c.ID > 0)
}

func TestPGCharacterRepositoryCreate(t *testing.T) {
	r := comic.NewPGPublisherRepository(testInstance)
	p, err := r.FindBySlug("marvel")
	assert.Nil(t, err)

	charRepo := comic.NewPGCharacterRepository(testInstance)
	c2 := &comic.Character{
		PublisherID:       p.ID,
		Name:              "  Emma Frost",
		VendorDescription: "asdsfsfd.",
		VendorID:          "3",
	}
	err = charRepo.Create(c2)
	assert.Nil(t, err)
	assert.Equal(t, "Emma Frost", c2.Name)
	assert.NotEqual(t, comic.CharacterSlug("emma-frost"), c2.Slug)
	assert.True(t, c2.ID > 0)
	assert.True(t, c2.PublisherID > 0)
	assert.True(t, c2.Publisher.ID > 0)
}

func TestPGCharacterRepositoryFindAll(t *testing.T) {
	r := comic.NewPGCharacterRepository(testInstance)
	c, err := r.FindAll(comic.CharacterCriteria{
		Slugs: []comic.CharacterSlug{comic.CharacterSlug("emma-frost"), comic.CharacterSlug("emma-frost-2")},
	})
	assert.Nil(t, err)
	assert.Len(t, c, 2)
}

func TestPGCharacterSyncLogRepositoryFindByIdReturnsNil(t *testing.T) {
	r := comic.NewPGCharacterSyncLogRepository(testInstance)
	sLog, err := r.FindByID(999)

	assert.Nil(t, err)
	assert.Nil(t, sLog)
}

func TestCharacterSyncLogRepositoryUpdate(t *testing.T) {
	syncLogRepo := comic.NewPGCharacterSyncLogRepository(testInstance)
	characterRepo := comic.NewPGCharacterRepository(testInstance)
	c, err := characterRepo.FindBySlug("emma-frost", true)
	assert.Nil(t, err)
	syncLogs, err := syncLogRepo.FindAllByCharacterID(c.ID)
	assert.Nil(t, err)
	assert.Len(t, syncLogs, 1)
	syncLog := syncLogs[0]
	status := syncLog.SyncStatus
	assert.Equal(t, string(comic.InProgress), string(status))
	syncLog.SyncStatus = comic.Success
	err = syncLogRepo.Update(syncLog)
	assert.Nil(t, err)
	assert.Equal(t, string(comic.Success), string(syncLog.SyncStatus))
}

func TestPGCharacterIssueRepositoryFindOneByReturnsNil(t *testing.T) {
	r := comic.NewPGCharacterIssueRepository(testInstance)
	res, err := r.FindOneBy(100, 100)
	assert.Nil(t, res)
	assert.Nil(t, err)
}

func TestCharacterIssueRepositoryCreateAndFindOneBy(t *testing.T) {
	r := comic.NewPGIssueRepository(testInstance)
	issue := &comic.Issue{
		PublicationDate:    time.Now(),
		SaleDate:           time.Now(),
		Format:             comic.FormatStandard,
		VendorPublisher:    "Marvel",
		VendorSeriesName:   "Uncanny X-Men",
		VendorSeriesNumber: "129",
		VendorID:           "321",
	}
	err := r.Create(issue)
	assert.Nil(t, err)
	assert.True(t, issue.ID > 0)

	cr := comic.NewPGCharacterRepository(testInstance)
	character, err := cr.FindBySlug("emma-frost", true)
	assert.Nil(t, err)
	ci := &comic.CharacterIssue{
		CharacterID:    character.ID,
		IssueID:        issue.ID,
		AppearanceType: comic.Alternate,
	}

	cir := comic.NewPGCharacterIssueRepository(testInstance)
	err = cir.Create(ci)
	assert.Nil(t, err)
	assert.Equal(t, ci.AppearanceType, comic.Alternate)
	// test the repo loads the ID.
	assert.True(t, ci.ID > 0)

	nci, err := cir.FindOneBy(character.ID, issue.ID)
	assert.Nil(t, err)
	assert.NotNil(t, nci)
	assert.Equal(t, comic.Alternate, nci.AppearanceType)
	assert.True(t, nci.ID > 0)
	assert.Equal(t, character.ID, nci.CharacterID)
	assert.Equal(t, issue.ID, nci.IssueID)
}

func TestPGIssueRepositoryFindByVendorIdReturnsNil(t *testing.T) {
	r := comic.NewPGIssueRepository(testInstance)
	result, err := r.FindByVendorID("98332")
	assert.Nil(t, result)
	assert.Nil(t, err)
}

func TestPGIssueRepositoryCreateAll(t *testing.T) {
	issue := &comic.Issue{
		PublicationDate:    time.Now(),
		SaleDate:           time.Now(),
		Format:             comic.FormatStandard,
		VendorPublisher:    "Marvel",
		VendorSeriesName:   "Uncanny X-Men",
		VendorSeriesNumber: "1000",
		VendorID:           "d7d6sd6",
	}
	issue2 := &comic.Issue{
		PublicationDate:    time.Now(),
		SaleDate:           time.Now(),
		Format:             comic.FormatStandard,
		VendorPublisher:    "Marvel",
		VendorSeriesName:   "Uncanny X-Men",
		VendorSeriesNumber: "10004",
		VendorID:           "d7d226sd6",
	}
	issues := []*comic.Issue{issue, issue2}
	r := comic.NewPGIssueRepository(testInstance)
	err := r.CreateAll(issues)
	assert.Nil(t, err)
	// test the ID gets loaded.
	assert.True(t, issue.ID > 0)
	assert.True(t, issue2.ID > 0)
	assert.Equal(t, issue.ID, issues[0].ID)
	assert.Equal(t, issue2.ID, issues[1].ID)
}

func TestPGCharacterIssueRepositoryCreateAll(t *testing.T) {
	cr := comic.NewPGCharacterRepository(testInstance)
	character, err := cr.FindBySlug("emma-frost", true)
	assert.Nil(t, err)

	issue := &comic.Issue{
		PublicationDate:    time.Now(),
		SaleDate:           time.Now(),
		Format:             comic.FormatStandard,
		VendorPublisher:    "Marvel",
		VendorSeriesName:   "X-Men",
		VendorSeriesNumber: "999",
		VendorID:           "93383",
	}
	issue2 := &comic.Issue{
		PublicationDate:    time.Now(),
		SaleDate:           time.Now(),
		Format:             comic.FormatStandard,
		VendorPublisher:    "Marvel",
		VendorSeriesName:   "X-Men",
		VendorSeriesNumber: "101",
		VendorID:           "3383",
	}
	issues := []*comic.Issue{issue, issue2}
	ir := comic.NewPGIssueRepository(testInstance)
	err = ir.CreateAll(issues)
	assert.Nil(t, err)
	assert.True(t, len(issues) > 1)
	assert.True(t, issues[0].ID > 0)
	assert.True(t, issues[1].ID > 0)

	characterIssues := []*comic.CharacterIssue{
		{CharacterID: character.ID, IssueID: issues[0].ID},
		{CharacterID: character.ID, IssueID: issues[1].ID},
	}
	cir := comic.NewPGCharacterIssueRepository(testInstance)
	err = cir.CreateAll(characterIssues)
	assert.Nil(t, err)
	// Test the ID is loaded.
	assert.True(t, characterIssues[0].ID > 0)
	assert.True(t, characterIssues[1].ID > 0)
}

// Test that the query returns both main and alternate appearances as slices.
func TestPGAppearanceRepositoryList(t *testing.T) {
	apy := comic.NewPGAppearancesPerYearRepository(testInstance)
	list, err := apy.List("emma-frost-2")

	assert.Nil(t, err)
	assert.Len(t, list.Aggregates, 41)

	// Test for blanks.
	bogus, err := apy.List("bogus")
	assert.Nil(t, err)
	assert.Len(t, bogus.Aggregates, 0)
}

func TestPGStatsRepository_Stats(t *testing.T) {
	s := comic.NewPGStatsRepository(testInstance)
	stats, err := s.Stats()
	assert.Nil(t, err)
	assert.NotNil(t, stats)

	assert.Equal(t, 8, stats.TotalIssues)
	assert.Equal(t, 6, stats.TotalAppearances)
	assert.Equal(t, 1979, stats.MinYear)
	assert.Equal(t, time.Now().Year(), stats.MaxYear)
}

func TestPGCharacterSyncLogRepositoryCreateAndFind(t *testing.T) {
	cr := comic.NewPGCharacterRepository(testInstance)
	characters, err := cr.FindAll(comic.CharacterCriteria{})
	assert.Nil(t, err)
	assert.True(t, len(characters) > 0)

	syncLog := comic.NewSyncLog(characters[0].ID, comic.Pending, comic.YearlyAppearances, nil)

	sl := comic.NewPGCharacterSyncLogRepository(testInstance)
	sl.Create(syncLog)
	// asserts it auto-increments
	assert.True(t, syncLog.ID > 0)

	id := syncLog.ID

	syncLog, err = sl.FindByID(syncLog.ID)
	assert.Nil(t, err)
	assert.Equal(t, id, syncLog.ID)
}

func TestPGCharacterRepositoryTotal(t *testing.T) {
	cr := comic.NewPGCharacterRepository(testInstance)
	total, err := cr.Total(comic.CharacterCriteria{
		IDs:           []comic.CharacterID{1},
		FilterSources: true,
	})
	assert.Nil(t, err)
	assert.True(t, total >= 0)
}

func TestPGPopularRepositoryRefreshAll(t *testing.T) {
	r := comic.NewPopularRefresher(testInstance)
	assert.Nil(t, r.RefreshAll())
}

func TestRedisCharacterThumbRepositoryThumbnails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	r := mock_comic.NewMockRedisClient(ctrl)
	val := "small:;medium:medium.jpg;large:large.jpg-small:small2.jpg;medium:medium2.jpg;large:large2.jpg"
	cmd := redis.NewStringResult(val, nil)
	r.EXPECT().Get("test:profile:thumbnails").Return(cmd)
	svc := comic.NewRedisCharacterThumbRepository(r)
	thmbs, err := svc.Thumbnails(comic.CharacterSlug("test"))
	assert.Nil(t, err)
	assert.NotNil(t, thmbs)
	img := thmbs.Image
	vendorImg := thmbs.VendorImage
	assert.NotNil(t, thmbs.Image)
	assert.NotNil(t, thmbs.VendorImage)
	assert.Equal(t, "small2.jpg", img.Small)
	assert.Equal(t, "medium2.jpg", img.Medium)
	assert.Equal(t, "large2.jpg", img.Large)
	assert.Equal(t, "", vendorImg.Small)
	assert.Equal(t, "medium.jpg", vendorImg.Medium)
	assert.Equal(t, "large.jpg", vendorImg.Large)
}

func TestRedisCharacterThumbRepositoryThumbnailsNil(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	r := mock_comic.NewMockRedisClient(ctrl)
	cmd := redis.NewStringResult("", redis.Nil)
	r.EXPECT().Get("test:profile:thumbnails").Return(cmd)
	svc := comic.NewRedisCharacterThumbRepository(r)
	thmbs, err := svc.Thumbnails(comic.CharacterSlug("test"))
	assert.Nil(t, err)
	assert.Nil(t, thmbs)
}

func TestRedisCharacterThumbRepositoryThumbnailsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	r := mock_comic.NewMockRedisClient(ctrl)
	val := "small:;medium:medium.jpg;large:large.jpg-small:small2.jpg;medium:medium2.jpg;large:large2.jpg"
	cmd := redis.NewStringResult(val, errors.New("error from redis client"))
	r.EXPECT().Get("test:profile:thumbnails").Return(cmd)
	svc := comic.NewRedisCharacterThumbRepository(r)
	thmbs, err := svc.Thumbnails(comic.CharacterSlug("test"))
	assert.Nil(t, thmbs)
	assert.Error(t, err)
}

func TestRedisCharacterThumbRepositoryAllThumbnails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	r := mock_comic.NewMockRedisClient(ctrl)
	cmd := redis.NewSliceCmd(nil)
	r.EXPECT().MGet(gomock.Any()).Return(cmd)
	svc := comic.NewRedisCharacterThumbRepository(r)
	thmbs, err := svc.AllThumbnails(comic.CharacterSlug("test1"), comic.CharacterSlug("test2"))
	assert.Nil(t, err)
	assert.Len(t, thmbs, 2)
}

func TestNewRedisCharacterThumbRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	r := mock_comic.NewMockRedisClient(ctrl)
	ctr := comic.NewRedisCharacterThumbRepository(r)
	assert.NotNil(t, ctr)
}

func TestNewRedisAppearancesPerYearRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	r := mock_comic.NewMockRedisClient(ctrl)
	ctr := comic.NewRedisAppearancesPerYearRepository(r)
	assert.NotNil(t, ctr)
}
