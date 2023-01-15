package middleware

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

// LogRequestBody logs request bodies to a file. Useful for dev debugging.
func LogRequestBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			buf, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.Fatal(err)
			}
			r.Body = ioutil.NopCloser(bytes.NewBuffer(buf))
			if len(buf) > 0 {
				// TODO: make the file name configurable
				f, err := os.OpenFile("request_bods.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					log.Fatal(err)
				}
				defer f.Close()
				f.Write(buf)
				f.WriteString("\n")
			}
		}
		next.ServeHTTP(w, r)
	})
}
