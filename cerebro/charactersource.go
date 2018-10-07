package cerebro

import (
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

// The search result from an external source with the local character and an error, if any.
type characterSearchResult struct {
	Character     *comic.Character
	SearchResult  externalissuesource.CharacterSearchResult
	SearchResults []externalissuesource.CharacterSearchResult
	Error         error
}

// Creates a source from a character and external link.
func (i *CharacterSourceImporter) createIfNotExists(character *comic.Character, link externalissuesource.CharacterLink) (*comic.CharacterSource, error) {
	source := comic.NewCharacterSource(link.Url, link.Name, character.ID, comic.VendorTypeCb)
	var otherName string
	// do another request and get the real name.
	retryConnectionError(func() (string, error) {
		cPage, err := i.externalSource.CharacterPage(link.Url)
		if err != nil {
			return link.Url, err
		}
		otherName = cPage.OtherName
		return link.Url, nil
	})
	source.VendorOtherName = otherName
	if s1, err := i.characterSvc.CreateSourceIfNotExists(source); err != nil {
		if err == comic.ErrAlreadyExists {
			updateMsg := "skipping source. already have it."
			// Update the vendor other name
			if s1.VendorOtherName != otherName {
				s1.VendorOtherName = otherName
				err := i.characterSvc.UpdateSource(s1)
				if err != nil {
					return nil, err
				}
				updateMsg = "updated character source's other name"
			}
			i.logger.Info(
				updateMsg,
				zap.String("vendor url", link.Url),
				zap.String("vendor name", source.VendorName),
				zap.String("character", character.Slug.Value()))
			return nil, nil

		}
		return nil, err
	}
	i.logger.Info(
		"created source",
		zap.String("vendor url", link.Url),
		zap.String("vendor name", source.VendorName),
		zap.String("character", character.Slug.Value()))
	return source, nil
}

// Creates a source if the link doesn't already exist in the character sources.
// Returns nil if nothing was created and there were no errors.
// Returns a `CharacterSourceImportItem` if a source was created or there was an error.
func (i *CharacterSourceImporter) importSources(character *comic.Character, link externalissuesource.CharacterLink) error {
	if _, err := i.createIfNotExists(character, link); err != nil {
		return err
	}
	// Now import other identities.
	pageChan := make(chan *externalissuesource.CharacterPage, 1)
	defer close(pageChan)
	err := retryConnectionError(func() (string, error) {
		i.logger.Info("requesting page...", zap.String("link", link.Url))
		cPage, err := i.externalSource.CharacterPage(link.Url)
		if err != nil {
			return link.Url, err
		}
		// send it over
		pageChan <- cPage
		return link.Url, nil
	})
	if err != nil {
		i.logger.Warn("error fetching url. quitting retry.", zap.String("link.Url", link.Url), zap.Error(err))
		return err
	}
	// Read from it
	page, ok := <-pageChan
	if !ok {
		return nil
	}
	for _, link := range page.OtherIdentities {
		if _, err := i.createIfNotExists(character, link); err != nil {
			i.logger.Error("error creating character.", zap.String("link.Url", link.Url), zap.Error(err))
		}
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
		case err := <-resultErrCh:
			return externalissuesource.CharacterSearchResult{}, err
		case result := <-resultCh:
			return result, nil
		}
	}
}

// Performs a search on a character received from the `characters` chan and sends the search result over to the `results` chan.
// The caller of the method is responsible for closing the channels.
func (i *CharacterSourceImporter) searchCharacter(workerID int, characters <-chan *comic.Character, results chan<- characterSearchResult) error {
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
			results <- characterSearchResult{Error: err, Character: c}
			continue
		}
		externalResults = append(externalResults, result)
		// now search by other name
		if c.OtherName != "" {
			otherNameResult, otherNameErr := i.retrySearchByName(c.OtherName)
			if err != nil {
				results <- characterSearchResult{Error: otherNameErr, Character: c}
				continue
			}
			externalResults = append(externalResults, otherNameResult)
		}
		results <- characterSearchResult{Character: c, SearchResults: externalResults}
	}
	return nil
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
	resultCh := make(chan characterSearchResult, characterLen)
	defer close(characterCh)
	defer close(resultCh)
	for w := 0; w < 10; w++ {
		// Start the goroutines.
		go i.searchCharacter(w, characterCh, resultCh)
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
		searchResults := result.SearchResults
		// loop over the search results
		for _, searchResult := range searchResults {
			// loop over the links of the results of the search result
			for _, link := range searchResult.Results {
				publisherName := ParsePublisherName(link.Name)
				// if not strict, import all results that matches the character's publisher
				if !isStrict && stringutil.EqualsIAny(publisherName, c.Publisher.Name, c.Publisher.Slug.Value()) {
					if err = i.importSources(c, link); err != nil {
						return err
					}
					// go to next item in loop
					continue
				}
				// non-strict search.
				characterName := ParseCharacterName(link.Name)
				// oh god ...
				if (c.Publisher.Slug == "marvel" &&
					stringutil.EqualsIAny(publisherName, c.Publisher.Name) &&
					strings.Index(c.Name, "(") != -1) ||
					(stringutil.EqualsIAny(publisherName, c.Publisher.Name, c.Publisher.Slug.Value()) &&
						stringutil.EqualsIAny(characterName, c.Name)) {
					if err = i.importSources(c, link); err != nil {
						return err
					}
				}
			}
		}
		// Now normalize sources for the character
		if err := i.characterSvc.NormalizeSources(c.ID); err != nil {
			i.logger.Error("error normalizing sources", zap.Error(err))
		} else {
			i.logger.Info("normalized character sources", zap.String("character", c.Slug.Value()))
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
