package routes

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

var pathParameterPattern = regexp.MustCompile(`:([A-Za-z0-9_]+)`)

func registerDocumentationRoutes(router *gin.Engine) {
	router.GET("/openapi.json", func(c *gin.Context) {
		c.JSON(http.StatusOK, buildOpenAPI(router.Routes()))
	})
	router.GET("/docs", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, swaggerPage)
	})
}

func buildOpenAPI(routes gin.RoutesInfo) gin.H {
	paths := gin.H{}
	for _, route := range routes {
		path := pathParameterPattern.ReplaceAllString(route.Path, `{$1}`)
		operations, ok := paths[path].(gin.H)
		if !ok {
			operations = gin.H{}
			paths[path] = operations
		}
		operation := gin.H{
			"operationId": operationID(route.Method, path),
			"tags":        []string{routeTag(path)},
			"responses": gin.H{
				"200": gin.H{"description": "Successful response"},
				"400": gin.H{"description": "Invalid request"},
				"401": gin.H{"description": "Authentication required"},
				"403": gin.H{"description": "Permission denied"},
				"500": gin.H{"description": "Internal server error"},
			},
		}
		parameters := pathParameters(path)
		if len(parameters) > 0 {
			operation["parameters"] = parameters
		}
		if route.Method != http.MethodGet && route.Method != http.MethodHead && route.Method != http.MethodDelete {
			operation["requestBody"] = gin.H{
				"required": false,
				"content": gin.H{"application/json": gin.H{
					"schema": gin.H{"type": "object"},
				}},
			}
		}
		if requiresBearer(path) {
			operation["security"] = []gin.H{{"bearerAuth": []string{}}}
		}
		operations[strings.ToLower(route.Method)] = operation
	}
	return gin.H{
		"openapi": "3.0.3",
		"info": gin.H{
			"title":       "WMS API",
			"version":     "1.0.0",
			"description": "Warehouse management API. Request and response examples are documented in the repository docs directory.",
		},
		"servers": []gin.H{{"url": "/"}},
		"paths":   paths,
		"components": gin.H{"securitySchemes": gin.H{
			"bearerAuth": gin.H{"type": "http", "scheme": "bearer"},
		}},
	}
}

func pathParameters(path string) []gin.H {
	matches := pathParameterPattern.FindAllStringSubmatch(strings.ReplaceAll(path, "{", ":"), -1)
	parameters := make([]gin.H, 0, len(matches))
	for _, match := range matches {
		parameters = append(parameters, gin.H{
			"name": match[1], "in": "path", "required": true,
			"schema": gin.H{"type": "string"},
		})
	}
	return parameters
}

func requiresBearer(path string) bool {
	return strings.HasPrefix(path, "/api/v1/") && path != "/api/v1/auth/login"
}

func routeTag(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 3 && parts[0] == "api" {
		return parts[2]
	}
	return "system"
}

func operationID(method, path string) string {
	value := strings.ToLower(method) + "_" + strings.Trim(path, "/")
	replacer := strings.NewReplacer("/", "_", "{", "", "}", "", "-", "_")
	return replacer.Replace(value)
}

const swaggerPage = `<!doctype html>
<html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">
<title>WMS API documentation</title><link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css"></head>
<body><div id="swagger-ui"></div><script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
<script>SwaggerUIBundle({url:'/openapi.json',dom_id:'#swagger-ui',deepLinking:true,persistAuthorization:true})</script></body></html>`
