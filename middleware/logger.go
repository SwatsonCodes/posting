package middleware

import (
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/sirupsen/logrus"
)

// LogRequest is middleware that logs incoming HTTP requests
func LogRequest(next http.Handler) http.Handler {
	return handlers.CombinedLoggingHandler(logrus.StandardLogger().Writer(), next)
}
