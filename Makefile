.PHONY: run-local
run-local:
	export `cat dev_vars.env | xargs` && go run .
