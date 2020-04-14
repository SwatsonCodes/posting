#! /bin/bash -eux

AWS_PAGER="" \
aws dynamodb create-table \
    --table-name posts \
    --attribute-definitions \
        AttributeName=post_id,AttributeType=S \
    --key-schema \
        AttributeName=post_id,KeyType=HASH \
--billing-mode PAY_PER_REQUEST \
--endpoint-url http://localhost:8000
