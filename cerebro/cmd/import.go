package cmd

import (
	"github.com/aimeelaplant/comiccruncher/cerebro"
	"github.com/aimeelaplant/comiccruncher/comic"
	"github.com/aimeelaplant/comiccruncher/internal/flagutil"
	"github.com/aimeelaplant/comiccruncher/internal/log"
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
		if importRunner, err := cerebro.NewImportRunner(); err != nil {
			log.CEREBRO().Fatal("could not instantiate import runner", zap.Error(err))
		} else {
			publishers := flagutil.Split(*cmd.Flag("publisher"), ",")
			if err := importRunner.Characters(publishers); err != nil {
				log.CEREBRO().Fatal("could not import characters", zap.Error(err))
			}
		}
	},
}

// The command for importing character sources.
var importCharacterSourcesCmd = &cobra.Command{
	Use:   "charactersources",
	Short: "Import character sources from an external source.",
	Run: func(cmd *cobra.Command, args []string) {
		if importRunner, err := cerebro.NewImportRunner(); err != nil {
			log.CEREBRO().Fatal("could not instantiate import runner", zap.Error(err))
		} else {
			slugs := flagutil.Split(*cmd.Flag("character.slug"), ",")
			strict := cmd.Flag("strict")
			var isStrict = true
			if strict != nil && strict.Value.String() == "false" {
				isStrict = false
			}
			if err := importRunner.CharacterSources(comic.NewCharacterSlugs(slugs...), isStrict); err != nil {
				log.CEREBRO().Fatal("could not import character sources", zap.Error(err))
			}
		}
	},
}

// The command for importing character issues.
var importCharacterIssuesCmd = &cobra.Command{
	Use:   "characterissues",
	Short: "Imports character issues from an external source.",
	Run: func(cmd *cobra.Command, args []string) {
		if importRunner, err := cerebro.NewImportRunner(); err != nil {
			log.CEREBRO().Fatal("could not instantiate import runner", zap.Error(err))
		} else {
			slugs := flagutil.Split(*cmd.Flag("character.slug"), ",")
			var reset bool
			doReset := cmd.Flag("reset")
			if doReset != nil && doReset.Value.String() == "true" {
				reset = true
			}
			if err = importRunner.CharacterIssues(comic.NewCharacterSlugs(slugs...), reset); err != nil {
				log.CEREBRO().Error("could not import character issues", zap.Error(err))
			}
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
