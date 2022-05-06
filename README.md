[![Go Report Card](https://goreportcard.com/badge/github.com/skx/rss2email)](https://goreportcard.com/report/github.com/skx/rss2email)
[![license](https://img.shields.io/github/license/skx/rss2email.svg)](https://github.com/skx/rss2email/blob/master/LICENSE)
[![Release](https://img.shields.io/github/release/skx/rss2email.svg)](https://github.com/skx/rss2email/releases/latest)

Table of Contents
=================

* [RSS2Email](#rss2email)
* [Installation](#installation)
  * [Build with Go Modules](#build-with-go-modules)
  * [bash completion](#bash-completion)
* [Feed Configuration](#feed-configuration)
* [Usage](#usage)
* [Daemon Mode](#daemon-mode)
* [Initial Run](#initial-run)
* [Assumptions](#assumptions)
* [Email Customization](#email-customization)
  * [Changing default From address](#changing-default-from-address)
* [Implementation Overview](#implementation-overview)
* [Github Setup](#github-setup)



# RSS2Email

This project began life as a naive port of the python-based [r2e](https://github.com/wking/rss2email) utility to golang.

Over time we've now gained a few more features:

* The ability to customize the email-templates which are generated and sent.
  * See [email customization](#email-customization) for details.
* The ability to send email via STMP, or via `/usr/sbin/sendmail`.
  * See [SMTP-setup](#smtp-setup) for details.
* The ability to include/exclude feed items from the emails.
  * For example receive emails only of feed items that contain the pattern "playstation".



# Installation

If you have golang installed you can fetch, build, and install the latest binary by running:

```sh
go install github.com/skx/rss2email@latest
```

If you prefer you can also fetch our latest binary release from [our release page](https://github.com/skx/rss2email/releases).

To install from source simply clone the repository and build in the usual manner:

```sh
git clone https://github.com/skx/rss2email
cd rss2email
go build .
go install .
```

**Version NOTES**:

* You'll need go version **1.16** or higher to build.
  * Because we use `go embed` to embed our (default) email-template within the binary.
* If you wish to run the included fuzz-tests against our configuration file parser you'll need at least version **1.18**.
  * See [configfile/FUZZING.md](configfile/FUZZING.md) for details.



## bash completion

The binary has integrated support for TAB-completion, for bash.  To enable this update your [dotfiles](https://github.com/skx/dotfiles/) to include the following:

```
source <(rss2email bash-completion)
```


# Feed Configuration

Once you have installed the application you'll need to configure the feeds to monitor, this could be done by editing the configuration file:

* `~/.rss2email/feeds.txt`

There are several built-in sub-commands for manipulating the feed-list, for example you can add a new feed to monitor via the `add` sub-command:

     $ rss2email add https://example.com/blog.rss

OPML files can be imported via the `import` sub-command:

     $ rss2email import feeds.opml

The list of feeds can be displayed via the `list` subcommand (note that adding the `-verbose` flag will fetch each of the feeds and that will be slow):

     $ rss2email list [-verbose]

Finally you can remove an entry from the feed-list via the `delete` sub-command:

     $ rss2email delete https://example.com/foo.rss

The configuration file in its simplest form is nothing more than a list of URLs, one per line.  However there is also support for adding per-feed options:

       https://foo.example.com/
        - key:value
       https://foo.example.com/
        - key2:value2

This is documented and explained in the integrated help:

    $ rss2email help config

Adding per-feed items allows excluding feed-entries by regular expression, for example this does what you'd expect:

       https://www.filfre.net/feed/rss/
        - exclude-title: The Analog Antiquarian



# Usage

Once you've populated your feed list, via a series of `rss2email add ..` commands, or by editing the configuration file directly, you are now ready to actually launch the application.

To run the application, announcing all new feed-items by email to `user@host.com` you'd run this:

    $ rss2email cron user@host.com

Once the feed-list has been fetched, and items processed, the application will terminate.  It is expected that you'll add an entry to your `crontab` file to ensure this runs regularly.  For example you might wish to run the check & email process once every 15 minutes, so you could add this:

     # Announce feed-changes via email four times an hour
     */15 * * * * $HOME/go/bin/rss2email cron recipient@example.com

When new items appear in the feeds they will then be sent to you via email.
Each email will be multi-part, containing both `text/plain` and `text/html`
versions of the new post(s).  There is a default template which should contain
the things you care about:

* A link to the item posted.
* The subject/title of the new feed item.
* The HTML and Text content of the new feed item.

If you wish you may customize the template which is used to generate the notification email, see [email-customization](#email-customization) for details.  It is also possible to run in a [daemon mode](#daemon-mode) which will leave the process running forever, rather than terminating after walking the feeds once.

The state of feed-entries is recorded beneath `~/.rss2email/state.db`, which is a [boltdb database](https://pkg.go.dev/go.etcd.io/bbolt).



# Daemon Mode

Typically you'd invoke `rss2email` with the `cron` sub-command as we documented above.  This works in the naive way you'd expect:

* Read the contents of each URL in the feed-list.
* For each feed-item which is new generate and send an email.
* Terminate.

The `daemon` process does a similar thing, however it does __not__ terminate.  Instead the process becomes:

* Read the contents of each URL in the feed-list.
* For each feed-item which is new generate and send an email.
* Sleep for 15 minutes by default.
  * Set the `SLEEP` environmental variable if you wish to change this.
  * e.g. "`export SLEEP=5`" will cause a five minute delay between restarts.
* Begin the process once more.

In short the process runs forever, in the foreground.  This is expected to be driven by `docker` or a systemd-service.  Creating the appropriate configuration is left as an exercise, but you might examine the following two files for inspiration:

* [Dockerfile](Dockerfile)
* [docker-compose.yml](docker-compose.yml)



# Initial Run

When you add a new feed all the items contained within that feed will initially be unseen/new, and this means you'll receive a flood of emails if you were to run:

     $ rss2email add https://blog.steve.fi/index.rss
     $ rss2email cron user@domain.com

To avoid this you can use the `-send=false` flag, which will merely
record each item as having been seen, rather than sending you emails:

     $ rss2email add https://blog.steve.fi/index.rss
     $ rss2email cron -send=false user@domain.com


# Assumptions

Because this application is so minimal there are a number of assumptions baked in:

* We assume that `/usr/sbin/sendmail` exists and will send email successfully.
  * You can cause emails to be sent via SMTP, see [SMTP-setup](#smtp-setup) for details.
* We assume the recipient and sender email addresses can be the same.
  * i.e. If you mail output to `bob@example.com` that will be used as the sender address.
  * You can change the default sender via the [email-customization](#email-customization) process described next if you prefer though.



# SMTP Setup

By default the outgoing emails we generate are piped to `/usr/sbin/sendmail` to be delivered.  If that is unavailable, or unsuitable, you can instead configure things such that SMTP is used directly.

To configure SMTP you need to setup the following environmental-variables (environmental variables were selected as they're natural to use within Docker and systemd-service files).


| Name              | Example Value     |
|-------------------|-------------------|
| **SMTP_HOST**     | `smtp.gmail.com`  |
| **SMTP_PORT**     | `587`             |
| **SMTP_USERNAME** | `bob@example.com` |
| **SMTP_PASSWORD** | `secret!value`    |

If those values are present then SMTP will be used, otherwise the email will be sent via the local MTA.



# Email Customization

By default the emails are sent using a template file which is embedded in the application.  You can override the template by creating the file `~/.rss2email/email.tmpl`, if that is present then it will be used instead of the default.

You can view the default template via the following command:

    $ rss2email list-default-template

You can copy the default-template to the right location by running the following, before proceeding to edit it as you wish:

    $ rss2email list-default-template > ~/.rss2email/email.tmpl

The default template contains a brief header documenting the available fields, and functions, which you can use.  As the template uses the standard Golang [text/template](https://golang.org/pkg/text/template/) facilities you can be pretty creative with it!

If you're a developer who wishes to submit changes to the embedded version you should carry out the following two-step process to make your change.

* Edit `template/template.txt`, which is the source of the template.
* Rebuild the application to update the embedded copy.

**NOTE**: If you read the earlier section on configuration you'll see that it is possible to add per-feed configuration values to the config file.  One of the supported options is to setup a feed-specific template-file.


## Changing default From address

As noted earlier when sending the notification emails the recipient address is used as the sender-address too.   There are no flags for changing the From: address used to send the emails, however using the section above you can [use a customized email-template](#email-customization), and simply update the template to read something like this:

```
From: my.sender@example.com
To: {{.To}}
Subject: [rss2email] {{.Subject}}
X-RSS-Link: {{.Link}}
X-RSS-Feed: {{.Feed}}
```

* i.e. Change the `{{.From}}` to your preferred sender-address.



# Implementation Overview

The two main commands are `cron` and `daemon` and they work in roughly the same way:

* They instantiate [processor/processor.go](processor/processor.go) to run the logic
  * That walks over the list of feeds from [configfile/configfile.go](configfile/configfile.go).
  * For each feed [httpfetch/httpfetch.go](httpfetch/httpfetch.go) is used to fetch the contents.
  * The result is a collection of `*gofeed.Feed` items, one for each entry in the remote feed.
    * These are wrapped via [withstate/feeditem.go](withstate/feeditem.go) so we can test if they're new.
    * [processor/emailer/emailer.go](processor/emailer/emailer.go) is used to send the email if necessary.
    * Either by SMTP or by executing `/usr/sbin/sendmail`

The other subcommands mostly just interact with the feed-list, via the use of [configfile/configfile.go](configfile/configfile.go) to add/delete/list the contents of the feed-list.


# Github Setup

This repository is configured to run tests upon every commit, and when
pull-requests are created/updated.  The testing is carried out via
[.github/run-tests.sh](.github/run-tests.sh) which is used by the
[github-action-tester](https://github.com/skx/github-action-tester) action.

Releases are automated in a similar fashion via [.github/build](.github/build),
and the [github-action-publish-binaries](https://github.com/skx/github-action-publish-binaries) action.

Steve
--
