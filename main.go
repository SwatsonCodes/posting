package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/gorillamux"
	"github.com/gorilla/mux"
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

var adapter *gorillamux.GorillaMuxAdapter
var router *mux.Router

func init() {
	log.Print("cold start")
	router = mux.NewRouter()
	router.HandleFunc("/", goAway)
	adapter = gorillamux.New(router)
}

func goAway(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, cowsay)
}

func main() {
	if _, isInsideLambda := os.LookupEnv("LAMBDA_TASK_ROOT"); isInsideLambda {
		lambda.Start(func(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
			return adapter.Proxy(req)
		})
	} else {
		log.Print("serving on port 8000")
		http.ListenAndServe(":8000", router)
	}
}
