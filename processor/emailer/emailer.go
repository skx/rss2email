// Package emailer is responsible for sending out a feed
// item via email.
//
// There are two ways emails are sent:
//
//  1. Via spawning /usr/sbin/sendmail.
//
//  2. Via SMTP.
//
// The choice is made based upon the presence of environmental
// variables.
package emailer

import (
	"bytes"
	"errors"
	"fmt"
	"html"
	"io"
	"log/slog"
	"mime/quotedprintable"
	"net/smtp"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/mmcdole/gofeed"
	"github.com/skx/rss2email/configfile"
	"github.com/skx/rss2email/state"
	emailtemplate "github.com/skx/rss2email/template"
	"github.com/skx/rss2email/withstate"
)

// Emailer stores our state
type Emailer struct {

	// Feed is the source feed from which this item came
	feed *gofeed.Feed

	// Item is the feed item itself
	item withstate.FeedItem

	// Config options for the feed.
	opts []configfile.Option

	// logger contains a dedicated logging object
	logger *slog.Logger
}

// New creates a new Emailer object.
//
// The arguments are the source feed, the feed item which is being notified,
// and any associated configuration values from the source feed.
func New(feed *gofeed.Feed, item withstate.FeedItem, opts []configfile.Option, log *slog.Logger) *Emailer {

	// Default options
	obj := &Emailer{feed: feed, item: item, opts: opts}

	// Create a new logger
	obj.logger = log.With(
		slog.Group("email",
			slog.String("link", item.Link),
			slog.String("title", item.Title),
		),
	)

	return obj
}

// env returns the contents of an environmental variable.
//
// This function exists to be used by our email-template.
func env(s string) string {
	return (os.Getenv(s))
}

// split converts a string to an array.
//
// This function exists to be used by our email-template.
func split(in string, delim string) []string {
	return strings.Split(in, delim)
}

// loadTemplate loads the template used for sending the email notification.
func (e *Emailer) loadTemplate() (*template.Template, error) {

	// Load the default template from the embedded resource.
	content := emailtemplate.EmailTemplate()

	// The directory within which we maintain state
	stateDir := state.Directory()

	// The path to the overridden template
	override := filepath.Join(stateDir, "email.tmpl")

	// If a per feed template was set, get it here.
	for _, opt := range e.opts {
		if opt.Name == "template" {
			override = filepath.Join(stateDir, opt.Value)
		}
	}

	// If the file exists, use it.
	_, err := os.Stat(override)
	if !os.IsNotExist(err) {
		content, err = os.ReadFile(override)
		if err != nil {

			e.logger.Debug("could not load template override file",
				slog.String("file", override),
				slog.String("error", err.Error()))

			return nil, fmt.Errorf("failed to read %s: %s", override, err.Error())
		}
	}

	//
	// Function map allows exporting functions to the template
	//
	funcMap := template.FuncMap{
		"env":            env,
		"quoteprintable": toQuotedPrintable,
		"split":          split,
		"encodeHeader":   encodeHeader,
	}

	tmpl := template.Must(template.New("email.tmpl").Funcs(funcMap).Parse(string(content)))

	return tmpl, nil
}

// toQuotedPrintable will convert the given input-string to a
// quoted-printable format.  This is required for our MIME-part
// body.
//
// NOTE: We use this function both directly, and from within our template.
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

// Encode email header entries to comply with the 7bit ASCII restriction
// of RFC 5322 according to RFC 2047.
//
// We use quotedprintable encoding only if necessary.
func encodeHeader(s string) string {
	se, err := toQuotedPrintable(s)
	if (err != nil) || (len(se) == len(s)) {
		return s
	}
	se = strings.Replace(strings.Replace(se, "?", "=3F", -1), " ", "=20", -1)
	se = strings.Replace(se, "=\r\n", "", -1) // remove soft line breaks
	return "=?utf-8?Q?" + se + "?="
}

// Sendmail is a simple function that emails the given address.
//
// We send a MIME message with both a plain-text and a HTML-version of the
// message.  This should be nicer for users.
func (e *Emailer) Sendmail(addresses []string, textstr string, htmlstr string) error {

	var err error

	//
	// Ensure we have a recipient.
	//
	if len(addresses) < 1 {

		e.logger.Error("missing recipient address")

		e := errors.New("empty recipient address, did you not setup a recipient?")
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
			From      string
			HTML      string
			Link      string
			Subject   string
			Tag       string
			Text      string
			To        string

			// In case people need access to fields
			// we've not wrapped/exported explicitly
			RSSFeed *gofeed.Feed
			RSSItem withstate.FeedItem
		}

		//
		// Populate it appropriately.
		//
		var x TemplateParms
		x.Feed = e.feed.Link
		x.FeedTitle = e.feed.Title
		x.From = addr
		x.Link = e.item.Link
		x.Subject = e.item.Title
		x.To = addr
		x.RSSFeed = e.feed
		x.RSSItem = e.item
		x.Tag = e.item.Tag

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
		// Load the template we're going to render.
		//
		var t *template.Template
		t, err = e.loadTemplate()
		if err != nil {
			return err
		}

		//
		// Render the template into the buffer.
		//
		buf := &bytes.Buffer{}
		err = t.Execute(buf, x)
		if err != nil {
			return err
		}

		//
		// Are we sending via SMTP?
		//
		if e.isSMTP() {

			e.logger.Debug("preparing to send email",
				slog.String("recipient", addr),
				slog.String("method", "smtp"))

			err := e.sendSMTP(addr, buf.Bytes())
			if err != nil {

				e.logger.Warn("error sending email",
					slog.String("recipient", addr),
					slog.String("method", "smtp"),
					slog.String("error", err.Error()))

				return err
			}

			e.logger.Debug("email sent",
				slog.String("recipient", addr),
				slog.String("method", "smtp"))

		} else {

			e.logger.Debug("preparing to send email",
				slog.String("recipient", addr),
				slog.String("method", "sendmail"))

			err := e.sendSendmail(addr, buf.Bytes())
			if err != nil {
				e.logger.Warn("error sending email",
					slog.String("recipient", addr),
					slog.String("method", "sendmail"),
					slog.String("error", err.Error()))
				return err
			}

			e.logger.Debug("email sent",
				slog.String("recipient", addr),
				slog.String("method", "sendmail"))

		}
	}

	e.logger.Debug("emails sent",
		slog.Int("recipients", len(addresses)))

	return nil
}

// isSMTP determines whether we should use SMTP to send the email.
//
// We just check to see that the obvious mandatory parameters are set in the
// environment.  If they're wrong we'll get an error at delivery time, as
// expected.
func (e *Emailer) isSMTP() bool {

	// Mandatory environmental variables
	vars := []string{"SMTP_HOST", "SMTP_USERNAME", "SMTP_PASSWORD"}

	for _, name := range vars {
		if os.Getenv(name) == "" {
			return false
		}
	}

	return true
}

// sendSMTP sends the content of the email to the destination address
// via SMTP.
func (e *Emailer) sendSMTP(to string, content []byte) error {

	// basics
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")

	p := 587
	if port != "" {
		n, err := strconv.Atoi(port)
		if err != nil {

			e.logger.Warn("error converting SMTP_PORT to integer",
				slog.String("port", port),
				slog.String("error", err.Error()))

			return err
		}
		p = n
	}

	// auth
	user := os.Getenv("SMTP_USERNAME")
	pass := os.Getenv("SMTP_PASSWORD")

	// Authenticate
	auth := smtp.PlainAuth("", user, pass, host)

	// Get the mailserver
	addr := fmt.Sprintf("%s:%d", host, p)

	// Send the mail
	err := smtp.SendMail(addr, auth, to, []string{to}, content)

	return err
}

// sendSendmail sends the content of the email to the destination address
// via /usr/sbin/sendmail
func (e *Emailer) sendSendmail(addr string, content []byte) error {

	// Get the command to run.
	sendmail := exec.Command("/usr/sbin/sendmail", "-i", "-f", addr, addr)
	stdin, err := sendmail.StdinPipe()
	if err != nil {

		e.logger.Warn("error creating STDIN pipe to sendmail",
			slog.String("recipient", addr),
			slog.String("error", err.Error()))

		return err
	}

	//
	// Get the output pipe.
	//
	stdout, err := sendmail.StdoutPipe()
	if err != nil {

		e.logger.Warn("error creating STDOUT pipe to sendmail",
			slog.String("recipient", addr),
			slog.String("error", err.Error()))

		return err
	}

	//
	// Run the command, and pipe in the rendered template-result
	//
	err = sendmail.Start()
	if err != nil {

		e.logger.Warn("error starting sendmail",
			slog.String("recipient", addr),
			slog.String("error", err.Error()))

		return err
	}
	_, err = stdin.Write(content)
	if err != nil {

		e.logger.Warn("error writing to sendmail pipe",
			slog.String("recipient", addr),
			slog.String("error", err.Error()))

		return err
	}
	stdin.Close()

	//
	// Read the output of Sendmail.
	//
	_, err = io.ReadAll(stdout)
	if err != nil {

		e.logger.Warn("error reading from sendmail pipe",
			slog.String("recipient", addr),
			slog.String("error", err.Error()))

		return err
	}

	//
	// Wait for the command to complete.
	//
	err = sendmail.Wait()
	if err != nil {

		e.logger.Warn("error awaiting sendmail completion",
			slog.String("recipient", addr),
			slog.String("error", err.Error()))

		return err
	}

	return err
}
