	package comic_test

import (
	"fmt"
	"github.com/aimeelaplant/comiccruncher/comic"
	"github.com/aimeelaplant/comiccruncher/internal/pgo"
	"github.com/go-pg/pg/orm"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

// The test database instance.
var testInstance = pgo.MustInstanceTest()
// The repository container with the injected db instance.
var testContainer = comic.NewPGRepositoryContainer(testInstance)

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
		VendorId:    "1",
		PublisherID: publisher.ID,
	}
	must(db.Model(character).Insert())
	source := &comic.CharacterSource{
		VendorName:  "Emma Frost (Marvel Universe)",
		CharacterID: comic.CharacterID(character.ID),
		VendorUrl:   "https://example.com",
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
		VendorId:    "2",
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

func TestPGPublisherRepository_FindBySlug(t *testing.T) {
	marvel, err := testContainer.PublisherRepository().FindBySlug("marvel")
	assert.Nil(t, err)
	assert.Equal(t, "marvel", string(marvel.Slug))
}

func TestNewPGPublisherRepository_FindBySlugReturnsNil(t *testing.T) {
	bogus, err := testContainer.PublisherRepository().FindBySlug("bogus")
	assert.Nil(t, err)
	assert.Nil(t, bogus)
}

func TestPGCharacterRepository_FindBySlugReturnsNil(t *testing.T) {
	bogus, err := testContainer.CharacterRepository().FindBySlug("bogus", true)
	assert.Nil(t, bogus)
	assert.Nil(t, err)
}

func TestCharacterRepository_FindBySlug(t *testing.T) {
	c, err := testContainer.CharacterRepository().FindBySlug("emma-frost", false)
	assert.Nil(t, err)
	assert.NotNil(t, c)
	assert.Equal(t, comic.CharacterSlug("emma-frost"), c.Slug)

	assert.True(t, c.ID > 0)
}

func TestPGCharacterRepository_Create(t *testing.T) {
	p, err := testContainer.PublisherRepository().FindBySlug("marvel")
	assert.Nil(t, err)

	charRepo := testContainer.CharacterRepository()
	c2 := &comic.Character{
		PublisherID:       p.ID,
		Name:              "  Emma Frost",
		VendorDescription: "asdsfsfd.",
		VendorId:          "3",
	}
	err = charRepo.Create(c2)
	assert.Nil(t, err)
	assert.Equal(t, "Emma Frost", c2.Name)
	assert.NotEqual(t, comic.CharacterSlug("emma-frost"), c2.Slug)
	assert.True(t, c2.ID > 0)
	assert.True(t, c2.PublisherID > 0)
	assert.True(t, c2.Publisher.ID > 0)
}

func TestPGCharacterRepository_FindAll(t *testing.T) {
	c, err := testContainer.CharacterRepository().FindAll(comic.CharacterCriteria{
		Slugs: []comic.CharacterSlug{comic.CharacterSlug("emma-frost"), comic.CharacterSlug("emma-frost-2")},
	})
	assert.Nil(t, err)
	assert.Len(t, c, 2)
}

func TestPGCharacterSyncLogRepository_FindByIdReturnsNil(t *testing.T) {
	sLog, err := testContainer.CharacterSyncLogRepository().FindById(999)

	assert.Nil(t, err)
	assert.Nil(t, sLog)
}

func TestCharacterSyncLogRepository_Update(t *testing.T) {
	syncLogRepo := testContainer.CharacterSyncLogRepository()
	characterRepo := testContainer.CharacterRepository()
	c, err := characterRepo.FindBySlug("emma-frost", true)
	assert.Nil(t, err)
		syncLogs, err := syncLogRepo.FindAllByCharacterId(c.ID)
	assert.Len(t, syncLogs, 1)
	syncLog := syncLogs[0]
	status := syncLog.SyncStatus
	assert.Equal(t, string(comic.InProgress), string(status))
	syncLog.SyncStatus = comic.Success
	err = syncLogRepo.Update(syncLog)
	assert.Nil(t, err)
	assert.Equal(t, string(comic.Success), string(syncLog.SyncStatus))
}

func TestPGCharacterIssueRepository_FindOneByReturnsNil(t *testing.T) {
	res, err := testContainer.CharacterIssueRepository().FindOneBy(100, 100)
	assert.Nil(t, res)
	assert.Nil(t, err)
}

func TestCharacterIssueRepository_Create_And_FindOneBy(t *testing.T) {
	issue := &comic.Issue{
		PublicationDate:    time.Now(),
		SaleDate:           time.Now(),
		Format:             comic.FormatStandard,
		VendorPublisher:    "Marvel",
		VendorSeriesName:   "Uncanny X-Men",
		VendorSeriesNumber: "129",
		VendorID:           "321",
	}
	err := testContainer.IssueRepository().Create(issue)
	assert.Nil(t, err)
	assert.True(t, issue.ID > 0)

	character, err := testContainer.CharacterRepository().FindBySlug("emma-frost", true)
	assert.Nil(t, err)
	ci := &comic.CharacterIssue{
		CharacterID:    character.ID,
		IssueID:        issue.ID,
		AppearanceType: comic.Alternate,
	}

	err = testContainer.CharacterIssueRepository().Create(ci)
	assert.Nil(t, err)
	assert.Equal(t, ci.AppearanceType, comic.Alternate)
	// test the repo loads the ID.
	assert.True(t, ci.ID > 0)

	nci, err := testContainer.CharacterIssueRepository().FindOneBy(character.ID, issue.ID)
	assert.Nil(t, err)
	assert.NotNil(t, nci)
	assert.Equal(t, comic.Alternate, nci.AppearanceType)
	assert.True(t, nci.ID > 0)
	assert.Equal(t, character.ID, nci.CharacterID)
	assert.Equal(t, issue.ID, nci.IssueID)
}

func TestPGIssueRepository_FindByVendorIdReturnsNil(t *testing.T) {
	result, err := testContainer.IssueRepository().FindByVendorId("98332")
	assert.Nil(t, result)
	assert.Nil(t, err)
}

func TestPGIssueRepository_CreateAll(t *testing.T) {
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
	err := testContainer.IssueRepository().CreateAll(issues)
	assert.Nil(t, err)
	// test the ID gets loaded.
	assert.True(t, issue.ID > 0)
	assert.True(t, issue2.ID > 0)
	assert.Equal(t, issue.ID, issues[0].ID)
	assert.Equal(t, issue2.ID, issues[1].ID)
}

func TestPGCharacterIssueRepository_CreateAll(t *testing.T) {
	character, err := testContainer.CharacterRepository().FindBySlug("emma-frost", true)
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
	err = testContainer.IssueRepository().CreateAll(issues)
	assert.Nil(t, err)
	assert.True(t, len(issues) > 1)
	assert.True(t, issues[0].ID > 0)
	assert.True(t, issues[1].ID > 0)

	characterIssues := []*comic.CharacterIssue{
		{CharacterID: character.ID, IssueID: issues[0].ID},
		{CharacterID: character.ID, IssueID: issues[1].ID},
	}
	err = testContainer.CharacterIssueRepository().CreateAll(characterIssues)
	assert.Nil(t, err)
	// Test the ID is loaded.
	assert.True(t, characterIssues[0].ID > 0)
	assert.True(t, characterIssues[1].ID > 0)
}

func TestNewPGRepositoryContainer(t *testing.T) {
	container := comic.NewPGRepositoryContainer(testInstance)

	assert.NotNil(t, container.PublisherRepository())
	assert.NotNil(t, container.IssueRepository())
	assert.NotNil(t, container.CharacterRepository())
	assert.NotNil(t, container.CharacterIssueRepository())
	assert.NotNil(t, container.CharacterSourceRepository())
	assert.NotNil(t, container.CharacterSyncLogRepository())
}

// Tests that the query returns the correct values for Main | Alternate appearances.
func TestPGAppearanceRepository_All(t *testing.T) {
	result, err := testContainer.AppearancesByYearsRepository().Both("emma-frost-2")

	assert.Nil(t, err)
	assert.Equal(t, "emma-frost-2", string(result.CharacterSlug))
	assert.Equal(t, comic.Main|comic.Alternate, result.Category)

	expected := make(map[int]int)
	now := time.Now()
	for i := 1979; i <= now.Year(); i++ {
		switch i {
		case 1979:
			expected[1979] = 1
		case 1980:
			expected[1980] = 2
		default:
			expected[i] = 0
		}
	}

	assert.Len(t, result.Aggregates, len(expected))

	for _, a := range result.Aggregates {
		val, ok := expected[a.Year]
		assert.True(t, ok)
		assert.Equal(t, val, a.Count)
	}

	// Test for blanks.
	bogus, err := testContainer.AppearancesByYearsRepository().Both("bogus")
	assert.Nil(t, err)
	assert.Nil(t, bogus.Aggregates)
}

// Tests that the query returns the correct values for Main appearances.
func TestPGAppearanceRepository_Main(t *testing.T) {
	result, err := testContainer.AppearancesByYearsRepository().Main("emma-frost-2")

	assert.Nil(t, err)
	assert.Equal(t, "emma-frost-2", string(result.CharacterSlug))
	assert.Equal(t, comic.Main, result.Category)
	expected := make(map[int]int)
	now := time.Now()
	for i := 1979; i <= now.Year(); i++ {
		switch i {
		case 1979:
			expected[1979] = 1
		case 1980:
			expected[1980] = 1
		default:
			expected[i] = 0
		}
	}

	assert.Len(t, result.Aggregates, len(expected))

	for _, a := range result.Aggregates {
		val, ok := expected[a.Year]
		assert.True(t, ok)
		assert.Equal(t, val, a.Count)
	}

	// Test for blanks.
	bogus, err := testContainer.AppearancesByYearsRepository().Main("bogus")
	assert.Nil(t, err)
	assert.Nil(t, bogus.Aggregates)
}

// Test that the query returns the correct values for Alternate appearances.
func TestPGAppearanceRepository_Alternate(t *testing.T) {
	result, err := testContainer.AppearancesByYearsRepository().Alternate("emma-frost-2")

	assert.Nil(t, err)
	assert.Equal(t, "emma-frost-2", string(result.CharacterSlug))
	assert.Equal(t, comic.Alternate, result.Category)

	expected := make(map[int]int)
	now := time.Now()
	for i := 1979; i <= now.Year(); i++ {
		switch i {
		case 1980:
			expected[1980] = 2
		default:
			expected[i] = 0
		}
	}

	assert.Len(t, result.Aggregates, len(expected))

	for _, a := range result.Aggregates {
		val, ok := expected[a.Year]
		assert.True(t, ok)
		assert.Equal(t, val, a.Count)
	}

	// Test for blanks.
	bogus, err := testContainer.AppearancesByYearsRepository().Alternate("bogus")
	assert.Nil(t, err)
	assert.Nil(t, bogus.Aggregates)
}

// Test that the query returns both main and alternate appearances as slices.
func TestPGAppearanceRepository_List(t *testing.T) {
	list, err := testContainer.AppearancesByYearsRepository().List("emma-frost-2")

	assert.Nil(t, err)
	assert.Len(t, list, 2)
	assert.Equal(t, comic.Main, list[0].Category)
	assert.Equal(t, comic.Alternate, list[1].Category)

	// Test for blanks.
	bogus, err := testContainer.AppearancesByYearsRepository().List("bogus")
	assert.Nil(t, err)
	assert.Len(t, bogus, 0)
}

func TestPGStatsRepository_Stats(t *testing.T) {
	stats, err := testContainer.StatsRepository().Stats()
	assert.Nil(t, err)
	assert.NotNil(t, stats)

	assert.Equal(t, 8, stats.TotalIssues)
	assert.Equal(t, 6, stats.TotalAppearances)
	assert.Equal(t, 1979, stats.MinYear)
	assert.Equal(t, time.Now().Year(), stats.MaxYear)
}

func TestPGCharacterSyncLogRepository_Create_and_Find(t *testing.T) {
	characters, err := testContainer.CharacterRepository().FindAll(comic.CharacterCriteria{})
	assert.Nil(t, err)
	assert.True(t, len(characters) > 0)

	syncLog := comic.NewSyncLog(characters[0].ID, comic.Pending, comic.YearlyAppearances, nil)
	testContainer.CharacterSyncLogRepository().Create(syncLog)
	// asserts it auto-increments
	assert.True(t, syncLog.ID > 0)

	id := syncLog.ID

	syncLog, err = testContainer.CharacterSyncLogRepository().FindById(syncLog.ID)
	assert.Nil(t, err)
	assert.Equal(t, id, syncLog.ID)
}


func TestPGCharacterRepository_Total(t *testing.T) {
	total, err := testContainer.CharacterRepository().Total(comic.CharacterCriteria{
		IDs: []comic.CharacterID{1},
		FilterSources: true,
	})
	assert.Nil(t, err)
	assert.True(t, total >= 0)
}
