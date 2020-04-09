package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type yesAuthorizer struct{}
type unAuthorizer struct{}

func (checker yesAuthorizer) IsRequestAuthorized(r *http.Request) bool {
	return true
}

func (checker unAuthorizer) IsRequestAuthorized(r *http.Request) bool {
	return false
}

func TestAuthCheckerMiddleware(t *testing.T) {
	accepter := authChecker{yesAuthorizer{}}
	refuser := authChecker{unAuthorizer{}}

	testcases := []struct {
		authorizer authChecker
		statusCode int
	}{
		{accepter, http.StatusOK},
		{refuser, http.StatusUnauthorized},
	}

	for _, testcase := range testcases {
		req, err := http.NewRequest(http.MethodGet, "/test", nil)
		if err != nil {
			t.Fatal(err)
		}
		recorder := httptest.NewRecorder()
		middleware := testcase.authorizer.Middleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
		middleware.ServeHTTP(recorder, req)

		if status := recorder.Code; status != testcase.statusCode {
			t.Errorf("handler returned wrong status code: got '%d' but expected '%d'",
				status, testcase.statusCode)
		}
	}
}
