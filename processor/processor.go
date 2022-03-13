// Package processor contains the code which will actually poll
// the list of URLs the user is watching, and send emails for those
// entries which are new.
//
// Items which are excluded are treated the same as normal items,
// in the sense they are processed once and then marked as having
// been seen - the only difference is no email is actually generated
// for them.
package processor

import (
	"fmt"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/k3a/html2text"
	"github.com/skx/rss2email/configfile"
	"github.com/skx/rss2email/httpfetch"
	"github.com/skx/rss2email/processor/emailer"
	"github.com/skx/rss2email/withstate"
	"go.etcd.io/bbolt"
)

// Processor stores our state
type Processor struct {

	// send controls whether we send emails, or just pretend to.
	send bool

	// verbose denotes how verbose we should be in execution.
	verbose bool

	// dbPath holds the path to the database
	dbPath string

	// database holds the db state
	dbHandle *bbolt.DB
}

// New creates a new Processor object.
//
// This might return an error if we fail to open the database we use
// for maintaining state.
func New() (*Processor, error) {
	db, err := bbolt.Open(dbGetPath(), 0666, nil)
	if err != nil {
		return nil, err
	}

	return &Processor{send: true, dbHandle: db}, nil
}

// Close should be called to cleanup our internal database-handle
func (p *Processor) Close() {
	p.dbHandle.Close()
}

// dbGetPath returns the path to use for the bolt database
func dbGetPath() string {

	// Default to using $HOME
	home := os.Getenv("HOME")

	if home == "" {
		// Get the current user, and use their home if possible.
		usr, err := user.Current()
		if err == nil {
			home = usr.HomeDir
		}
	}

	// Return the path
	return filepath.Join(home, ".rss2email", "state.db")
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

	// Keep track of the previous hostname from which we fetched a feed
	prev := ""

	// For each feed contained in the configuration file
	for _, entry := range entries {

		// Create a bucket to hold the entry-state here,
		// if we've not done so previously.
		//
		// We do this because we store the "seen" vs "unseen"
		// state in BoltDB database.
		//
		// BoltDB has a concept of "Buckets", which contain
		// key=value entries.
		//
		// Since we process feeds it seems logical to create
		// a bucket for each Feed URL, and then store the
		// URLs we've seen with a random value.
		//
		err := p.dbHandle.Update(func(tx *bbolt.Tx) error {
			_, err := tx.CreateBucketIfNotExists([]byte(entry.URL))
			if err != nil {
				return fmt.Errorf("create bucket failed: %s", err)
			}
			return nil
		})

		// If we have a DB-error then we return, this shouldn't happen.
		if err != nil {
			errors = append(errors, fmt.Errorf("error creating bucket for %s: %s", entry.URL, err))
			return (errors)
		}

		// Should we sleep before getting this feed?
		sleep := 0

		// We default to notifying the global recipient-list.
		//
		// But there might be a per-feed set of recipients which
		// we'll prefer if available.
		feedRecipients := recipients

		// parse the hostname form the URL
		host := ""
		u, err := url.Parse(entry.URL)
		if err == nil {
			host = u.Host
		}

		// Are we fetching from the same host as the previous feed?
		// If so then we'll add a delay to try to avoid annoying that
		// host.
		if host == prev {
			p.message(fmt.Sprintf("Fetching from same host as previous feed, %s, adding 5s delay", host))
			sleep = 5
		}

		// For each option
		for _, opt := range entry.Options {

			// Is it a set of recipients?
			if opt.Name == "notify" {

				// Save the values
				feedRecipients = strings.Split(opt.Value, ",")

				// But trim leading/trailing space
				for i := range feedRecipients {
					feedRecipients[i] = strings.TrimSpace(feedRecipients[i])
				}
			}

			// Sleep setting?
			if opt.Name == "sleep" {

				// Convert the value, and if there was
				// no error save it away.
				num, nErr := strconv.Atoi(opt.Value)
				if nErr != nil {
					fmt.Printf("WARNING: %s:%s - failed to parse as sleep-delay %s\n", opt.Name, opt.Value, nErr.Error())
				} else {
					sleep = num
				}
			}
		}

		// If we're supposed to sleep, do so
		if sleep != 0 {
			time.Sleep(time.Duration(sleep) * time.Second)
		}

		// Process this specific entry.
		err = p.processFeed(entry, feedRecipients)
		if err != nil {
			errors = append(errors, fmt.Errorf("error processing %s - %s", entry.URL, err))
		}

		// Now update with our current host.
		prev = host
	}

	return errors
}

// message shows a message if our verbose flag is set.
//
// NOTE: This appends a newline to the message.
func (p *Processor) message(msg string) {
	if p.verbose {
		fmt.Printf("%s\n", msg)
	}
}

// processFeed takes a configuration entry as input, fetches the appropriate
// remote contents, and then processes each feed item found within it.
//
// Feed items which are new/unread will generate an email, unless they are
// specifically excluded by the per-feed options.
func (p *Processor) processFeed(entry configfile.Feed, recipients []string) error {

	// Show what we're doing.
	p.message(fmt.Sprintf("Fetching feed: %s", entry.URL))

	// Fetch the feed for the input URL
	helper := httpfetch.New(entry)
	feed, err := helper.Fetch()
	if err != nil {
		return err
	}

	// Show how many entries we've found in the feed.
	p.message(fmt.Sprintf("\tFeed contains %d entries", len(feed.Items)))

	// Count how many seen/unseen items there were.
	seen := 0
	unseen := 0

	// Keep track of all the items in the feed.
	items := []string{}

	// For each entry in the feed ..
	for _, xp := range feed.Items {

		// Wrap the feed-item in a class of our own,
		// so that we can use our helper methods to mark
		// read-state.
		//
		// TODO: Remove this stuff in the near-future.
		item := withstate.FeedItem{Item: xp}

		// Keep track of the fact that we saw this feed-item.
		//
		// This is used for pruning the BoltDB state file
		items = append(items, item.Link)

		// If we've not already notified about this one.
		//
		// Check legacy-state first, then the new-state.
		isNew := true

		if p.seenItem(entry.URL, item.Link) {
			isNew = false
		} else {
			if !item.IsNew() {
				isNew = false
			}
		}

		if isNew {

			// Bump the count
			unseen++

			// Show the new item.
			p.message(fmt.Sprintf("\t\tNew entry in feed: %s", item.Title))
			p.message(fmt.Sprintf("\t\t\t%s", item.Link))

			// If we're supposed to send email then do that.
			if p.send {

				// Get the content of the feed-item.
				//
				// This has to be done ahead of sending email,
				// as we can use this to skip entries via
				// regular expression on the title/body contents.
				content := ""
				content, err = item.HTMLContent()
				if err != nil {
					content = item.RawContent()
				}

				// Should we skip this entry?
				//
				// Skipping here means that we don't send an email,
				// however we do mark it as read - so it will only
				// be processed once.
				if !p.shouldSkip(entry, item.Title, content) {

					// Convert the content to text.
					text := html2text.HTML2Text(content)

					// Send the mail
					helper := emailer.New(feed, item, entry.Options)
					err = helper.Sendmail(recipients, text, content)
					if err != nil {
						return err
					}
				}
			}
		} else {

			// Bump the count
			seen++

		}

		// Mark the item as having been seen, after the email was (probably) sent.
		//
		// This does run the risk that sending mail fails, due to error, and
		// that keeps happening forever...
		err = p.recordItem(entry.URL, item.Link)
		if err != nil {
			return err
		}

		// Since we've marked this item as being seen we can remove the legacy
		// state-file
		item.RemoveLegacy()
	}

	// Show how many entries we've found in the feed.
	p.message(fmt.Sprintf("\t%02d entries already seen", seen))
	p.message(fmt.Sprintf("\t%02d entries not seen before", unseen))

	// Now prune the items in this feed
	err = p.pruneFeed(entry.URL, items)
	if err != nil {
		return fmt.Errorf("error pruning boltdb for %s: %s", entry.URL, err)
	}

	return nil
}

// seenItem returns true if we've seen this item.
//
// It does this by checking the BoltDB in which we record state.
func (p *Processor) seenItem(feed string, entry string) bool {
	val := ""

	p.dbHandle.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(feed))
		v := b.Get([]byte(entry))
		if v != nil {
			val = string(v)
		}
		return nil
	})

	if val == "" {
		return false
	}
	return true
}

// recordItem marks an URL as having been seen.
//
// It does this by updating the BoltDB in which we record state.
func (p *Processor) recordItem(feed string, entry string) error {
	err := p.dbHandle.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(feed))
		err := b.Put([]byte(entry), []byte("seen"))
		return err
	})
	return err
}

// pruneFeed will remove unknown items from our state database
func (p *Processor) pruneFeed(feed string, items []string) error {

	// A list of items to remove
	toRemove := []string{}

	// Create a map of the items we've already seen
	seen := make(map[string]bool)
	for _, str := range items {
		seen[str] = true
	}

	// Select the bucket, which we know will exist,
	// and see if we should remove any of the keys
	// that are present.
	//
	// (i.e. Remove the ones that are not in teh map above)
	p.dbHandle.View(func(tx *bbolt.Tx) error {

		// Select the bucket, which we know must exist
		b := tx.Bucket([]byte(feed))

		c := b.Cursor()

		for k, _ := c.First(); k != nil; k, _ = c.Next() {

			key := string(k)
			_, ok := seen[key]
			if !ok {
				toRemove = append(toRemove, key)
			}
		}
		return nil
	})

	// Now for each item we should remove we should .. remove it.
	for _, ent := range toRemove {
		p.message(fmt.Sprintf("expiring feed entry %s", ent))

		err := p.dbHandle.Update(func(tx *bbolt.Tx) error {
			b := tx.Bucket([]byte(feed))
			return b.Delete([]byte(ent))
		})
		if err != nil {
			return fmt.Errorf("failed to remove %s - %s", ent, err)
		}

	}

	return nil

	// We
}

// shouldSkip returns true if this entry should be skipped/ignored.
//
// Our configuration file allows a series of per-feed configuration items,
// and those allow skipping the entry by regular expression matches on
// the item title or body.
//
// Similarly there is an `include` setting which will ensure we only
// email items matching a particular regular expression.
//
// Note that if an entry should be skipped it is still marked as
// having been read, but no email is sent.
func (p *Processor) shouldSkip(config configfile.Feed, title string, content string) bool {

	// Walk over the options to see if there are any exclude* options
	// specified.
	for _, opt := range config.Options {

		// Exclude by title?
		if opt.Name == "exclude-title" {
			match, _ := regexp.MatchString(opt.Value, title)
			if match {
				p.message(fmt.Sprintf("\t\t\tSkipping due to 'exclude-title' match of '%s'.", opt.Value))

				// True: skip/ignore this entry
				return true
			}
		}

		// Exclude by body/content?
		if opt.Name == "exclude" {

			match, _ := regexp.MatchString(opt.Value, content)
			if match {
				p.message(fmt.Sprintf("\t\t\tSkipping due to 'exclude' match of %s.", opt.Value))

				// True: skip/ignore this entry
				return true
			}
		}
	}

	// If we have an include-setting then we must skip the entry unless
	// it matches.
	//
	// There might be more than one include setting and a match against
	// any will suffice.
	//
	include := false

	for _, opt := range config.Options {
		if opt.Name == "include-title" {

			// We found (at least one) include option
			include = true

			// OK we've found a `include` setting,
			// so we MUST skip unless there is a match
			match, _ := regexp.MatchString(opt.Value, title)
			if match {
				p.message(fmt.Sprintf("\t\t\tIncluding as this entry's title matches %s.", opt.Value))

				// False: Do not skip/ignore this entry
				return false
			}
		}
		if opt.Name == "include" {

			// We found (at least one) include option
			include = true

			// OK we've found a `include` setting,
			// so we MUST skip unless there is a match
			match, _ := regexp.MatchString(opt.Value, content)
			if match {
				p.message(fmt.Sprintf("\t\t\tIncluding as this entry matches %s.", opt.Value))

				// False: Do not skip/ignore this entry
				return false
			}
		}
	}

	// If we had at least one "include" setting and we reach here
	// the we had no match.
	//
	// i.e. The entry did not include a string we regarded as mandatory.
	if include {
		p.message("\t\t\tExcluding entry, as it didn't match any include, or include-title, patterns")

		// True: skip/ignore this entry
		return true
	}

	// False: Do not skip/ignore this entry
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
