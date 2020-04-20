package middleware

import (
	"net/http"
)

func SetURLHost(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Host == "" {
			r.URL.Host = r.Host
		}
		next.ServeHTTP(w, r)
	})
}
