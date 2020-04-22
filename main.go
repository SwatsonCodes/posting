package main

import (
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/gorillamux"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/swatsoncodes/very-nice-website/db"
	"github.com/swatsoncodes/very-nice-website/middleware"
)

const bodySizeLimit middleware.RequestBodyLimitBytes = 32 * 1024 // 32KiB
const tableName = "posts"                                        // TODO: make this configurable

func isRunningInLambda() bool {
	_, inLambda := os.LookupEnv("LAMBDA_TASK_ROOT")
	return inLambda
}

func main() {
	var dynamoEndpoint *string
	var sender, twilioToken string
	var ok bool
	log.SetFormatter(&log.JSONFormatter{DisableHTMLEscape: true})
	log.Info("hello")

	if sender, ok = os.LookupEnv("ALLOWED_SENDER"); !ok {
		log.Fatal("env var ALLOWED_SENDER not set")
	}
	if twilioToken, ok = os.LookupEnv("TWILIO_AUTH_TOKEN"); !ok {
		log.Fatal("env var TWILIO_AUTH_TOKEN not set")
	}
	if de, ok := os.LookupEnv("DYNAMODB_ENDPOINT"); ok {
		dynamoEndpoint = &de
		log.Infof("using custom DynamoDB endpoint: '%s'", de)
	}

	db, err := db.New(tableName, dynamoEndpoint)
	if err != nil {
		log.WithError(err).Fatal("failed to initialize db")
	}
	poster := Poster{
		AllowedSender:   sender,
		TwilioAuthToken: twilioToken,
		DB:              db,
	}
	router := mux.NewRouter()

	router.Handle("/posts",
		bodySizeLimit.LimitRequestBody( // guard against giant posts
			middleware.AuthChecker(poster.IsRequestAuthorized).CheckAuth( // make sure posters are authorized
				http.HandlerFunc(poster.CreatePost)))).
		Methods(http.MethodPost).
		Headers("Content-Type", "application/x-www-form-urlencoded")
	router.HandleFunc("/posts", poster.GetPosts).Methods(http.MethodGet)
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/posts", http.StatusMovedPermanently)
	}).Methods(http.MethodGet)

	router.NotFoundHandler = http.HandlerFunc(GoAway)

	router.Use(middleware.LogRequest)
	adapter := gorillamux.New(router)

	if isRunningInLambda() {
		lambda.Start(func(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
			return adapter.Proxy(req)
		})
	} else {
		log.Info("serving on port 8008")
		http.ListenAndServe(":8008", router)
	}
}
