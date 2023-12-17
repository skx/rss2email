//
// Delete a feed from our feed-list.
//

package main

import (
	"flag"
	"log/slog"

	"github.com/skx/rss2email/configfile"
	"github.com/skx/subcommands"
)

// Structure for our options and state.
type delCmd struct {

	// We embed the NoFlags option, because we accept no command-line flags.
	subcommands.NoFlags

	// Configuration file, used for testing
	config *configfile.ConfigFile
}

// Arguments handles argument-flags we might have.
//
// In our case we use this as a hook to setup our configuration-file,
// which allows testing.
func (d *delCmd) Arguments(flags *flag.FlagSet) {
	d.config = configfile.New()
}

// Info is part of the subcommand-API
func (d *delCmd) Info() (string, string) {
	return "delete", `Remove a feed from our feed-list.

Remove one or more specified URLs from the configuration file.

To see details of the configuration file, including the location,
please run:

   $ rss2email help config

Example:

    $ rss2email delete https://blog.steve.fi/index.rss
`
}

// Entry-point.
func (d *delCmd) Execute(args []string) int {

	// Parse the existing file
	_, err := d.config.Parse()
	if err != nil {
		logger.Error("failed to parse configuration file",
			slog.String("configfile", d.config.Path()),
			slog.String("error", err.Error()))
		return 1
	}

	changed := false

	// For each argument remove it from the list, if present.
	for _, entry := range args {
		d.config.Delete(entry)
		changed = true
	}

	// Save the list.
	if changed {
		err = d.config.Save()
		if err != nil {
			logger.Error("failed to save the updated feed list", slog.String("error", err.Error()))
			return 1
		}
	}

	// All done, with no errors.
	return 0

}
