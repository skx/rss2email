package main

import (
	"bytes"
	"strings"
	"testing"
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
		"only exists to show help",
		"rss2email help config",
	}

	// The text written to stdout
	output := out.(*bytes.Buffer).String()

	for _, txt := range expected {
		if !strings.Contains(output, txt) {
			t.Fatalf("Failed to find expected output")
		}
	}
}
