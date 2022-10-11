package commands

import (
	"context"
	"time"

	"github.com/dustin/go-humanize/english"

	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/fs"
)

// LabelsCommand registers the index cli command.
var LabelsCommand = cli.Command{
	Name:      "labels",
	Usage:     "Run image classify original media files",
	Subcommands: []cli.Command{
		{
			Name:   "index",
			Usage:  "run label action",
			Action: labelIndexAction,
		},
		{
			Name:   "reset",
			Usage:  "reset all labels",
			Action: labelResetAction,
		},
	},
}


// indexIndexAction indexes all photos in originals directory (photo library)
func labelIndexAction(ctx *cli.Context) error {
	start := time.Now()

	conf := config.NewConfig(ctx)
	service.SetConfig(conf)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := conf.Init(); err != nil {
		return err
	}

	conf.InitDb()


	if conf.ReadOnly() {
		log.Infof("config: read-only mode enabled")
	}

	var indexed fs.Done

	subPath := ""

	if w := service.Index(); w != nil {
		opt := photoprism.IndexOptions{
			Path:    subPath,
			Rescan:  true,
			Convert: conf.Settings().Index.Convert && conf.SidecarWritable(),
			Stack:   true,
			LabelsOnly: true,
		}

		indexed = w.Start(opt)
	}

	if w := service.Purge(); w != nil {
		purgeStart := time.Now()
		opt := photoprism.PurgeOptions{
			Path:   subPath,
			Ignore: indexed,
		}

		if files, photos, err := w.Start(opt); err != nil {
			log.Error(err)
		} else if len(files) > 0 || len(photos) > 0 {
			log.Infof("purge: removed %s and %s [%s]", english.Plural(len(files), "file", "files"), english.Plural(len(photos), "photo", "photos"), time.Since(purgeStart))
		}
	}


	elapsed := time.Since(start)

	log.Infof("indexed %s in %s", english.Plural(len(indexed), "file", "files"), elapsed)

	conf.Shutdown()

	return nil
}

// labelResetAction resets face clusters and matches.
func labelResetAction(ctx *cli.Context) error {
	start := time.Now()

	conf := config.NewConfig(ctx)
	service.SetConfig(conf)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := conf.Init(); err != nil {
		return err
	}

	conf.InitDb()

	if err := query.ResetAllLabels(); err != nil {
		return err
	} else {
		elapsed := time.Since(start)

		log.Infof("completed in %s", elapsed)
	}

	conf.Shutdown()

	return nil
}
