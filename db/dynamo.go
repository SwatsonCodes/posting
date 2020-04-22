package db

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/swatsoncodes/very-nice-website/models"
)

type dynamo struct {
	Table string
	svc   *dynamodbiface.DynamoDBAPI
}

func New(table string, endpoint *string) (*dynamo, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}
	config := aws.Config{Endpoint: endpoint}
	var svc dynamodbiface.DynamoDBAPI = dynamodb.New(sess, &config)
	return &dynamo{table, &svc}, nil
}

func (dynamo dynamo) PutPost(post models.Post) (err error) {
	item, err := dynamodbattribute.MarshalMap(post)
	if err != nil {
		return
	}

	// TODO: dont allow items to be overwritten
	_, err = (*dynamo.svc).PutItem(&dynamodb.PutItemInput{
		TableName: &dynamo.Table,
		Item:      item,
	})
	return
}

func (dynamo dynamo) GetPosts() (*[]models.Post, error) {
	// TODO: pagination stuff
	// TODO: get posts in order by time
	result, err := (*dynamo.svc).Scan(&dynamodb.ScanInput{TableName: &dynamo.Table})
	if err != nil {
		return nil, err
	}
	if *result.Count <= 0 {
		return &[]models.Post{}, nil
	}

	posts := make([]models.Post, *result.Count)
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &posts)
	if err != nil {
		return nil, err
	}
	return &posts, nil
}
