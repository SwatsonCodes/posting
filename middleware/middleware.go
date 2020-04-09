package middleware

import (
	"fmt"
	"net/http"
)

type BodyLimiter struct {
	MaxBodySizeBytes int64
}

func (bodyLimiter BodyLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ContentLength > bodyLimiter.MaxBodySizeBytes {
			http.Error(w, fmt.Sprintf("request body exceeds %d bytes", bodyLimiter.MaxBodySizeBytes), http.StatusBadRequest)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, r.ContentLength)
		next.ServeHTTP(w, r)
	})
}
