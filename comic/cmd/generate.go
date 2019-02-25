package cmd

import (
	"github.com/aimeelaplant/comiccruncher/comic"
	"github.com/aimeelaplant/comiccruncher/imaging"
	"github.com/aimeelaplant/comiccruncher/internal/flagutil"
	"github.com/aimeelaplant/comiccruncher/internal/log"
	"github.com/aimeelaplant/comiccruncher/internal/pgo"
	"github.com/aimeelaplant/comiccruncher/internal/rediscache"
	"github.com/aimeelaplant/comiccruncher/storage"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "The command for generating stuff",
}

// The import command.
var generateThumbsCmd = &cobra.Command{
	Use:   "thumbs",
	Short: "The command to generate thumbnails for characters.",
	Run: func(cmd *cobra.Command, args []string) {
		s := flagutil.Split(*cmd.Flag("character.slug"), ",")
		slugs := comic.NewCharacterSlugs(s...)
		svc := comic.NewCharacterServiceFactory(pgo.MustInstance())
		strg, err := storage.NewS3StorageFromEnv()
		if err != nil {
			log.COMIC().Fatal("error getting storage from env", zap.Error(err))
		}
		thmbr := comic.NewCharacterThumbnailService(rediscache.Instance(), imaging.NewS3ThumbnailUploader(strg, imaging.NewInMemoryThumbnailer()))
		characters, err := svc.CharactersWithSources(slugs, 0, 0)
		if err != nil {
			log.COMIC().Fatal("error getting characters", zap.Error(err))
		}
		for _, c := range characters {
			_, err = thmbr.Upload(c)
			slug := c.Slug.Value()
			if err != nil {
				log.COMIC().Error("error uploading thumbnails for character", zap.String("character", slug), zap.Error(err))
			} else {
				log.COMIC().Info("done uploading thumbnails for character", zap.String("character", slug))
			}
		}
		log.COMIC().Info("done generating thumbnails")
	},
}

func init() {
	generateThumbsCmd.Flags().StringP("character.slug", "s", "", "Filter by characters slugs to generate thumbs. For example: `character.slug=jean-grey,scarlet-witch`")
	generateCmd.AddCommand(generateThumbsCmd)
	RootCmd.AddCommand(generateCmd)
}
