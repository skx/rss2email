//
// List our configured-feeds.
//

package main

import (
	"flag"
	"fmt"
	"log/slog"
	"time"

	"github.com/skx/rss2email/configfile"
	"github.com/skx/rss2email/httpfetch"
)

var (
	maxInt = int(^uint(0) >> 1)
)

// Structure for our options and state.
type listCmd struct {

	// Configuration file, used for testing
	config *configfile.ConfigFile

	// verbose controls whether our feed-list contains information
	// about feed entries and their ages
	verbose bool
}

// Arguments handles argument-flags we might have.
//
// In our case we use this as a hook to setup our configuration-file,
// which allows testing.
func (l *listCmd) Arguments(flags *flag.FlagSet) {

	// Setup configuration file
	l.config = configfile.New()

	// Are we listing verbosely?
	flags.BoolVar(&l.verbose, "verbose", false, "Show extra information about each feed (slow)?")
}

// Info is part of the subcommand-API
func (l *listCmd) Info() (string, string) {
	return "list", `Output the list of feeds which are being polled.

This subcommand lists the feeds which are specified in the
configuration file.

To see details of the configuration file, including the location,
please run:

   $ rss2email help config


You can add '-verbose' to see details about the feed contents, but note
that this will require downloading the contents of each feed and will
thus be slow - a simpler way of showing history would be to run:

    $ rss2email seen

Example:

    $ rss2email list
`
}

func (l *listCmd) showFeedDetails(entry configfile.Feed) {

	// Fetch the details
	helper := httpfetch.New(entry)
	feed, err := helper.Fetch()
	if err != nil {
		fmt.Fprintf(out, "# %s\n%s\n", err.Error(), entry.URL)
		return
	}

	// Handle single vs. plural entries
	entriesString := "entries"
	if len(feed.Items) == 1 {
		entriesString = "entry"
	}

	// get the age-range of the feed-entries
	oldest := -1
	newest := maxInt
	for _, item := range feed.Items {
		if item.PublishedParsed == nil {
			break
		}

		age := int(time.Since(*item.PublishedParsed) / (24 * time.Hour))
		if age > oldest {
			oldest = age
		}

		if age < newest {
			newest = age
		}
	}

	// Now show the details, which is a bit messy.
	fmt.Fprintf(out, "# %d %s, aged %d-%d days\n", len(feed.Items), entriesString, newest, oldest)
	fmt.Fprintf(out, "%s\n", entry.URL)
}

// Entry-point.
func (l *listCmd) Execute(args []string) int {

	// Now do the parsing
	entries, err := l.config.Parse()
	if err != nil {
		logger.Error("failed to parse configuration file", slog.String("error", err.Error()))
		return 1
	}

	// Show the feeds
	for _, entry := range entries {

		if l.verbose {
			l.showFeedDetails(entry)
		} else {
			fmt.Fprintf(out, "%s\n", entry.URL)
		}
	}

	return 0
}
