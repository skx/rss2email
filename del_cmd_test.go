package main

import (
	"os"
	"testing"

	"github.com/skx/rss2email/configfile"
)

func TestDel(t *testing.T) {

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

	del := delCmd{}
	del.Arguments(nil) // only for coverage

	config := configfile.NewWithPath(tmpfile.Name())
	del.config = config

	// Delete an entry
	del.Execute([]string{"https://example.net/"})

	// Open the file and confirm only one entry.
	x := configfile.NewWithPath(tmpfile.Name())
	entries, err := x.Parse()
	if err != nil {
		t.Fatalf("Error parsing written file")
	}

	if len(entries) != 1 {
		t.Fatalf("Expected only one entry")
	}
	if entries[0].URL != "https://example.org/" {
		t.Fatalf("Wrong item deleted")
	}
	if len(entries[0].Options) != 0 {
		t.Fatalf("We have orphaned parameters")
	}

	os.Remove(tmpfile.Name())
}
