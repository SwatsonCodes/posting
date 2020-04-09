package middleware

import (
	"fmt"
	"net/http"
)

type bodyLimiter struct {
	maxBodySizeBytes int64
}

func NewBodyLimiter(maxBodySizeBytes int64) *bodyLimiter {
	limiter := bodyLimiter{maxBodySizeBytes}
	return &limiter
}

func (limiter bodyLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ContentLength > limiter.maxBodySizeBytes {
			http.Error(w, fmt.Sprintf("request body exceeds %d bytes", limiter.maxBodySizeBytes), http.StatusBadRequest)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, r.ContentLength)
		next.ServeHTTP(w, r)
	})
}
