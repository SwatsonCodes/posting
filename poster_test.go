package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/swatsoncodes/very-nice-website/db"
	"github.com/swatsoncodes/very-nice-website/models"
)

type mockPostsDB struct{ shouldErr bool }

func (m mockPostsDB) PutPost(post models.Post) (err error) {
	if m.shouldErr {
		return errors.New("i am a failure")
	}
	return
}

func (m mockPostsDB) GetPosts() (posts *[]models.Post, err error) {
	if m.shouldErr {
		return nil, errors.New("i am a failure")
	}
	return &[]models.Post{models.Post{}}, nil
}

func TestGoAway(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GoAway)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, rr.Body.String(), cowsay)
	assert.Equal(t, rr.Code, http.StatusOK)
}

func TestCreatePost(t *testing.T) {
	happyDB := mockPostsDB{shouldErr: false}
	sadDB := mockPostsDB{shouldErr: true}
	testcases := []struct {
		body        string
		contentType string
		db          db.PostsDB
		statusCode  int
	}{
		{
			url.Values{
				"SmsSid":   []string{"abc123"},
				"Body":     []string{"hello"},
				"NumMedia": []string{"0"},
			}.Encode(),
			"application/x-www-form-urlencoded",
			happyDB,
			http.StatusCreated,
		},
		{
			"bad bod",
			"application/x-www-form-urlencoded",
			happyDB,
			http.StatusBadRequest,
		},
		{
			url.Values{
				"SmsSid":   []string{"abc123"},
				"Body":     []string{"hello"},
				"NumMedia": []string{"0"},
			}.Encode(),
			"text/plain; boundary=",
			happyDB,
			http.StatusBadRequest,
		},
		{
			url.Values{
				"SmsSid":   []string{"abc123"},
				"Body":     []string{"hello"},
				"NumMedia": []string{"0"},
			}.Encode(),
			"application/x-www-form-urlencoded",
			sadDB,
			http.StatusInternalServerError,
		},
	}

	for _, testcase := range testcases {
		poster := Poster{DB: testcase.db}
		req, _ := http.NewRequest(http.MethodPost, "/test", strings.NewReader(testcase.body))
		req.Header.Set("Content-Type", testcase.contentType)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(poster.CreatePost)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, testcase.statusCode, rr.Code)
	}

}

func TestGetPosts(t *testing.T) {
	happyDB := mockPostsDB{shouldErr: false}
	sadDB := mockPostsDB{shouldErr: true}
	testcases := []struct {
		db         db.PostsDB
		statusCode int
	}{
		{happyDB, http.StatusOK},
		{sadDB, http.StatusInternalServerError},
	}

	for _, testcase := range testcases {
		poster := Poster{DB: testcase.db}
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(poster.GetPosts)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, testcase.statusCode, rr.Code)
		if testcase.statusCode == http.StatusOK {
			bod, _ := ioutil.ReadAll(rr.Body)
			var posts []models.Post
			err := json.Unmarshal(bod, &posts)
			assert.Nil(t, err)
			assert.NotNil(t, posts)
		}
	}
}

func TestIsPosterAuthorized(t *testing.T) {
	poster := Poster{
		TwilioAccountID: "abc123",
		AllowedSender:   "+15558675309",
	}

	testcases := []struct {
		postForm   *map[string][]string
		authorized bool
	}{
		{&map[string][]string{
			"AccountSid": []string{poster.TwilioAccountID},
			"From":       []string{poster.AllowedSender},
		}, true},
		{&map[string][]string{
			"AccountSid": []string{"bad"},
			"From":       []string{poster.AllowedSender},
		}, false},
		{&map[string][]string{
			"AccountSid": []string{poster.TwilioAccountID},
			"From":       []string{"bad"},
		}, false},
		{&map[string][]string{
			"AccountSid": []string{"bad"},
			"From":       []string{"bad"},
		}, false},
		{&map[string][]string{
			"From": []string{poster.AllowedSender},
		}, false},
		{&map[string][]string{
			"AccountSid": []string{poster.TwilioAccountID},
		}, false},
		{&map[string][]string{
			"AccountSid": []string{poster.TwilioAccountID},
		}, false},
	}

	for _, testcase := range testcases {
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.PostForm = *testcase.postForm
		isAuthorized := poster.IsRequestAuthorized(req)
		if isAuthorized != testcase.authorized {
			t.Errorf("expected '%v' but got '%v'", testcase.authorized, isAuthorized)
		}
	}
}
