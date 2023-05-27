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
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/k3a/html2text"
	"github.com/skx/rss2email/configfile"
	"github.com/skx/rss2email/httpfetch"
	"github.com/skx/rss2email/processor/emailer"
	"github.com/skx/rss2email/state"
	"github.com/skx/rss2email/withstate"
	"go.etcd.io/bbolt"
)

// Processor stores our state
type Processor struct {

	// send controls whether we send emails, or just pretend to.
	send bool

	// verbose denotes how verbose we should be in execution.
	verbose bool

	// database holds a handle to the BoltDB database we use to
	// store feed-entry state within.
	dbHandle *bbolt.DB
}

// New creates a new Processor object.
//
// This might return an error if we fail to open the database we use
// for maintaining state.
func New() (*Processor, error) {

	// Ensure we have a state-directory.
	dir := state.Directory()
	errM := os.MkdirAll(dir, 0666)
	if errM != nil {
		return nil, errM
	}

	// Now create the database, if missing, or open it if it exists.
	db, err := bbolt.Open(filepath.Join(dir, "state.db"), 0666, nil)
	if err != nil {
		return nil, err
	}

	return &Processor{send: true, dbHandle: db}, nil
}

// Close should be called to cleanup our internal database-handle.
func (p *Processor) Close() {
	p.dbHandle.Close()
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

	// Now do the parsing
	entries, err := conf.Parse()
	if err != nil {
		errors = append(errors, fmt.Errorf("error with config-file %s - %s", conf.Path(), err))
		return errors
	}

	// Keep track of the previous hostname from which we fetched a feed
	prev := ""

	// Keep track of each feed we've processed
	feeds := []string{}

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
		err = p.dbHandle.Update(func(tx *bbolt.Tx) error {
			_, err2 := tx.CreateBucketIfNotExists([]byte(entry.URL))
			if err2 != nil {
				return fmt.Errorf("create bucket failed: %s", err)
			}
			return nil
		})

		// If we have a DB-error then we return, this shouldn't happen.
		if err != nil {
			errors = append(errors, fmt.Errorf("error creating bucket for %s: %s", entry.URL, err))
			return (errors)
		}

		// Record the URL of the feed in our list,
		// which is used for reaping obsolete feeds
		feeds = append(feeds, entry.URL)

		// Should we sleep before getting this feed?
		sleep := 0

		// We default to notifying the global recipient-list.
		//
		// But there might be a per-feed set of recipients which
		// we'll prefer if available.
		feedRecipients := recipients

		// parse the hostname form the URL
		//
		// We do this because some remote sites, such as Reddit,
		// will apply rate-limiting if we make too many consecutive
		// requests in a short period of time.
		host := ""
		u, err2 := url.Parse(entry.URL)
		if err2 == nil {
			host = u.Host
		}

		// Are we fetching from the same host as the previous feed?
		// If so then we'll add a delay to try to avoid annoying that
		// host.
		if host == prev {
			p.message(fmt.Sprintf("Fetching from same host as previous feed, %s, adding 5s delay", host))
			sleep = 5
		}

		// Now look at each per-feed option, if any are set.
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

	// Reap feeds which are obsolete.
	err = p.pruneUnknownFeeds(feeds)
	if err != nil {
		errors = append(errors, err)
	}

	// All feeds were processed, return any errors we found along the way
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

	// Is there a tag set for this feed?
	tag := ""

	// Look at each per-feed option to determine that
	for _, opt := range entry.Options {
		if strings.ToLower(opt.Name) == "tag" {
			tag = opt.Value
		}
	}

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
		// so that we can get access to the content easily.
		//
		// Specifically here we turn relative URLs into absolute
		// ones, using the feed link as the base.
		//
		// We have some legacy code for determining "new" vs "seen",
		// but that will go away in the future.
		item := withstate.FeedItem{Item: xp}

		// Set the tag for the item, if present.
		if tag != "" {
			item.Tag = tag
		}

		// Keep track of the fact that we saw this feed-item.
		//
		// This is used for pruning the BoltDB state file.
		items = append(items, item.Link)

		// Assume this feed-entry is new, and we've not seen it
		// in the past.
		isNew := true

		// Is this link already in the BoltDB?
		//
		// If so it's not new.
		if p.seenItem(entry.URL, item.Link) {
			isNew = false
		}

		// If this entry is new then we must notify, unless
		// the entry is excluded for some reason.
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

				// check for regular expressions
				skip := p.shouldSkip(entry, item.Title, content)

				// check for age (exclude-older)
				skip = skip || p.shouldSkipOlder(entry, item.Published)

				if !skip {
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

		// Mark the item as having been seen, after the email
		// was (probably) sent.
		//
		// This does run the risk that sending mail fails,
		// due to error, and that keeps happening forever...
		err = p.recordItem(entry.URL, item.Link)
		if err != nil {
			return err
		}
	}

	// Show how many entries we've found in the feed.
	p.message(fmt.Sprintf("\t%02d entries already seen", seen))
	p.message(fmt.Sprintf("\t%02d entries not seen before", unseen))

	// Now prune the items in this feed.
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

	err := p.dbHandle.View(func(tx *bbolt.Tx) error {

		// Select the feed-bucket
		b := tx.Bucket([]byte(feed))

		// Get the entry with key of the feed URL.
		v := b.Get([]byte(entry))
		if v != nil {
			val = string(v)
		}
		return nil
	})
	if err != nil {
		fmt.Printf("seenItem failed:%s\n", err)
	}

	return val != ""
}

// recordItem marks an URL as having been seen.
//
// It does this by updating the BoltDB in which we record state.
func (p *Processor) recordItem(feed string, entry string) error {

	err := p.dbHandle.Update(func(tx *bbolt.Tx) error {

		// Select the feed-bucket
		b := tx.Bucket([]byte(feed))

		// Set a value "seen" to the key of the feed item link
		err := b.Put([]byte(entry), []byte("seen"))
		return err
	})
	return err
}

// pruneFeed will remove unknown items from our state database.
//
// Here we are given the URL of the feed, and a set of feed-item links,
// we remove items which are no longer in the remote feed.
//
// See also `pruneUnknownFeeds` for removing feeds which are no longer
// fetched at all.
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
	// (i.e. Remove the ones that are not in the map above)
	err := p.dbHandle.View(func(tx *bbolt.Tx) error {

		// Select the bucket, which we know must exist
		b := tx.Bucket([]byte(feed))

		// Get a cursor to the key=value entries in the bucket
		c := b.Cursor()

		// Iterate over the key/value pairs.
		for k, _ := c.First(); k != nil; k, _ = c.Next() {

			// Convert the key to a string
			key := string(k)

			// Is this in our list of seen entries?
			_, ok := seen[key]
			if !ok {
				// If not remove the key/value pair
				toRemove = append(toRemove, key)
			}
		}
		return nil
	})

	if err != nil {
		fmt.Printf("failed to iterate over keys:%s\n", err)
		return err
	}

	// Remove each entry that we were supposed to remove.
	for _, ent := range toRemove {
		p.message(fmt.Sprintf("expiring feed entry %s", ent))

		err := p.dbHandle.Update(func(tx *bbolt.Tx) error {

			// Select the bucket
			b := tx.Bucket([]byte(feed))

			// Delete the key=value pair.
			return b.Delete([]byte(ent))
		})
		if err != nil {
			return fmt.Errorf("failed to remove %s - %s", ent, err)
		}
	}

	return nil
}

// pruneUnknownFeeds removes feeds from our database which are no longer
// contained within our configuration file.
//
// To recap BoltDB has a notion of buckets, which are used to store key=value
// pairs.  We create a bucket for every feed which is present in our
// configuration value, then use the URL of feed-items as the keys.
//
// Here we remove buckets which are obsolete.
func (p *Processor) pruneUnknownFeeds(feeds []string) error {

	// Create a map for lookup
	seen := make(map[string]bool)
	for _, str := range feeds {
		seen[str] = true
	}

	// Now walk the database and see which buckets should be removed.
	toRemove := []string{}

	err := p.dbHandle.View(func(tx *bbolt.Tx) error {

		return tx.ForEach(func(bucketName []byte, _ *bbolt.Bucket) error {

			// Does this name exist in our map?
			_, ok := seen[string(bucketName)]

			// If not then it should be removed.
			if !ok {
				toRemove = append(toRemove, string(bucketName))
			}
			return nil
		})
	})

	if err != nil {
		fmt.Printf("failed to find orphaned buckets;%s\n", err)
		return err
	}
	// For each bucket we need to remove, remove it
	for _, bucket := range toRemove {

		// We're going to remove the bucket
		p.message(fmt.Sprintf("removing feed-bucket %s", bucket))

		err := p.dbHandle.Update(func(tx *bbolt.Tx) error {

			// Select the bucket, which we know must exist
			b := tx.Bucket([]byte(bucket))

			// Get a cursor to the key=value entries in the bucket
			c := b.Cursor()

			// Iterate over the key/value pairs and delete them.
			for k, _ := c.First(); k != nil; k, _ = c.Next() {

				err := b.Delete(k)
				if err != nil {
					return (fmt.Errorf("failed to delete bucket key %s:%s - %s", bucket, k, err))
				}
			}

			// Now delete the bucket itself
			err := tx.DeleteBucket([]byte(bucket))
			if err != nil {
				return fmt.Errorf("failed to remove bucket %s: %s", bucket, err)
			}

			return nil
		})
		if err != nil {
			return fmt.Errorf("error removing bucket %s: %s", bucket, err)
		}
	}

	return nil
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

// shouldSkipOlder returns true if this entry should be skipped due to age.
//
// Age is configured with "exclude-older" in days.
func (p *Processor) shouldSkipOlder(config configfile.Feed, published string) bool {

	// Walk over the options to see if there are any exclude-age options
	// specified.
	for _, opt := range config.Options {

		if opt.Name == "exclude-older" {
			pubTime, err := time.Parse(time.RFC1123, published)
			if err != nil {
				p.message(fmt.Sprintf("exclude-older: skipped due to failed parse of item.published as date %s", err))
				return true
			}
			f, err := strconv.ParseFloat(opt.Value, 32)
			if err != nil {
				p.message(fmt.Sprintf("exclude-older: failed to parse config option exclude-older as float %s", err))
				return false
			}

			delta := time.Second * time.Duration(f*24*60*60)
			if pubTime.Add(delta).Before(time.Now()) {
				p.message(fmt.Sprintf("\t\t\tSkipping due to 'exclude-older' (age %.1f days)", time.Since(pubTime).Hours()/24))
				return true
			}
		}
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
