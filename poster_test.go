package main

import (
	"net/http"
	"testing"
)

func TestIsPosterAuthorized(t *testing.T) {
	poster := Poster{
		TwilioAccountID: "abc123",
		AllowedSender:   "+15558675309",
	}

	// TODO: add test case for mangled body
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
		req, err := http.NewRequest(http.MethodGet, "/test", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.PostForm = *testcase.postForm
		isAuthorized := poster.IsRequestAuthorized(req)
		if isAuthorized != testcase.authorized {
			t.Errorf("expected '%v' but got '%v'", testcase.authorized, isAuthorized)
		}
	}
}
