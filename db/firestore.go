package db

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/swatsoncodes/posting/models"
)

type firestoreClient struct {
	ctx   context.Context
	posts *firestore.CollectionRef
}

func NewFirestoreClient(projectID, postsCollectionName string) (client *firestoreClient, err error) {
	ctx := context.Background()
	fc, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return
	}
	return &firestoreClient{ctx, fc.Collection(postsCollectionName)}, nil
}

func (client *firestoreClient) PutPost(post models.Post) (err error) {
	_, err = client.posts.Doc(post.ID).Create(client.ctx, post)
	return
}

func (client *firestoreClient) GetPosts(offset, limit int) (*[]models.Post, error) {
	docs, err := client.posts.
		Select("post_id", "body", "created_at", "media_urls").
		OrderBy("created_at", firestore.Desc).
		Offset(offset).
		Limit(limit).
		Documents(client.ctx).
		GetAll()
	if err != nil {
		return nil, err
	}

	posts := make([]models.Post, len(docs))
	for i, doc := range docs {
		if err = doc.DataTo(&posts[i]); err != nil {
			return nil, err
		}
	}

	return &posts, nil
}
