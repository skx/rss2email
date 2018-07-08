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

This project, being built in go, is self-contained and easy to deploy without the need for additional external libraries.


## Installation

Assuming you have a working golang installation you can install the binary
via:

     go get -u github.com/skx/rss2email
     go install github.com/skx/rss2email

If you prefer you can fetch a binary from [our release page](github.com/skx/rss2email/releases).  Currently there is only a binary for Linux (amd64) due to the use of `cgo` in our dependencies.


## Configuration

Once you have a binary you'll need to configure your list of feeds, you do
that by adding them one by one:

     $ rss2email add https://example.com/blog.rss
     $ rss2email add https://example.net/index.rss
     $ rss2email add https://example.com/foo.rss

You can see the configured feeds via:

     $ rss2email list

Or delete a feed by specifying it:

     $ rss2email delete https://example.com/foo.rss

> **NOTE**: Feeds are stored in `~/.rss2email/feeds`, you might prefer to edit that directly.  Just add one URI per line.


Once you've added your feeds you should then add the binary to your
`crontab`, to ensure it runs regularly, via a line such as this:

     # Announce feed-changes to email
     */15 * * * * $HOME/go/bin/rss2email cron

When the feeds are updated to include new-entries they will be sent to you
via email.  Emails will be multi-part, containing both `text/plain` and
`text/html` items for each feed update.


## Assumptions

Because this application is so minimal there are a number of assumptions baked in:

* We assume that `/usr/sbin/sendmail` exists and will send email to the local user `steve` when invoked like this:
   * "`/usr/sbin/sendmail -f steve steve`"
* We assume that you'll invoke it via `cron`.
  * `$LOGIN` will be used to determine where the email is sent to.
* The sender of the email address will be `user@rss2email.invalid`.
  * This matches `r2e` meaning my existing mail filter(s) accept it and file appropriately.

Steve
--
