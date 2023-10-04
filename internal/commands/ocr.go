package commands

import (
	"context"
	"time"

	"github.com/dustin/go-humanize/english"

	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/pkg/fs"
)

var OcrCommand = cli.Command{
	Name:  "ocr",
	Usage: "Run ocr for images",
	Subcommands: []cli.Command{
		{
			Name:   "index",
			Usage:  "run label action",
			Action: ocrIndexAction,
		},
	},
}

func ocrIndexAction(ctx *cli.Context) error {
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
			OcrOnly: true,
		}

		indexed = w.Start(opt)
	}

	elapsed := time.Since(start)

	log.Infof("indexed %s in %s", english.Plural(len(indexed), "file", "files"), elapsed)

	conf.Shutdown()

	return nil
}
