package main

import (
	"bytes"
	"fmt"
	"github.com/davexre/sitescan/mocks"
	"github.com/davexre/sitescan/webhandler"
	"github.com/davexre/synceddata"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"testing"
)

func TestCompareMaps(t *testing.T) {
	// implement the map variables
	sitename := "X"
	var map1 = make(map[string]string)
	var map2 = make(map[string]string)

	map1["string1"] = "string1map"
	map1["string2"] = "string2map"
	map2["string1"] = "string1map"
	map2["string3"] = "string3map"

	expectedOutput := []byte("Files/directories only at X:\n============================\n\nstring2\n\n\n")

	tmpfile, err := ioutil.TempFile("", "output")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	oldStdout := os.Stdout
	os.Stdout = tmpfile

	compareMaps(&map1, &map2, sitename)

	os.Stdout = oldStdout

	if _, err := tmpfile.Seek(0, 0); err != nil {
		log.Fatal(err)
	}

	output, err := ioutil.ReadFile(tmpfile.Name())
	if err != nil {
		log.Fatal(err)
	}

	if err := tmpfile.Close(); err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, output[:], expectedOutput[:])
}

// Test site structure
// someurl.com/
//             "Name"
//             dir1/
//             dir1/file11
//             dir2/
//             dir2/file21
//             file3
func TestWalkLink(t *testing.T) {

	response := ""
	url := "http://someurl.com/"
	var testmap = make(map[string]string)
	var counter synceddata.Counter

	webhandler.Client = &mocks.MockClient{}
	mocks.GetDoFunc = func(req *http.Request) (*http.Response, error) {
		urlReq := req.URL.String()
		switch {
		case urlReq == url:
			response = `<a href="name">Name</a><a href="dir1/">dir1</a><a href="dir2/">dir2/</a><a href="file3.mp4">file3.mp4</a>`
		case urlReq == url+"dir1/":
			response = `<a href="file11.mp3">file11.mp3</a>`
		case urlReq == url+"dir2/":
			response = `<a href="file21.jpg">file21.jpg</a>`
		default:
			fmt.Printf("TestWalkLink - Invalid test URL - exiting\n")
			os.Exit(1)
		}
		r := ioutil.NopCloser(bytes.NewReader([]byte(response)))
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}

	walkLink(url, "", "", &testmap, "", "", &counter)

	/// now, check our map!
	assert.Equal(t, testmap["dir1/"], "dir1/", "map entry incorrect")
	assert.Equal(t, testmap["dir1/file11.mp3"], "dir1/file11.mp3", "map entry incorrect")
	assert.Equal(t, testmap["dir2/"], "dir2/", "map entry incorrect")
	assert.Equal(t, testmap["dir2/file21.jpg"], "dir2/file21.jpg", "map entry incorrect")
	assert.Equal(t, testmap["file3.mp4"], "file3.mp4", "map entry incorrect")

}
