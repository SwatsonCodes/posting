package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
	"github.com/swatsoncodes/very-nice-website/db"
	"github.com/swatsoncodes/very-nice-website/models"
)

const cowsay string = `
 ____________
< GO AWAY <3 >
 ------------
        \   ^__^
         \  (oo)\_______
            (__)\       )\/\
                ||----w |
                ||     ||
`

type Poster struct {
	AllowedSender           string
	TwilioAccountID         string
	MaxRequestBodySizeBytes int64
	DB                      db.PostsDB
}

func GoAway(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, cowsay)
}

func (poster Poster) CreatePost(w http.ResponseWriter, r *http.Request) {
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

	// TODO: verify that post ID corresponds to a message ID via Twilio API
	if err := poster.DB.PutPost(*post); err != nil {
		log.WithError(err).Error("failed to put post to DB")
		http.Error(w, "unable to save post", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (poster Poster) GetPosts(w http.ResponseWriter, r *http.Request) {
	posts, err := poster.DB.GetPosts()
	if err != nil {
		log.WithError(err).Error("failed to get posts from db")
		http.Error(w, "unable to retrieve posts", http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(*posts)
	if err != nil {
		log.WithError(err).Error("failed to marshal posts to json")
		http.Error(w, "unable to retrieve posts", http.StatusInternalServerError)
		return
	}

	// TODO: don't just serve raw json as response
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

func (poster Poster) IsRequestAuthorized(r *http.Request) bool {
	err := r.ParseForm()
	if err != nil {
		log.WithError(err).Error("failed to parse form body")
		return false
	}

	if accountID, aOK := (r.PostForm)["AccountSid"]; aOK {
		if sender, sOK := (r.PostForm)["From"]; sOK {
			return accountID[0] == poster.TwilioAccountID && sender[0] == poster.AllowedSender
		}
		return false
	}
	return false
}
