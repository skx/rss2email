//
// Import an OPML feedlist.
//

package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"

	"github.com/skx/rss2email/feedlist"
	"github.com/skx/subcommands"
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

// Structure for our options and state.
type importCmd struct {

	// We embed the NoFlags option, because we accept no command-line flags.
	subcommands.NoFlags
}

// Info is part of the subcommand-API
func (i *importCmd) Info() (string, string) {
	return "import", `Import a list of feeds via an OPML file.

This command imports a series of feeds from the specified OPML
file into the configuration file this application uses.

To see details of the configuration file, including the location,
please run:

   $ rss2email help config

Example:

    $ rss2email import file1.opml file2.opml .. fileN.opml
`
}

// Execute is invoked if the user specifies `import` as the subcommand.
func (i *importCmd) Execute(args []string) int {

	// Get the feed-list, from the default location.
	list := feedlist.New("")

	added := 0

	// For each file on the command-line
	for _, file := range args {

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
	return 0
}
