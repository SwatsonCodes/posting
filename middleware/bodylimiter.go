package middleware

import (
	"fmt"
	"net/http"
)

func LimitRequestBody(maxRequestBodySizeBytes int64, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ContentLength > maxRequestBodySizeBytes {
			http.Error(w, fmt.Sprintf("request body exceeds %d bytes", maxRequestBodySizeBytes), http.StatusBadRequest)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, r.ContentLength)
		next.ServeHTTP(w, r)
	})
}
