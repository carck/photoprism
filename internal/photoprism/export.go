package photoprism

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/photoprism/photoprism/pkg/fs"
)

func ExportDB(force bool) (err error, file string) {
	cfg := Config()
	file = filepath.Join(os.TempDir(), "up.db")

	if force || !fs.FileExists(file) {
		cmd := exec.Command(cfg.ExportCommand(), file)
		if err = cmd.Run(); err != nil {
			return err, file
		}
	}

	return nil, file
}
