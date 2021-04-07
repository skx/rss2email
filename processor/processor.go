// Package processor contains the code which will actually poll
// the list of URLs the user is watching, and send emails for those
// entries which are new.
package processor

import (
	"fmt"
	"regexp"

	"github.com/k3a/html2text"
	"github.com/skx/rss2email/configfile"
	"github.com/skx/rss2email/httpfetch"
	"github.com/skx/rss2email/processor/emailer"
	"github.com/skx/rss2email/withstate"
)

// Processor stores our state
type Processor struct {

	// send controls whether we send emails, or just pretend to.
	send bool

	// verbose denotes how verbose we should be in execution.
	verbose bool
}

// New creates a new Processor object
func New() *Processor {
	return &Processor{send: true}
}

// ProcessFeeds is the main workhorse here, we process each feed and send
// emails appropriately.
func (p *Processor) ProcessFeeds(recipients []string) []error {

	//
	// If we receive errors we'll store them here,
	// so we can keep processing subsequent URIs.
	//
	var errors []error

	// Get the configuration-file
	conf := configfile.New()

	// Upgrade it if necessary
	conf.Upgrade()

	// Now do the parsing
	entries, err := conf.Parse()
	if err != nil {
		errors = append(errors, fmt.Errorf("error with config-file %s - %s", conf.Path(), err))
		return errors
	}

	// For each entry in the list ..
	for _, entry := range entries {

		// Handle it.
		err := p.processFeed(entry, recipients)
		if err != nil {
			errors = append(errors, fmt.Errorf("error processing %s - %s", entry.URL, err))
		}
	}

	// Prune old state files
	prunedCount, pruneErrors := withstate.PruneStateFiles()

	// If we got any errors propagate them
	errors = append(errors, pruneErrors...)

	// Show what we did, if we should
	if p.verbose && prunedCount > 0 {
		fmt.Printf("Pruned %d entry state files\n", prunedCount)
	}

	return errors
}

// processFeed takes a configuration entry as input, fetches the appropriate
// remote contents, and then processes each feed item found within it.
//
// Feed items which are new/unread will generate an email.
func (p *Processor) processFeed(entry configfile.Feed, recipients []string) error {

	// Show what we're doing.
	if p.verbose {
		fmt.Printf("Fetching feed: %s\n", entry.URL)
	}

	// Fetch the feed for the input URL
	helper := httpfetch.New(entry)
	feed, err := helper.Fetch()
	if err != nil {
		return err
	}

	if p.verbose {
		fmt.Printf("\tFeed contains %d entries\n", len(feed.Items))
	}

	// For each entry in the feed ..
	for _, xp := range feed.Items {

		// Wrap it so we can use our helper methods
		item := withstate.FeedItem{Item: xp}

		// If we've not already notified about this one.
		if item.IsNew() {

			// Show the new item.
			if p.verbose {
				fmt.Printf("\t\tFeed entry: %s\n", item.Title)
			}

			// If we're supposed to send email then do that
			if p.send {

				content, err := item.HTMLContent()
				if err != nil {
					content = item.RawContent()
				}

				if !p.shouldSkip(entry, item, content) {

					// Convert the content to text.
					text := html2text.HTML2Text(content)

					// Send the mail
					helper := emailer.New(feed, item)
					err = helper.Sendmail(recipients, text, content)
					if err != nil {
						return err
					}
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

// shouldSkip returns true if this entry should be skipped.
//
// If an entry should be skipped it is still marked as read, but no email is sent.
func (p *Processor) shouldSkip(config configfile.Feed, item withstate.FeedItem, content string) bool {

	// Walk over the options for this feed to see if the new entry
	// should be skipped.
	for _, opt := range config.Options {

		// Exclude by title?
		if opt.Name == "exclude-title" {
			match, _ := regexp.MatchString(opt.Value, item.Title)
			if match {
				if p.verbose {
					fmt.Printf("\t\t\tSkipping due to 'exclude-title' match of '%s'.\n", opt.Value)
				}
				return true
			}
		}

		// Exclude by body/content?
		if opt.Name == "exclude" {

			match, _ := regexp.MatchString(opt.Value, content)
			if match {
				if p.verbose {
					fmt.Printf("\t\t\tSkipping due to 'exclude' match of %s.\n", opt.Value)
				}
				return true
			}
		}
	}

	return false
}

// SetVerbose updates the verbosity state of this object.
func (p *Processor) SetVerbose(state bool) {
	p.verbose = state
}

// SetSendEmail updates the state of this object, when the send-flag
// is false zero emails are generated.
func (p *Processor) SetSendEmail(state bool) {
	p.send = state
}
