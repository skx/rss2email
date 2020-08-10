//
// Import an OPML feedlist.
//

package main

import (
	"context"
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/google/subcommands"
	"github.com/skx/rss2email/feedlist"
)

type opml struct {
	XMLName   xml.Name  `xml:"opml"`
	Version   string    `xml:"version,attr"`
	OpmlTitle string    `xml:"head>title"`
	Outlines  []outline `xml:"body>outline"`
}

type outline struct {
	Text    string `xml:"text,attr"`
	Title   string `xml:"title,attr"`
	Type    string `xml:"type,attr"`
	XMLURL  string `xml:"xmlUrl,attr"`
	HTMLURL string `xml:"htmlUrl,attr"`
	Favicon string `xml:"rssfr-favicon,attr"`
}

//
// The options set by our command-line flags.
//
type importCmd struct {
}

//
// Glue
//
func (*importCmd) Name() string     { return "import" }
func (*importCmd) Synopsis() string { return "Import an OPML feed-list." }
func (*importCmd) Usage() string {
	return `Import a list of feeds via an OPML file.

Example:

    $ rss2email import file1.opml file2.opml .. fileN.opml
`
}

//
// Flag setup: NOP
//
func (p *importCmd) SetFlags(f *flag.FlagSet) {
}

//
// Entry-point.
//
func (p *importCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	// Get the feed-list, from the default location.
	list := feedlist.New("")

	added := 0

	// For each file on the command-line
	for _, file := range f.Args() {

		// Read content
		data, err := ioutil.ReadFile(file)
		if err != nil {
			fmt.Printf("failed to read %s: %s\n", file, err.Error())
			continue
		}

		// Parse
		o := opml{}
		err = xml.Unmarshal(data, &o)
		if err != nil {
			fmt.Printf("failed to parse %s: %s\n", file, err.Error())
			continue
		}
		entries := make([]string, len(o.Outlines))
		for i, outline := range o.Outlines {

			if outline.XMLURL != "" {
				fmt.Printf("Adding %s\n", outline.XMLURL)
				entries[i] = outline.XMLURL
				added++
			}
		}
		errors := list.Add(entries...)
		for _, err := range errors {
			fmt.Printf("%s\n", (err.Error()))
		}
	}

	// Did we make a change?  Then add them.
	if added > 0 {
		err := list.Save()
		if err != nil {
			fmt.Printf("failed to update feed list: %s\n", err.Error())
		}
	}

	// All done.
	return subcommands.ExitSuccess
}
