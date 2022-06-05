// Package withstate provides a simple wrapper of the gofeed.Item, which
// allows simple tracking of the seen vs. unseen (new vs. old) state of
// an RSS feeds' entry.
//
// State for a feed-item is stored upon the local filesystem.
package withstate

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/mmcdole/gofeed"
)


// FeedItem is a structure wrapping a gofeed.Item, to allow us to record
// state.
type FeedItem struct {

	// Wrapped structure
	*gofeed.Item

	// Tag is a field that can be set for this feed item,
	// inside our configuration file.
	Tag string
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
