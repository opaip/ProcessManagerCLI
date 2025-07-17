package api

import (
	"log/slog"
	"net/http"
)

// logRequests is a middleware that logs every incoming request.
func (api *ProcessAPI) logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		api.Logger.Info("api request received", "method", r.Method, "path", r.URL.Path, "remote_addr", r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

// authMiddleware is a middleware that checks for a valid API key.
func (api *ProcessAPI) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the API key from the request header
		apiKey := r.Header.Get("X-API-KEY")

		if apiKey == "" {
			api.Logger.Warn("API key is missing", "remote_addr", r.RemoteAddr)
			respondWithError(w, http.StatusUnauthorized, "API Key is missing")
			return
		}

		// Check if the provided key is valid
		if !api.isKeyValid(apiKey) {
			api.Logger.Warn("Invalid API key provided", "remote_addr", r.RemoteAddr)
			respondWithError(w, http.StatusForbidden, "Invalid API Key")
			return
		}

		// If the key is valid, proceed to the next handler
		next.ServeHTTP(w, r)
	})
}

// isKeyValid checks if a given key exists in the configured list of API keys.
func (api *ProcessAPI) isKeyValid(providedKey string) bool {
	for _, validKey := range api.Config.ApiKeys {
		if providedKey == validKey {
			return true
		}
	}
	return false
}