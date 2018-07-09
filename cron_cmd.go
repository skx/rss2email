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
func (p *cronCmd) ProcessURL(input string) {

	if p.verbose {
		fmt.Printf("Fetching %s\n", input)
	}

	// Fetch the URL
	txt, err := p.FetchFeed(input)
	if err != nil {
		fmt.Printf("Error processing %s - %s\n", input, err.Error())
		return
	}

	// Parse it
	fp := gofeed.NewParser()
	feed, err := fp.ParseString(txt)
	if err != nil {
		fmt.Printf("Error parsing %s contents: %s\n", input, err.Error())
		return
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
			}

			// If we're supposed to send email then do that
			if p.send {

				// Convert the body to text.
				text := html2text.HTML2Text(i.Content)

				// Send the mail
				err := SendMail(os.Getenv("LOGNAME"), i.Title, text, i.Content)

				// Assuming no errors then this item
				// has been processed.
				if err == nil {
					RecordSeen(i)
				}
			} else {

				// We're not sending email, so just record
				// this item as having been processed.
				RecordSeen(i)
			}
		}
	}
}

// The options set by our command-line flags.
type cronCmd struct {
	// Should we be verbose in operation?
	verbose bool

	// Should we send emails?
	send bool
}

//
// Glue
//
func (*cronCmd) Name() string     { return "cron" }
func (*cronCmd) Synopsis() string { return "Process each of the feeds." }
func (*cronCmd) Usage() string {
	return `cron :
  Read the list of feeds and send email for each new item found in them.
`
}

//
// Flag setup: NOP
//
func (p *cronCmd) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&p.verbose, "verbose", false, "Should we be extra verbose?")
	f.BoolVar(&p.send, "send", true, "Should we send emails, or just pretend to?")
}

//
// Entry-point.
//
func (p *cronCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	//
	// Create the helper
	//
	list := NewFeed()

	//
	// For each entry in the list ..
	//
	for _, uri := range list.Entries() {

		//
		// Handle it.
		//
		p.ProcessURL(uri)
	}

	//
	// All done.
	//
	return subcommands.ExitSuccess
}
