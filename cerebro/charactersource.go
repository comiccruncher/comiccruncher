package cerebro

import (
	"errors"
	"github.com/aimeelaplant/comiccruncher/comic"
	"github.com/aimeelaplant/comiccruncher/internal/log"
	"github.com/aimeelaplant/comiccruncher/internal/stringutil"
	"github.com/aimeelaplant/externalissuesource"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"sync"
)

// CharacterSourceImporter is responsible for importing a characters' sources into a persistence layer.
type CharacterSourceImporter struct {
	httpClient     *http.Client
	characterSvc   comic.CharacterServicer
	externalSource externalissuesource.ExternalSource
	logger         *zap.Logger
	mu             sync.Mutex
}

// CharacterSourceImportItem represents issue sources whose import was attempted and an error if an unexpected event occurred.
type CharacterSourceImportItem struct {
	Source *comic.CharacterSource
	Error  error
}

// searchResults contains the corresponding character and any search results for the character.
type searchResults struct {
	Character     *comic.Character
	SearchResults []externalissuesource.CharacterSearchResult
	Error         error
}

// Retries to get a character page.
func (i *CharacterSourceImporter) retryCharacterPage(url string) (*externalissuesource.CharacterPage, error) {
	pageChan := make(chan *externalissuesource.CharacterPage, 1)
	defer close(pageChan)
	err := retryConnectionError(func() (string, error) {
		i.logger.Info("requesting page...", zap.String("link", url))
		cPage, err := i.externalSource.CharacterPage(url)
		if err != nil {
			return url, err
		}
		// send it over
		pageChan <- cPage
		return url, nil
	})
	if err != nil {
		i.logger.Error("error fetching url", zap.String("url", url), zap.Error(err))
		return nil, err
	}
	// Read from it
	page, ok := <-pageChan
	if !ok {
		return nil, errors.New("error on channel")
	}
	return page, nil
}

// Creates a source from a character and external link if it doesn't already exist.
// If the source wasn't created, then it returns `nil`.
// This also does a recursive call to create sources from other identities if they exist.
func (i *CharacterSourceImporter) createIfNotExists(c *comic.Character, l externalissuesource.CharacterLink, depth ...int) error {
	// Check if we have the source first before requesting it.
	src, err := i.characterSvc.Source(c.ID, l.Url)
	if err != nil {
		return err
	}
	if src != nil {
		i.logger.Info(
			"source already exists for character. skipping",
			zap.String("source", src.VendorURL),
			zap.String("character", c.Slug.Value()))
		return nil
	}
	// request the character page so we can get the other name for the character.
	// why do we need the other name? makes it easier to disable crap we don't need.
	page, err := i.retryCharacterPage(l.Url)
	if err != nil {
		return err
	}
	src = comic.NewCharacterSource(l.Url, l.Name, c.ID, comic.VendorTypeCb)
	// Update the source with the other name if `page.OtherName` is blank
	trimmedOn := strings.Trim(strings.TrimSpace(page.OtherName), ".")
	if src.VendorOtherName == "" && trimmedOn != "" {
		src.VendorOtherName = trimmedOn
	}
	err = i.characterSvc.CreateSource(src)
	if err == nil {
		i.logger.Info(
			"created source",
			zap.String("vendor url", l.Url),
			zap.String("vendor name", src.VendorName),
			zap.String("character", c.Slug.Value()))
	}

	// now go for other identities.
	if len(page.OtherIdentities) > 0 && len(depth) == 0 {
		i.logger.Info("getting other identities")
		for _, o := range page.OtherIdentities {
			// recursive call to create the other identities
			// pass in 1 because we just want to make 1 recursive call for now.
			if err2 := i.createIfNotExists(c, o, 1); err2 != nil {
				return err
			}
		}
	}
	return nil
}

// Creates a source if the link doesn't already exist in the character sources.
func (i *CharacterSourceImporter) importSources(c *comic.Character, l externalissuesource.CharacterLink) error {
	if err := i.createIfNotExists(c, l); err != nil {
		i.logger.Error("error importing sources", zap.String("character", c.Slug.Value()), zap.Error(err))
		return err
	}
	return nil
}

// SearchableName returns a search-friendly name for the external source and removes any parentheses and periods.
// If parensIndex is > -1, it will get the name within the parentheses.
func SearchableName(s string, parensIndex int) string {
	if parensIndex == -1 {
		return strings.Replace(strings.Replace(strings.Replace(s, "(", "", -1), ")", "", -1), ".", "", -1)
	}
	return strings.Replace(s[parensIndex+1:strings.Index(s, ")")], ".", "", -1)
}

// Searches for a name until there is no connection error.
func (i *CharacterSourceImporter) retrySearchByName(name string) (externalissuesource.CharacterSearchResult, error) {
	resultCh := make(chan externalissuesource.CharacterSearchResult, 1)
	defer close(resultCh)
	resultErrCh := make(chan error, 1)
	defer close(resultErrCh)
	err := retryConnectionError(func() (string, error) {
		i.logger.Info("searching for character name", zap.String("query", name))
		// Gonna have to lock this resource to avoid race conditions.
		// Or I can just pass in a copy of the external source.
		// TODO: make more intuitive later.
		i.mu.Lock()
		result, err := i.externalSource.SearchCharacter(name)
		i.mu.Unlock()
		if err != nil {
			return name, err
		}
		resultCh <- result
		return name, nil
	})
	if err != nil {
		resultErrCh <- err
	}
	for {
		select {
		case errC := <-resultErrCh:
			return externalissuesource.CharacterSearchResult{}, errC
		case result := <-resultCh:
			return result, nil
		}
	}
}

// Performs a search on a character received from the `characters` chan and sends the search result over to the `results` chan.
// The caller of the method is responsible for closing the channels.
func (i *CharacterSourceImporter) gatherSearchResults(workerID int, characters <-chan *comic.Character, results chan<- searchResults) {
	for c := range characters {
		var searchName string
		// if the name has a parentheses and it's a marvel character, we wanna search the name within the parens
		if parenIndex := strings.Index(c.Name, "("); parenIndex != -1 && c.Publisher.Slug == "marvel" {
			searchName = SearchableName(c.Name, parenIndex)
		} else {
			searchName = SearchableName(c.Name, -1)
		}
		var externalResults []externalissuesource.CharacterSearchResult
		result, err := i.retrySearchByName(searchName)
		if err != nil {
			results <- searchResults{Error: err, Character: c}
			continue
		}
		externalResults = append(externalResults, result)
		// now search by other name
		if c.OtherName != "" {
			otherNameResult, otherNameErr := i.retrySearchByName(c.OtherName)
			if err != nil {
				results <- searchResults{Error: otherNameErr, Character: c}
				continue
			}
			externalResults = append(externalResults, otherNameResult)
		}
		results <- searchResults{Character: c, SearchResults: externalResults}
	}
}

// Import with the specified character criteria, concurrently imports character sources from an external source.
// If strict is set to true, it will import sources whose name _exactly_ matches the character's name (case insensitive).
// Otherwise, it will import all sources that match the search result.
func (i *CharacterSourceImporter) Import(slugs []comic.CharacterSlug, isStrict bool) error {
	characters, err := i.characterSvc.Characters(slugs, len(slugs), 0)
	if err != nil {
		return err
	}
	characterLen := len(characters)
	characterCh := make(chan *comic.Character, characterLen)
	resultCh := make(chan searchResults, characterLen)
	defer close(characterCh)
	defer close(resultCh)
	for w := 0; w < 10; w++ {
		// Start the goroutines.
		go i.gatherSearchResults(w, characterCh, resultCh)
	}
	// Send the work over.
	for _, c := range characters {
		characterCh <- c
	}
	// Now collect the results of the work.
	for x := 0; x < characterLen; x++ {
		result := <-resultCh
		c := result.Character
		if result.Error != nil {
			i.logger.Error("got error. skipping.", zap.String("character", c.Slug.Value()), zap.Error(result.Error))
			continue
		}
		searches := result.SearchResults
		var sourceErr error
		// loop over the search results
		for _, search := range searches {
			// loop over the links of the results of the search result
			for _, link := range search.Results {
				// match the publisher
				if !stringutil.EqualsIAny(ParsePublisherName(link.Name), c.Publisher.Slug.Value(), c.Publisher.Name) {
					// continue if no match.
					continue
				}
				// if not strict, import all results that get returned
				if !isStrict {
					sourceErr = i.importSources(c, link)
				}
				// strict. match the character first.
				characterName := ParseCharacterName(link.Name)
				if strings.Contains(c.Name, characterName) || c.OtherName != "" && strings.Contains(c.OtherName, characterName) {
					sourceErr = i.importSources(c, link)
				}
			}
		}
		// Now normalize sources for the character if no error from importing sources.
		if sourceErr == nil {
			if errN := i.characterSvc.NormalizeSources(c.ID); errN != nil {
				i.logger.Error("error normalizing sources", zap.Error(errN), zap.String("character", c.Slug.Value()))
			} else {
				i.logger.Info("normalized character sources", zap.String("character", c.Slug.Value()))
			}
		}
	}
	i.logger.Info("Done!")
	return nil
}

// ParsePublisherName parses the publisher name from the given string.
func ParsePublisherName(s string) string {
	firstParen := strings.Index(s, "(")
	secondParen := strings.Index(s, ")")
	if firstParen != -1 && secondParen != -1 {
		return s[firstParen+1 : secondParen]
	}
	return ""
}

// ParseCharacterName parses the character name from the given string.
func ParseCharacterName(s string) string {
	// error here
	firstParenIdx := strings.Index(s, "(")
	if firstParenIdx != -1 && firstParenIdx > 0 {
		return s[:firstParenIdx-1]
	}
	return s
}

// NewCharacterSourceImporter returns a new instance of the importer.
func NewCharacterSourceImporter(
	httpClient *http.Client,
	container *comic.PGRepositoryContainer,
	cbExternalSource externalissuesource.ExternalSource,
) *CharacterSourceImporter {
	return &CharacterSourceImporter{
		httpClient:     httpClient,
		characterSvc:   comic.NewCharacterService(container),
		externalSource: cbExternalSource,
		logger:         log.CEREBRO(),
	}
}
