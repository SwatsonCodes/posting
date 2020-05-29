package main

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"text/template"

	log "github.com/sirupsen/logrus"
	"github.com/swatsoncodes/posting/db"
	"github.com/swatsoncodes/posting/models"
	"github.com/swatsoncodes/posting/upstream/imgur"
	"github.com/swatsoncodes/posting/upstream/twilio"
)

const cowsay string = `
 _____
< no  >
 -----
        \   ^__^
         \  (oo)\_______
            (__)\       )\/\
                ||----w |
                ||     ||
`
const postsTemplate string = "posts.html"

type Poster struct {
	AllowedSender   string
	TwilioAuthToken string
	ImgurUploader   imgur.Uploader
	DB              *db.PostsDB
	PageSize        int
	PostsTemplate   *template.Template
}

func NewPoster(allowedSender, twilioAuthToken, imgurClientID, templatesPath string, pageSize int, postsDB *db.PostsDB) (*Poster, error) {
	template, err := template.ParseFiles(filepath.Join(templatesPath, postsTemplate))
	if err != nil {
		return nil, err
	}
	return &Poster{allowedSender, twilioAuthToken, imgur.Uploader{ClientID: imgurClientID}, postsDB, pageSize, template}, nil
}

func GoAway(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, cowsay)
}

func (poster Poster) CreatePost(w http.ResponseWriter, r *http.Request) {
	// TODO: reply with twilio-friendly response
	err := r.ParseForm()
	if err != nil {
		log.WithError(err).Warn("failed to parse form body")
		http.Error(w, "unable to parse request form body", http.StatusBadRequest)
		return
	}

	post, err := models.ParsePost(&r.PostForm)
	if err != nil {
		log.WithError(err).Warn("got bad post")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := post.RehostImagesOnImgur(poster.ImgurUploader); err != nil {
		log.WithError(err).Error("failed to upload images to imgur")
		http.Error(w, "unable to save post", http.StatusInternalServerError)
		return
	}

	if err := (*poster.DB).PutPost(*post); err != nil {
		log.WithError(err).Error("failed to put post to DB")
		http.Error(w, "unable to save post", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
}

func (poster Poster) GetPosts(w http.ResponseWriter, r *http.Request) {
	pageNum := getPageNum(r)
	posts, isMore, err := (*poster.DB).GetPosts(pageNum*poster.PageSize, poster.PageSize)
	if err != nil {
		log.WithError(err).Error("failed to get posts from db")
		http.Error(w, "unable to retrieve posts", http.StatusInternalServerError)
		return
	}

	nextPage := -1
	if isMore {
		nextPage = pageNum + 1
	}
	templatePayload := struct {
		Posts              []models.Post
		NextPage, PrevPage int
	}{
		*posts,
		nextPage,
		pageNum - 1,
	}

	if err = poster.PostsTemplate.Execute(w, templatePayload); err != nil {
		log.WithError(err).Error(err.Error())
		http.Error(w, "unable to render posts html", http.StatusInternalServerError)
		return
	}
}

func getPageNum(r *http.Request) (offset int) {
	if page, ok := r.URL.Query()["page"]; ok {
		if len(page) == 0 {
			return
		}
		if p, err := strconv.Atoi(page[0]); err == nil {
			if p < 0 {
				return
			}
			return p
		}
	}
	return
}

func (poster Poster) IsRequestAuthorized(r *http.Request) bool {
	return twilio.IsRequestSigned(r, poster.TwilioAuthToken) && r.PostForm.Get("From") == poster.AllowedSender
}
