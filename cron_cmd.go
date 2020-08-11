//
// This is the cron-subcommand.
//

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/google/subcommands"
	"github.com/k3a/html2text"
	"github.com/skx/rss2email/feedlist"
	"github.com/skx/rss2email/withstate"
)

// ProcessURL takes an URL as input, fetches the contents, and then
// processes each feed item found within it.
//
// Feed items which are new/unread will generate an email.
func (p *cronCmd) ProcessURL(input string) error {

	// Show what we're doing.
	if p.verbose {
		fmt.Printf("Fetching: %s\n", input)
	}

	// Fetch the feed for the input URL
	feed, err := feedlist.Feed(input)
	if err != nil {
		return err
	}

	if p.verbose {
		fmt.Printf("\tFound %d entries\n", len(feed.Items))
	}

	// For each entry in the feed ..
	for _, xp := range feed.Items {

		// Wrap it so we can use our helper methods
		item := withstate.FeedItem{Item: xp}

		// If we've not already notified about this one.
		if item.IsNew() {

			// Show the new item.
			if p.verbose {
				fmt.Printf("\t\tNew Entry: %s\n", item.Title)
			}

			// If we're supposed to send email then do that
			if p.send {

				// The body should be stored in the
				// "Content" field.
				content := item.Content

				// If the Content field is empty then
				// use the Description instead, if it
				// is non-empty itself.
				if (content == "") && item.Description != "" {
					content = item.Description
				}

				// Convert the content to text.
				text := html2text.HTML2Text(content)

				// Send the mail
				err := SendMail(feed, item, p.emails, text, content)
				if err != nil {
					return err
				}
			}

			// Mark the item as having been seen, after the
			// email was sent.
			//
			// This does run the risk that sending mail
			// fails, due to error, and that keeps happening
			// forever...
			item.RecordSeen()
		}
	}

	return nil
}

// The options set by our command-line flags.
type cronCmd struct {
	// Should we be verbose in operation?
	verbose bool

	// Emails has the list of emails to which we should send our
	// notices
	emails []string

	// Should we send emails?
	send bool
}

//
// Glue
//
func (*cronCmd) Name() string     { return "cron" }
func (*cronCmd) Synopsis() string { return "Send emails for each new entry in our feed lists." }
func (*cronCmd) Usage() string {
	return `This sub-command polls all configured feeds, sending an email for
each item which is new.

State is maintained beneath '~/.rss2email/seen/', and the feed list
itself is read from '~/.rss2email/feeds'.


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

//
// Flag setup
//
func (p *cronCmd) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&p.verbose, "verbose", false, "Should we be extra verbose?")
	f.BoolVar(&p.send, "send", true, "Should we send emails, or just pretend to?")
}

//
// Entry-point.
//
func (p *cronCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	// No argument?  That's a bug
	if len(f.Args()) == 0 {
		fmt.Printf("Usage: rss2email cron email1@example.com .. emailN@example.com\n")
		return subcommands.ExitFailure
	}

	// Save each argument away, checking it is fully-qualified.
	for _, email := range f.Args() {
		if strings.Contains(email, "@") {
			p.emails = append(p.emails, email)
		} else {
			fmt.Printf("Usage: rss2email cron [flags] email1 .. emailN\n")
			return subcommands.ExitFailure
		}
	}

	// Get the feed-list, from the default location.
	list := feedlist.New("")

	//
	// If we receive errors we'll store them here,
	// so we can keep processing subsequent URIs.
	//
	var errors []string

	//
	// For each entry in the list ..
	//
	for _, uri := range list.Entries() {

		//
		// Handle it.
		//
		err := p.ProcessURL(uri)
		if err != nil {
			errors = append(errors, fmt.Sprintf("error processing %s - %s\n", uri, err))
		}
	}

	//
	// If we found errors then handle that.
	//
	if len(errors) > 0 {

		// Show each error to STDERR
		for _, err := range errors {
			fmt.Fprintln(os.Stderr, err)
		}

		// Use a suitable exit-code.
		return subcommands.ExitFailure
	}

	//
	// All good.
	//
	return subcommands.ExitSuccess
}
