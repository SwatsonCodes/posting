package db

import "github.com/swatsoncodes/posting/models"

type PostsDB interface {
	PutPost(post models.Post) (err error)
	GetPosts(offset, limit int) (posts *[]models.Post, err error)
}
