package cmd

import (
	"github.com/aimeelaplant/comiccruncher/cerebro"
	"github.com/aimeelaplant/comiccruncher/comic"
	"github.com/aimeelaplant/comiccruncher/internal/flagutil"
	"github.com/aimeelaplant/comiccruncher/internal/listutil"
	"github.com/aimeelaplant/comiccruncher/internal/log"
	"github.com/aimeelaplant/comiccruncher/internal/pgo"
	"github.com/aimeelaplant/comiccruncher/internal/rediscache"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// The import command.
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "The command to import resources from an external source.",
}

// The command for importing characters.
var importCharactersCmd = &cobra.Command{
	Use:   "characters",
	Short: "Import characters from an external source.",
	Run: func(cmd *cobra.Command, args []string) {
		publishers := flagutil.Split(*cmd.Flag("publisher"), ",")
		db := pgo.MustInstance()
		if len(publishers) == 0 || listutil.StringInSlice(publishers, "marvel") {
			mi := cerebro.NewMarvelCharactersImporter(db)
			err := mi.ImportAll()
			if err != nil {
				log.WEB().Fatal("error importing characters from marvel", zap.Error(err))
			}
		}
		if len(publishers) == 0 || listutil.StringInSlice(publishers, "dc") {
			dcImporter := cerebro.NewDCCharactersImporter(db)
			err := dcImporter.ImportAll()
			if err != nil {
				if err != nil {
					log.WEB().Fatal("error importing characters from dc", zap.Error(err))
				}
			}
		}
	},
}

// The command for importing character sources.
var importCharacterSourcesCmd = &cobra.Command{
	Use:   "charactersources",
	Short: "Import character sources from an external source.",
	Run: func(cmd *cobra.Command, args []string) {
		db := pgo.MustInstance()
		cs := cerebro.NewCharacterSourceImporter(db)
		slugs := flagutil.Split(*cmd.Flag("character.slug"), ",")
		strict := cmd.Flag("strict")
		var isStrict = true
		if strict != nil && strict.Value.String() == "false" {
			isStrict = false
		}
		if err := cs.Import(comic.NewCharacterSlugs(slugs...), isStrict); err != nil {
			log.CEREBRO().Fatal("could not import character sources", zap.Error(err))
		}
	},
}

// The command for importing character issues.
var importCharacterIssuesCmd = &cobra.Command{
	Use:   "characterissues",
	Short: "Imports character issues from an external source.",
	Run: func(cmd *cobra.Command, args []string) {
		db := pgo.MustInstance()
		redis := rediscache.Instance()
		ci := cerebro.NewCharacterIssueImporter(db, redis)
		slugs := flagutil.Split(*cmd.Flag("character.slug"), ",")
		var reset bool
		doReset := cmd.Flag("reset")
		if doReset != nil && doReset.Value.String() == "true" {
			reset = true
		}
		if err := ci.ImportAll(comic.NewCharacterSlugs(slugs...), reset); err != nil {
			log.CEREBRO().Error("could not import character issues", zap.Error(err))
		}
	},
}

// Init scripts.
func init() {
	importCharacterIssuesCmd.Flags().StringP("character.slug", "s", "", "Filter by characters slugs to import only those, for example: `character.slug=jean-grey,scarlet-witch`")
	importCharacterIssuesCmd.Flags().Bool("reset", false, "Reset all the associated issues for the specified characters, including the character issues stored in Postgres and the Redis appearances. Defaults to false.")
	importCharacterSourcesCmd.Flags().StringP("character.slug", "s", "", "Filter by characters slugs to import only those, for example: `character.slug=jean-grey,scarlet-witch`")
	// Default is true for strict mode.
	importCharacterSourcesCmd.Flags().Bool("strict", true, "If true, import sources whose name _exactly_ matches the character's name (case insensitive). Otherwise, it will import all sources that match the search result. Default is true.")
	importCharactersCmd.Flags().StringP("publisher", "p", "", "Filter by a publisher to import characters, for example: `--publisher=dc,marvel`")
	importCmd.AddCommand(importCharactersCmd, importCharacterSourcesCmd, importCharacterIssuesCmd)
	RootCmd.AddCommand(importCmd)
}
