package main

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
	"github.com/swatsoncodes/posting/db"
	"github.com/swatsoncodes/posting/models"
)

type mockPostsDB struct{ shouldErr bool }

func (m mockPostsDB) PutPost(post models.Post) (err error) {
	if m.shouldErr {
		return errors.New("i am a failure")
	}
	return
}

func (m mockPostsDB) GetPosts(offset, limit int) (posts *[]models.Post, isMore bool, err error) {
	if m.shouldErr {
		return nil, true, errors.New("i am a failure")
	}
	return &[]models.Post{models.Post{}}, true, nil
}

type mockUploader struct{ shouldErr bool }

func (m mockUploader) UploadMedia(media io.Reader) (string, error) {
	if m.shouldErr {
		return "", errors.New("i am bad at uploading media")
	}
	return "http://www.example.com", nil
}

func makeFormData(fields map[string]string, files map[string][]byte) (io.Reader, string) {
	bod := &bytes.Buffer{}
	writer := multipart.NewWriter(bod)
	defer writer.Close()
	for k, v := range fields {
		writer.WriteField(k, v)
	}
	for k, f := range files {
		w, _ := writer.CreateFormFile(k, k)
		io.Copy(w, bytes.NewReader(f))
	}
	return bod, writer.FormDataContentType()
}

func TestCreatePost(t *testing.T) {
	var happyDB db.PostsDB = mockPostsDB{shouldErr: false}
	var sadDB db.PostsDB = mockPostsDB{shouldErr: true}
	var happyUploader models.Uploader = mockUploader{shouldErr: false}
	var sadUploader models.Uploader = mockUploader{shouldErr: true}
	testcases := []struct {
		fields     map[string]string
		files      map[string][]byte
		db         *db.PostsDB
		uploader   models.Uploader
		statusCode int
	}{
		{
			map[string]string{
				"Body": "hello",
			},
			nil,
			&happyDB,
			happyUploader,
			http.StatusOK,
		},
		{
			nil,
			map[string][]byte{
				"Pics": []byte("a"),
			},
			&happyDB,
			happyUploader,
			http.StatusOK,
		},
		{
			map[string]string{
				"Body": "hello",
			},
			map[string][]byte{
				"Pics": []byte("a"),
			},
			&happyDB,
			happyUploader,
			http.StatusOK,
		},
		{
			nil,
			nil,
			&happyDB,
			happyUploader,
			http.StatusBadRequest,
		},
		{
			map[string]string{
				"Body": "hello",
			},
			nil,
			&sadDB,
			happyUploader,
			http.StatusInternalServerError,
		},
		{
			map[string]string{
				"Body": "hello",
			},
			nil,
			&happyDB,
			sadUploader,
			http.StatusOK,
		},
		{
			nil,
			map[string][]byte{
				"Pics": []byte("a"),
			},
			&happyDB,
			sadUploader,
			http.StatusInternalServerError,
		},
		{
			map[string]string{
				"Body": "hello",
			},
			nil,
			&sadDB,
			sadUploader,
			http.StatusInternalServerError,
		},
	}

	for _, testcase := range testcases {
		poster := Poster{
			DB:            testcase.db,
			Uploader:      testcase.uploader,
			PostsTemplate: template.Must(template.ParseFiles("templates/posts.html")),
		}
		bod, ct := makeFormData(testcase.fields, testcase.files)
		req, _ := http.NewRequest(http.MethodPost, "/test", bod)
		req.Header.Set("Content-Type", ct)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(poster.CreatePost)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, testcase.statusCode, rr.Code)
	}

}

func TestGetPosts(t *testing.T) {
	var happyDB db.PostsDB = mockPostsDB{shouldErr: false}
	var sadDB db.PostsDB = mockPostsDB{shouldErr: true}
	testcases := []struct {
		db         *db.PostsDB
		statusCode int
	}{
		{&happyDB, http.StatusOK},
		{&sadDB, http.StatusInternalServerError},
	}

	for _, testcase := range testcases {
		poster := Poster{
			DB:            testcase.db,
			PostsTemplate: template.Must(template.ParseFiles("templates/posts.html")),
		}
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(poster.GetPosts)
		handler.ServeHTTP(rr, req)
		assert.Equal(t, testcase.statusCode, rr.Code)
	}
}
