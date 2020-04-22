package middleware

import (
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/sirupsen/logrus"
)

func LogRequest(next http.Handler) http.Handler {
	return handlers.CombinedLoggingHandler(logrus.StandardLogger().Writer(), next)
}
