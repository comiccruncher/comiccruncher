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

// ExternalCharacter represents a character from a remote source, such as the Marvel API.
type ExternalCharacter struct {
	Publisher    string
	VendorID     string // The vendor's unique identifier of the external character.
	Name         string
	Description  string
	ThumbnailURL string
	URL          string
}

// importer is the base structure for importing a remote character to a repository.
type importer struct {
	publisherSvc comic.PublisherServicer
	characterSvc comic.CharacterServicer
	storage      storage.Storage
	logger       *zap.Logger
}

// CharacterImporter is the interface for importing characters.
type CharacterImporter interface {
	ImportAll() error
}

// MarvelCharactersImporter imports characters from the Marvel API into a local repository.
type MarvelCharactersImporter struct {
	importer  *importer
	marvelAPI *marvel.API
}

// DcCharactersImporter imports characters from the DC API to the local repository.
type DcCharactersImporter struct {
	dcAPI    *dc.Api
	importer *importer
}

// CharacterImportResult represents the result of an import, with the `ExternalCharacter` being the external source
// and the `LocalCharacter` being the object that was imported from the external source.
// An error could happen, so the `LocalCharacter` returns `nil` with an `Error` set if the character
// wasn't imported.
type CharacterImportResult struct {
	ExternalCharacter *ExternalCharacter // Returns nil if the external character wasn't fetched.
	LocalCharacter    *comic.Character   // Returns nil if the character wasn't imported.
	Error             error              // Returns an error if something bad happened.
}

// createNewCharacter creates a new character and sync log for the character.
func (importer *importer) createNewCharacter(ec ExternalCharacter, publisher comic.Publisher) (*comic.Character, error) {
	vt, err := vendorType(ec)
	if err != nil {
		return nil, err
	}
	newChar := comic.NewCharacter(ec.Name, publisher.ID, vt, ec.VendorID)
	newChar.VendorDescription = ec.Description
	newChar.VendorUrl = ec.URL
	if shouldUploadImage(ec) {
		file, errU := importer.storage.UploadFromRemote(ec.ThumbnailURL, remoteImageDir)
		if errU != nil {
			return nil, errU
		}
		newChar.VendorImage = file.Pathname
		newChar.VendorImageMd5 = file.MD5Hash
	}
	if errC := importer.characterSvc.Create(newChar); errC != nil {
		return nil, errC
	}
	now := time.Now()
	if _, errI := importer.characterSvc.CreateSyncLogP(newChar.ID, comic.Success, comic.Characters, &now); errI != nil {
		importer.logger.Error("error creating sync log", zap.Uint("id", newChar.ID.Value()), zap.Error(errI))
	}
	return newChar, nil
}

// updateCharacter updates the character if any of the fields have changed and updates the sync log.
// Returns a boolean indicating if the character was updated and an error, if any.
func (importer importer) updateCharacter(ec ExternalCharacter, character *comic.Character) (bool, error) {
	isUpdated := false
	// don't update the character names!!! just check the vendor description and vendor image
	if character.VendorDescription != ec.Description {
		character.VendorDescription = ec.Description
		isUpdated = true
	}
	if character.VendorImage == "" && shouldUploadImage(ec) {
		ul, err := importer.storage.UploadFromRemote(ec.ThumbnailURL, remoteImageDir)
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

// Import imports a single character into the persistence layer.
// This method is responsible for the logic of importing a character into our persistence layer.
// It either creates or updates a Marvel or DC character.
func (importer *importer) Import(ec ExternalCharacter, publisher comic.Publisher) (*comic.Character, error) {
	vt, err := vendorType(ec)
	if err != nil {
		return nil, err
	}
	vendorID := ec.VendorID
	// Include disabled character so we don't import it again.
	character, err := importer.characterSvc.CharacterByVendor(vendorID, vt, true)
	if err != nil {
		return nil, err
	}
	// If we don't have the character
	if character == nil {
		// Create it ...
		return importer.createNewCharacter(ec, publisher)
	}
	// we have the character and it's not disabled
	if character.IsDisabled != false {
		// We do have the character, so update it.
		if _, errU := importer.updateCharacter(ec, character); errU != nil {
			return nil, errU
		}
		return character, nil
	}
	importer.logger.Info("skipped disabled character", zap.String("character", character.Name))
	return nil, nil
}

// ImportAll launches goroutines to import characters from the Marvel API.
// Returns an error if there is a system error or an error fetching from the API.
func (mci *MarvelCharactersImporter) ImportAll() error {
	limit := 100
	var wg sync.WaitGroup
	totalCharacters, err := mci.marvelAPI.TotalCharacters()
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
			resultWrapper, resultErr, errA := mci.marvelAPI.Characters(&marvel.Criteria{
				Limit:   limit,
				Offset:  offset,
				OrderBy: "name",
			})
			// ughh, this here below is so gross....
			if errA != nil {
				mci.importer.logger.Error("error getting characters from the api", zap.Error(errA))
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
				localCharacter, errI := mci.importer.Import(externalCharacter, *publisher)
				if errI != nil {
					mci.importer.logger.Error(
						"error importing external character",
						zap.String("externalCharacter", externalCharacter.Name),
						zap.Error(errI))
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

// ImportAll launches goroutines to import characters from the DC API.
// Returns an error if there is a system error or an error fetching from the API.
func (dci *DcCharactersImporter) ImportAll() error {
	var wg sync.WaitGroup
	totalCharacters, err := dci.dcAPI.TotalCharacters()
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
			result, errF := dci.dcAPI.FetchCharacters(currentPageNumber)
			if errF != nil {
				dci.importer.logger.Error("error fetching characters from DC API", zap.Error(errF))
				return
			}
			for _, dcCharacter := range result.Results {
				externalCharacter := fromDcCharacter(dcCharacter)
				localCharacter, errI := dci.importer.Import(externalCharacter, *publisher)
				if errI != nil {
					dci.importer.logger.Error("error importing external character", zap.String("character", externalCharacter.Name), zap.Error(err))
				} else if localCharacter == nil {
					dci.importer.logger.Info("did not import anything. no changes or nothing to import.")
				} else {
					dci.importer.logger.Info(
						"imported character from external character",
						zap.String("localCharacter", localCharacter.Name),
						zap.String("externalCharacter", externalCharacter.Name))
				}
			}
		}(currentPageNumber, &wg, publisher)
	}
	wg.Wait() // done goroutines
	return nil
}

// fromMarvelCharacter returns an external character object from a Marvel character.
func fromMarvelCharacter(mc *marvel.Character) ExternalCharacter {
	ec := ExternalCharacter{
		VendorID:    strconv.Itoa(mc.ID),
		Name:        mc.Name,
		Description: html.UnescapeString(policy.Sanitize(strings.TrimSpace(mc.Description))),
		Publisher:   publisherMarvel,
	}
	if mc.Thumbnail.Extension != "" && mc.Thumbnail.Path != "" {
		ec.ThumbnailURL = fmt.Sprintf("%s.%s", mc.Thumbnail.Path, mc.Thumbnail.Extension)
	}
	for _, v := range mc.Urls {
		if v.Type != "detail" {
			continue
		}
		qstnMarkIdx := strings.LastIndex(v.Url, "?")
		if qstnMarkIdx != -1 {
			ec.URL = strings.Replace(v.Url[:qstnMarkIdx], "http", "https", -1)
		} else {
			ec.URL = strings.Replace(v.Url, "http", "https", -1)
		}
	}
	return ec
}

// fromDcCharacter returns an external character object from a DC character.
func fromDcCharacter(dcCharacter *dc.CharacterResult) ExternalCharacter {
	ec := ExternalCharacter{
		VendorID:  dcCharacter.Id,
		Name:      strings.TrimSpace(dcCharacter.Fields.Name),
		Publisher: publisherDc,
		URL:       fmt.Sprintf("%s%s", dc.ApiUrl, dcCharacter.Fields.Url),
	}
	if len(dcCharacter.Fields.Body) > 0 {
		ec.Description = html.UnescapeString(policy.Sanitize(strings.TrimSpace(dcCharacter.Fields.Body[0])))
	}
	if len(dcCharacter.Fields.ProfilePicture) > 0 {
		ec.ThumbnailURL = dc.ApiUrl + dcCharacter.Fields.ProfilePicture[0]
	}
	return ec
}

// shouldUploadImage determines whether we should upload the character photo or not.
func shouldUploadImage(ec ExternalCharacter) bool {
	if ec.Publisher == publisherMarvel &&
		ec.ThumbnailURL != "" &&
		!strings.Contains(strings.ToLower(ec.ThumbnailURL), "image_not_available") {
		return true
	}
	if ec.Publisher == publisherDc && ec.ThumbnailURL != "" {
		return true
	}
	return false
}

// vendorType determines the vendor type based on the external character.
func vendorType(ec ExternalCharacter) (comic.VendorType, error) {
	if ec.Publisher == publisherMarvel {
		return comic.VendorTypeMarvel, nil
	} else if ec.Publisher == publisherDc {
		return comic.VendorTypeDC, nil
	}
	return comic.VendorType(0), fmt.Errorf("unknown publisher %s", ec.Publisher)
}

// NewMarvelCharactersImporter returns a new instance of the Marvel character importer.
func NewMarvelCharactersImporter(
	marvelAPI *marvel.API,
	container *comic.PGRepositoryContainer,
	storage storage.Storage) CharacterImporter {
	imp := &importer{
		publisherSvc: comic.NewPublisherService(container),
		characterSvc: comic.NewCharacterService(container),
		storage:      storage,
		logger:       log.MARVELIMPORTER(),
	}
	return &MarvelCharactersImporter{
		marvelAPI: marvelAPI,
		importer:  imp,
	}
}

// NewDcCharactersImporter returns a new instance of the DC character importer.
func NewDcCharactersImporter(
	dcAPI *dc.Api,
	container *comic.PGRepositoryContainer,
	storage storage.Storage) CharacterImporter {
	imp := &importer{
		publisherSvc: comic.NewPublisherService(container),
		characterSvc: comic.NewCharacterService(container),
		storage:      storage,
		logger:       log.DCIMPORTER(),
	}
	return &DcCharactersImporter{
		dcAPI:    dcAPI,
		importer: imp,
	}
}
