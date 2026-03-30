package openapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler handles OpenAPI documentation requests
type Handler struct {
	specJSON []byte
}

// NewHandler creates a new OpenAPI handler with pre-generated spec
func NewHandler() *Handler {
	jsonBytes, err := GetSpecJSON()
	if err != nil {
		jsonBytes = []byte(`{"error":"failed to generate spec"}`)
	}
	return &Handler{
		specJSON: jsonBytes,
	}
}

// ServeJSON handles GET /openapi.json
func (h *Handler) ServeJSON(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Data(http.StatusOK, "application/json", h.specJSON)
}

// ServeDocs handles GET /docs — returns Swagger UI HTML page
func (h *Handler) ServeDocs(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, docsHTML)
}

const docsHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>WeCom Gateway API 文档</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
    <style>
        body { margin: 0; padding: 0; }
        .topbar { display: none; }
        .swagger-ui .info { margin: 20px 0; }
        .swagger-ui .info .title { font-size: 2em; }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            SwaggerUIBundle({
                url: "/openapi.json",
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                layout: "StandaloneLayout",
                persistAuthorization: true,
                displayRequestDuration: true,
                filter: true,
                tryItOutEnabled: true
            });
        }
    </script>
</body>
</html>`
