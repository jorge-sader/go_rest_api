package middlewares

import (
	"net/http"
	"slices"
)

var allowedOrigins = []string{
	"https://localhost:3000",
	"https://127.0.0.1:3000",
	"https://your-domain.com",
}

func Cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("origin")

		if slices.Contains(allowedOrigins, origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			http.Error(w, "Not allowed by CORS", http.StatusForbidden)
			return
		}

		w.Header().Set("access-control-allow-headers", "Content-type, Authorization")
		w.Header().Set("access-control-expose-headers", "Authorization")
		w.Header().Set("access-control-allow-methods", "GET, POST, PUT, PATCH, DELETE")
		w.Header().Set("access-control-allow-credentials", "true")
		w.Header().Set("access-control-max-age", "3600")

		w.Header().Set("cross-origin-resource-policy", "same-origin")
		w.Header().Set("cross-origin-opener-policy", "same-origin")
		w.Header().Set("cross-origin-embedder-policy", "require-corp")

		if r.Method == http.MethodOptions {
			return
		}

		next.ServeHTTP(w, r)
	})
}
