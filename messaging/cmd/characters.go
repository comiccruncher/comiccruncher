package cmd

import (
	"fmt"
	"github.com/aimeelaplant/comiccruncher/comic"
	"github.com/aimeelaplant/comiccruncher/internal/flagutil"
	"github.com/aimeelaplant/comiccruncher/internal/pgo"
	"github.com/aimeelaplant/comiccruncher/messaging"
	"github.com/spf13/cobra"
	"os"
)

var charactersCmd = &cobra.Command{
	Use:   "characters",
	Short: "The command for sending characters with issue sources to the queue.",
	Run: func(cmd *cobra.Command, args []string) {
		slugs := flagutil.Split(*cmd.Flag("slug"), ",")
		criteria := comic.CharacterCriteria{FilterSources: true}
		if len(slugs) > 0 {
			criteria.Slugs = comic.NewCharacterSlugs(slugs...)
		}
		messenger := messaging.NewJSONSqsMessengerFromEnv()
		svc := messaging.NewCharacterMessageService(messenger, comic.NewPGRepositoryContainer(pgo.MustInstance()))
		if err := svc.Send(criteria); err != nil {
			fmt.Println(fmt.Sprintf("got error from service: %s", err))
			os.Exit(-1)
		}
	},
}

func init() {
	charactersCmd.Flags().StringP("slug", "s", "", "Filter by character slugs, for example: `slug=jean-grey,scarlet-witch`. If this flag isn't set, it will send all characters with issue sources to the queue.")
	RootCmd.AddCommand(charactersCmd)
}
