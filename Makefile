
.PHONY: docker-dynamo-run
docker-dynamo-run:
	docker run -d -p 8000:8000 --name dynamo_nice --rm amazon/dynamodb-local

.PHONY: docker-dynamo-init
docker-dynamo-init:
	./create_posts_table_local.sh

.PHONY: docker-dynamo-local
docker-dynamo-local: docker-dynamo-run docker-dynamo-init

.PHONY: docker-dynamo-kill
docker-dynamo-kill:
	docker kill dynamo_nice

.PHONY: run-local
run-local:
	export `cat dev_vars.env | xargs` && go run main.go
