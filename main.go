package main

import (
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

const cowsay string = `
 _________
< GO AWAY >
 ---------
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
			},
			nil
	})
}
