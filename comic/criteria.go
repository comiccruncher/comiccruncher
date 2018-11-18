package comic

// IssueCriteria for querying issues.
type IssueCriteria struct {
	Ids        []IssueID
	VendorIds  []string
	VendorType VendorType
	Formats    []Format
	Limit      int
	Offset     int
}

// CharacterSourceCriteria for querying character sources.
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

// CharacterCriteria for querying characters.
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

// PopularSortCriteria is criteria for sorting popular characters.
type PopularSortCriteria string

const (
	// MostIssues sorts by the most issues for a character.
	MostIssues PopularSortCriteria = "issue_count_rank"
	// AverageIssuesPerYear sorts by the highest average issues per year for each character.
	AverageIssuesPerYear = "average_rank"
)

// PopularCriteria is for querying ranked and popular characters.
type PopularCriteria struct {
	AppearanceType AppearanceType
	SortBy         PopularSortCriteria
	Limit          int
	Offset         int
}

// TrendingCriteria is for querying characters who are trending.
type TrendingCriteria struct {
	PublisherID PublisherID
	Limit       int
	Offset      int
}
