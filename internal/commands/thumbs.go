package commands

import (
	"strings"
	"time"

	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/pkg/sanitize"
)

// ThumbsCommand registers the resample cli command.
var ThumbsCommand = cli.Command{
	Name:      "thumbs",
	Usage:     "Generates thumbnails using the current settings",
	ArgsUsage: "[ORIGINALS SUB-FOLDER]",
	Flags: []cli.Flag{
		cli.StringSliceFlag{
			Name:  "ext, e",
			Usage: "only process files with the specified extensions, e.g. mp4",
		},
		cli.BoolFlag{
			Name:  "force, f",
			Usage: "replace existing thumbnails",
		},
	},
	Action: thumbsAction,
}

// thumbsAction pre-renders thumbnail images.
func thumbsAction(ctx *cli.Context) error {
	start := time.Now()

	conf := config.NewConfig(ctx)
	service.SetConfig(conf)

	if err := conf.Init(); err != nil {
		return err
	}

	subPath := strings.TrimSpace(ctx.Args().First())

	log.Infof("creating thumbnails for %s in %s", sanitize.Log(subPath), sanitize.Log(conf.ThumbPath()))

	rs := service.Resample()

	if err := rs.Start(subPath, ctx.StringSlice("ext"), ctx.Bool("force")); err != nil {
		log.Error(err)
		return err
	}

	log.Infof("thumbnails created in %s", time.Since(start))

	return nil
}
