package main

import (
	"bytes"
	"flag"
	"os"
	"strings"
	"testing"

	"github.com/skx/rss2email/configfile"
)

// TestList confirms that listing the feed-list works as expected
func TestList(t *testing.T) {

	bak := out
	out = new(bytes.Buffer)
	defer func() { out = bak }()

	// Create an instance of the command, and setup a default
	// configuration file

	content := `# Comment here
https://example.org/
https://example.net/index.rss
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

	list := listCmd{}
	flags := flag.NewFlagSet("test", flag.ContinueOnError)
	list.Arguments(flags)
	config := configfile.NewWithPath(tmpfile.Name())
	list.config = config

	ret := list.Execute([]string{})
	if ret != 0 {
		t.Fatalf("unexpected error running list")
	}

	output := out.(*bytes.Buffer).String()

	// We should have two URLs
	if !strings.Contains(output, "https://example.org/") {
		t.Errorf("List didn't contain expected output")
	}
	if !strings.Contains(output, "https://example.net/index.rss") {
		t.Errorf("List didn't contain expected output")
	}

	// We should not have comments, or parameters
	if strings.Contains(output, "foo") {
		t.Errorf("We found a parameter we didn't expect")
	}
	if strings.Contains(output, "#") {
		t.Errorf("We found a comment we didn't expect")
	}

	os.Remove(tmpfile.Name())
}
