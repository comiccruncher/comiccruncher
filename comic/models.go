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
	FormatPrestige     Format = "prestige"
	FormatOther        Format = "other"
)

// The types of appearances for a character issue.
// Bitwise values to represent appearance types.
const (
	// Main is their main universe(s)
	Main AppearanceType = 1 << 0
	// Alternate is an alternate reality appearance or whatever.
	Alternate AppearanceType = 1 << 1
)

// Consts for a later feature.
// The available importance types.
const (
	// Cameo - they just make a cameo appearance
	Cameo Importance = iota + 1
	// Minor - meh, minor
	Minor
	// Major  character in issue
	Major
)

// Constants for character sync log types.
const (
	// YearlyAppearances is the syncing for yearly appearances.
	YearlyAppearances CharacterSyncLogType = iota + 1
	// Characters is the syncing for characters.
	Characters
)

// Constants for character sync log statuses.
const (
	// Pending - when a sync is pending and waiting in the queue.
	Pending CharacterSyncLogStatus = iota + 1
	// InProgress - when a sync is currently in progress and tallying appearances.
	InProgress
	// Fail - when a sync failed.
	Fail
	// Success - when a sync succeeded.
	Success
)

// A map for the string values of appearance types.
var categoryToString = map[AppearanceType]string{
	Main:             "main",
	Alternate:        "alternate",
	Main | Alternate: "all",
}

var cdnURL = os.Getenv("CC_CDN_URL")

// PublisherID is the PK identifier for the publisher.
type PublisherID uint

// PublisherSlug is the unique string identifier for the publisher.
type PublisherSlug string

// IssueID is the PK identifier for the issue.
type IssueID uint

// CharacterID is the PK identifier for the character.
type CharacterID uint

// CharacterSlug is the unique slug for the character.
type CharacterSlug string

// CharacterIssueID is the PK identifier for a character issue.
type CharacterIssueID uint

// CharacterSourceID is the PK identifier for the character source struct.
type CharacterSourceID uint

// CharacterSyncLogID is the PK identifier for character sync logs.
type CharacterSyncLogID uint

// CharacterSyncLogType is the type of sync that occurred for the character.
type CharacterSyncLogType int

// CharacterSyncLogStatus is the status of the sync.
type CharacterSyncLogStatus int

// Format is the format for the issue.
type Format string

// VendorType is type of vendor from an external source for an issue.
type VendorType int

// IssueCountRankID is the ranking for the number of issues for a character.
type IssueCountRank uint

// AvgIssuesPerYearRank is the rank for average issues per year.
type AvgIssuesPerYearRank uint

// AppearanceType is a type of appearance, such as an alternate universe or main character appearance.
// A bitwise enum representing the types of appearances.
// Main is 001
// Alternate is 100
// Both Main and Alternate would be 101 so: `Main | Alternate`
type AppearanceType uint8

// Importance -- for a later feature -- ranks a character issue by the character's importance in the issue.
type Importance int

// AppearancesByYears represents the key, category, and appearances categorized per year for a character.
type AppearancesByYears struct {
	CharacterSlug CharacterSlug     `json:"slug"` // The unique identifier for the character.
	Category      AppearanceType    `json:"category"`
	Aggregates    []YearlyAggregate `json:"aggregates"`
}

// YearlyAggregate is the aggregated year and count of an appearance for that year.
type YearlyAggregate struct {
	Year  int `json:"year"`
	Count int `json:"count"`
}

// Publisher is a publisher is an entity that publishes comics and characters.
type Publisher struct {
	tableName struct{}      `pg:",discard_unknown_columns"`
	ID        PublisherID   `json:"-"`
	Name      string        `json:"name" sql:",notnull"`
	Slug      PublisherSlug `json:"slug" sql:",notnull,unique:uix_publisher_slug"`
	CreatedAt time.Time     `json:"-" sql:",notnull,default:NOW()"`
	UpdatedAt time.Time     `sql:",notnull,default:NOW()" json:"-"`
}

// Issue is an issue with details about its publication and on sale dates.
type Issue struct {
	tableName          struct{} `pg:",discard_unknown_columns"`
	ID                 IssueID
	PublicationDate    time.Time `sql:",notnull"`
	SaleDate           time.Time `sql:",notnull"` // @TODO: add an index.
	IsVariant          bool      `sql:",notnull"`
	MonthUncertain     bool      `sql:",notnull"`
	Format             Format    `sql:",notnull"`
	VendorPublisher    string    `sql:",notnull"`
	VendorSeriesName   string    `sql:",notnull"`
	VendorSeriesNumber string    `sql:",notnull"`
	// IsReprint means the issue is a full reprint with no original story. (So something like Classic X-Men 7 would not count).
	IsReprint  bool       `sql:"default:false,notnull"`
	VendorType VendorType `sql:",notnull,unique:uix_vendor_type_vendor_id,type:smallint"`
	VendorID   string     `sql:",notnull,unique:uix_vendor_type_vendor_id"`
	CreatedAt  time.Time  `sql:",notnull,default:NOW()" json:"-"`
	UpdatedAt  time.Time  `sql:",notnull,default:NOW()" json:"-"`
}

// Character - A model for a character.
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
	VendorID          string        `sql:",notnull,unique:uix_vendor_type_vendor_id" json:"-"`
	VendorImage       string        `json:"vendor_image"`
	VendorImageMd5    string        `sql:",type:varchar(32)," json:"-"`
	VendorURL         string        `json:"vendor_url"`
	VendorDescription string        `json:"vendor_description"`
	IsDisabled        bool          `json:"-" sql:",notnull"`
	CreatedAt         time.Time     `sql:",notnull,default:NOW()" json:"-"`
	UpdatedAt         time.Time     `sql:",notnull,default:NOW()" json:"-"`
}

// CharacterSource contains external profile links to the character.
type CharacterSource struct {
	tableName       struct{}          `pg:",discard_unknown_columns"`
	ID              CharacterSourceID `json:"id"`
	Character       *Character        // Pointer. Could be nil. Not eager-loaded.
	CharacterID     CharacterID       `pg:",fk:character_id" sql:",notnull,unique:uix_character_id_vendor_url,on_delete:CASCADE" json:"character_id"`
	VendorType      VendorType        `sql:",notnull,type:smallint" json:"type"`
	VendorURL       string            `sql:",notnull,unique:uix_character_id_vendor_url"`
	VendorName      string            `sql:",notnull"`
	VendorOtherName string
	IsDisabled      bool      `sql:",notnull"`
	IsMain          bool      `sql:",notnull"`
	CreatedAt       time.Time `sql:",default:NOW(),notnull" json:"-"`
	UpdatedAt       time.Time `sql:",notnull,default:NOW()" json:"-"`
}

// CharacterSyncLog contains information pertaining to syncs for the character.
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

// CharacterIssue references an issue for a character.
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

// Stats represents general stats about the db.
type Stats struct {
	TotalCharacters  int `json:"total_characters"`
	TotalAppearances int `json:"total_appearances"`
	MinYear          int `json:"min_year"`
	MaxYear          int `json:"max_year"`
	TotalIssues      int `json:"total_issues"`
}

// RankedCharacter represents a character who has its rank and issue count accounted for
// with its appearances attached..
type RankedCharacter struct {
	ID                CharacterID          `json:"-"`
	Publisher         Publisher            `json:"publisher"`
	PublisherID       PublisherID          `json:"-"`
	AvgRankID         AvgIssuesPerYearRank `json:"average_issues_per_year_rank"`
	AvgRank           float64              `json:"average_issues_per_year"`
	IssueCountRankID  IssueCountRank       `json:"issue_count_rank"`
	IssueCount        uint                 `json:"issue_count"`
	Name              string               `json:"name"`
	OtherName         string               `json:"other_name"`
	Description       string               `json:"description"`
	Image             string               `json:"image"`
	Slug              CharacterSlug        `json:"slug"`
	VendorImage       string               `json:"vendor_image"`
	VendorURL         string               `json:"vendor_url"`
	VendorDescription string               `json:"vendor_description"`
	Appearances       []AppearancesByYears `json:"appearances"`
}

// MarshalJSON overrides the image and vendor image for the CDN url.
func (c *RankedCharacter) MarshalJSON() ([]byte, error) {
	strctImage := ""
	if c.Image != "" {
		strctImage = fmt.Sprintf("%s/%s", cdnURL, c.Image)
	}
	strctVendorImage := ""
	if c.VendorImage != "" {
		strctVendorImage = fmt.Sprintf("%s/%s", cdnURL, c.VendorImage)
	}
	type Alias RankedCharacter
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
// HasAny checks that the category has any of the given flags.
func (u AppearanceType) HasAny(flags AppearanceType) bool {
	return u&flags > 0
}

// HasAll checks that the category has all of the given flags.
func (u AppearanceType) HasAll(flags AppearanceType) bool {
	flagValues := uint8(flags)
	return uint8(u)&flagValues == flagValues
}

// MarshalJSON returns the JSON string representation.
func (u AppearanceType) MarshalJSON() ([]byte, error) {
	return json.Marshal(categoryToString[u])
}

// Scan for the ORM converting the enum.
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

// Value for the ORM converting the enum.
func (u AppearanceType) Value() (driver.Value, error) {
	return fmt.Sprintf("%08b", byte(u)), nil
}

// AddAppearance adds an appearance to the appearances for the character.
func (c *AppearancesByYears) AddAppearance(appearance YearlyAggregate) *AppearancesByYears {
	c.Aggregates = append(c.Aggregates, appearance)
	return c
}

// Total returns the total number of appearances per year.
func (c *AppearancesByYears) Total() int {
	total := 0
	for _, a := range c.Aggregates {
		total += a.Count
	}
	return total
}

// Value returns the raw value
func (id IssueID) Value() uint {
	return uint(id)
}

// Value returns the raw value.
func (id CharacterID) Value() uint {
	return uint(id)
}

// Value returns the raw value.
func (slug CharacterSlug) Value() string {
	return string(slug)
}

// Value returns the raw value.
func (id CharacterSyncLogID) Value() uint {
	return uint(id)
}

// Value returns the raw value.
func (slug PublisherSlug) Value() string {
	return string(slug)
}

// MarshalJSON overrides JSON marshaling for CDN url.
func (c *Character) MarshalJSON() ([]byte, error) {
	strctImage := ""
	if c.Image != "" {
		strctImage = fmt.Sprintf("%s/%s", cdnURL, c.Image)
	}
	strctVendorImage := ""
	if c.VendorImage != "" {
		strctVendorImage = fmt.Sprintf("%s/%s", cdnURL, c.VendorImage)
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

// NewCharacterSlugs creates character slugs from the specified `strs` string.
func NewCharacterSlugs(strs ...string) []CharacterSlug {
	slugs := make([]CharacterSlug, len(strs))
	for i := range strs {
		slugs[i] = CharacterSlug(strs[i])
	}
	return slugs
}

// NewCharacter Creates a new character.
func NewCharacter(name string, publisherID PublisherID, vendorType VendorType, vendorID string) *Character {
	return &Character{
		Name:        name,
		PublisherID: PublisherID(publisherID),
		VendorType:  vendorType,
		VendorID:    vendorID,
	}
}

// NewCharacterSyncLog creates a new sync log object for syncing characters.
func NewCharacterSyncLog(id CharacterID, status CharacterSyncLogStatus, syncedAt *time.Time) *CharacterSyncLog {
	return &CharacterSyncLog{
		CharacterID: id,
		SyncType:    Characters,
		SyncStatus:  status,
		SyncedAt:    syncedAt,
	}
}

// NewIssue creates a new issue struct.
func NewIssue(
	vendorID, vendorPublisher, vendorSeriesName, vendorSeriesNumber string,
	publicationDate, saleDate time.Time,
	isVariant, monthUncertain, isReprint bool,
	format Format) *Issue {
	return &Issue{
		VendorID:           vendorID,
		VendorSeriesName:   vendorSeriesName,
		VendorSeriesNumber: vendorSeriesNumber,
		VendorPublisher:    vendorPublisher,
		PublicationDate:    publicationDate,
		SaleDate:           saleDate,
		IsVariant:          isVariant,
		IsReprint:          isReprint,
		VendorType:         VendorTypeCb, // Make this default for now.
		MonthUncertain:     monthUncertain,
		Format:             format,
	}
}

// NewSyncLog Creates a pointer to a new sync log object for the yearly appearances category.
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

// NewSyncLogPending creates a new pending sync log struct for the specified type.
func NewSyncLogPending(
	id CharacterID,
	syncLogType CharacterSyncLogType) *CharacterSyncLog {
	return &CharacterSyncLog{
		CharacterID: id,
		SyncStatus:  Pending,
		SyncType:    syncLogType}
}

// NewCharacterSource creates a new character source struct.
func NewCharacterSource(url, name string, id CharacterID, vendorType VendorType) *CharacterSource {
	return &CharacterSource{
		VendorURL:   url,
		VendorName:  name,
		CharacterID: id,
		VendorType:  vendorType,
	}
}

// NewCharacterIssue creates a new character issue struct.
func NewCharacterIssue(characterID CharacterID, id IssueID, appearanceType AppearanceType) *CharacterIssue {
	return &CharacterIssue{
		CharacterID:    characterID,
		IssueID:        id,
		AppearanceType: appearanceType,
	}
}

// NewAppearancesByYears creates a new struct with the parameters.
func NewAppearancesByYears(slug CharacterSlug, cat AppearanceType, aggs []YearlyAggregate) AppearancesByYears {
	return AppearancesByYears{
		CharacterSlug: slug,
		Category:      cat,
		Aggregates:    aggs,
	}
}
