package swaggerui

import (
	"fmt"
	"net/http"
)

// Handler returns an HTTP handler that serves an interactive OpenAPI documentation UI (Scalar / Swagger)
// based on the provided OpenAPI specification URL or JSON/YAML content.
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
      body { margin: 0; padding: 0; }
    </style>
  </head>
  <body>
    <script
      id="api-reference"
      data-url="./openapi.yaml">
    </script>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
  </body>
</html>`, title)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(html))
	}
}
