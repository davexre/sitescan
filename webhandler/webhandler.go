package webhandler

import (
	"fmt"
	"net/http"
	"net/url"
)

// HTTPClient interface will allow for substituting a mock HTTP client for testing purposes
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// MockClient can be used for mocking an HTTPClient response
type MockClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

// Do satisfies the interface's requriement for a Do function, and returns the results
// of the function pointed to by GetDoFunc below.
func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	return GetDoFunc(req)
}

var (
	// Client defines which HTTP interface will be used by HTTPHandler. By default, this is
	// set to http.Client{} as part of the init function, but it can be changed to provide
	// a mock HTTP response for testing purposes
	Client HTTPClient

	// GetDoFunc fetches the mock client's `Do` func - essentially, you can set this
	// variable to an anonymous function within the test that receives an http.Request
	// pointer, and returns an http.Response pointer and/or an error
	GetDoFunc func(req *http.Request) (*http.Response, error)
)

func init() {
	Client = &http.Client{}
}

// ValidateURL will double check a given string to ensure that it's actually a valid
// URL and will highlight any problems with it.
func ValidateURL(u string) error {

	url, err := url.Parse(u)
	switch {
	case err != nil:
		return err
	case url.Scheme == "" || (url.Scheme != "http" && url.Scheme != "https"):
		return fmt.Errorf("ERROR: URL must begin with http or https: <%s>", u)
	case url.Host == "":
		return fmt.Errorf("ERROR: URL has no host specified: <%s>", u)
	default:
		return nil
	}

}

// HTTPHandler retrieves a given URL, and can support basic HTTP authentication. Keeping this
// code separated in a handler function allows for easier testing of several other pieces.
func HTTPHandler(url, user, pass string) (*http.Response, error) {

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if user != "" || pass != "" {
		req.SetBasicAuth(user, pass)
	}

	return (Client.Do(req))
}
