
.PHONY: docker-dynamo-run
docker-dynamo-run:
	docker run -d -p 8000:8000 --name dynamo_nice --rm amazon/dynamodb-local

.PHONY: docker-dynamo-init
docker-dynamo-init:
	./db/create_dynamo_posts_table_local.sh

.PHONY: docker-dynamo-local
docker-dynamo-local: docker-dynamo-run docker-dynamo-init

.PHONY: docker-dynamo-kill
docker-dynamo-kill:
	docker kill dynamo_nice

.PHONY: run-local
run-local:
	export `cat dev_vars.env | xargs` && go run .

.PHONY: clean
clean:
	rm main
	rm nice_lambda.zip

.PHONY: build
build:
	GOOS=linux go build main.go poster.go

.PHONY: package
package:
	zip -r nice_lambda.zip main gcloud_poster_creds.json templates/

.PHONY: lambda-update
lambda-update: clean build package
	aws lambda update-function-code --function-name very_nice --zip-file fileb://nice_lambda.zip --publish
