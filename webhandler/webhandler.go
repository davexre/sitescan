package webhandler

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

var (
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
		fmt.Printf("ERROR: URL must begin with http or https: <%s>\n", u)
		return errors.New(fmt.Sprintf("ERROR: URL must begin with http or https: <%s>\n", u))
	case url.Host == "":
		fmt.Printf("ERROR: URL has no host specified: <%s>\n", u)
		return errors.New(fmt.Sprintf("ERROR: URL has no host specified: <%s>\n", u))
	default:
		return nil
	}

}

func HttpHandler(url, user, pass string) (*http.Response, error) {

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if user != "" || pass != "" {
		req.SetBasicAuth(user, pass)
	}

	return (Client.Do(req))
}
