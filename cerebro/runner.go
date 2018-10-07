package cerebro

import (
	"time"
	"github.com/aimeelaplant/comiccruncher/internal/listutil"
	"github.com/aimeelaplant/comiccruncher/internal/log"
	"fmt"
	"go.uber.org/zap"
	"github.com/aimeelaplant/comiccruncher/comic"
	"github.com/aimeelaplant/comiccruncher/internal/pgo"
	"net/http"
	"github.com/aimeelaplant/comiccruncher/storage"
	"github.com/aimeelaplant/comiccruncher/internal/rediscache"
	"github.com/aimeelaplant/externalissuesource"
	"github.com/aimeelaplant/comiccruncher/marvel"
	"github.com/aimeelaplant/comiccruncher/dc"
)

// Basically a container to run the different imports for the commands here. Way cleaner than instantiating them
// individually in each cmd struct.
type ImportRunner struct {
	marvelImporter          CharacterImporter
	dcImporter              CharacterImporter
	pgContainer             comic.PGRepositoryContainer
	characterIssueImporter  CharacterIssueImporter
	characterSourceImporter CharacterSourceImporter
}

// Imports DC and Marvel characters.
func (r ImportRunner) Characters(publishers []string) error {
	start := time.Now()
	if len(publishers) == 0 || listutil.StringInSlice(publishers, "dc") {
		dcErrCh := make(chan error, 1)
		go func() {
			dcErr := r.dcImporter.ImportAll()
			if dcErr != nil {
				dcErrCh <- dcErr
			} else {
				close(dcErrCh)
			}
		}()
		dcErr := <-dcErrCh
		if dcErr != nil {
			log.CEREBRO().Error(fmt.Sprintf("Error from DC importer: %s", dcErr))
		}
	}
	if len(publishers) == 0 || listutil.StringInSlice(publishers, "marvel") {
		marvelErrCh := make(chan error, 1)
		go func() {
			marvelErr := r.marvelImporter.ImportAll()
			if marvelErr != nil {
				marvelErrCh <- marvelErr
			} else {
				close(marvelErrCh)
			}
		}()
		marvelErr := <-marvelErrCh
		if marvelErr != nil {
			log.CEREBRO().Error(fmt.Sprintf("Error from Marvel importer: %s", marvelErr))
		}
	}
	log.CEREBRO().Info("Finished imports", zap.Duration("duration", time.Since(start)))
	return nil
}

// Imports character sources.
func (r ImportRunner) CharacterSources(slugs []comic.CharacterSlug, isStrict bool) error {
	return r.characterSourceImporter.Import(slugs, isStrict)
}

// Imports character issues and creates a sync log for each character that gets imported.
func (r ImportRunner) CharacterIssues(slugs []comic.CharacterSlug) error {
	return r.characterIssueImporter.ImportAll(slugs)
}

// Imports an existing character and existing sync log by their slug and sync log id.
func (r ImportRunner) CharacterIssuesWithCharacterAndLog(slug comic.CharacterSlug, id comic.CharacterSyncLogID) error {
	if character, err := r.pgContainer.CharacterRepository().FindBySlug(slug, false); err != nil {
		return err
	} else {
		if syncLog, err := r.pgContainer.CharacterSyncLogRepository().FindById(id); err != nil {
			return err
		} else {
			return r.characterIssueImporter.ImportWithSyncLog(*character, syncLog)
		}

	}
	return nil
}

// Returns a new import runner.
func NewImportRunner() (ImportRunner, error) {
	db, err := pgo.Instance()
	if err != nil {
		return ImportRunner{}, err
	}
	container := comic.NewPGRepositoryContainer(db)
	httpClient := http.DefaultClient
	s3Storage, err := storage.NewS3StorageFromEnv()
	redisRepository := comic.NewRedisAppearancesPerYearRepository(rediscache.Instance())
	appearancesSyncer := comic.NewAppearancesSyncer(container, redisRepository)
	// Use the http client provided from the external source.
	externalSource := externalissuesource.NewCbExternalSource(externalissuesource.NewHttpClient(), &externalissuesource.CbExternalSourceConfig{})
	return ImportRunner{
		marvelImporter:          NewMarvelCharactersImporter(marvel.NewMarvelAPI(httpClient), container, s3Storage),
		dcImporter:              NewDcCharactersImporter(dc.NewDcApi(httpClient), container, s3Storage),
		characterIssueImporter:  *NewCharacterIssueImporter(container, appearancesSyncer, externalSource),
		characterSourceImporter: *NewCharacterSourceImporter(httpClient, container, externalSource),
		pgContainer: *container,
	}, err
}
