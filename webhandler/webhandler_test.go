package webhandler

import (
	"bytes"
	"github.com/davexre/sitescan/mocks"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
)

func init() {
	Client = &mocks.MockClient{}
}

func TestValidateURL(t *testing.T) {
	assert := assert.New(t)

	var tests = []struct {
		input       string
		expectError bool
	}{
		{"", true},
		{"someurl.com", true},
		{"file://somefile", true},
		{"http:some/file/path", true},
		{"\"http://www.somehost.com/path\"", true},
		{"http://www.somehost.com/path", false},
		{"https://www.somehost.com/path", false},
	}
	for _, test := range tests {
		if test.expectError {
			assert.NotNil(ValidateURL(test.input))
		} else {
			assert.Nil(ValidateURL(test.input))
		}
	}

}

func TestHTTPHandler(t *testing.T) {
	assert := assert.New(t)

	var tests = []struct {
		url         string
		expectError bool
		response    string
	}{
		{"http://testurl.com", false, `<a href="dir1">dir1</a><a href=dir2>dir2</a>`},
		{"\"http://bogus.com\"", true, `<a href="dir1">dir1</a><a href=dir2>dir2</a>`},
	}
	for _, test := range tests {

		r := ioutil.NopCloser(bytes.NewReader([]byte(test.response)))
		mocks.GetDoFunc = func(*http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       r,
			}, nil
		}

		res, err := HTTPHandler(test.url, "", "")
		if test.expectError {
			assert.NotNil(err)
			assert.Nil(res)
		} else {
			assert.NotNil(res)
			assert.Nil(err)
		}
	}
}
