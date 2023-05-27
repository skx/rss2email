package main

import (
	"flag"
	"github.com/skx/rss2email/configfile"
	"os"
	"testing"
)

// TestUsage just calls the usage-function for each of our handlers,
// and passes some bogus flags to the arguments handler.
func TestUsage(t *testing.T) {

	add := addCmd{}
	add.Info()
	add.Arguments(flag.NewFlagSet("test", flag.ContinueOnError))

	cron := cronCmd{}
	cron.Info()
	cron.Arguments(flag.NewFlagSet("test", flag.ContinueOnError))

	config := configCmd{}
	config.Info()
	config.Arguments(flag.NewFlagSet("test", flag.ContinueOnError))

	daemon := daemonCmd{}
	daemon.Info()
	daemon.Arguments(flag.NewFlagSet("test", flag.ContinueOnError))

	del := delCmd{}
	del.Info()
	del.Arguments(flag.NewFlagSet("test", flag.ContinueOnError))

	export := exportCmd{}
	export.Info()
	export.Arguments(flag.NewFlagSet("test", flag.ContinueOnError))

	imprt := importCmd{}
	imprt.Info()
	imprt.Arguments(flag.NewFlagSet("test", flag.ContinueOnError))

	list := listCmd{}
	list.Info()
	list.Arguments(flag.NewFlagSet("test", flag.ContinueOnError))

	ldt := listDefaultTemplateCmd{}
	ldt.Info()
	ldt.Arguments(flag.NewFlagSet("test", flag.ContinueOnError))

	seen := seenCmd{}
	seen.Info()
	seen.Arguments(flag.NewFlagSet("test", flag.ContinueOnError))

	unse := unseeCmd{}
	unse.Info()
	unse.Arguments(flag.NewFlagSet("test", flag.ContinueOnError))

	vers := versionCmd{}
	vers.Info()
	vers.Arguments(flag.NewFlagSet("test", flag.ContinueOnError))
}

// TestBrokenConfig is used to test that the commands which assume
// a broken configuration file do so.
func TestBrokenConfig(t *testing.T) {

	data := []byte(`# This is bogus, options must follow URLs
 - foo:bar`)
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

	// Test the various commands.
	a := addCmd{}
	a.config = configfile.NewWithPath(tmpfile.Name())
	res := a.Execute([]string{})
	if res != 1 {
		t.Fatalf("expected error with config file")
	}

	d := delCmd{}
	d.config = configfile.NewWithPath(tmpfile.Name())
	res = d.Execute([]string{})
	if res != 1 {
		t.Fatalf("expected error with config file")
	}

	e := exportCmd{}
	e.config = configfile.NewWithPath(tmpfile.Name())
	res = e.Execute([]string{})
	if res != 1 {
		t.Fatalf("expected error with config file")
	}

	i := importCmd{}
	i.config = configfile.NewWithPath(tmpfile.Name())
	res = i.Execute([]string{})
	if res != 1 {
		t.Fatalf("expected error with config file")
	}

	l := listCmd{}
	l.config = configfile.NewWithPath(tmpfile.Name())
	res = l.Execute([]string{})
	if res != 1 {
		t.Fatalf("expected error with config file")
	}

	// TODO : error-match

	os.Remove(tmpfile.Name())
}
