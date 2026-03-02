package orchestrate

import (
	"net/http"
	"os"
	"path/filepath"
)

// ServeWebUI serves the web UI from the filesystem (not embedded for now).
func (s *Server) ServeWebUI() http.Handler {
	// Try to serve from web/dist directory
	distPath := filepath.Join("web", "dist")
	if _, err := os.Stat(distPath); err == nil {
		return http.FileServer(http.Dir(distPath))
	}

	// Fallback HTML if UI not built
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
    <title>AgentRuntime</title>
    <style>
        body { font-family: system-ui; padding: 2rem; background: #0f0f0f; color: #e0e0e0; }
        .container { max-width: 600px; margin: 0 auto; text-align: center; }
        h1 { color: #3b82f6; }
        code { background: #1a1a1a; padding: 0.25rem 0.5rem; border-radius: 0.25rem; }
    </style>
</head>
<body>
    <div class="container">
        <h1>AgentRuntime Control Panel</h1>
        <p>Web UI is available but not built yet.</p>
        <p>To build and enable the UI:</p>
        <pre><code>cd web && npm install && npm run build</code></pre>
        <p>API is available at <code>/v1/*</code></p>
    </div>
</body>
</html>`))
	})
}

// SetupWebUIRoutes adds web UI routes to the server.
func (s *Server) SetupWebUIRoutes(mux *http.ServeMux) {
	handler := s.ServeWebUI()

	// Serve static assets
	mux.Handle("/assets/", handler)

	// Serve index.html for all other paths (SPA routing)
	mux.Handle("/", handler)
}
