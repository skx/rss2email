//
// This is the cron-subcommand.
//
// It is a little ropy as it was ported in a hurry.
//

package main

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"html"
	"io/ioutil"
	"mime/quotedprintable"
	"net/http"
	"os"
	"os/exec"
	"text/template"

	"github.com/google/subcommands"
	"github.com/k3a/html2text"
	"github.com/mmcdole/gofeed"
)

// Template is our text/template which is designed used to send an
// email to the local user.  We're using a template such that we
// can send both HTML and Text versions of the RSS feed item.
var Template = `Content-Type: multipart/mixed; boundary=21ee3da964c7bf70def62adb9ee1a061747003c026e363e47231258c48f1
From: {{.From}}
To: {{.To}}
Subject: [rss2email] {{.Subject}}
Mime-Version: 1.0

--21ee3da964c7bf70def62adb9ee1a061747003c026e363e47231258c48f1
Content-Type: multipart/related; boundary=76a1282373c08a65dd49db1dea2c55111fda9a715c89720a844fabb7d497

--76a1282373c08a65dd49db1dea2c55111fda9a715c89720a844fabb7d497
Content-Type: multipart/alternative; boundary=4186c39e13b2140c88094b3933206336f2bb3948db7ecf064c7a7d7473f2

--4186c39e13b2140c88094b3933206336f2bb3948db7ecf064c7a7d7473f2
Content-Type: text/plain; charset=UTF-8
Content-Transfer-Encoding: quoted-printable

{{.Text}}
--4186c39e13b2140c88094b3933206336f2bb3948db7ecf064c7a7d7473f2
Content-Type: text/html; charset=UTF-8
Content-Transfer-Encoding: quoted-printable

{{.HTML}}
--4186c39e13b2140c88094b3933206336f2bb3948db7ecf064c7a7d7473f2--

--76a1282373c08a65dd49db1dea2c55111fda9a715c89720a844fabb7d497--
--21ee3da964c7bf70def62adb9ee1a061747003c026e363e47231258c48f1--
`

// toQuotedPrintable will convert the given input-string to a
// quoted-printable format.  This is required for our MIME-part
// body.
func (p *cronCmd) toQuotedPrintable(s string) (string, error) {
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
func (p *cronCmd) SendMail(addr string, subject string, textstr string, htmlstr string) error {
	var err error

	//
	// Ensure we have a recipient.
	//
	if addr == "" {
		e := errors.New("Empty recipient address, is '$LOGNAME' set?")
		fmt.Printf("%s\n", e.Error())
		return e
	}

	//
	// Here is a temporary structure we'll use to popular our email
	// template.
	//
	type TemplateParms struct {
		To      string
		From    string
		Text    string
		HTML    string
		Subject string
	}

	//
	// Populate it appropriately.
	//
	var x TemplateParms
	x.To = addr
	x.From = "user@rss2email.invalid"
	x.Text, err = p.toQuotedPrintable(textstr)
	if err != nil {
		return err
	}
	x.HTML, err = p.toQuotedPrintable(html.UnescapeString(htmlstr))
	if err != nil {
		return err
	}
	x.Subject = subject

	//
	// Render our template into a buffer.
	//
	src := string(Template)
	t := template.Must(template.New("tmpl").Parse(src))
	buf := &bytes.Buffer{}
	err = t.Execute(buf, x)

	//
	// Prepare to run sendmail, with a pipe we can write our
	// message to.
	//
	sendmail := exec.Command("/usr/sbin/sendmail", "-f", addr, addr)
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
	var out []byte
	out, err = ioutil.ReadAll(stdout)
	if err != nil {
		fmt.Printf("Error reading mail output: %s\n", err.Error())
		return nil
	}

	if p.verbose {
		fmt.Printf("%s\n", out)
	}
	err = sendmail.Wait()

	if err != nil {
		fmt.Printf("Waiting for process to terminate failed: %s\n", err.Error())
	}
	return nil
}

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

// GUID2Hash converts a GUID into something we can use on the filesystem,
// via the use of the SHA1-hash.
func (p *cronCmd) GUID2Hash(guid string) string {
	hasher := sha1.New()
	hasher.Write([]byte(guid))
	hashBytes := hasher.Sum(nil)

	// Hexadecimal conversion
	hexSha1 := hex.EncodeToString(hashBytes)

	return hexSha1
}

// HasSeen will return true if we've previously emailed this feed-entry.
func (p *cronCmd) HasSeen(item *gofeed.Item) bool {
	sha := p.GUID2Hash(item.GUID)
	if _, err := os.Stat(os.Getenv("HOME") + "/.rss2email/seen/" + sha); os.IsNotExist(err) {
		return false
	}
	return true
}

// RecordSeen will update our state to record the given GUID as having
// been seen.
func (p *cronCmd) RecordSeen(item *gofeed.Item) {
	dir := os.Getenv("HOME") + "/.rss2email/seen"
	os.MkdirAll(dir, os.ModePerm)

	d1 := []byte(item.Link)
	sha := p.GUID2Hash(item.GUID)
	_ = ioutil.WriteFile(dir+"/"+sha, d1, 0644)
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
		if !p.HasSeen(i) {

			if p.verbose {
				fmt.Printf("New item: %s\n", i.GUID)
			}

			// Convert the body to text.
			text := html2text.HTML2Text(i.Content)

			// Send the email
			err := p.SendMail(os.Getenv("LOGNAME"), i.Title, text, i.Content)

			// Only then record this item as having been seen
			if err == nil {
				p.RecordSeen(i)
			}
		}
	}
}

// The options set by our command-line flags.
//
type cronCmd struct {
	verbose bool
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
	return subcommands.ExitSuccess
}
