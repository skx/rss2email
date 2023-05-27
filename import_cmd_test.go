package main

import (
	"os"
	"testing"

	"github.com/skx/rss2email/configfile"
)

func TestImport(t *testing.T) {

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

	if _, err = tmpfile.Write(data); err != nil {
		t.Fatalf("Error writing to config file")
	}
	if err = tmpfile.Close(); err != nil {
		t.Fatalf("Error creating temporary file")
	}

	// Create an OPML file to use as input
	opml, err := os.CreateTemp("", "opml")
	if err != nil {
		t.Fatalf("Error creating temporary file for OMPL input")
	}
	d1 := []byte(`
<?xml version="1.0" encoding="utf-8"?>
<opml version="1.0">
<head>
<title>Feed Value</title>
</head>
<body>
<outline xmlUrl="http://floooh.github.io/feed.xml"/>
<outline xmlUrl="http://feeds.feedburner.com/24ways"/>
<outline xmlUrl="http://feeds.feedburner.com/2ality"/>
<outline xmlUrl="http://feeds.feedburner.com/AJAXMagazine"/>
<outline xmlUrl="http://alexsexton.com/?feed=rss2"/>
<outline xmlUrl="http://www.broken-links.com/feed/"/>
<outline xmlUrl="http://www.wait-till-i.com/feed/"/>
</body>
</opml>
`)
	err = os.WriteFile(opml.Name(), d1, 0644)
	if err != nil {
		t.Fatalf("failed to write OPML file")
	}

	// Create an instance of the command, and setup the config file
	im := importCmd{}
	im.Arguments(nil)
	config := configfile.NewWithPath(tmpfile.Name())
	im.config = config

	// Run the import
	im.Execute([]string{opml.Name()})

	// Look for the new entries in the feed.
	entries, err2 := config.Parse()
	if err2 != nil {
		t.Errorf("error parsing the (updated) config file")
	}
	if len(entries) != 9 {
		t.Fatalf("found %d entries", len(entries))
	}

	// Cleanup
	os.Remove(tmpfile.Name())
	os.Remove(opml.Name())
}
