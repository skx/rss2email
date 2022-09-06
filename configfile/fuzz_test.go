//go:build go1.18
// +build go1.18

package configfile

import (
	"os"
	"strings"
	"testing"
)

func FuzzParser(f *testing.F) {
	f.Add([]byte(""))
	f.Add([]byte("https://example.com"))
	f.Add([]byte("https://example.com\r"))
	f.Add([]byte("https://example.com\r\n"))
	f.Add([]byte(`
https://example.com
  - foo:bar
  - bar:baz
https://example.com
  - foo:bar
  - bar:baz`))

	f.Fuzz(func(t *testing.T, input []byte) {
		// Create a temporary file
		tmpfile, _ := os.CreateTemp("", "example")

		// Cleanup when we're done
		defer os.Remove(tmpfile.Name())

		// Write it out
		_, err := tmpfile.Write(input)
		if err != nil {
			t.Fatalf("failed to write temporary file %s", err)
		}

		tmpfile.Close()

		// Create a new config-reader
		c := NewWithPath(tmpfile.Name())

		// Parse, looking for errors
		_, err = c.Parse()
		if err != nil {

			// This is a known error, we expect to get
			if !strings.Contains(err.Error(), "option outside a URL") {
				t.Errorf("Input gave bad error: %s %s\n", input, err)
			}

		}
	})
}
