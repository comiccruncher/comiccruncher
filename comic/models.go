package comic

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

// Vendor types for the characters and character sources.
const (
	VendorTypeCb VendorType = iota
	VendorTypeMarvel
	VendorTypeDC
)

// The format types for the issue.
const (
	FormatUnknown      Format = "unknown"
	FormatStandard     Format = "standard"
	FormatTPB          Format = "tpb"
	FormatManga        Format = "manga"
	FormatHC           Format = "hc"
	FormatOGN          Format = "ogn"
	FormatWeb          Format = "web"
	FormatAnthology    Format = "anthology"
	FormatMagazine     Format = "magazine"
	FormatDigitalMedia Format = "digital"
	FormatMiniComic    Format = "mini"
	FormatFlipbook     Format = "flipbook"
	FormatOther        Format = "other"
)

// The types of appearances for a character issue.
const (
	// Bitwise values to represent appearance types.
	Main      AppearanceType = 1 << 0
	Alternate AppearanceType = 1 << 1
)

// Consts for a later feature.
const (
	// The available importance types.
	Cameo Importance = iota + 1
	Minor
	Major
)

// Constants for character sync log types.
const (
	// The syncing for yearly appearances.
	YearlyAppearances CharacterSyncLogType = iota + 1
	// The syncing for characters.
	Characters
)

// Constants for character sync log statuses.
const (
	// When a sync is pending and waiting in the queue.
	Pending CharacterSyncLogStatus = iota + 1
	// When a sync is currently in progress and tallying appearances.
	InProgress
	// When a sync failed.
	Fail
	// When a sync succeeded.
	Success
)

// A map for the string values of appearance types.
var categoryToString = map[AppearanceType]string{
	Main:             "main",
	Alternate:        "alternate",
	Main | Alternate: "all",
}

var cdnUrl = os.Getenv("CC_CDN_URL")

// The PK identifier for the publisher.
type PublisherID uint

// The unique string identifier for the publisher.
type PublisherSlug string

// The PK identifier for the issue.
type IssueID uint

// The PK identifier for the character.
type CharacterID uint

// The unique slug for the character.
type CharacterSlug string

// The PK identifier for a character issue.
type CharacterIssueID uint

// The PK identifier for the character source struct.
type CharacterSourceID uint

// The PK identifier for character sync logs.
type CharacterSyncLogID uint

// The type of sync that occurred for the character.
type CharacterSyncLogType int

// The status of the sync.
type CharacterSyncLogStatus int

// The format for the issue.
type Format string

// A type of vendor from an external source for an issue or ____ TODO.
type VendorType int

// A type of appearance, such as an alternate universe or main character appearance.
// A bitwise enum representing the types of appearances.
// Main is 001
// Alternate is 100
// Both Main and Alternate would be 101 so: `Main | Alternate`
type AppearanceType uint8

// For a later feature -- ranks a character issue by the character's importance in the issue.
type Importance int

// Represents the key, category, and appearances categorized per year for a character.
type AppearancesByYears struct {
	CharacterSlug CharacterSlug     `json:"slug"` // The unique identifier for the character.
	Category      AppearanceType    `json:"category"`
	Aggregates    []YearlyAggregate `json:"aggregates"`
}

// The aggregated year and count of an appearance for that year.
type YearlyAggregate struct {
	Year  int `json:"year"`
	Count int `json:"count"`
}

// A publisher is an entity that publishes comics and characters.
type Publisher struct {
	tableName struct{}      `pg:",discard_unknown_columns"`
	ID        PublisherID   `json:"-"`
	Name      string        `json:"name" sql:",notnull"`
	Slug      PublisherSlug `json:"slug" sql:",notnull,unique:uix_publisher_slug"`
	CreatedAt time.Time     `json:"-" sql:",notnull,default:NOW()"`
	UpdatedAt time.Time     `sql:",notnull,default:NOW()" json:"-"`
}

// An issue with details about its publication and on sale dates.
type Issue struct {
	tableName          struct{} `pg:",discard_unknown_columns"`
	ID                 IssueID
	PublicationDate    time.Time  `sql:",notnull"`
	SaleDate           time.Time  `sql:",notnull"` // @TODO: add an index.
	IsVariant          bool       `sql:",notnull"`
	MonthUncertain     bool       `sql:",notnull"`
	Format             Format     `sql:",notnull"`
	VendorPublisher    string     `sql:",notnull"`
	VendorSeriesName   string     `sql:",notnull"`
	VendorSeriesNumber string     `sql:",notnull"`
	VendorType         VendorType `sql:",notnull,unique:uix_vendor_type_vendor_id,type:smallint"`
	VendorID           string     `sql:",notnull,unique:uix_vendor_type_vendor_id"`
	CreatedAt          time.Time  `sql:",notnull,default:NOW()" json:"-"`
	UpdatedAt          time.Time  `sql:",notnull,default:NOW()" json:"-"`
}

// Criteria for querying issues.
type IssueCriteria struct {
	Ids        []IssueID
	VendorIds  []string
	VendorType VendorType
	Limit      int
	Offset     int
}

// A model for a character.
type Character struct {
	tableName         struct{}      `pg:",discard_unknown_columns"`
	ID                CharacterID   `json:"-"`
	Publisher         Publisher     `json:"publisher"`
	PublisherID       PublisherID   `pg:",fk:publisher_id" sql:",notnull,on_delete:CASCADE" json:"-"`
	Name              string        `sql:",notnull" json:"name"`
	OtherName         string        `json:"other_name"`
	Description       string        `json:"description"`
	Image             string        `json:"image"`
	Slug              CharacterSlug `sql:",notnull,unique:uix_character_slug" json:"slug"`
	VendorType        VendorType    `sql:",notnull,unique:uix_vendor_type_vendor_id" json:"-"`
	VendorId          string        `sql:",notnull,unique:uix_vendor_type_vendor_id" json:"-"`
	VendorImage       string        `json:"vendor_image"`
	VendorImageMd5    string        `sql:",type:varchar(32)," json:"-"`
	VendorUrl         string        `json:"vendor_url"`
	VendorDescription string        `json:"vendor_description"`
	IsDisabled        bool          `json:"-" sql:",notnull"`
	CreatedAt         time.Time     `sql:",notnull,default:NOW()" json:"-"`
	UpdatedAt         time.Time     `sql:",notnull,default:NOW()" json:"-"`
}

// A model that contains external profile links to the character.
type CharacterSource struct {
	tableName       struct{}          `pg:",discard_unknown_columns"`
	ID              CharacterSourceID `json:"id"`
	Character       *Character        // Pointer. Could be nil. Not eager-loaded.
	CharacterID     CharacterID       `pg:",fk:character_id" sql:",notnull,unique:uix_character_id_vendor_url,on_delete:CASCADE" json:"character_id"`
	VendorType      VendorType        `sql:",notnull,type:smallint" json:"type"`
	VendorUrl       string            `sql:",notnull,unique:uix_character_id_vendor_url"`
	VendorName      string            `sql:",notnull"`
	VendorOtherName string
	IsDisabled      bool      `sql:",notnull"`
	IsMain          bool      `sql:",notnull"`
	CreatedAt       time.Time `sql:",default:NOW(),notnull" json:"-"`
	UpdatedAt       time.Time `sql:",notnull,default:NOW()" json:"-"`
}

// Criteria for querying character sources.
type CharacterSourceCriteria struct {
	CharacterIDs []CharacterID
	VendorUrls   []string
	VendorType   VendorType
	// If IsMain is null, it will  return both.
	IsMain *bool
	// Include sources that are disabled. By default it does not include disabled sources.
	IncludeIsDisabled bool
	Limit             int
	Offset            int
}

// A model that contains information pertaining to syncs for the character.
type CharacterSyncLog struct {
	tableName   struct{}               `pg:",discard_unknown_columns"`
	ID          CharacterSyncLogID     `json:"id"`
	SyncType    CharacterSyncLogType   `sql:",notnull,type:smallint" json:"type"`
	SyncStatus  CharacterSyncLogStatus `sql:",notnull,type:smallint" json:"status"`
	Message     string
	SyncedAt    *time.Time  `json:"synced_at"`
	Character   *Character  // Not eager-loaded, could be nil.
	CharacterID CharacterID `pg:",fk:character_id" sql:",notnull,on_delete:CASCADE" json:"character_id"`
	CreatedAt   time.Time   `sql:",notnull,default:NOW()" json:"-"`
	UpdatedAt   time.Time   `sql:",notnull,default:NOW()" json:"-"`
}

// A model that references an issue for a character.
type CharacterIssue struct {
	tableName      struct{} `pg:",discard_unknown_columns"`
	ID             CharacterIssueID
	Character      *Character     // Not eager-loaded. Could be nil.
	CharacterID    CharacterID    `pg:",fk:character_id" sql:",notnull,unique:uix_character_id_issue_id,on_delete:CASCADE"`
	Issue          *Issue         // Not eager-loaded. Could be nil.
	IssueID        IssueID        `pg:",fk:issue_id" sql:",notnull,unique:uix_character_id_issue_id,on_delete:CASCADE"`
	AppearanceType AppearanceType `sql:",notnull,type:bit(8),default:B'00000001'"`
	Importance     *Importance    `sql:",type:smallint"`
	CreatedAt      time.Time      `sql:",notnull,default:NOW()" json:"-"`
	UpdatedAt      time.Time      `sql:",notnull,default:NOW()" json:"-"`
}

// Criteria for querying characters.
type CharacterCriteria struct {
	IDs               []CharacterID
	Slugs             []CharacterSlug
	PublisherIDs      []PublisherID
	PublisherSlugs    []PublisherSlug
	FilterSources     bool         // Filter characters that only have sources. If false it returns characters regardless.
	FilterIssues      bool         // Filter characters that only have issues. If false it returns characters regardless.
	VendorTypes       []VendorType // Include characters that are disabled. By default it does not.
	IncludeIsDisabled bool
	VendorIds         []string
	Limit             int
	Offset            int
}

// Represents general stats about the db.
type Stats struct {
	TotalCharacters  int `json:"total_characters"`
	TotalAppearances int `json:"total_appearances"`
	MinYear          int `json:"min_year"`
	MaxYear          int `json:"max_year"`
	TotalIssues      int `json:"total_issues"`
}

// Checks that the category has any of the given flags.
func (u AppearanceType) HasAny(flags AppearanceType) bool {
	return u&flags > 0
}

// Checks that the category has all of the given flags.
func (u AppearanceType) HasAll(flags AppearanceType) bool {
	flagValues := uint8(flags)
	return uint8(u)&flagValues == flagValues
}

// Returns the JSON string representation.
func (u AppearanceType) MarshalJSON() ([]byte, error) {
	return json.Marshal(categoryToString[u])
}

// For the ORM converting the enum.
func (u *AppearanceType) Scan(value interface{}) error {
	val, ok := value.([]byte)
	if ok {
		val, err := strconv.ParseUint(string(val), 2, 8)
		if err != nil {
			return err
		}
		*u = AppearanceType(val)
	} else {
		return errors.New("byte not returned. this is bad")
	}
	return nil
}

// For the ORM converting the enum.
func (u AppearanceType) Value() (driver.Value, error) {
	return fmt.Sprintf("%08b", byte(u)), nil
}

// Adds an appearance to the appearances for the character.
func (c *AppearancesByYears) AddAppearance(appearance YearlyAggregate) *AppearancesByYears {
	c.Aggregates = append(c.Aggregates, appearance)
	return c
}

// Returns the total number of appearances per year.
func (c *AppearancesByYears) Total() int {
	total := 0
	for _, a := range c.Aggregates {
		total += a.Count
	}
	return total
}

// Returns the raw value.
func (id CharacterID) Value() uint {
	return uint(id)
}

// Returns the raw value.
func (slug CharacterSlug) Value() string {
	return string(slug)
}

// Returns the raw value.
func (id CharacterSyncLogID) Value() uint {
	return uint(id)
}

// Returns the raw value.
func (slug PublisherSlug) Value() string {
	return string(slug)
}

// Overrides JSON marshaling for CDN url.
func (c *Character) MarshalJSON() ([]byte, error) {
	strctImage := ""
	if c.Image != "" {
		strctImage = fmt.Sprintf("%s/%s", cdnUrl, c.Image)
	}
	strctVendorImage := ""
	if c.VendorImage != "" {
		strctVendorImage = fmt.Sprintf("%s/%s", cdnUrl, c.VendorImage)
	}
	type Alias Character
	return json.Marshal(&struct {
		*Alias
		Image       string `json:"image"`
		VendorImage string `json:"vendor_image"`
	}{
		Alias:       (*Alias)(c),
		Image:       strctImage,
		VendorImage: strctVendorImage,
	})
}

// Creates character slugs from the specified `strs` string.
func NewCharacterSlugs(strs ...string) []CharacterSlug {
	slugs := make([]CharacterSlug, len(strs))
	for i := range strs {
		slugs[i] = CharacterSlug(strs[i])
	}
	return slugs
}

// Creates a new character.
func NewCharacter(name string, publisherId PublisherID, vendorType VendorType, vendorId string) *Character {
	return &Character{
		Name:        name,
		PublisherID: PublisherID(publisherId),
		VendorType:  vendorType,
		VendorId:    vendorId,
	}
}

// Creates a new sync log object for syncing characters.
func NewCharacterSyncLog(id CharacterID, status CharacterSyncLogStatus, syncedAt *time.Time) *CharacterSyncLog {
	return &CharacterSyncLog{
		CharacterID: id,
		SyncType:    Characters,
		SyncStatus:  status,
		SyncedAt:    syncedAt,
	}
}

// Creates a new issue struct.
func NewIssue(
	vendorId, vendorPublisher, vendorSeriesName, vendorSeriesNumber string,
	publicationDate, saleDate time.Time,
	isVariant, monthUncertain bool,
	format Format) *Issue {
	return &Issue{
		VendorID:           vendorId,
		VendorSeriesName:   vendorSeriesName,
		VendorSeriesNumber: vendorSeriesNumber,
		VendorPublisher:    vendorPublisher,
		PublicationDate:    publicationDate,
		SaleDate:           saleDate,
		IsVariant:          isVariant,
		VendorType:         VendorTypeCb, // Make this default for now.
		MonthUncertain:     monthUncertain,
		Format:             format,
	}
}

// Creates a pointer to a new sync log object for the yearly appearances category.
func NewSyncLog(
	id CharacterID,
	status CharacterSyncLogStatus,
	t CharacterSyncLogType,
	syncedAt *time.Time) *CharacterSyncLog {
	return &CharacterSyncLog{
		CharacterID: id,
		SyncType:    t,
		SyncStatus:  status,
		SyncedAt:    syncedAt,
	}
}

// Creates a new pending sync log struct for the specified type.
func NewSyncLogPending(
	id CharacterID,
	syncLogType CharacterSyncLogType) *CharacterSyncLog {
	return &CharacterSyncLog{
		CharacterID: id,
		SyncStatus:  Pending,
		SyncType:    syncLogType}
}

// Creates a new character source struct.
func NewCharacterSource(url, name string, id CharacterID, vendorType VendorType) *CharacterSource {
	return &CharacterSource{
		VendorUrl:   url,
		VendorName:  name,
		CharacterID: id,
		VendorType:  vendorType,
	}
}

// Creates a new character issue struct.
func NewCharacterIssue(characterID CharacterID, id IssueID, appearanceType AppearanceType) *CharacterIssue {
	return &CharacterIssue{
		CharacterID:    characterID,
		IssueID:        id,
		AppearanceType: appearanceType,
	}
}
