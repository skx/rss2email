package main

import (
	"bytes"
	"flag"
	"strings"
	"testing"
)

func TestDefaultTemplate(t *testing.T) {

	bak := out
	out = new(bytes.Buffer)
	defer func() { out = bak }()

	s := listDefaultTemplateCmd{}

	//
	// Call the Arguments function for coverage.
	//
	flags := flag.NewFlagSet("test", flag.ContinueOnError)
	s.Arguments(flags)

	//
	// Call the handler.
	//
	s.Execute([]string{})

	//
	// Look for some lines in the output
	//
	expected := []string{
		"X-RSS-Link: {{.Link}}",
		"Content-Type: multipart/alternative;",
		"the default template which is used by default to generate emails",
	}

	// The text written to stdout
	output := out.(*bytes.Buffer).String()

	for _, txt := range expected {
		if !strings.Contains(output, txt) {
			t.Fatalf("Failed to find expected output")
		}
	}
}
