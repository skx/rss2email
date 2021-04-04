//
// Export our feeds in a standard format
//

package main

import (
	"fmt"
	"os"
	"text/template"

	"github.com/skx/rss2email/feedlist"
	"github.com/skx/subcommands"
)

// Structure for our options and state.
type exportCmd struct {

	// We embed the NoFlags option, because we accept no command-line flags.
	subcommands.NoFlags
}

// Info is part of the subcommand-API
func (e *exportCmd) Info() (string, string) {
	return "export", `Export the feed list as an OPML file.

This command exports the list of configured feeds as an OPML file.

To see details of the configuration file, including the location,
please run:

   $ rss2email help config

Example:

    $ rss2email export
`
}

// Execute is invoked if the user specifies `add` as the subcommand.
func (e *exportCmd) Execute(args []string) int {

	// Individual feed URL
	type Feed struct {
		URL string
	}

	// Template Data
	type TemplateData struct {
		Entries []Feed
	}
	data := TemplateData{}

	// Get the feed-list, from the default location.
	list := feedlist.New("")

	for _, entry := range list.Entries() {
		data.Entries = append(data.Entries, Feed{URL: entry})
	}

	// Template
	tmpl := `<?xml version="1.0" encoding="utf-8"?>
<opml version="1.0">
<head>
<title>Feed Export</title>
</head>
<body>
{{range .Entries}}<outline xmlUrl="{{.URL}}"/>
{{end}}
</body>
</opml>
`
	// Compile the template and write to STDOUT
	t := template.Must(template.New("tmpl").Parse(tmpl))
	err := t.Execute(os.Stdout, data)
	if err != nil {
		fmt.Printf("error rendering template: %s\n", err.Error())
		return 1
	}

	return 0
}
