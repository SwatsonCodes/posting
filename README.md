GOOS=linux go build main.go
zip nice_lambda.zip main
aws lambda update-function-code --function-name very_nice --zip-file fileb://nice_lambda.zip --publish
