package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type yesAuthorizer struct{}
type unAuthorizer struct{}

func authorizeRequest(r *http.Request) bool {
	return true
}

func unauthorizeRequest(r *http.Request) bool {
	return false
}

func TestAuthCheckerMiddleware(t *testing.T) {

	testcases := []struct {
		authorizer AuthChecker
		statusCode int
	}{
		{authorizeRequest, http.StatusOK},
		{unauthorizeRequest, http.StatusUnauthorized},
	}

	for _, testcase := range testcases {
		req, err := http.NewRequest(http.MethodGet, "/test", nil)
		if err != nil {
			t.Fatal(err)
		}
		recorder := httptest.NewRecorder()
		middleware := testcase.authorizer.CheckAuth(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
		middleware.ServeHTTP(recorder, req)

		if status := recorder.Code; status != testcase.statusCode {
			t.Errorf("handler returned wrong status code: got '%d' but expected '%d'",
				status, testcase.statusCode)
		}
	}
}
