package web

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

//go:embed all:dist
var staticFS embed.FS

var contentTypes = map[string]string{
	".html": "text/html; charset=utf-8",
	".js":   "application/javascript",
	".css":  "text/css; charset=utf-8",
	".json": "application/json",
	".svg":  "image/svg+xml",
	".png":  "image/png",
	".ico":  "image/x-icon",
	".woff": "font/woff",
	".woff2": "font/woff2",
}

func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	dist, err := fs.Sub(staticFS, "dist")
	if err != nil {
		s.serveFallbackHTML(w)
		return
	}

	urlPath := r.URL.Path
	if urlPath == "/" {
		urlPath = "/index.html"
	}

	filePath := strings.TrimPrefix(urlPath, "/")
	data, err := fs.ReadFile(dist, filePath)
	if err != nil {
		data, err = fs.ReadFile(dist, "index.html")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		urlPath = "/index.html"
	}

	ext := path.Ext(urlPath)
	contentType := contentTypes[ext]
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	w.Header().Set("Content-Type", contentType)
	if ext == ".js" || ext == ".css" {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	}
	w.Write(data)
}

func (s *Server) serveFallbackHTML(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>BAR Web UI</title>
  <style>
    body { font-family: system-ui, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
    h1 { color: #333; }
    .endpoint { background: #f5f5f5; padding: 10px; margin: 10px 0; border-radius: 4px; }
    code { background: #e0e0e0; padding: 2px 6px; border-radius: 3px; }
  </style>
</head>
<body>
  <h1>BAR Web UI</h1>
  <p>Frontend not built. Run <code>cd web && pnpm build</code> first.</p>
  <h2>API Endpoints</h2>
  <div class="endpoint"><code>GET /api/health</code> - Health check</div>
  <div class="endpoint"><code>GET /api/tasks</code> - List all tasks</div>
  <div class="endpoint"><code>GET /api/tasks/:id</code> - Get task detail</div>
  <div class="endpoint"><code>GET /api/ledger/:task_id</code> - Get ledger entries</div>
  <div class="endpoint"><code>GET /api/diff/:task_id/:step_id</code> - Get diff content</div>
  <div class="endpoint"><code>GET /api/status</code> - Get current status</div>
  <div class="endpoint"><code>WS /ws</code> - WebSocket for real-time updates</div>
</body>
</html>`))
}
