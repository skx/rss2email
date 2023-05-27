package main

import (
	"os"
	"testing"

	"github.com/skx/rss2email/configfile"
)

func TestAdd(t *testing.T) {

	// Create an instance of the command, and setup a default
	// configuration file

	content := `# Comment here
https://example.org/
https://example.net/
 - foo: bar
`
	data := []byte(content)
	tmpfile, err := os.CreateTemp("", "example")
	if err != nil {
		t.Fatalf("Error creating temporary file")
	}

	if _, err = tmpfile.Write(data); err != nil {
		t.Fatalf("Error writing to config file")
	}
	if err = tmpfile.Close(); err != nil {
		t.Fatalf("Error creating temporary file")
	}

	add := addCmd{}
	add.Arguments(nil)
	config := configfile.NewWithPath(tmpfile.Name())
	add.config = config

	// Add an entry
	add.Execute([]string{"https://blog.steve.fi/index.rss"})

	// Open the file and confirm it has the content we expect
	x := configfile.NewWithPath(tmpfile.Name())
	entries, err := x.Parse()
	if err != nil {
		t.Fatalf("Error parsing written file")
	}

	found := false
	for _, entry := range entries {
		if entry.URL == "https://blog.steve.fi/index.rss" {
			found = true
		}
	}

	if !found {
		t.Fatalf("Adding the entry failed")
	}

	os.Remove(tmpfile.Name())
}
