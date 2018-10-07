package main

import (
	"fmt"
	"github.com/aimeelaplant/comiccruncher/comic"
	"github.com/aimeelaplant/comiccruncher/internal/log"
	"github.com/aimeelaplant/comiccruncher/internal/pgo"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"go.uber.org/zap"
	"os"
)

// Logs the error as fatal and exits.
func fatalIfError(err error) {
	if err != nil {
		log.MIGRATIONS().Fatal("error", zap.Error(err))
	}
}

// Logs the error as fatal and exits.
func fatalResult(_ orm.Result, err error) {
	if err != nil {
		log.MIGRATIONS().Fatal("error", zap.Error(err))
	}
}

// Logs an error if there's an error instantiating the db.
// Or logs when there's no error.
func logFatalIfError(err error, env string) {
	if err != nil {
		log.MIGRATIONS().Fatal("error instantiating database", zap.Error(err))
	}
	log.MIGRATIONS().Info("instantiated connection", zap.String("environment", env))
}

// Generates SQL for creating an `updated_at` column for the given `tableName`.
func updatedAtTrigger(tableName string) string {
	return fmt.Sprintf(`
		CREATE OR REPLACE FUNCTION update_updated_at_column() 
		RETURNS TRIGGER AS $$
		BEGIN
    		NEW.updated_at = now();
    	RETURN NEW; 
		END;
		$$ language 'plpgsql';
		DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_%[1]s_updated_at_column') THEN
				CREATE TRIGGER update_%[1]s_updated_at_column
				BEFORE UPDATE ON %[1]s FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
			END IF;
		END;
		$$`, tableName)
}

func mustInstance() *pg.DB {
	env := os.Getenv("CC_ENVIRONMENT")
	if env == "test" {
		db, err := pgo.InstanceTest()
		logFatalIfError(err, env)
		return db
	} else if env == "development" || env == "production" {
		db, err := pgo.Instance()
		logFatalIfError(err, env)
		return db
	}
	return nil
}

func main() {
	tx := mustInstance()
	fatalIfError(tx.RunInTransaction(func(tx *pg.Tx) error {
		if os.Getenv("CC_ENVIRONMENT") != "production" {
			fatalResult(tx.Exec("CREATE EXTENSION IF NOT EXISTS pg_trgm;"))
		}
		opts := &orm.CreateTableOptions{
			IfNotExists:   true,
			FKConstraints: true,
		}
		fatalIfError(tx.CreateTable(&comic.Publisher{}, opts))
		fatalIfError(tx.CreateTable(&comic.Character{}, opts))
		fatalIfError(tx.CreateTable(&comic.CharacterSource{}, opts))
		fatalIfError(tx.CreateTable(&comic.CharacterSyncLog{}, opts))
		fatalIfError(tx.CreateTable(&comic.Issue{}, opts))
		fatalIfError(tx.CreateTable(&comic.CharacterIssue{}, opts))
		updatedAtTriggers := []string{
			updatedAtTrigger("publishers"),
			updatedAtTrigger("characters"),
			updatedAtTrigger("character_sources"),
			updatedAtTrigger("character_sync_logs"),
			updatedAtTrigger("issues"),
			updatedAtTrigger("character_issues"),
		}
		for _, t := range updatedAtTriggers {
			fatalResult(tx.Exec(t))
		}
		fatalResult(tx.Exec(`
			CREATE INDEX IF NOT EXISTS characters_publisher_id_idx ON characters(publisher_id) WHERE is_disabled = false;
			CREATE INDEX IF NOT EXISTS characters_name_odx ON characters(name) WHERE is_disabled = false;
			CREATE INDEX IF NOT EXISTS character_sources_character_id_idx ON character_sources(character_id) WHERE is_disabled = false;
			CREATE INDEX IF NOT EXISTS character_sync_logs_character_id_idx ON character_sync_logs(character_id);
			CREATE INDEX IF NOT EXISTS characters_name_idx_gin on characters USING GIN(name gin_trgm_ops) WHERE is_disabled = false;
			CREATE INDEX IF NOT EXISTS characters_other_name_idx_gin ON characters USING GIN(other_name gin_trgm_ops) WHERE is_disabled = false AND (other_name IS NOT NULL AND other_name != '');
		`))
		fatalResult(tx.Exec(`
			ALTER TABLE IF EXISTS character_sources
  			ADD COLUMN IF NOT EXISTS vendor_other_name text NULL
		`))
		if os.Getenv("CC_ENVIRONMENT") != "test" {
			fatalResult(tx.Exec("INSERT INTO publishers (name, slug, created_at, updated_at) VALUES (?, ?, now(), now()) ON CONFLICT DO NOTHING;", "Marvel", "marvel"))
			fatalResult(tx.Exec("INSERT INTO publishers (name, slug, created_at, updated_at) VALUES (?, ?, now(), now()) ON CONFLICT DO NOTHING;", "DC Comics", "dc"))
		}
		return nil
	}))
	log.MIGRATIONS().Info("done")
}
