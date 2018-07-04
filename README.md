[![Travis CI](https://img.shields.io/travis/skx/rss2email/master.svg?style=flat-square)](https://travis-ci.org/skx/rss2email)
[![Go Report Card](https://goreportcard.com/badge/github.com/skx/rss2email)](https://goreportcard.com/report/github.com/skx/rss2email)
[![license](https://img.shields.io/github/license/skx/rss2email.svg)](https://github.com/skx/rss2email/blob/master/LICENSE)
[![Release](https://img.shields.io/github/release/skx/rss2email.svg)](https://github.com/skx/rss2email/releases/latest)

# RSS2Email

This project is a naive port of the [r2e](https://github.com/wking/rss2email) project to golang.


## Rationale

I prefer to keep my server(s) pretty minimal, and replacing `r2e` allowed
me to remove a bunch of Python packages I otherwise have no need for:

      steve@ssh ~ $ sudo dpkg --purge rss2email
      Removing rss2email (1:3.9-2.1) ...

      ssh ~ # apt-get autoremove
      Reading package lists... Done
      Building dependency tree
      Reading state information... Done
      The following packages will be REMOVED:
       python-xdg python3-bs4 python3-chardet python3-feedparser python3-html2text
       python3-html5lib python3-lxml python3-six python3-webencodings
       upgraded, 0 newly installed, 9 to remove and 0 not upgraded.



## Installation

Assuming you have a working golang installation you can install the binary
via:

     go get -u github.com/skx/rss2email
     go install github.com/skx/rss2email

Once installed you'll want to configure a list of feeds, and add it to your
`crontab`.


## Configuration

RSS2Email is designed to be invoked by `cron`, and when executed it will
process a list of feed-URLs.  Each URL will be processed in turn, and the items
in the feed examined one by one:

* If the item in the feed has been seen before it will be ignored.
* If the item in the feed has not been seen:
  * An email will be sent.
  * The item will be recorded as having been notified already.

The the list of URLs to be fetched should be stored in the file:

* `~/.rss2email/feeds`
   * One URL per line.

Assuming you've added your entries you should then add the binary to
your `crontab`, via a line such as this:

     # Announce feed-changes to email
     */15 * * * * $HOME/go/bin/rss2email

The emails will be sent to your user when they appear, via the environmental
variable `USER`.


## Assumptions

Because this application is so minimal there are a number of assumptions
baked in:

* We assume that `/usr/sbin/sendmail` exists and will send email to the local user `steve` when invoked like this:
   * "`/usr/sbin/sendmail -f steve steve`"
* We assume that you'll invoke it via `cron`.
  * `$LOGIN` will be used to determine where the email is sent to.
* The sender of the email address will be `user@rss2email.invalid`.
  * This matches `r2e` meaning my existing mail filter(s) accept it and file appropriately.

Steve
--
