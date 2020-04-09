package middleware

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

type bytesReadCounter struct {
	bytesRead int64
}

func (bytesCounter *bytesReadCounter) ReaderHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	bytesCounter.bytesRead = int64(len(body))
}

func TestBodyLimiterMiddleware(t *testing.T) {
	testcases := []struct {
		method           string
		bodySizeBytes    int
		contentLength    int64
		maxBodySizeBytes int64
		statusCode       int
	}{
		{http.MethodGet, 1, 1, 8, http.StatusOK},
		{http.MethodPut, 1, 1, 8, http.StatusOK},
		{http.MethodPost, 1, 1, 8, http.StatusOK},
		{http.MethodGet, 20, 20, 8, http.StatusBadRequest},
		{http.MethodPut, 20, 20, 8, http.StatusBadRequest},
		{http.MethodPost, 20, 20, 8, http.StatusBadRequest},
		{http.MethodPost, 8, 8, 8, http.StatusOK},
		{http.MethodPost, 0, 0, 0, http.StatusOK},
		{http.MethodPost, 1, 1, 0, http.StatusBadRequest},
		{http.MethodPost, 0, 0, -1, http.StatusBadRequest},
		{http.MethodPost, 20, 8, 10, http.StatusOK},
		{http.MethodPost, 20, 0, 10, http.StatusOK},
		{http.MethodPost, 20, 30, 10, http.StatusBadRequest},
		{http.MethodPost, 20, 10, 10, http.StatusOK},
	}

	for _, testcase := range testcases {
		req, err := http.NewRequest(testcase.method, "/test", bytes.NewReader(bytes.Repeat([]byte("a"), testcase.bodySizeBytes)))
		if err != nil {
			t.Fatal(err)
		}
		req.ContentLength = testcase.contentLength
		recorder := httptest.NewRecorder()
		bytesCounter := bytesReadCounter{}
		middleware := bodyLimiter{testcase.maxBodySizeBytes}.Middleware(http.HandlerFunc(bytesCounter.ReaderHandler))
		middleware.ServeHTTP(recorder, req)

		if status := recorder.Code; status != testcase.statusCode {
			t.Errorf("handler returned wrong status code: got '%d' but expected '%d'",
				status, testcase.statusCode)
		}

		if testcase.statusCode == http.StatusBadRequest {
			expectedResponse := fmt.Sprintf("request body exceeds %d bytes\n", testcase.maxBodySizeBytes)
			if recorder.Body.String() != expectedResponse {
				t.Errorf("handler returned unexpected body: got '%s' but expected '%s'",
					recorder.Body.String(), expectedResponse)
			}
			if bytesCounter.bytesRead > 0 {
				t.Errorf("final handler should not have run, but it read %d bytes from the request body", bytesCounter.bytesRead)
			}
			continue
		}

		if bytesCounter.bytesRead > testcase.contentLength {
			t.Errorf("the request Content-Length was %d bytes, but the final handler read %d bytes from the body", testcase.contentLength, bytesCounter.bytesRead)
		}

	}

}
