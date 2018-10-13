package cmd

import (
	"github.com/aimeelaplant/comiccruncher/cerebro"
	"github.com/aimeelaplant/comiccruncher/comic"
	"github.com/aimeelaplant/comiccruncher/internal/log"
	"github.com/aimeelaplant/comiccruncher/messaging"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// The start command.
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "The command for starting a long-running process.",
}

// The start command for syncing character issues from a message queue.
var startCharacterIssuesCmd = &cobra.Command{
	Use:   "characterissues",
	Short: "The command for starting the long-running process of consuming messages to sync character issues.",
	Run: func(cmd *cobra.Command, args []string) {
		if importer, err := cerebro.NewImportRunner(); err != nil {
			log.CEREBRO().Fatal("error instantiating import runner", zap.Error(err))
		} else {
			var handler messaging.SyncMessageFunc
			handler = func(message *messaging.SyncMessage) {
				if err := importer.CharacterIssuesWithCharacterAndLog(
					comic.CharacterSlug(message.CharacterSlug), comic.CharacterSyncLogID(message.CharacterSyncLogID)); err != nil {
					log.CEREBRO().Fatal("error importing character issues",
						zap.Error(err),
						zap.String("character", message.CharacterSlug),
						zap.Uint("sync log", message.CharacterSyncLogID))
				}
			}
			consumer := messaging.NewSyncMessageConsumerFromEnv(10, 10, handler)
			if err := consumer.Consume(true); err != nil {
				log.CEREBRO().Fatal("error consuming message", zap.Error(err))
			}
		}
	},
}

// Init scripts.
func init() {
	startCmd.AddCommand(startCharacterIssuesCmd)
	RootCmd.AddCommand(startCmd)
}
