package db

import "github.com/swatsoncodes/posting/models"

type PostsDB interface {
	PutPost(post models.Post) (err error)
	GetPosts() (posts *[]models.Post, err error)
}
