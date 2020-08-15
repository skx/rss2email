//
// This is the cron-subcommand.
//

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/k3a/html2text"
	"github.com/skx/rss2email/feedlist"
	"github.com/skx/rss2email/withstate"
)

// ProcessURL takes an URL as input, fetches the contents, and then
// processes each feed item found within it.
//
// Feed items which are new/unread will generate an email.
func (c *cronCmd) ProcessURL(input string) error {

	// Show what we're doing.
	if c.verbose {
		fmt.Printf("Fetching: %s\n", input)
	}

	// Fetch the feed for the input URL
	feed, err := feedlist.Feed(input)
	if err != nil {
		return err
	}

	if c.verbose {
		fmt.Printf("\tFound %d entries\n", len(feed.Items))
	}

	// For each entry in the feed ..
	for _, xp := range feed.Items {

		// Wrap it so we can use our helper methods
		item := withstate.FeedItem{Item: xp}

		// If we've not already notified about this one.
		if item.IsNew() {

			// Show the new item.
			if c.verbose {
				fmt.Printf("\t\tNew Entry: %s\n", item.Title)
			}

			// If we're supposed to send email then do that
			if c.send {

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
				err := SendMail(feed, item, c.emails, text, content)
				if err != nil {
					return err
				}
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

	return nil
}

// Structure for our options and state.
type cronCmd struct {
	// Should we be verbose in operation?
	verbose bool

	// Emails has the list of emails to which we should send our
	// notices
	emails []string

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

	// Save each argument away, checking it is fully-qualified.
	for _, email := range args {
		if strings.Contains(email, "@") {
			c.emails = append(c.emails, email)
		} else {
			fmt.Printf("Usage: rss2email cron [flags] email1 .. emailN\n")
			return 1
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
		err := c.ProcessURL(uri)
		if err != nil {
			errors = append(errors, fmt.Sprintf("error processing %s - %s\n", uri, err))
		}
	}

	prunedCount, pruneErrors := withstate.PruneStateFiles()
	for _, err := range pruneErrors {
		errors = append(errors, err.Error())
	}

	if c.verbose && prunedCount > 0 {
		fmt.Printf("Pruned %d entry state files\n", prunedCount)
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
		return 1
	}

	//
	// All good.
	//
	return 0
}
