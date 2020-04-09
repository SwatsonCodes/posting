package main

import (
	"net/http"
	"testing"
)

func TestIsPosterAuthorized(t *testing.T) {
	app := niceApp{
		TwilioAccountID: "abc123",
		AllowedSender:   "+15558675309",
	}

	// TODO: add test case for mangled body
	testcases := []struct {
		postForm   *map[string][]string
		authorized bool
	}{
		{&map[string][]string{
			"AccountSid": []string{app.TwilioAccountID},
			"From":       []string{app.AllowedSender},
		}, true},
		{&map[string][]string{
			"AccountSid": []string{"bad"},
			"From":       []string{app.AllowedSender},
		}, false},
		{&map[string][]string{
			"AccountSid": []string{app.TwilioAccountID},
			"From":       []string{"bad"},
		}, false},
		{&map[string][]string{
			"AccountSid": []string{"bad"},
			"From":       []string{"bad"},
		}, false},
		{&map[string][]string{
			"From": []string{app.AllowedSender},
		}, false},
		{&map[string][]string{
			"AccountSid": []string{app.TwilioAccountID},
		}, false},
		{&map[string][]string{
			"AccountSid": []string{app.TwilioAccountID},
		}, false},
	}

	for _, testcase := range testcases {
		req, err := http.NewRequest(http.MethodGet, "/test", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.PostForm = *testcase.postForm
		isAuthorized := app.isRequestAuthorized(req)
		if isAuthorized != testcase.authorized {
			t.Errorf("expected '%v' but got '%v'", testcase.authorized, isAuthorized)
		}
	}
}
