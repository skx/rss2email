package configfile

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

// Test a file exists
func TestExists(t *testing.T) {

	// Create a temporary file
	tmpfile, err := ioutil.TempFile("", "example")
	if err != nil {
		t.Fatalf("error creating temporary file")
	}

	// Create the config-reader and pass it the
	// name of our temporary file
	conf := New()
	conf.path = tmpfile.Name()

	if !conf.Exists() {
		t.Fatalf("The config file doesn't exist, and it should!")
	}

	// Same again with the different constructor.
	conf2 := NewWithPath(tmpfile.Name())
	if !conf2.Exists() {
		t.Fatalf("The config file doesn't exist, and it should!")
	}

	// Remove it
	os.Remove(conf.path)

	if conf.Exists() {
		t.Fatalf("Config file exists, but we just deleted it!")
	}
	if conf2.Exists() {
		t.Fatalf("Config file exists, but we just deleted it!")
	}

	// Parsing should return an error, when the file doesn't exist
	_, err = conf.Parse()
	if err == nil {
		t.Fatalf("Expected an error parsing a missing file, got none!")
	}
}

// TestHome ensures that the result of Home() is a directory, and that our
// default configuration has that as a prefix.
func TestHome(t *testing.T) {

	config := New()
	home := config.Home()
	fi, err := os.Stat(home)
	if err != nil {
		t.Fatalf("Failed to stat()")
	}

	mode := fi.Mode()

	if !mode.IsDir() {
		t.Fatalf("Home() resulted in a non-directory")
	}

	// Test the path starts beneath home
	path := config.Path()
	if !strings.HasPrefix(path, home) {
		t.Fatalf("config file doesn't seem to exist beneath home")
	}

	//
	// Same again, but unset the environmental variable too
	//
	os.Unsetenv("HOME")
	home = config.Home()
	fi, err = os.Stat(home)
	if err != nil {
		t.Fatalf("Failed to stat()")
	}

	mode = fi.Mode()

	if !mode.IsDir() {
		t.Fatalf("Home() resulted in a non-directory")
	}
}

// TestBasicFile tests parsing a basic file.
func TestBasicFile(t *testing.T) {

	c := ParserHelper(t, `https://example.com/
https://example.net/


https://example.org`)

	out, err := c.Parse()
	if err != nil {
		t.Fatalf("Error parsing file: %v", err)
	}

	if len(out) != 3 {
		t.Fatalf("parsed wrong number of entries, got %d\n%v", len(out), out)
	}

	// All options should have no params
	for _, entry := range out {
		if len(entry.Options) != 0 {
			t.Fatalf("Found entry with unexpected parameters:%s\n", entry.URL)
		}
	}

	os.Remove(c.path)
}

// TestEmptyFile tests parsing an empty file.
func TestEmptyFile(t *testing.T) {

	c := ParserHelper(t, ``)

	out, err := c.Parse()
	if err != nil {
		t.Fatalf("Error parsing file: %v", err)
	}

	if len(out) != 0 {
		t.Fatalf("parsed wrong number of entries, got %d\n%v", len(out), out)
	}

	os.Remove(c.path)
}

// TestEmptyFileComment tests parsing a file empty of everything but comments
func TestEmptyFileComment(t *testing.T) {

	c := ParserHelper(t, `# Comment1
#Comment2`)

	out, err := c.Parse()
	if err != nil {
		t.Fatalf("Error parsing file: %v", err)
	}

	if len(out) != 0 {
		t.Fatalf("parsed wrong number of entries, got %d\n%v", len(out), out)
	}

	os.Remove(c.path)
}

// TestOptions tests parsing a file with one URL with options
func TestOptions(t *testing.T) {

	c := ParserHelper(t, `
http://example.com/
 - foo:bar
 - retry: 7
#Comment2`)

	out, err := c.Parse()
	if err != nil {
		t.Fatalf("Error parsing file: %v", err)
	}

	// One entry
	if len(out) != 1 {
		t.Fatalf("parsed wrong number of entries, got %d\n%v", len(out), out)
	}

	// We should have two options
	if len(out[0].Options) != 2 {
		t.Fatalf("Found wrong number of options, got %d", len(out[0].Options))
	}

	for _, opt := range out[0].Options {
		if opt.Name != "foo" &&
			opt.Name != "retry" {
			t.Fatalf("found bogus option %v", opt)
		}
	}

	os.Remove(c.path)
}

// TestBrokenOptions looks for options outside an URL
func TestBrokenOptions(t *testing.T) {

	c := ParserHelper(t, `# https://example.com/index.rss
 - foo: bar`)

	_, err := c.Parse()
	if err == nil {
		t.Fatalf("Expected an error, got none!")
	}
	if !strings.Contains(err.Error(), "outside") {
		t.Fatalf("Got an error, but not the correct one:%s", err.Error())
	}

	os.Remove(c.path)

}

// TestAdd tests adding an entry works
func TestAdd(t *testing.T) {

	c := ParserHelper(t, ``)

	entries, err := c.Parse()
	if err != nil {
		t.Fatalf("unexpected error")
	}
	if len(entries) != 0 {
		t.Fatalf("expected no entries, but got some")
	}

	// add multiple times
	c.Add("https://example.com/")
	c.Add("https://example.com/")

	// Save
	err = c.Save()
	if err != nil {
		t.Fatalf("error saving")
	}

	// parse now we've saved
	entries, err = c.Parse()
	if err != nil {
		t.Fatalf("unexpected error")
	}
	if len(entries) != 1 {
		t.Fatalf("expected one entry, got %d", len(entries))
	}

	os.Remove(c.path)
}

// TestAddProperties tests adding to a file with properties doesn't fail
func TestAddProperties(t *testing.T) {

	c := ParserHelper(t, `
http://example.com/
 - foo:bar
 - retry: 7
#Comment2`)

	var out []Feed
	var err error

	_, err = c.Parse()
	if err != nil {
		t.Fatalf("Error parsing file: %v", err)
	}

	// Add another entry
	c.Add("https://blog.steve.fi/index.rss")

	// Now save and reload
	err = c.Save()
	if err != nil {
		t.Fatalf("Error saving file")
	}

	// Reparse
	out, err = c.Parse()
	if err != nil {
		t.Fatalf("Error parsing file: %v", err)
	}

	// Two entries now
	if len(out) != 2 {
		t.Fatalf("parsed wrong number of entries, got %d\n%v", len(out), out)
	}

	// We should have two options
	if len(out[0].Options) != 2 {
		t.Fatalf("Found wrong number of options, got %d", len(out[0].Options))
	}

	for _, opt := range out[0].Options {
		if opt.Name != "foo" &&
			opt.Name != "retry" {
			t.Fatalf("found bogus option %v", opt)
		}
	}

	os.Remove(c.path)
}

// TestDelete tests removing an entry.
func TestDelete(t *testing.T) {

	c := ParserHelper(t, `
http://example.com/
 - foo:bar
 - retry: 7
#Comment2
https://bob.com/index.rss`)

	var out []Feed
	var err error

	_, err = c.Parse()
	if err != nil {
		t.Fatalf("Error parsing file: %v", err)
	}

	// Add another entry
	c.Delete("https://bob.com/index.rss")

	// Now save and reload
	err = c.Save()
	if err != nil {
		t.Fatalf("Error saving file")
	}

	// Reparse
	out, err = c.Parse()
	if err != nil {
		t.Fatalf("Error parsing file: %v", err)
	}

	// One entries now
	if len(out) != 1 {
		t.Fatalf("parsed wrong number of entries, got %d\n%v", len(out), out)
	}

	// We should have two options
	if len(out[0].Options) != 2 {
		t.Fatalf("Found wrong number of options, got %d", len(out[0].Options))
	}

	for _, opt := range out[0].Options {
		if opt.Name != "foo" &&
			opt.Name != "retry" {
			t.Fatalf("found bogus option %v", opt)
		}
	}

	os.Remove(c.path)
}

// TestSaveBogusFile ensures that saving to a bogus file results in an error
func TestSaveBogusFile(t *testing.T) {

	// Create an empty config
	c := ParserHelper(t, ``)

	// Remove the path, and setup something bogus
	os.Remove(c.path)
	c.path = "/dev/null/fsdf/C:/3ljs3"

	err := c.Save()
	if err == nil {
		t.Fatalf("Saving to a bogus file worked, and it shouldn't!")
	}

}

// TestFuzz is a fake-test just to get our coverage increased
func TestFuzz(t *testing.T) {
	Fuzz([]byte("- foo:bar"))
}

// ParserHelper writes the specified text to a configuration-file
// and configures that path to be used for a ConfigFile object
func ParserHelper(t *testing.T, content string) *ConfigFile {

	data := []byte(content)
	tmpfile, err := ioutil.TempFile("", "example")
	if err != nil {
		t.Fatalf("Error creating temporary file")
	}

	if _, err := tmpfile.Write(data); err != nil {
		t.Fatalf("Error writing to config file")
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("Error creating temporary file")
	}

	// Create a new config-reader
	c := New()
	c.path = tmpfile.Name()

	return c
}
