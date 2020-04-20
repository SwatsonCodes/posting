package middleware

import (
	"fmt"
	"net/http"
)

type RequestBodyLimitBytes int64

func (limit RequestBodyLimitBytes) LimitRequestBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if l := int64(limit); r.ContentLength > l {
			http.Error(w, fmt.Sprintf("request body exceeds %d bytes", l), http.StatusBadRequest)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, r.ContentLength)
		next.ServeHTTP(w, r)
	})
}
