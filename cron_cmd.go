//
// This is the cron-subcommand.
//

package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/google/subcommands"
	"github.com/k3a/html2text"
	"github.com/mmcdole/gofeed"
)

// FetchFeed fetches a feed from the remote URL.
//
// We must use this instead of the URL handler that the feed-parser supports
// because reddit, and some other sites, will just return a HTTP error-code
// if we're using a standard "spider" User-Agent.
//
func (p *cronCmd) FetchFeed(url string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "rss2email (https://github.com/skx/rss2email)")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	output, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// ProcessURL takes an URL as input, fetches the contents, and then
// processes each feed item found within it.
func (p *cronCmd) ProcessURL(input string) error {

	if p.verbose {
		fmt.Printf("Fetching %s\n", input)
	}

	// Fetch the URL
	txt, err := p.FetchFeed(input)
	if err != nil {
		return fmt.Errorf("error processing %s - %s", input, err.Error())
	}

	// Parse it
	fp := gofeed.NewParser()
	feed, err := fp.ParseString(txt)
	if err != nil {
		return fmt.Errorf("error parsing %s contents: %s", input, err.Error())
	}

	if p.verbose {
		fmt.Printf("\tFound %d entries\n", len(feed.Items))
	}

	// For each entry in the feed ..
	for _, i := range feed.Items {

		// If we've not already notified about this one.
		if !HasSeen(i) {

			if p.verbose {
				fmt.Printf("New item: %s\n", i.GUID)
				fmt.Printf("\tTitle: %s\n", i.Title)
			}

			// Mark the item as having been seen.
			RecordSeen(i)

			// Expand the subject format-string
			subject := p.subject

			// Feed
			subject = strings.ReplaceAll(subject, "#{FEED.TITLE}", feed.Title)
			subject = strings.ReplaceAll(subject, "#{FEED.LINK}", feed.Link)

			// Item
			subject = strings.ReplaceAll(subject, "#{ITEM.TITLE}", i.Title)
			subject = strings.ReplaceAll(subject, "#{ITEM.LINK}", i.Link)

			// Author
			subject = strings.ReplaceAll(subject, "#{ITEM.AUTHOR.NAME}", i.Author.Name)
			subject = strings.ReplaceAll(subject, "#{ITEM.AUTHOR.EMAIL}", i.Author.Email)

			// If we're supposed to send email then do that
			if p.send {

				// The body should be stored in the
				// "Content" field.
				content := i.Content

				// If the Content field is empty then
				// use the Description instead, if it
				// is non-empty itself.
				if (content == "") && i.Description != "" {
					content = i.Description
				}

				// Convert the content to text.
				text := html2text.HTML2Text(content)

				// Send the mail
				err := SendMail(input, p.fromAddr, p.emails, subject, i.Link, text, content)
				if err != nil {
					return err
				}
			}
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

	// The address all emails should be sent from.  If omitted,
	// the From: address is the same as the To: address.
	fromAddr string

	// Template-string which is expanded into the email subjects
	subject string
}

//
// Glue
//
func (*cronCmd) Name() string     { return "cron" }
func (*cronCmd) Synopsis() string { return "Process each of the feeds." }
func (*cronCmd) Usage() string {
	return `cron :
Read the list of feeds and send email for each new item found in them.

By default emails are sent to the address specified upon the command-line,
with the same sender, however you can use the '--from' flag to change that.

Emails are constructed with a static template, however it is possible to
customize the subject via a series of template-variables.  The default
subjects will be:

   --subject="[rss2email] #{ITEM.TITLE}"

Valid template values include:

  #{FEED.TITLE}
  #{FEED.LINK}
  #{ITEM.TITLE}
  #{ITEM.LINK}
  #{ITEM.AUTHOR.NAME}
  #{ITEM.AUTHOR.EMAIL}

`
}

//
// Flag setup: NOP
//
func (p *cronCmd) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&p.verbose, "verbose", false, "Should we be extra verbose?")
	f.BoolVar(&p.send, "send", true, "Should we send emails, or just pretend to?")
	f.StringVar(&p.fromAddr, "from", "", "Specify the sending email address to use.")
	f.StringVar(&p.subject, "subject", "[rss2email] $ITEM.TITLE", "A format string used to generate the email subject.")

}

//
// Entry-point.
//
func (p *cronCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	//
	// No argument?  That's a bug
	//
	if len(f.Args()) == 0 {
		fmt.Printf("Usage: rss2email cron email1@example.com .. emailN@example.com\n")
		return subcommands.ExitFailure
	}

	//
	// Save each argument away, checking it is fully-qualified.
	//
	for _, email := range f.Args() {
		if strings.Contains(email, "@") {
			p.emails = append(p.emails, email)
		} else {
			fmt.Printf("Usage: rss2email cron email1 .. emailN\n")
			return subcommands.ExitFailure
		}
	}

	//
	// Create the helper
	//
	list := NewFeed()

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
