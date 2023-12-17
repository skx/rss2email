//
// This is the cron-subcommand.
//

package main

import (
	"flag"
	"fmt"
	"log/slog"
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
new item in those feeds.

The list of feeds is read from '~/.rss2email/feeds'.

We record details of all the feed-items which have been seen beneath
 '~/.rss2email/seen/', and these entries will be expired automatically
when the corresponding entries have fallen out of the source feed.

Example:

    $ rss2email cron user1@example.com user2@example.com


Email Sending:

By default we pipe outgoing messages through '/usr/sbin/sendmail' for delivery,
however it is possible to use SMTP for sending emails directly.  If you
wish to use SMTP you need to configure the following environmental variables:

    SMTP_HOST       (e.g. "smtp.gmail.com")
    SMTP_PORT       (e.g. "587")
    SMTP_USERNAME   (e.g. "user@domain.com")
    SMTP_PASSWORD   (e.g. "secret!word#here")


Email Template:

An embedded template is used to generate the emails which are sent, you
may create a local override for this, for more details see :

    $ rss2email help list-default-template
`
}

// Arguments handles our flag-setup.
func (c *cronCmd) Arguments(f *flag.FlagSet) {
	f.BoolVar(&c.verbose, "verbose", false, "Should we be extra verbose?")
	f.BoolVar(&c.send, "send", true, "Should we send emails, or just pretend to?")
}

// Entry-point
func (c *cronCmd) Execute(args []string) int {

	// verbose will change the log-level of our logger
	if c.verbose {
		loggerLevel.Set(slog.LevelDebug)
	}

	// No argument?  That's a bug
	if len(args) == 0 {
		fmt.Printf("Usage: rss2email cron email1@example.com .. emailN@example.com\n")
		return 1
	}

	// The list of addresses to notify, unless overridden by a per-feed
	// configuration option.
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
	p, err := processor.New()
	if err != nil {
		fmt.Printf("Error creating feed processor: %s\n", err.Error())
		return 1
	}

	// Close the database handle, once processed.
	defer p.Close()

	// Setup the state
	p.SetSendEmail(c.send)
	p.SetLogger(logger)

	errors := p.ProcessFeeds(recipients)

	// If we found errors then show them.
	if len(errors) != 0 {
		for _, err := range errors {
			fmt.Fprintln(os.Stderr, err.Error())
		}

		return 1
	}

	// All good.
	return 0
}
