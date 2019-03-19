[![Go Report Card](https://goreportcard.com/badge/github.com/skx/rss2email)](https://goreportcard.com/report/github.com/skx/rss2email)
[![license](https://img.shields.io/github/license/skx/rss2email.svg)](https://github.com/skx/rss2email/blob/master/LICENSE)
[![Release](https://img.shields.io/github/release/skx/rss2email.svg)](https://github.com/skx/rss2email/releases/latest)

Table of Contents
=================

* [RSS2Email](#rss2email)
  * [Rationale](#rationale)
* [Installation](#installation)
  * [Build without Go Modules (Go before 1.11)](#build-without-go-modules-go-before-111)
  * [Build with Go Modules (Go 1.11 or higher)](#build-with-go-modules-go-111-or-higher)
* [Configuration](#configuration)
  * [Initial Run](#initial-run)
* [Assumptions](#assumptions)
* [Github Setup](#github-setup)


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
      0 upgraded, 0 newly installed, 9 to remove and 0 not upgraded.

This project, being built in go, is self-contained and easy to deploy without the need for additional external libraries.




# Installation

There are two ways to install this project from source, which depend on the version of the [go](https://golang.org/) version you're using.

If you prefer you can fetch a binary from [our release page](https://github.com/skx/rss2email/releases).  Currently there is only a binary for Linux (amd64) due to the use of `cgo` in our dependencies.

## Build without Go Modules (Go before 1.11)

    go get -u github.com/skx/rss2email

## Build with Go Modules (Go 1.11 or higher)

    git clone https://github.com/skx/rss2email ;# make sure to clone outside of GOPATH
    cd rss2email
    go install


# Configuration

Once you have a binary you'll need to configure the feeds to monitor. To
add a new feed use the `add` sub-command:

     $ rss2email add https://example.com/blog.rss
     $ rss2email add https://example.net/index.rss
     $ rss2email add https://example.com/foo.rss

You can view the currently configured feeds via the `list` subcommand:

     $ rss2email list

Or delete a feed by specifying the item to remove:

     $ rss2email delete https://example.com/foo.rss

> **NOTE**: Feeds are stored in `~/.rss2email/feeds`, you might prefer to edit that directly.  Just add one URI per line.

Once you've added your feeds you should then add the binary to your
`crontab`, to ensure it runs regularly to actually send you the emails.
You should add something similar to this to your `crontab`:

     # Announce feed-changes via email
     */15 * * * * $HOME/go/bin/rss2email cron

When new items appear in the feeds they will then be sent to you via email.
Each email will be multi-part, containing both `text/plain` and `text/html`
versions of the new post(s).


## Initial Run

When you add a new feed all the items will initially be unseen/new, and
this means you'll receive a flood of emails if you were to run:

     $ rss2email add https://blog.steve.fi/index.rss
     $ rss2email cron

To avoid this you can use the `-send=false` flag, which will merely
record each item as having been seen, rather than sending you emails:

     $ rss2email add https://blog.steve.fi/index.rss
     $ rss2email cron -send=false


# Assumptions

Because this application is so minimal there are a number of assumptions baked in:

* We assume that `/usr/sbin/sendmail` exists and will send email to the local user `steve` when invoked like this:
   * "`/usr/sbin/sendmail -f steve steve`"
* We assume that you'll invoke it via `cron`.
  * `$LOGNAME` will be used to determine where the email is sent to.
* The sender of the email address will be `user@rss2email.invalid`.
  * This matches `r2e` meaning my existing mail filter(s) accept it and file appropriately.


# Github Setup

This repository is configured to run tests upon every commit, and when
pull-requests are created/updated.  The testing is carried out via
[.github/run-tests.sh](.github/run-tests.sh) which is used by the
[github-action-tester](https://github.com/skx/github-action-tester) action.

Releases are automated in a similar fashion via [.github/build](.github/build),
and the [github-action-publish-binaries](https://github.com/skx/github-action-publish-binaries) action.

Steve
--
