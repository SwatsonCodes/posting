// Package db provides tooling for interacting with a database that stores Posts.
package db

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/swatsoncodes/posting/models"
	"google.golang.org/api/iterator"
)

type firestoreClient struct {
	ctx   context.Context
	posts *firestore.CollectionRef
}

// NewFirestoreClient returns a reference to a Firestore DB client
func NewFirestoreClient(projectID, postsCollectionName string) (client *firestoreClient, err error) {
	ctx := context.Background()
	fc, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return
	}
	return &firestoreClient{ctx, fc.Collection(postsCollectionName)}, nil
}

// PutPost adds a new post to the DB
func (client *firestoreClient) PutPost(post models.Post) (err error) {
	_, err = client.posts.Doc(post.ID).Create(client.ctx, post)
	return
}

// GetPosts retrieves Posts from the DB, starting from the given offset and up to the given limit.
func (client *firestoreClient) GetPosts(offset, limit int) (posts *[]models.Post, isMore bool, err error) {
	var poasts []models.Post
	isMore = true

	docs := client.posts.
		Select("post_id", "body", "created_at", "media_urls").
		OrderBy("created_at", firestore.Desc).
		Offset(offset).
		Limit(limit + 1). // ask for one more so we can tell if we've reached the end
		Documents(client.ctx)
	defer docs.Stop()

	for i := 0; i < limit; i++ {
		var post models.Post
		var doc *firestore.DocumentSnapshot
		doc, err = docs.Next()
		if err == iterator.Done {
			return &poasts, false, nil
		}
		if err != nil {
			return
		}

		if err = doc.DataTo(&post); err != nil {
			return
		}
		poasts = append(poasts, post)
	}
	if _, err = docs.Next(); err == iterator.Done {
		isMore = false
	}

	return &poasts, isMore, nil
}
