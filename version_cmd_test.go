package main

import (
	"bytes"
	"flag"
	"runtime"
	"testing"
)

func TestVersion(t *testing.T) {
	bak := out
	out = new(bytes.Buffer)
	defer func() { out = bak }()

	//
	// Expected
	//
	expected := "unreleased\n"

	s := versionCmd{}

	//
	// Call the Arguments function for coverage.
	//
	flags := flag.NewFlagSet("test", flag.ContinueOnError)
	s.Arguments(flags)

	//
	// Call the handler.
	//
	s.Execute([]string{})

	if out.(*bytes.Buffer).String() != expected {
		t.Errorf("Expected '%s' received '%s'", expected, out)
	}
}

func TestVersionVerbose(t *testing.T) {
	bak := out
	out = new(bytes.Buffer)
	defer func() { out = bak }()

	//
	// Expected
	//
	expected := "unreleased\nBuilt with " + runtime.Version() + "\n"

	s := versionCmd{verbose: true}

	//
	// Call the handler.
	//
	s.Execute([]string{})

	if out.(*bytes.Buffer).String() != expected {
		t.Errorf("Expected '%s' received '%s'", expected, out)
	}
}
