package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/skx/rss2email/configfile"
)

func TestExport(t *testing.T) {

	// Replace the STDIO handle
	bak := out
	out = new(bytes.Buffer)
	defer func() { out = bak }()

	// Create a simple configuration file
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

	if _, err := tmpfile.Write(data); err != nil {
		t.Fatalf("Error writing to config file")
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("Error creating temporary file")
	}

	// Create an instance of the command, and setup the config file
	ex := exportCmd{}
	ex.Arguments(nil)
	config := configfile.NewWithPath(tmpfile.Name())
	ex.config = config

	// Run the export
	ex.Execute([]string{})

	//
	// Look for some lines in the output
	//
	expected := []string{
		"Feed Export",
		"xmlUrl=\"https://example.org/\"",
		"</opml>",
	}

	// The text written to stdout
	output := out.(*bytes.Buffer).String()

	for _, txt := range expected {
		if !strings.Contains(output, txt) {
			t.Fatalf("Failed to find expected output")
		}
	}

	// Cleanup
	os.Remove(tmpfile.Name())
}
