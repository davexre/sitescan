package mocks

import (
	"net/http"
)

// MockClient can be used for mocking an HTTPClient response
type MockClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

var (
	// GetDoFunc fetches the mock client's `Do` func - essentially, you can set this
	// variable to an anonymous function within the test that receives an http.Request
	// pointer, and returns an http.Response pointer and/or an error
	GetDoFunc func(req *http.Request) (*http.Response, error)
)

// Do satisfies the interface's requriement for a Do function, and returns the results
// of the function pointed to by GetDoFunc below.
func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	return GetDoFunc(req)
}
