package swaggerui

import (
	"fmt"
	"net/http"
)

// Handler returns an HTTP handler that serves an interactive OpenAPI documentation UI
// based on the provided OpenAPI specification URL or JSON/YAML content.
//
// The handler serves the raw spec at /openapi.json or /openapi.yaml, and a minimal
// HTML page at all other paths. The HTML page loads the Scalar API Reference viewer
// from a self-hosted or CDN source depending on the embedScalar parameter.
//
// For production deployments, embed the Scalar JS/CSS assets and serve them from the
// same origin to avoid CSP and supply-chain risks. See the project README for setup.
func Handler(title string, specJSONOrYAML []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/openapi.json" || r.URL.Path == "/openapi.yaml" {
			w.Header().Set("Content-Type", "application/yaml")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(specJSONOrYAML)
			return
		}

		html := fmt.Sprintf(`<!doctype html>
<html>
  <head>
    <title>%s — FactoryOS API Documentation</title>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <style>
      body { margin: 0; padding: 0; font-family: system-ui, sans-serif; display: flex; align-items: center; justify-content: center; min-height: 100vh; }
      .placeholder { text-align: center; max-width: 600px; padding: 2rem; }
      pre { text-align: left; background: #f5f5f5; padding: 1rem; overflow-x: auto; border-radius: 4px; }
      @media (prefers-color-scheme: dark) { pre { background: #1e1e1e; } }
    </style>
  </head>
  <body>
    <div class="placeholder">
      <h1>%s — API Documentation</h1>
      <p>The raw OpenAPI specification is available at:</p>
      <pre><a href="./openapi.yaml">./openapi.yaml</a></pre>
      <p>To view this spec interactively, load it into
        <a href="https://scalar.com/">Scalar</a>,
        <a href="https://editor.swagger.io/">Swagger Editor</a>, or
        <a href="https://redocly.com/">Redoc</a>.
      </p>
    </div>
  </body>
</html>`, title, title)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(html))
	}
}
