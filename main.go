package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/gorillamux"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/swatsoncodes/very-nice-website/middleware"
)

const cowsay string = `
 ____________
< GO AWAY <3 >
 ------------
        \   ^__^
         \  (oo)\_______
            (__)\       )\/\
                ||----w |
                ||     ||
`
const defaultMaxRequestBodySizeBytes int64 = 32 * 1024 // 32KiB
var adapter *gorillamux.GorillaMuxAdapter
var router *mux.Router

type niceApp struct {
	AllowedSender           string
	TwilioAccountID         string
	MaxRequestBodySizeBytes int64
}

func goAway(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, cowsay)
}

func (app niceApp) postsHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.WithError(err).Error("failed to parse form body")
		http.Error(w, "unable to parse request form body", http.StatusBadRequest)
	}
	for k, v := range r.PostForm {
		log.Infof("%s: %s", k, v[0])
	}
	w.WriteHeader(http.StatusCreated)
}

func (app niceApp) isRequestAuthorized(r *http.Request) bool {
	err := r.ParseForm()
	if err != nil {
		log.WithError(err).Error("failed to parse form body")
		return false
	}
	if accountID, aOK := (r.PostForm)["AccountSid"]; aOK {
		if sender, sOK := (r.PostForm)["From"]; sOK {
			log.Infof("accountID: %s, sender: %s", accountID, sender)
			return accountID[0] == app.TwilioAccountID && sender[0] == app.AllowedSender
		}
		return false
	}
	return false
}

func initRouter() {
	sender, ok := os.LookupEnv("ALLOWED_SENDER")
	if !ok {
		log.Fatal("env var ALLOWED_SENDER not set")
	}
	accountID, ok := os.LookupEnv("TWILIO_ACCOUNT_ID")
	if !ok {
		log.Fatal("env var TWILIO_ACCOUNT_ID not set")
	}
	maxBody := os.Getenv("MAX_REQUEST_BODY_SIZE_BYTES")
	max, err := strconv.ParseInt(maxBody, 10, 0)
	if err != nil {
		log.Warnf("env var MAX_REQUEST_BODY_SIZE_BYTES not set or invalid. using default value of %d", defaultMaxRequestBodySizeBytes)
		max = defaultMaxRequestBodySizeBytes
	}

	app := niceApp{
		AllowedSender:           sender,
		TwilioAccountID:         accountID,
		MaxRequestBodySizeBytes: max,
	}
	router = mux.NewRouter()
	router.HandleFunc("/", goAway).Methods(http.MethodGet)
	router.Handle("/posts", middleware.LimitRequestBody(max, middleware.CheckAuth(app.isRequestAuthorized, http.HandlerFunc(app.postsHandler)))).Methods(http.MethodPost)
	adapter = gorillamux.New(router)
}

func isRunningInLambda() bool {
	_, inLambda := os.LookupEnv("LAMBDA_TASK_ROOT")
	return inLambda
}

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	log.Info("hello")
	initRouter()
	if isRunningInLambda() {
		lambda.Start(func(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
			return adapter.Proxy(req)
		})
	} else {
		log.Info("serving on port 8000")
		http.ListenAndServe(":8000", router)
	}
}
