//
// Export our feeds in a standard format
//

package main

import (
	"flag"
	"log/slog"
	"text/template"

	"github.com/skx/rss2email/configfile"
	"github.com/skx/subcommands"
)

// Structure for our options and state.
type exportCmd struct {

	// We embed the NoFlags option, because we accept no command-line flags.
	subcommands.NoFlags

	// Configuration file, used for testing
	config *configfile.ConfigFile
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

// Arguments handles argument-flags we might have.
//
// In our case we use this as a hook to setup our configuration-file,
// which allows testing.
func (e *exportCmd) Arguments(flags *flag.FlagSet) {
	e.config = configfile.New()
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

	// Now do the parsing
	entries, err := e.config.Parse()
	if err != nil {
		logger.Error("failed to parse configuration file",
			slog.String("configfile", e.config.Path()),
			slog.String("error", err.Error()))
		return 1
	}

	// Populate our template variables
	for _, entry := range entries {
		data.Entries = append(data.Entries, Feed{URL: entry.URL})
	}

	// Template
	tmpl := `<?xml version="1.0" encoding="utf-8"?>
<opml version="1.0">
<head>
<title>Feed Export</title>
</head>
<body>
{{range .Entries}}<outline xmlUrl="{{.URL}}"/>
{{end}}</body>
</opml>
`
	// Compile the template and write to STDOUT
	t := template.Must(template.New("tmpl").Parse(tmpl))
	err = t.Execute(out, data)
	if err != nil {
		logger.Error("error rendering template", slog.String("error", err.Error()))
		return 1
	}

	return 0
}
