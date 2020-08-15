//
// Export our feeds in a standard format
//

package main

import (
	"context"
	"flag"
	"os"
	"text/template"

	"github.com/google/subcommands"
	"github.com/skx/rss2email/feedlist"
)

type exportCmd struct {
}

//
// Glue
//
func (*exportCmd) Name() string     { return "export" }
func (*exportCmd) Synopsis() string { return "Export the feed list as an OPML file." }
func (*exportCmd) Usage() string {
	return `This command exports the list of configured feeds as an OPML file.
`
}

//
// Flag setup
//
func (p *exportCmd) SetFlags(f *flag.FlagSet) {
}

//
// Entry-point.
//
func (p *exportCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

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
	t.Execute(os.Stdout, data)

	return subcommands.ExitSuccess
}
