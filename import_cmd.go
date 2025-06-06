//
// Import an OPML feedlist.
//

package main

import (
	"encoding/xml"
	"flag"
	"log/slog"
	"os"

	"github.com/skx/rss2email/configfile"
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

	// Configuration file, used for testing
	config *configfile.ConfigFile
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

// Arguments handles argument-flags we might have.
//
// In our case we use this as a hook to setup our configuration-file,
// which allows testing.
func (i *importCmd) Arguments(flags *flag.FlagSet) {
	i.config = configfile.New()
}

// Execute is invoked if the user specifies `import` as the subcommand.
func (i *importCmd) Execute(args []string) int {

	_, err := i.config.Parse()
	if err != nil {
		logger.Error("failed to parse configuration file",
			slog.String("configfile", i.config.Path()),

			slog.String("error", err.Error()))
		return 1
	}

	// For each file on the command-line
	for _, file := range args {

		// Read content
		var data []byte
		data, err = os.ReadFile(file)
		if err != nil {
			logger.Error("failed to read file", slog.String("file", file), slog.String("error", err.Error()))
			continue
		}

		// Parse
		o := opml{}
		err = xml.Unmarshal(data, &o)
		if err != nil {
			logger.Error("failed to parse XML file", slog.String("file", file), slog.String("error", err.Error()))
			continue
		}

		for _, outline := range o.Outlines {

			if outline.XMLURL != "" {
				logger.Debug("Adding entry from file", slog.String("file", file), slog.String("url", outline.XMLURL))
				i.config.Add(outline.XMLURL)
			}
		}

	}

	// Did we make a change?  Then add them.
	err = i.config.Save()
	if err != nil {
		logger.Error("failed to save the updated feed list", slog.String("error", err.Error()))
		return 1
	}

	// All done.
	return 0
}
