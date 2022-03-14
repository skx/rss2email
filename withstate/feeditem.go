// Package withstate provides a simple wrapper of the gofeed.Item, which
// allows simple tracking of the seen vs. unseen (new vs. old) state of
// an RSS feeds' entry.
//
// State for a feed-item is stored upon the local filesystem.
package withstate

import (
	"crypto/sha1"
	"fmt"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/mmcdole/gofeed"
)

// statePrefix holds the prefix directory, and is used to
// allow changes during testing
//
// TODO: Remove this, as legacy.
var statePrefix string

// FeedItem is a structure wrapping a gofeed.Item, to allow us to record
// state.
type FeedItem struct {

	// Wrapped structure
	*gofeed.Item
}

// IsNew reports whether this particular feed-item is new.
//
// TODO: Remove this, as legacy.
func (item *FeedItem) IsNew() bool {

	file := item.path()
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return true
	}
	return false
}

// RawContent provides content or fallback to description
func (item *FeedItem) RawContent() string {
	// The body should be stored in the
	// "Content" field.
	content := item.Item.Content

	// If the Content field is empty then
	// use the Description instead, if it
	// is non-empty itself.
	if (content == "") && item.Item.Description != "" {
		content = item.Item.Description
	}

	return content
}

// HTMLContent provides processed HTML
func (item *FeedItem) HTMLContent() (string, error) {
	rawContent := item.RawContent()

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(rawContent))
	if err != nil {
		return rawContent, err
	}
	doc.Find("a, img").Each(func(i int, e *goquery.Selection) {
		var attr string
		switch e.Get(0).Data {
		case "a":
			attr = "href"
		case "img":
			attr = "src"
			e.RemoveAttr("loading")
			e.RemoveAttr("srcset")
		}

		ref, _ := e.Attr(attr)
		switch {
		case ref == "":
			return
		case strings.HasPrefix(ref, "data:"):
			return
		case strings.HasPrefix(ref, "http://"):
			return
		case strings.HasPrefix(ref, "https://"):
			return
		default:
			e.SetAttr(attr, item.patchReference(ref))
		}
	})
	doc.Find("iframe").Each(func(i int, iframe *goquery.Selection) {
		src, _ := iframe.Attr("src")
		if src == "" {
			iframe.Remove()
		} else {
			iframe.ReplaceWithHtml(fmt.Sprintf(`<a href="%s">%s</a>`, src, src))
		}
	})
	doc.Find("script").Each(func(i int, script *goquery.Selection) {
		script.Remove()
	})

	return doc.Html()
}

func (item *FeedItem) patchReference(ref string) string {
	resURL, err := url.Parse(ref)
	if err != nil {
		return ref
	}

	itemURL, err := url.Parse(item.Item.Link)
	if err != nil {
		return ref
	}

	if resURL.Host == "" {
		resURL.Host = itemURL.Host
	}
	if resURL.Scheme == "" {
		resURL.Scheme = itemURL.Scheme
	}

	return resURL.String()
}

// stateDirectory returns the directory beneath which we store state
//
// TODO: Remove this, as legacy.
func stateDirectory() string {

	// If we've found it already, or we've mocked it, then
	// return the appropriate value
	if statePrefix != "" {
		return statePrefix
	}

	// Default to using $HOME
	home := os.Getenv("HOME")

	if home == "" {
		// Get the current user, and use their home if possible.
		usr, err := user.Current()
		if err == nil {
			home = usr.HomeDir
		}
	}

	// Store the path for the future, and return it.
	statePrefix = filepath.Join(home, ".rss2email", "seen")
	return statePrefix
}

// path returns an appropriate marker-file, which is used to record
// the seen vs. unseen state of a particular entry.
//
// TODO: Remove this, as legacy.
func (item *FeedItem) path() string {

	guid := item.GUID
	if guid == "" {
		guid = item.Link
	}

	// Hash the item GUID and convert to hexadecimal
	hexSha1 := fmt.Sprintf("%x", sha1.Sum([]byte(guid)))

	// Finally join the path
	out := filepath.Join(stateDirectory(), hexSha1)
	return out

}

// RemoveLegacy removes the file that was used to record this
// entries state - because it is now stored in boltdb
//
// TODO: Remove this, as legacy.
func (item *FeedItem) RemoveLegacy() {

	// Remove the file - ignoring errors.
	os.Remove(item.path())
}
