package cerebro

import (
	"fmt"
	"github.com/aimeelaplant/comiccruncher/comic"
	"github.com/aimeelaplant/comiccruncher/dc"
	"github.com/aimeelaplant/comiccruncher/internal/listutil"
	"github.com/aimeelaplant/comiccruncher/internal/log"
	"github.com/aimeelaplant/comiccruncher/internal/pgo"
	"github.com/aimeelaplant/comiccruncher/internal/rediscache"
	"github.com/aimeelaplant/comiccruncher/marvel"
	"github.com/aimeelaplant/comiccruncher/storage"
	"github.com/aimeelaplant/externalissuesource"
	"go.uber.org/zap"
	"net/http"
	"time"
)

// ImportRunner is basically a container to run the different imports for the commands here. Way cleaner than instantiating them
// individually in each cmd struct.
type ImportRunner struct {
	marvelImporter          CharacterImporter
	dcImporter              CharacterImporter
	pgContainer             comic.PGRepositoryContainer
	characterIssueImporter  CharacterIssueImporter
	characterSourceImporter CharacterSourceImporter
}

// Characters imports DC and Marvel characters.
func (r *ImportRunner) Characters(publishers []string) error {
	start := time.Now()
	if len(publishers) == 0 || listutil.StringInSlice(publishers, "dc") {
		dcErr := r.dcImporter.ImportAll()
		if dcErr != nil {
			log.CEREBRO().Error(fmt.Sprintf("Error from DC importer: %s", dcErr))
		}
	}
	if len(publishers) == 0 || listutil.StringInSlice(publishers, "marvel") {
		marvelErr := r.marvelImporter.ImportAll()
		if marvelErr != nil {
			log.CEREBRO().Error(fmt.Sprintf("Error from Marvel importer: %s", marvelErr))
		}
	}
	log.CEREBRO().Info("Finished imports", zap.Duration("duration", time.Since(start)))
	return nil
}

// CharacterSources imports character sources.
func (r *ImportRunner) CharacterSources(slugs []comic.CharacterSlug, isStrict bool) error {
	return r.characterSourceImporter.Import(slugs, isStrict)
}

// CharacterIssues imports character issues and creates a sync log for each character that gets imported.
func (r *ImportRunner) CharacterIssues(slugs []comic.CharacterSlug, doReset bool) error {
	return r.characterIssueImporter.ImportAll(slugs, doReset)
}

// NewImportRunner returns a new import runner.
func NewImportRunner() (*ImportRunner, error) {
	db, err := pgo.Instance()
	if err != nil {
		return nil, err
	}
	container := comic.NewPGRepositoryContainer(db)
	httpClient := http.DefaultClient
	s3Storage, err := storage.NewS3StorageFromEnv()
	r := rediscache.Instance()
	redisRepository := comic.NewRedisAppearancesPerYearRepository(r)
	appearancesSyncer := comic.NewAppearancesSyncer(container, redisRepository)
	// Use the http client provided from the external source.
	externalSource := externalissuesource.NewCbExternalSource(externalissuesource.NewHttpClient(), &externalissuesource.CbExternalSourceConfig{})
	statsSyncer := comic.NewCharacterStatsSyncer(r, container.CharacterRepository(), comic.NewPGPopularRepository(db, comic.NewRedisCharacterThumbRepository(r)))
	return &ImportRunner{
		marvelImporter:          NewMarvelCharactersImporter(marvel.NewMarvelAPI(httpClient), container, s3Storage),
		dcImporter:              NewDcCharactersImporter(dc.NewDcAPI(httpClient), container, s3Storage),
		characterIssueImporter:  *NewCharacterIssueImporter(container, appearancesSyncer, externalSource, statsSyncer, redisRepository),
		characterSourceImporter: *NewCharacterSourceImporter(container, externalSource),
		pgContainer:             *container,
	}, err
}
