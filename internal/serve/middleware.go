package serve

import (
	"net/http"
)

const (
	// Standard HTTP headers.
	HeaderReferrerPolicy      = "Referrer-Policy"
	HeaderXContentTypeOptions = "X-Content-Type-Options"
	HeaderXFrameOptions       = "X-Frame-Options"
	HeaderXRobotsTag          = "X-Robots-Tag"
	HeaderXXSSProtection      = "X-XSS-Protection"
)

// securityHeaders sets HTTP headers for security purposes.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(HeaderReferrerPolicy, "no-referrer")
		w.Header().Set(HeaderXContentTypeOptions, "nosniff")
		w.Header().Set(HeaderXFrameOptions, "deny")
		w.Header().Set(HeaderXRobotsTag, "noindex, nofollow")
		w.Header().Set(HeaderXXSSProtection, "1; mode=block")

		next.ServeHTTP(w, r)
	})
}
