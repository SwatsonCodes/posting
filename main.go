package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/swatsoncodes/posting/db"
	"github.com/swatsoncodes/posting/middleware"
)

// TODO: make these configurable
const bodySizeLimit middleware.RequestBodyLimitBytes = 256 << 20 // 256MiB
const collectionName = "posts"
const pageSize = 5

func main() {
	var imgurClientID, gcloudID, username, password, port string
	var ok bool
	templatesPath := "templates"
	log.SetFormatter(&log.JSONFormatter{DisableHTMLEscape: true})
	log.Info("hello")

	if imgurClientID, ok = os.LookupEnv("IMGUR_CLIENT_ID"); !ok {
		log.Fatal("env var IMGUR_CLIENT_ID not set")
	}
	if gcloudID, ok = os.LookupEnv("GCLOUD_PROJECT_ID"); !ok {
		log.Fatal("env var GCLOUD_PROJECT_ID not set")
	}
	if username, ok = os.LookupEnv("BASIC_AUTH_USERNAME"); !ok {
		log.Fatal("env var BASIC_AUTH_USERNAME not set") // TODO: consider making auth optional
	}
	if password, ok = os.LookupEnv("BASIC_AUTH_PASSWORD"); !ok {
		log.Fatal("env var BASIC_AUTH_PASSWORD not set")
	}
	if port, ok = os.LookupEnv("PORT"); !ok {
		port = "8008"
	}

	postsDB, err := db.NewFirestoreClient(gcloudID, collectionName)
	if err != nil {
		log.WithError(err).Fatal("failed to initialize db")
	}
	var pdb db.PostsDB = postsDB
	poster, err := NewPoster(imgurClientID, templatesPath, pageSize, int64(bodySizeLimit), &pdb)
	router := mux.NewRouter().StrictSlash(true)
	auth := middleware.BasicAuthorizer{[]byte(username), []byte(password)}

	router.Handle("/posts",
		bodySizeLimit.LimitRequestBody( // guard against giant posts
			auth.BasicAuth( // make sure posters are authorized
				http.HandlerFunc(poster.CreatePost)))).
		Methods(http.MethodPost)
	router.HandleFunc("/posts", poster.GetPosts).Methods(http.MethodGet)
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/posts", http.StatusMovedPermanently)
	}).Methods(http.MethodGet)
	router.Handle("/new",
		auth.BasicAuth( // require auth on new Post upload page
			func(w http.ResponseWriter, r *http.Request) {
				http.ServeFile(w, r, "static/new")
			}),
	)
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("static/")))

	router.Use(middleware.LogRequest)
	if env := os.Getenv("POSTER_ENV"); env == "DEV" {
		router.Use(middleware.LogRequestBody)
	}
	log.Infof("serving on port %s", port)
	http.ListenAndServe(fmt.Sprintf(":%s", port), router)
}
