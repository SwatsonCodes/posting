// Package db provides tooling for interacting with a database that stores Posts.
package db

import "github.com/swatsoncodes/posting/models"

// PostsDB specifies methods for storing and retrieving Posts from a database.
type PostsDB interface {
	PutPost(post models.Post) (err error)
	GetPosts(offset, limit int) (posts *[]models.Post, isMore bool, err error)
}
