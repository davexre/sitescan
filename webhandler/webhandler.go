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

var (
	// Client defines which HTTP interface will be used by HTTPHandler. By default, this is
	// set to http.Client{} as part of the init function, but it can be changed to provide
	// a mock HTTP response for testing purposes
	Client HTTPClient
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
		fmt.Printf("ERROR: invalid URL: <%s>\n", u)
		fmt.Printf("%v\n", err)
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
