package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

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
const maxRequestBodySizeBytes int64 = 32 * 1024 // 32KiB
var adapter *gorillamux.GorillaMuxAdapter
var router *mux.Router

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.Info("cold start")
	router = mux.NewRouter()
	router.HandleFunc("/", goAway).Methods(http.MethodGet)
	router.HandleFunc("/posts", postsHandler).Methods(http.MethodPost)
	router.Use(middleware.BodyLimiter{MaxBodySizeBytes: maxRequestBodySizeBytes}.Middleware)
	adapter = gorillamux.New(router)
}

func goAway(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, cowsay)
}

func postsHandler(w http.ResponseWriter, r *http.Request) {
	bod, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.WithError(err).Error("failed to read post body")
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
	log.Info(string(bod))
}

func main() {
	if _, isInsideLambda := os.LookupEnv("LAMBDA_TASK_ROOT"); isInsideLambda {
		lambda.Start(func(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
			return adapter.Proxy(req)
		})
	} else {
		log.Info("serving on port 8000")
		http.ListenAndServe(":8000", router)
	}
}
