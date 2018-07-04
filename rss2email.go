//
// RSS2Email.
//
// When launched read ~/.rss2email/feeds which will contain a list of URLS
// to fetch.
//
// For each feed send new entries via email.
//
//

package main

import (
	"bufio"
	"crypto/sha1"
	"encoding/hex"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strings"

	"os"

	"github.com/k3a/html2text"
	"github.com/mmcdole/gofeed"
)

// VERBOSE is a value set via a command-line flag, and controls how
// noisy we should be.
var VERBOSE = false

// VERSION is our version, as set via CI.
var version = "master/unreleased"

// SendMail is a simple function that emails the given address.
//
// This is done via `/usr/sbin/sendmail` rather than via the use of SMTP.
//
func SendMail(addr string, subject string, body string) {
	sendmail := exec.Command("/usr/sbin/sendmail", "-f", addr, addr)
	stdin, err := sendmail.StdinPipe()
	if err != nil {
		fmt.Printf("Error sending email: %s\n", err.Error())
		return
	}

	stdout, err := sendmail.StdoutPipe()
	if err != nil {
		fmt.Printf("Error sending email: %s\n", err.Error())
		return
	}

	// What we'll send
	msg := ""
	msg += "To: " + addr + "\n"
	msg += "From: user@rss2email.invalid\n"
	msg += "Subject: [rss2email] " + subject + "\n"
	msg += "\n"
	msg += body

	sendmail.Start()
	stdin.Write([]byte(msg))
	stdin.Close()
	_, _ = ioutil.ReadAll(stdout)
	sendmail.Wait()
}

// FetchFeed fetches a feed from the remote URL.
//
// We must use this instead of the URL handler that the feed-parser supports
// because reddit, and some other sites, will just return a HTTP error-code
// if we're using a standard "spider" User-Agent.
//
func FetchFeed(url string) (string, error) {
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
func GUID2Hash(guid string) string {
	hasher := sha1.New()
	hasher.Write([]byte(guid))
	hashBytes := hasher.Sum(nil)

	// Hexadecimal conversion
	hexSha1 := hex.EncodeToString(hashBytes)

	return hexSha1
}

// HasSeen will return true if we've previously emailed this feed-entry.
func HasSeen(item *gofeed.Item) bool {
	sha := GUID2Hash(item.GUID)
	if _, err := os.Stat(os.Getenv("HOME") + "/.rss2email/seen/" + sha); os.IsNotExist(err) {
		return false
	}
	return true
}

// RecordSeen will update our state to record the given GUID as having
// been seen.
func RecordSeen(item *gofeed.Item) {
	dir := os.Getenv("HOME") + "/.rss2email/seen"
	os.MkdirAll(dir, os.ModePerm)

	d1 := []byte(item.Link)
	sha := GUID2Hash(item.GUID)
	_ = ioutil.WriteFile(dir+"/"+sha, d1, 0644)
}

// ProcessURL takes an URL as input, fetches the contents, and then
// processes each feed item found within it.
func ProcessURL(input string) {

	if VERBOSE {
		fmt.Printf("Fetching %s\n", input)
	}

	// Fetch the URL
	txt, err := FetchFeed(input)
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

	if VERBOSE {
		fmt.Printf("Found %d entries\n", len(feed.Items))
	}

	// For each entry in the feed ..
	for _, i := range feed.Items {

		// If we've not already notified about this one.
		if !HasSeen(i) {

			if VERBOSE {
				fmt.Printf("New item: %s\n", i.GUID)
			}

			// Convert the body to text.
			text := html2text.HTML2Text(i.Content)

			// Send the email
			SendMail(os.Getenv("USER"), i.Title, text)

			// Only then record this item as having been seen
			RecordSeen(i)
		}
	}
}

// main is our entry-point
func main() {

	verbose := flag.Bool("verbose", false, "Should we be verbose?")
	showver := flag.Bool("version", false, "Show our version and terminate.")
	flag.Parse()

	if *showver {
		fmt.Printf("rss2email %s\n", version)
		os.Exit(0)
	}

	// Update our global variables appropriately
	VERBOSE = *verbose

	//
	// Open our input-file
	//
	path := os.Getenv("HOME") + "/.rss2email/feeds"
	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("Error opening %s - %s\n", path, err.Error())
		return
	}
	defer file.Close()

	//
	// Process it line by line.
	//
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		tmp := scanner.Text()
		tmp = strings.TrimSpace(tmp)

		//
		// Skip lines that begin with a comment.
		//
		if (tmp != "") && (!strings.HasPrefix(tmp, "#")) {
			ProcessURL(tmp)
		}
	}
}
