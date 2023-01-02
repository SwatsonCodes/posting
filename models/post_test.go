package models

import (
	"mime/multipart"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func makeForm(fields map[string][]string) multipart.Form {
	f := multipart.Form{
		Value: fields,
	}
	// TODO: figure out a good way to add files
	return f
}

func TestDecodePost(t *testing.T) {
	testcases := []struct {
		form      multipart.Form
		shouldErr bool
		body      string
	}{
		{
			makeForm(map[string][]string{"Body": []string{"hello"}}),
			false,
			"hello",
		},
		{
			makeForm(map[string][]string{"Body": []string{"hello", "hi"}}),
			false,
			"hello",
		},
		{
			makeForm(map[string][]string{"Body": []string{}}),
			true,
			"",
		},
		{
			makeForm(map[string][]string{"Body": []string{""}}),
			true,
			"",
		},
		{
			makeForm(map[string][]string{}),
			true,
			"",
		},
		{
			makeForm(map[string][]string{"bad": []string{"field"}}),
			true,
			"",
		},
	}

	for _, testcase := range testcases {
		post, err := ParsePost(testcase.form)
		if testcase.shouldErr {
			assert.Error(t, err)
			assert.Nil(t, post)
			continue
		}

		assert.Nil(t, err)
		_, err = uuid.Parse(post.ID)
		assert.NoError(t, err)
		assert.Equal(t, testcase.body, post.Body)
		assert.NotNil(t, post.CreatedAt)
	}
}
