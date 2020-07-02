package poster

// PostsDB TODO
type PostsDB interface {
	PutPost(post Post) (err error)
	GetPosts(offset, limit int) (posts *[]Post, isMore bool, err error)
}
