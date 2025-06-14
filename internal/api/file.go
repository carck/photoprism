package api

import (
	"net/http"

	"github.com/photoprism/photoprism/pkg/sanitize"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/query"
)

// GetFile returns file details as JSON.
//
// Route: GET /api/v1/files/:hash
// Params:
// - hash (string) SHA-1 hash of the file
func GetFile(router *gin.RouterGroup) {
	router.GET("/files/:hash", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourceFiles, acl.ActionRead)

		if s.Invalid() {
			AbortUnauthorized(c)
			return
		}

		p, err := query.FileByHash(sanitize.Token(c.Param("hash")))

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		c.JSON(http.StatusOK, p)
	})
}

// GetFileByRootAndNameQuery returns file details as JSON
//  by file root and file name via query params.
//
// Route: GET /api/v1/files/by-path?root=/&name=filename.jpg
func GetFileByRootAndNameQuery(router *gin.RouterGroup) {
	router.GET("/files/by-path", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourceFiles, acl.ActionRead)

		if s.Invalid() {
			AbortUnauthorized(c)
			return
		}

		root := c.Query("root")
		name := c.Query("name")

		root = sanitize.Path(root)
		name = sanitize.Path(name)

		p, err := query.FileByRootAndName(root, name)
		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		c.JSON(http.StatusOK, p)
	})
}
