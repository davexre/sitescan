package webhandler

import (
	"github.com/davexre/sitescan/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

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
