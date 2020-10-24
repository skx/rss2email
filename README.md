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
  * [bash completion](#bash-completion)
* [Configuration & Usage](#configuration--usage)
  * [Initial Run](#initial-run)
* [Assumptions](#assumptions)
* [Customization](#customization)
* [Github Setup](#github-setup)


# RSS2Email

This project is a naive port of the python-based [r2e](https://github.com/wking/rss2email) I used for many years to golang.


## Rationale

I prefer to keep my server(s) pretty minimal, and replacing `r2e` allowed
me to remove a bunch of Python packages I otherwise had no need for:

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


## bash completion

The binary has integrated support for TAB-completion, for bash.  To enable this update your [dotfiles](https://github.com/skx/dotfiles/) to include the following:

```
source <(./rss2email bash-completion)
```


# Configuration & Usage

Once you have installed the application you'll need to configure the feeds to monitor. To add a new feed use the `add` sub-command:

     $ rss2email add https://example.com/blog.rss
     $ rss2email add https://example.net/index.rss
     $ rss2email add https://example.com/foo.rss

If you have a list of feeds stored in an OPML file you can import them in a single-step via the `import` sub-command:

     $ rss2email import feeds.opml

The list of feeds can be displayed via the `list` subcommand:

     $ rss2email list

> **NOTE**: You can add `-verbose` to list the number of entries present in each feed, and get an idea of the age of entries.

Removing a feed from the list is done by specifying the item to remove:

     $ rss2email delete https://example.com/foo.rss

> **NOTE**: Feeds are stored in `~/.rss2email/feeds`, you might prefer to edit that directly.  Just add one URI per line.

Once you've added your feeds you should then add the binary to your
`crontab`, to ensure it runs regularly to actually send you the emails.
You should add something similar to this to your `crontab`:

     # Announce feed-changes via email four times an hour
     */15 * * * * $HOME/go/bin/rss2email cron recipient@example.com

When new items appear in the feeds they will then be sent to you via email.
Each email will be multi-part, containing both `text/plain` and `text/html`
versions of the new post(s).  There is a default template which should contain
the things you care about:

* A link to the item posted.
* The subject/title of the new feed item.
* The HTML and Text content of the new feed item.

If you wish you may customize the template which is used to generate the notification email, see [customization](#customization) for details.

We record the state of feed-entries beneath `~/.rss2email/seen`, and these entries are automatically pruned over time.


## Initial Run

When you add a new feed all the items will initially be unseen/new, and
this means you'll receive a flood of emails if you were to run:

     $ rss2email add https://blog.steve.fi/index.rss
     $ rss2email cron user@domain.com

To avoid this you can use the `-send=false` flag, which will merely
record each item as having been seen, rather than sending you emails:

     $ rss2email add https://blog.steve.fi/index.rss
     $ rss2email cron -send=false user@domain.com


# Assumptions

Because this application is so minimal there are a number of assumptions baked in:

* We assume that `/usr/sbin/sendmail` exists and will send email successfully.
* We assume the recipient and sender email addresses can be the same.
  * i.e. If you mail output to `bob@example.com` that will be used as the sender address.
  * You can change the default sender via the [customization](#customization) process described next if you prefer though.



# Customization

By default the emails are sent using a template file which is embedded in the application.  You can override the template by creating the file `~/.rss2email/email.tmpl`, if that is present then it will be used instead of the default.

You can copy the default-template to the right location by running the following, before proceeding to edit it as you wish:

    $ rss2email list-default-template > ~/.rss2email/email.tmpl

You can view the default template via the following command:

    $ rss2email list-default-template

The default template contains a brief header documenting the available fields, and functions, which you can use.  As the template uses the standard Golang [text/template](https://golang.org/pkg/text/template/) facilities you can be pretty creative with it!

If you're a developer who wishes to submit changes to the embedded version you should carry out the three-step process to make your change.

* First of all edit `data/email.tmpl`, this is the source of the template.
* Next run the [implant](https://github.com/skx/implant) tool.
  * This is responsible for reading the file, and embedding it into the file `static.go`, which is then included in the binary when the project is compiled.
  * You'll run `implant -package template -output ./template/static.go`
* Rebuild the application to update the embedded copy `go build .`
  * This will ensure that the changes you made to `data/email.tmpl` are actually contained within your binary, and will be used the next time you launch it.



# Github Setup

This repository is configured to run tests upon every commit, and when
pull-requests are created/updated.  The testing is carried out via
[.github/run-tests.sh](.github/run-tests.sh) which is used by the
[github-action-tester](https://github.com/skx/github-action-tester) action.

Releases are automated in a similar fashion via [.github/build](.github/build),
and the [github-action-publish-binaries](https://github.com/skx/github-action-publish-binaries) action.

Steve
--
