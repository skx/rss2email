package main

import (
	"bytes"
	"flag"
	"os"
	"strings"
	"testing"

	"github.com/skx/rss2email/configfile"
)

func TestConfig(t *testing.T) {

	bak := out
	out = new(bytes.Buffer)
	defer func() { out = bak }()

	s := configCmd{}

	//
	// Call the handler.
	//
	s.Execute([]string{})

	//
	// Look for some lines in the output
	//
	expected := []string{
		"release 3.x",
		"RSS2Email is a simple",
	}

	// The text written to stdout
	output := out.(*bytes.Buffer).String()

	for _, txt := range expected {
		if !strings.Contains(output, txt) {
			t.Fatalf("Failed to find expected output")
		}
	}
}

// TestMissingConfig ensures we see a warning if the configuration
// file is not present.
func TestMissingConfig(t *testing.T) {

	// Create a temporary file, so we get a name of something
	// that doesn't exist
	tmpfile, err := os.CreateTemp("", "example")
	if err != nil {
		t.Fatalf("failed to create temporary file")
	}
	os.Remove(tmpfile.Name())

	//
	// Setup a configuration-file, which doesn't exist.
	//
	s := configCmd{}
	flags := flag.NewFlagSet("test", flag.ContinueOnError)
	s.Arguments(flags)
	config := configfile.NewWithPath(tmpfile.Name())
	s.config = config

	// Get the documentation
	_, doc := s.Info()

	//
	// Look for some lines in the output
	//
	expected := []string{
		tmpfile.Name(),

		// known configuration options
		"include ",
		"include-title",
		"exclude ",
		"exclude-title",
		"exclude-older",
	}

	for _, txt := range expected {
		if !strings.Contains(doc, txt) {
			t.Fatalf("Failed to find expected output: %s", txt)
		}
	}
}
