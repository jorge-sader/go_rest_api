package middlewares

import "net/http"

func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-DNS-Prefetch-Control", "off")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1;mode=block")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Strict-Transport-Security", "max-age=6372000;includeSubDomains;preload")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-Powered-By", "Django") // set to wrong language as decoy
		w.Header().Set("Server", "")
		w.Header().Set("x-permitted-cross-domain-policies", "none")
		w.Header().Set("cache-control", "no-store, no-cache, must-revalidate, max-age=0")

		w.Header().Set("permissions-policy", "geo-location=(self), microphone=()")

		next.ServeHTTP(w, r)
	})
}
