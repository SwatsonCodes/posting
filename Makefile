run-local:
	export `cat dev_vars.env | xargs` && go run .

test:
	go test

docker-firestore-run:
	docker run -d -p 8001:8080 --name firestore_poster --rm ridedott/firestore-emulator:latest

docker-firestore-kill:
	docker kill firestore_poster

gcloud-deploy:
	gcloud app deploy

gcloud-logs:
	gcloud app logs tail -s default
