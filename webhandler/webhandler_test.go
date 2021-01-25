package webhandler

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidateURL(t *testing.T) {
	assert := assert.New(t)

	var tests = []struct {
		input    string
		expected bool
	}{
		{"", false},
		{"someurl.com", false},
		{"file://somefile", false},
		{"http:some/file/path", false},
		{"\"http://www.somehost.com/path\"", false},
		{"http://www.somehost.com/path", true},
		{"https://www.somehost.com/path", true},
	}
	for _, test := range tests {
		assert.Equal(ValidateURL(test.input), test.expected)
	}

}
