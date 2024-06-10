//
// This is the daemon-subcommand.
//

package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/skx/rss2email/processor"
)

// Structure for our options and state.
type daemonCmd struct {

	// Should we be verbose in operation?
	verbose bool
}

// Info is part of the subcommand-API.
func (d *daemonCmd) Info() (string, string) {
	return "daemon", `Send emails for each new entry in our feed lists.

This sub-command polls all configured feeds, sending an email for
each item which is new.  Once the list of feeds has been processed
the command will pause for 15 minutes, before beginning again.

To see details of the configuration file, including the location, please
run:

   $ rss2email help config

In terms of implementation this command follows everything documented
in the 'cron' sub-command.  The only difference is this one never
terminates - even if email-generation fails.


Example:

    $ rss2email daemon user1@example.com user2@example.com
`
}

// Arguments handles our flag-setup.
func (d *daemonCmd) Arguments(f *flag.FlagSet) {
	f.BoolVar(&d.verbose, "verbose", false, "Should we be extra verbose?")
}

// Entry-point
func (d *daemonCmd) Execute(args []string) int {

	// If running verbosely change our log-level
	if d.verbose {
		loggerLevel.Set(slog.LevelDebug)
	}

	// No argument?  That's a bug
	if len(args) == 0 {
		fmt.Printf("Usage: rss2email daemon email1@example.com .. emailN@example.com\n")
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
			fmt.Printf("Usage: rss2email daemon [flags] email1 .. emailN\n")
			return 1
		}
	}

	for {

		// Create the helper
		p, err := processor.New()

		if err != nil {
			logger.Error("failed to create feed processor",
				slog.String("error", err.Error()))
			return 1
		}

		// Ensure we send our version
		p.SetVersion(version)

		// Setup the state - note we ALWAYS send emails in this mode.
		p.SetSendEmail(true)
		p.SetLogger(logger)

		// Process all the feeds
		errors := p.ProcessFeeds(recipients)

		// If we found errors then show them.
		if len(errors) != 0 {
			for _, err := range errors {
				fmt.Fprintln(os.Stderr, err.Error())
			}
		}

		// Close the database handle, once processed.
		p.Close()

		// Default time to sleep - in minutes
		n := 5

		logger.Warn("sleeping before polling feeds again",
			slog.Int("delay.minutes", n))

		time.Sleep(time.Duration(n) * time.Minute)
	}
}
