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
	"os/exec"
	"text/template"
)

var (
	// The compiled template we use to generate our email
	tmpl *template.Template
)

// setupTemplate loads the template we use for generating the email
// notification.
func setupTemplate() *template.Template {

	// Already setup?  Return it.
	if tmpl != nil {
		return tmpl
	}

	// Otherwise create it now from a hardwired string.
	src := `Content-Type: multipart/mixed; boundary=21ee3da964c7bf70def62adb9ee1a061747003c026e363e47231258c48f1
From: {{.From}}
To: {{.To}}
Subject: [rss2email] {{.Subject}}
X-RSS-Link: {{.Link}}
X-RSS-Feed: {{.Feed}}
Mime-Version: 1.0

--21ee3da964c7bf70def62adb9ee1a061747003c026e363e47231258c48f1
Content-Type: multipart/related; boundary=76a1282373c08a65dd49db1dea2c55111fda9a715c89720a844fabb7d497

--76a1282373c08a65dd49db1dea2c55111fda9a715c89720a844fabb7d497
Content-Type: multipart/alternative; boundary=4186c39e13b2140c88094b3933206336f2bb3948db7ecf064c7a7d7473f2

--4186c39e13b2140c88094b3933206336f2bb3948db7ecf064c7a7d7473f2
Content-Type: text/plain; charset=UTF-8
Content-Transfer-Encoding: quoted-printable

{{quoteprintable .Link}}

{{.Text}}

{{quoteprintable .Link}}
--4186c39e13b2140c88094b3933206336f2bb3948db7ecf064c7a7d7473f2
Content-Type: text/html; charset=UTF-8
Content-Transfer-Encoding: quoted-printable

<p><a href=3D"{{quoteprintable .Link}}">{{quoteprintable .Subject}}</a></p>
{{.HTML}}
<p><a href=3D"{{quoteprintable .Link}}">{{quoteprintable .Subject}}</a></p>
--4186c39e13b2140c88094b3933206336f2bb3948db7ecf064c7a7d7473f2--

--76a1282373c08a65dd49db1dea2c55111fda9a715c89720a844fabb7d497--
--21ee3da964c7bf70def62adb9ee1a061747003c026e363e47231258c48f1--
`

	//
	// Function map allows exporting functions to the template
	//
	funcMap := template.FuncMap{
		"quoteprintable": toQuotedPrintable,
	}

	tmpl = template.Must(template.New("tmpl").Funcs(funcMap).Parse(src))

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
func SendMail(feedURL string, fromAddr string, addresses []string, subject string, link string, textstr string, htmlstr string) error {
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
			Feed    string
			To      string
			From    string
			Text    string
			HTML    string
			Subject string
			Link    string
		}

		//
		// Populate it appropriately.
		//
		var x TemplateParms
		x.Feed = feedURL
		x.From = addr
		x.Link = link
		x.Subject = subject
		x.To = addr

		// Sender-address might be overridden.
		if fromAddr != "" {
			x.From = fromAddr
		}

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
