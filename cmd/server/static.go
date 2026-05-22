package main

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"todo/web"
)

func registerStaticRoutes(r *gin.Engine, logger *zap.Logger) {
	distFS, err := fs.Sub(web.Files, "dist")
	if err != nil {
		logger.Fatal("无法加载静态资源", zap.Error(err))
	}
	assetsFS, err := fs.Sub(distFS, "assets")
	if err != nil {
		logger.Fatal("无法加载 assets 资源", zap.Error(err))
	}
	indexHTML, err := fs.ReadFile(distFS, "index.html")
	if err != nil {
		logger.Fatal("无法加载 index.html", zap.Error(err))
	}

	// Root-level static files
	r.StaticFS("/assets", http.FS(assetsFS))
	r.StaticFileFS("/favicon.svg", "favicon.svg", http.FS(distFS))
	r.StaticFileFS("/icons.svg", "icons.svg", http.FS(distFS))

	// /admin prefix mirrors the same files so that relative paths
	// (base: './') resolve correctly when the page URL is /admin/...
	r.StaticFS("/admin/assets", http.FS(assetsFS))
	r.StaticFileFS("/admin/favicon.svg", "favicon.svg", http.FS(distFS))
	r.StaticFileFS("/admin/icons.svg", "icons.svg", http.FS(distFS))

	r.NoRoute(func(c *gin.Context) {
		if isAPIRoute(c.Request.URL.Path) {
			writeAPINotFound(c)
			return
		}

		path := c.Request.URL.Path

		// Non-root paths need the hash router to handle them.
		// Redirect /admin/login → /#/admin/login so Vue Router can match.
		if path != "/" {
			c.Redirect(http.StatusFound, "/#"+path)
			return
		}

		// Root: serve index.html directly.
		// Do not use FileFromFS("index.html"): http.FileServer redirects /index.html
		// to ./, which can loop at the root route.
		c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
	})
}

// normalizeAssetPath converts /admin/assets/... to /assets/... for logging.
func normalizeAssetPath(path string) string {
	if strings.HasPrefix(path, "/admin/") {
		return path[len("/admin"):]
	}
	return path
}
