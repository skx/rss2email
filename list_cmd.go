//
// List our configured-feeds.
//

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/skx/rss2email/feedlist"
	"github.com/skx/rss2email/template"
)

// Structure for our options and state.
type listCmd struct {

	// Should we list the template-contents, rather than the feed list?
	template bool

	// Should we show extra information about a feed?
	verbose bool
}

// Info is part of the subcommand-API
func (l *listCmd) Info() (string, string) {
	return "list", `Output the list of feeds which are being polled.

By default this subcommand lists the configured feeds which will be
polled, however it also allows you to dump the default email-template.


Example:

    $ rss2email list
    $ rss2email list -verbose
    $ rss2email list -template


Flags:

`
}

// Arguments handles our flag-setup.
func (l *listCmd) Arguments(f *flag.FlagSet) {
	f.BoolVar(&l.template, "template", false, "Show the contents of the default template?")
	f.BoolVar(&l.verbose, "verbose", false, "Show extra information about each feed?")
}

//
// Entry-point.
//
func (l *listCmd) Execute(args []string) int {

	if l.template {

		// Load the default template from the embedded resource.
		content, err := template.EmailTemplate()
		if err != nil {
			fmt.Printf("failed to load embedded resource: %s\n", err.Error())
			os.Exit(1)
		}

		fmt.Printf("%s\n", string(content))
		return 0
	}

	// Get the feed-list, from the default location.
	list := feedlist.New("")

	list.WriteAllEntriesIncludingComments(os.Stdout, l.verbose)

	return 0
}
