//
// List our configured-feeds.
//

package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/google/subcommands"
	"github.com/skx/rss2email/feedlist"
)

//
// The options set by our command-line flags.
//
type listCmd struct {

	// Should we list the template-contents, rather than the feed list?
	template bool
}

//
// Glue
//
func (*listCmd) Name() string     { return "list" }
func (*listCmd) Synopsis() string { return "List configured feeds." }
func (*listCmd) Usage() string {
	return `Output the list of feeds which are being polled.

By default this subcommand lists the configured feeds which will be
polled, however it also allows you to dump the default email-template.

Example:

    $ rss2email list
    $ rss2email list -template
`
}

//
// Flag setup: NOP
//
func (p *listCmd) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&p.template, "template", false, "Show the contents of the default template?")
}

//
// Entry-point.
//
func (p *listCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	if p.template {

		// Load the default template from the embedded resource.
		content, err := getResource("data/email.tmpl")
		if err != nil {
			fmt.Printf("failed to load embedded resource: %s\n", err.Error())
			os.Exit(1)
		}

		fmt.Printf("%s\n", string(content))
		return subcommands.ExitSuccess
	}

	//
	// Create the helper
	//
	list := feedlist.New("")

	//
	// For each entry in the list ..
	//
	for _, uri := range list.Entries() {

		//
		// Print it
		//
		fmt.Printf("%s\n", uri)
	}
	return subcommands.ExitSuccess
}
