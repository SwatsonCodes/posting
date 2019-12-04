package main

import (
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

const cowsay string = `
 __________
< GO AWAY. >
 ----------
        \   ^__^
         \  (oo)\_______
            (__)\       )\/\
                ||----w |
                ||     ||
`

func main() {
	lambda.Start(func() (events.APIGatewayProxyResponse, error) {
		return events.APIGatewayProxyResponse{
				StatusCode: http.StatusOK,
				Body:       cowsay,
				Headers: map[string]string{
					"Content-Type":  "text/plain",
					"Cache-Control": "no-cache",
				},
			},
			nil
	})
}
