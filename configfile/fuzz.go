// fuzz_test.go - Simple test-cases

package configfile

import (
	"io/ioutil"
	"os"
	"strings"
)

// Fuzz is used for fuzz-testing
func Fuzz(data []byte) int {

	// Create a temporary file
	tmpfile, _ := ioutil.TempFile("", "example")

	// Cleanup when we're done
	defer os.Remove(tmpfile.Name())

	// Write it out
	tmpfile.Write(data)
	tmpfile.Close()

	// Create a new config-reader
	c := New()
	c.path = tmpfile.Name()

	// Parse, looking for errors
	_, err := c.Parse()
	if err != nil {

		// This is a known error, we expect to get
		if !strings.Contains(err.Error(), "option outside a URL") {
			panic(err)
		}
	}

	return 1
}
