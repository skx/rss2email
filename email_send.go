// This file contains the code which is used to send email.
//
// Emails are sent via a simple text/template, and each email will
// have both a text and a HTML part to it.
//

package main

import (
	"bytes"
	"errors"
	"fmt"
	"html"
	"io/ioutil"
	"mime/quotedprintable"
	"os"
	"os/exec"
	"os/user"
	"path"
	"text/template"

	"github.com/mmcdole/gofeed"
	"github.com/skx/rss2email/withstate"
)

var (
	// The compiled template we use to generate our email
	tmpl *template.Template
)

// setupTemplate loads the template we use for generating the email
// notification.
func setupTemplate() *template.Template {

	// Already setup?  Return the template
	if tmpl != nil {
		return tmpl
	}

	// Load the default template from the embedded resource.
	content, err := getResource("data/email.tmpl")
	if err != nil {
		fmt.Printf("failed to load embedded resource: %s\n", err.Error())
		os.Exit(1)
	}

	//
	// Is there an on-disk template instead?  If so use it.
	//
	home := os.Getenv("HOME")

	// If that fails then get the current user, and use
	// their home if possible.
	if home == "" {
		usr, errr := user.Current()
		if errr == nil {
			home = usr.HomeDir
		}
	}

	// The path to the overridden template
	override := path.Join(home, ".rss2email", "email.tmpl")

	// If the file exists, use it.
	_, err = os.Stat(override)
	if !os.IsNotExist(err) {
		content, err = ioutil.ReadFile(override)
		if err != nil {
			fmt.Printf("failed to read %s: %s\n", override, err.Error())
			os.Exit(1)
		}
	}

	//
	// Function map allows exporting functions to the template
	//
	funcMap := template.FuncMap{
		"quoteprintable": toQuotedPrintable,
	}

	tmpl = template.Must(template.New("email.tmpl").Funcs(funcMap).Parse(string(content)))

	return tmpl
}

// toQuotedPrintable will convert the given input-string to a
// quoted-printable format.  This is required for our MIME-part
// body.
//
// NOTE: We use this function both directly, and from within our
// template.
func toQuotedPrintable(s string) (string, error) {
	var ac bytes.Buffer
	w := quotedprintable.NewWriter(&ac)
	_, err := w.Write([]byte(s))
	if err != nil {
		return "", err
	}
	err = w.Close()
	if err != nil {
		return "", err
	}
	return ac.String(), nil
}

// SendMail is a simple function that emails the given address.
//
// This is done via `/usr/sbin/sendmail` rather than via the use of SMTP.
//
// We send a MIME message with both a plain-text and a HTML-version of the
// message.  This should be nicer for users.
func SendMail(feed *gofeed.Feed, item withstate.FeedItem, addresses []string, textstr string, htmlstr string) error {
	var err error

	//
	// Ensure we have a recipient.
	//
	if len(addresses) < 1 {
		e := errors.New("empty recipient address, did you not setup a recipient?")
		fmt.Printf("%s\n", e.Error())
		return e
	}

	//
	// Process each address
	//
	for _, addr := range addresses {

		//
		// Here is a temporary structure we'll use to popular our email
		// template.
		//
		type TemplateParms struct {
			Feed      string
			FeedTitle string
			To        string
			From      string
			Text      string
			HTML      string
			Subject   string
			Link      string

			// In case people need access to fields
			// we've not wrapped/exported explicitly
			RSSFeed *gofeed.Feed
			RSSItem withstate.FeedItem
		}

		//
		// Populate it appropriately.
		//
		var x TemplateParms
		x.Feed = feed.Link
		x.FeedTitle = feed.Title
		x.From = addr
		x.Link = item.Link
		x.Subject = item.Title
		x.To = addr
		x.RSSFeed = feed
		x.RSSItem = item

		// The real meat of the mail is the text & HTML
		// parts.  They need to be encoded, unconditionally.
		x.Text, err = toQuotedPrintable(textstr)
		if err != nil {
			return err
		}
		x.HTML, err = toQuotedPrintable(html.UnescapeString(htmlstr))
		if err != nil {
			return err
		}

		//
		// Render our template into a buffer.
		//
		t := setupTemplate()
		buf := &bytes.Buffer{}
		err = t.Execute(buf, x)
		if err != nil {
			return err
		}

		//
		// Prepare to run sendmail, with a pipe we can write our
		// message to.
		//
		sendmail := exec.Command("/usr/sbin/sendmail", "-i", "-f", addr, addr)
		stdin, err := sendmail.StdinPipe()
		if err != nil {
			fmt.Printf("Error sending email: %s\n", err.Error())
			return err
		}

		//
		// Get the output pipe.
		//
		stdout, err := sendmail.StdoutPipe()
		if err != nil {
			fmt.Printf("Error sending email: %s\n", err.Error())
			return err
		}

		//
		// Run the command, and pipe in the rendered template-result
		//
		sendmail.Start()
		_, err = stdin.Write(buf.Bytes())
		if err != nil {
			fmt.Printf("Failed to write to sendmail pipe: %s\n", err.Error())
			return err
		}
		stdin.Close()

		//
		// Read the output of Sendmail.
		//
		_, err = ioutil.ReadAll(stdout)
		if err != nil {
			fmt.Printf("Error reading mail output: %s\n", err.Error())
			return nil
		}

		err = sendmail.Wait()

		if err != nil {
			fmt.Printf("Waiting for process to terminate failed: %s\n", err.Error())
		}
	}
	return nil
}
