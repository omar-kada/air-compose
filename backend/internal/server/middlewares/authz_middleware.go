// Package middlewares provides HTTP middleware functionality.
package middlewares

import (
	"net/http"
	"strings"

	"omar-kada/air-compose/api"
	"omar-kada/air-compose/models"
)

// AuthorizationMiddleware checks if the requested API endpoint is disabled based on feature flags.
func AuthorizationMiddleware(next http.Handler) http.Handler {
	features := models.LoadFeatures()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url, ok := strings.CutPrefix(r.URL.Path, "/api/")
		if !ok {
			next.ServeHTTP(w, r)
			return
		}
		if isDisabled(r.Method, url, features) {
			sendError(w, api.ErrorCodeNOTALLOWED)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func isDisabled(method, url string, features models.Features) bool {
	switch url {
	case "config":
		return method == http.MethodPost && !features.EditConfig || method == http.MethodGet && !features.DisplayConfig
	case "settings":
		return method == http.MethodPost && !features.EditSettings
	}
	return false
}
