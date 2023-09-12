package server

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/sanitize"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/auto"
	"github.com/photoprism/photoprism/internal/config"
	"golang.org/x/net/webdav"
)

const WebDAVOriginals = "/originals"
const WebDAVImport = "/import"
const WebDAVCargo = "/cargo"

// MarkUploadAsFavorite sets the favorite flag for newly uploaded files.
func MarkUploadAsFavorite(fileName string) {
	yamlName := fs.AbsPrefix(fileName, false) + fs.YamlExt

	// Abort if YAML file already exists to avoid overwriting metadata.
	if fs.FileExists(yamlName) {
		log.Warnf("webdav: %s already exists", sanitize.Log(filepath.Base(yamlName)))
		return
	}

	// Make sure directory exists.
	if err := os.MkdirAll(filepath.Dir(yamlName), os.ModePerm); err != nil {
		log.Errorf("webdav: %s", err.Error())
		return
	}

	// Write YAML data to file.
	if err := os.WriteFile(yamlName, []byte("Favorite: true\n"), os.ModePerm); err != nil {
		log.Errorf("webdav: %s", err.Error())
		return
	}

	// Log success.
	log.Infof("webdav: marked %s as favorite", sanitize.Log(filepath.Base(fileName)))
}

func SetFileTime(fileName string, tHeader string) {
	if msec, err := strconv.ParseInt(tHeader, 10, 64); err != nil {
		log.Errorf("webdav: parse file time error %s", err.Error())
	} else {
		t := time.UnixMilli(msec)
		if err = os.Chtimes(fileName, t, t); err != nil {
			log.Errorf("webdav: change file time error: %v", err)
		}
	}
}

// WebDAV handles any requests to /originals|import/*
func WebDAV(path string, router *gin.RouterGroup, conf *config.Config) {
	if router == nil {
		log.Error("webdav: router is nil")
		return
	}

	if conf == nil {
		log.Error("webdav: conf is nil")
		return
	}

	f := webdav.Dir(path)

	srv := &webdav.Handler{
		Prefix:     router.BasePath(),
		FileSystem: f,
		LockSystem: webdav.NewMemLS(),
		Logger: func(r *http.Request, err error) {
			if err != nil {
				switch r.Method {
				case MethodPut, MethodPost, MethodPatch, MethodDelete, MethodCopy, MethodMove:
					log.Errorf("webdav: %s in %s %s", sanitize.Log(err.Error()), sanitize.Log(r.Method), sanitize.Log(r.URL.String()))
				case MethodPropfind:
					log.Tracef("webdav: %s in %s %s", sanitize.Log(err.Error()), sanitize.Log(r.Method), sanitize.Log(r.URL.String()))
				default:
					log.Debugf("webdav: %s in %s %s", sanitize.Log(err.Error()), sanitize.Log(r.Method), sanitize.Log(r.URL.String()))
				}

			} else {
				var fileName string
				if router.BasePath() == WebDAVOriginals {
					fileName = filepath.Join(conf.OriginalsPath(), strings.TrimPrefix(r.URL.Path, router.BasePath()))
				} else if router.BasePath() == WebDAVImport {
					fileName = filepath.Join(conf.ImportPath(), strings.TrimPrefix(r.URL.Path, router.BasePath()))
				} else if router.BasePath() == WebDAVCargo {
					fileName = filepath.Join(conf.CargoPath(), strings.TrimPrefix(r.URL.Path, router.BasePath()))
				}

				// Mark uploaded files as favorite if X-Favorite HTTP header is "1".
				if r.Method == MethodPut && r.Header.Get("X-Favorite") == "1" {
					MarkUploadAsFavorite(fileName)
				}

				switch r.Method {
				case MethodPut, MethodPost, MethodPatch, MethodDelete, MethodCopy, MethodMove:
					log.Infof("webdav: %s %s", sanitize.Log(r.Method), sanitize.Log(r.URL.String()))

					if r.Header.Get("X-PS-TIME") != "" {
						SetFileTime(fileName, r.Header.Get("X-PS-TIME"))
					}

					if router.BasePath() == WebDAVOriginals {
						auto.ShouldIndex()
					} else if router.BasePath() == WebDAVImport {
						auto.ShouldImport()
					}
				default:
					log.Tracef("webdav: %s %s", sanitize.Log(r.Method), sanitize.Log(r.URL.String()))
				}
			}
		},
	}

	handler := func(c *gin.Context) {
		w := c.Writer
		r := c.Request

		srv.ServeHTTP(w, r)
	}

	router.Handle(MethodHead, "/*path", handler)
	router.Handle(MethodGet, "/*path", handler)
	router.Handle(MethodPut, "/*path", handler)
	router.Handle(MethodPost, "/*path", handler)
	router.Handle(MethodPatch, "/*path", handler)
	router.Handle(MethodDelete, "/*path", handler)
	router.Handle(MethodOptions, "/*path", handler)
	router.Handle(MethodMkcol, "/*path", handler)
	router.Handle(MethodCopy, "/*path", handler)
	router.Handle(MethodMove, "/*path", handler)
	router.Handle(MethodLock, "/*path", handler)
	router.Handle(MethodUnlock, "/*path", handler)
	router.Handle(MethodPropfind, "/*path", handler)
	router.Handle(MethodProppatch, "/*path", handler)
}
