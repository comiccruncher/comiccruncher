package cerebro

import (
	"errors"
	"fmt"
	"github.com/aimeelaplant/comiccruncher/comic"
	"github.com/aimeelaplant/comiccruncher/dc"
	"github.com/aimeelaplant/comiccruncher/internal/log"
	"github.com/aimeelaplant/comiccruncher/marvel"
	"github.com/aimeelaplant/comiccruncher/storage"
	"github.com/microcosm-cc/bluemonday"
	"go.uber.org/zap"
	"html"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	publisherMarvel = "Marvel"
	publisherDc     = "DC"
	remoteImageDir  = "images/characters"
)

var policy = bluemonday.UGCPolicy()

// Represents a character from a remote source, such as the Marvel API.
type ExternalCharacter struct {
	Publisher    string
	VendorId     string // The vendor's unique identifier of the external character.
	Name         string
	Description  string
	ThumbnailUrl string
	Url          string
}

// The base structure for importing a remote character to a repository.
type importer struct {
	publisherSvc        comic.PublisherServicer
	characterSvc        comic.CharacterServicer
	storage             storage.Storage
	logger              *zap.Logger
}

// The interface for importing characters.
type CharacterImporter interface {
	ImportAll() error
}

// Imports characters from the Marvel API into a local repository.
type MarvelCharactersImporter struct {
	importer  *importer
	marvelApi *marvel.API
}

// Imports characters from the DC API to the local repository.
type DcCharactersImporter struct {
	dcApi    *dc.Api
	importer *importer
}

// Represents the result of an import, with the `ExternalCharacter` being the external source
// and the `LocalCharacter` being the object that was imported from the external source.
// An error could happen, so the `LocalCharacter` returns `nil` with an `Error` set if the character
// wasn't imported.
type CharacterImportResult struct {
	ExternalCharacter *ExternalCharacter // Returns nil if the external character wasn't fetched.
	LocalCharacter    *comic.Character   // Returns nil if the character wasn't imported.
	Error             error              // Returns an error if something bad happened.
}

// Creates a new character and sync log for the character.
func (importer *importer) createNewCharacter(ec ExternalCharacter, publisher comic.Publisher) (*comic.Character, error) {
	vendorType, err := vendorType(ec)
	if err != nil {
		return nil, err
	}
	newChar := comic.NewCharacter(ec.Name, publisher.ID, vendorType, ec.VendorId)
	newChar.VendorDescription = ec.Description
	newChar.VendorUrl = ec.Url
	if shouldUploadImage(ec) {
		file, err := importer.storage.UploadFromRemote(ec.ThumbnailUrl, remoteImageDir)
		if err != nil {
			return nil, err
		}
		newChar.VendorImage = file.Pathname
		newChar.VendorImageMd5 = file.MD5Hash
	}
	if err := importer.characterSvc.Create(newChar); err != nil {
		return nil, err
	}
	now := time.Now()
	if _, err := importer.characterSvc.CreateSyncLogP(newChar.ID, comic.Success, comic.Characters, &now); err != nil {
		importer.logger.Error("error creating sync log", zap.Uint("id", newChar.ID.Value()), zap.Error(err))
	}
	return newChar, nil
}

// Updates the character if any of the fields have changed and updates the sync log.
// Returns a boolean indicating if the character was updated and an error, if any.
func (importer importer) updateCharacter(ec ExternalCharacter, character *comic.Character) (bool, error) {
	isUpdated := false
	// don't update the character names!!! just check the vendor description and vendor image
	if character.VendorDescription != ec.Description {
		character.VendorDescription = ec.Description
		isUpdated = true
	}
	if character.VendorImage == "" && shouldUploadImage(ec) {
		ul, err := importer.storage.UploadFromRemote(ec.ThumbnailUrl, remoteImageDir)
		if err != nil {
			return false, err
		}
		character.VendorImageMd5 = ul.MD5Hash
		character.VendorImage = ul.Pathname
		isUpdated = true
	}
	if !isUpdated {
		return isUpdated, nil
	}
	err := importer.characterSvc.Update(character)
	if err != nil {
		importer.logger.Error("Error updating character", zap.String("character", ec.Name), zap.Error(err))
		return false, err
	}
	now := time.Now()
	_, err = importer.characterSvc.CreateSyncLogP(character.ID, comic.Success, comic.Characters, &now)
	if err != nil {
		importer.logger.Error("Error creating character sync log for character", zap.Uint("id", character.ID.Value()), zap.Error(err))
	}
	return isUpdated, nil
}

// Import a single character into the persistence layer.
// This method is responsible for the logic of importing a character into our persistence layer.
// It either creates or updates a Marvel or DC character.
func (importer *importer) Import(ec ExternalCharacter, publisher comic.Publisher) (*comic.Character, error) {
	vendorType, err := vendorType(ec)
	if err != nil {
		return nil, err
	}
	vendorID := ec.VendorId
	// Include disabled character so we don't import it again.
	character, err := importer.characterSvc.CharacterByVendor(vendorID, vendorType, true)
	if err != nil {
		return nil, err
	}
	// If we don't have the character
	if character == nil {
		// Create it ...
		return  importer.createNewCharacter(ec, publisher)
	} else if character.IsDisabled == false {
		// We do have the character, so update it.
		if _, err := importer.updateCharacter(ec, character); err != nil {
			return nil, err
		} else {
			return character, nil
		}
	} else {
		importer.logger.Info("skipped disabled character", zap.String("character", character.Name))
	}
	return nil, nil
}



// Launches goroutines to import characters from the Marvel API.
// Returns an error if there is a system error or an error fetching from the API.
func (mci *MarvelCharactersImporter) ImportAll() error {
	limit := 100
	var wg sync.WaitGroup
	totalCharacters, err := mci.marvelApi.TotalCharacters()
	if err != nil {
		return err
	}
	publisher, err := mci.importer.publisherSvc.Publisher("marvel")
	if err != nil {
		return err
	}
	// This should never happen, but be safe!
	if publisher == nil {
		return errors.New("no marvel publisher to associate a character")
	}
	for offset := 0; offset < totalCharacters; offset += limit {
		wg.Add(1)
		go func(offset, limit int, wg *sync.WaitGroup, publisher *comic.Publisher) {
			defer wg.Done()
			resultWrapper, resultErr, err := mci.marvelApi.Characters(&marvel.Criteria{
				Limit:   limit,
				Offset:  offset,
				OrderBy: "name",
			})
			if err != nil {
				mci.importer.logger.Error("error getting characters from the api", zap.Error(err))
				return
			} else if resultErr != nil {
				mci.importer.logger.Error(
					"Error returned from the Marvel API.",
					zap.String("code", resultErr.Code),
					zap.String("message", resultErr.Message))
				return
			} else if resultWrapper.Code != 200 {
				mci.importer.logger.Error(
					"Unexpected status returned from API.",
					zap.Int("code", resultWrapper.Code),
					zap.String("status", resultWrapper.Status))
				return
			}
			for j := 0; j < len(resultWrapper.Results); j++ {
				marvelCharacter := resultWrapper.CharactersResultContainer.Results[j]
				if marvelCharacter.Comics.Available < 25 {
					// This character probably isn't important enough, so don't import.
					mci.importer.logger.Info("skipping character. not enough comics.", zap.String("character", marvelCharacter.Name))
					continue
				}
				externalCharacter := fromMarvelCharacter(marvelCharacter)
				localCharacter, err := mci.importer.Import(externalCharacter, *publisher)
				if err != nil {
					mci.importer.logger.Error(
						"error importing external character",
						zap.String("externalCharacter", externalCharacter.Name),
						zap.Error(err))
				} else if localCharacter != nil {
					mci.importer.logger.Info(
						"imported local character from external character",
						zap.String("localCharacter", localCharacter.Name),
						zap.String("externalCharacter", externalCharacter.Name))
				} else {
					mci.importer.logger.Info(
						"did not import anything. nothing to import or no changes to make.",
						zap.String("externalCharacter", externalCharacter.Name))
				}
			}
		}(offset, limit, &wg, publisher)
	}
	wg.Wait() // done goroutines.
	return nil
}

// Launches goroutines to import characters from the DC API.
// Returns an error if there is a system error or an error fetching from the API.
func (dci *DcCharactersImporter) ImportAll() error {
	var wg sync.WaitGroup
	totalCharacters, err := dci.dcApi.TotalCharacters()
	if err != nil {
		return err
	}
	totalPages := totalCharacters / 25 // they only return 25 characters per page.
	publisher, err := dci.importer.publisherSvc.Publisher("dc")
	if err != nil {
		return err
	}
	// This, again, should never happen, but be safe!
	if publisher == nil {
		return errors.New("no dc publisher to associate a character")
	}
	for currentPageNumber := 1; currentPageNumber <= totalPages; currentPageNumber++ {
		wg.Add(1)
		go func(currentPageNumber int, wg *sync.WaitGroup, publisher *comic.Publisher) {
			defer wg.Done()
			result, err := dci.dcApi.FetchCharacters(currentPageNumber)
			if err != nil {
				dci.importer.logger.Error("error fetching characters from DC API", zap.Error(err))
				return
			}
			for _, dcCharacter := range result.Results {
				externalCharacter := fromDcCharacter(dcCharacter)
				localCharacter, err := dci.importer.Import(externalCharacter, *publisher)
				if err != nil {
					dci.importer.logger.Error("error importing external character", zap.String("character", externalCharacter.Name), zap.Error(err))
				} else if localCharacter != nil {
					dci.importer.logger.Info(
						"imported character from external character",
						zap.String("localCharacter", localCharacter.Name),
						zap.String("externalCharacter", externalCharacter.Name))
				} else {
					dci.importer.logger.Info("did not import anything. no changes or nothing to import.")
				}
			}
		}(currentPageNumber, &wg, publisher)
	}
	wg.Wait() // done goroutines
	return nil
}

// Returns an external character object from a Marvel character.
func fromMarvelCharacter(marvelCharacter *marvel.Character) ExternalCharacter {
	ec := ExternalCharacter{
		VendorId:    strconv.Itoa(marvelCharacter.ID),
		Name:        marvelCharacter.Name,
		Description: html.UnescapeString(policy.Sanitize(strings.TrimSpace(marvelCharacter.Description))),
		Publisher:   publisherMarvel,
	}
	if marvelCharacter.Thumbnail.Extension != "" && marvelCharacter.Thumbnail.Path != "" {
		ec.ThumbnailUrl = fmt.Sprintf("%s.%s", marvelCharacter.Thumbnail.Path, marvelCharacter.Thumbnail.Extension)
	}
	for _, v := range marvelCharacter.Urls {
		if v.Type != "detail" {
			continue
		}
		questionMarkIndex := strings.LastIndex(v.Url, "?")
		if questionMarkIndex != -1 {
			ec.Url = strings.Replace(v.Url[:questionMarkIndex], "http", "https", -1)
		} else {
			ec.Url = strings.Replace(v.Url, "http", "https", -1)
		}
	}
	return ec
}

// Returns an external character object from a DC character.
func fromDcCharacter(dcCharacter *dc.CharacterResult) ExternalCharacter {
	ec := ExternalCharacter{
		VendorId:  dcCharacter.Id,
		Name:      strings.TrimSpace(dcCharacter.Fields.Name),
		Publisher: publisherDc,
		Url:       fmt.Sprintf("%s%s", dc.ApiUrl, dcCharacter.Fields.Url),
	}
	if len(dcCharacter.Fields.Body) > 0 {
		ec.Description = html.UnescapeString(policy.Sanitize(strings.TrimSpace(dcCharacter.Fields.Body[0])))
	}
	if len(dcCharacter.Fields.ProfilePicture) > 0 {
		ec.ThumbnailUrl = dc.ApiUrl + dcCharacter.Fields.ProfilePicture[0]
	}
	return ec
}

// Determines whether we should upload the character photo or not.
func shouldUploadImage(ec ExternalCharacter) bool {
	if ec.Publisher == publisherMarvel && ec.ThumbnailUrl != "" &&
		!strings.Contains(strings.ToLower(ec.ThumbnailUrl), "image_not_available") {
		return true
	}
	if ec.Publisher == publisherDc && ec.ThumbnailUrl != "" {
		return true
	}
	return false
}

// Determines the vendor type based on the external character.
func vendorType(ec ExternalCharacter) (comic.VendorType, error) {
	var vendorType comic.VendorType
	if ec.Publisher == publisherMarvel {
		vendorType = comic.VendorTypeMarvel
	} else if ec.Publisher == publisherDc {
		vendorType = comic.VendorTypeDC
	} else {
		return vendorType, errors.New(fmt.Sprintf("unknown publisher %s", ec.Publisher))
	}
	return vendorType, nil
}

// Returns a new instance of the Marvel character importer.
func NewMarvelCharactersImporter(
	marvelApi *marvel.API,
	container *comic.PGRepositoryContainer,
	storage storage.Storage) CharacterImporter {
	importer := &importer{
		publisherSvc: comic.NewPublisherService(container),
		characterSvc: comic.NewCharacterService(container),
		storage:             storage,
		logger:              log.MARVELIMPORTER(),
	}
	return &MarvelCharactersImporter{
		marvelApi: marvelApi,
		importer:  importer,
	}
}

// Returns a new instance of the DC character importer.
func NewDcCharactersImporter(
	dcApi *dc.Api,
	container *comic.PGRepositoryContainer,
	storage storage.Storage) CharacterImporter {
	importer := &importer{
		publisherSvc: comic.NewPublisherService(container),
		characterSvc: comic.NewCharacterService(container),
		storage:             storage,
		logger:              log.DCIMPORTER(),
	}
	return &DcCharactersImporter{
		dcApi:    dcApi,
		importer: importer,
	}
}
