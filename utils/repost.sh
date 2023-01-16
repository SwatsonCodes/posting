#! /bin/bash -eux
# re-posts all requests found in request_bods.log to the localhost.
# useful for debugging and testing.

cat request_bods.log | while read bod; do
	curl -X POST 'http://localhost:8008/posts' -d "$bod"
done
