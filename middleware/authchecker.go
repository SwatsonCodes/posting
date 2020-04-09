package middleware

import "net/http"

type requestAuthorizer interface {
	IsRequestAuthorized(*http.Request) bool
}

type authChecker struct {
	requestAuthorizer requestAuthorizer
}

func NewAuthChecker(authorizer requestAuthorizer) *authChecker {
	checker := authChecker{authorizer}
	return &checker
}

func (checker authChecker) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !checker.requestAuthorizer.IsRequestAuthorized(r) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
