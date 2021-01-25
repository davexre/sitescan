package main

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
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
