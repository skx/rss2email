//
// This is the cron-subcommand.
//

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/skx/rss2email/processor"
)

// Structure for our options and state.
type cronCmd struct {
	// Should we be verbose in operation?
	verbose bool

	// Should we send emails?
	send bool
}

// Info is part of the subcommand-API.
func (c *cronCmd) Info() (string, string) {
	return "cron", `Send emails for each new entry in our feed lists.

This sub-command polls all configured feeds, sending an email for
each item which is new.

We record details of all the feed-items which have been seen beneath
 '~/.rss2email/seen/', and these entries will be expired automatically
when the corresponding entries have fallen out of the source feed.

The list of feeds is read from '~/.rss2email/feeds'.


Example:

    $ rss2email cron user1@example.com user2@example.com

Customization:

An embedded template is used to generate the emails which are sent, this
may be overridden via the creation of a local template-file located at
'~/.rss2email/email.tmpl'.  The default template can be exported and
modified like so:

    $ rss2email list -template > ~/.rss2email/email.tmpl


Flags:

`
}

// Arguments handles our flag-setup.
func (c *cronCmd) Arguments(f *flag.FlagSet) {
	f.BoolVar(&c.verbose, "verbose", false, "Should we be extra verbose?")
	f.BoolVar(&c.send, "send", true, "Should we send emails, or just pretend to?")
}

//
// Entry-point
//
func (c *cronCmd) Execute(args []string) int {

	// No argument?  That's a bug
	if len(args) == 0 {
		fmt.Printf("Usage: rss2email cron email1@example.com .. emailN@example.com\n")
		return 1
	}

	// The list of addresses to which we should send our notices.
	recipients := []string{}

	// Save each argument away, checking it is fully-qualified.
	for _, email := range args {
		if strings.Contains(email, "@") {
			recipients = append(recipients, email)
		} else {
			fmt.Printf("Usage: rss2email cron [flags] email1 .. emailN\n")
			return 1
		}
	}

	// Create the helper
	p := processor.New()

	// Setup the state
	p.SetVerbose(c.verbose)
	p.SetSendEmail(c.send)

	errors := p.ProcessFeeds(recipients)

	// If we found errors then show them.
	if len(errors) > 0 {
		for _, err := range errors {
			fmt.Fprintln(os.Stderr, err.Error())
		}

		return 1
	}

	// All good.
	return 0
}
