package db

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/stretchr/testify/assert"
	"github.com/swatsoncodes/very-nice-website/models"
)

func TestPutPost(t *testing.T) {
	var md dynamodbiface.DynamoDBAPI = mockDynamo{}
	mockamo := Dynamo{"mock", &md}
	assert.NoError(t, mockamo.putPost(models.Post{}), "got unexpected error")
}

func TestGetPosts(t *testing.T) {
	var md dynamodbiface.DynamoDBAPI = mockDynamo{}
	mockamo := Dynamo{"mock", &md}
	posts, err := mockamo.getPosts()
	assert.NoError(t, err, "got unexpected error")
	assert.NotNil(t, posts, "expected some posts, but got nil")
}
