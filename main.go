package main

import (
	"net/http"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/gorillamux"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/swatsoncodes/very-nice-website/db"
	"github.com/swatsoncodes/very-nice-website/middleware"
)

const defaultMaxRequestBodySizeBytes int64 = 32 * 1024 // 32KiB
const tableName = "posts"                              // TODO: make this configurable

func isRunningInLambda() bool {
	_, inLambda := os.LookupEnv("LAMBDA_TASK_ROOT")
	return inLambda
}

func main() {
	var dynamoEndpoint *string
	log.SetFormatter(&log.JSONFormatter{DisableHTMLEscape: true})
	log.Info("hello")

	sender, ok := os.LookupEnv("ALLOWED_SENDER")
	if !ok {
		log.Fatal("env var ALLOWED_SENDER not set")
	}
	accountID, ok := os.LookupEnv("TWILIO_ACCOUNT_ID")
	if !ok {
		log.Fatal("env var TWILIO_ACCOUNT_ID not set")
	}
	if de, ok := os.LookupEnv("DYNAMODB_ENDPOINT"); ok {
		dynamoEndpoint = &de
		log.Infof("using custom DynamoDB endpoint: '%s'", de)
	}
	if !ok {
		log.Fatal("env var TWILIO_ACCOUNT_ID not set")
	}
	maxBody := os.Getenv("MAX_REQUEST_BODY_SIZE_BYTES")
	max, err := strconv.ParseInt(maxBody, 10, 0)
	if err != nil {
		log.Warnf("env var MAX_REQUEST_BODY_SIZE_BYTES not set or invalid. using default value of %d", defaultMaxRequestBodySizeBytes)
		max = defaultMaxRequestBodySizeBytes
	}

	db, err := db.New(tableName, dynamoEndpoint)
	if err != nil {
		panic(err)
	}
	poster := Poster{
		AllowedSender:           sender,
		TwilioAccountID:         accountID,
		MaxRequestBodySizeBytes: max,
		DB:                      db,
	}
	router := mux.NewRouter()

	router.HandleFunc("/", GoAway).Methods(http.MethodGet)
	router.Handle("/posts",
		middleware.LimitRequestBody(max, middleware.CheckAuth(
			poster.IsRequestAuthorized, http.HandlerFunc(poster.CreatePost)))).
		Methods(http.MethodPost).
		Headers("Content-Type", "application/x-www-form-urlencoded")
	router.HandleFunc("/posts", poster.GetPosts).Methods(http.MethodGet)
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
