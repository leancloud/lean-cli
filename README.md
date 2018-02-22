# lean(1)

[![Build Status](https://travis-ci.org/leancloud/lean-cli.svg?branch=master)](https://travis-ci.org/leancloud/lean-cli) [![GoDoc](https://godoc.org/github.com/leancloud/lean-cli?status.svg)](https://godoc.org/github.com/leancloud/lean-cli)

Command-line tool to develop and manage [LeanCloud](https://leancloud.cn) apps.

## Install

- via `homebrew`: `$ brew install lean-cli`
- via `https://releases.leanapp.cn/#/leancloud/lean-cli/releases`(In case of your connection with GitHub cracked)

## Develop

Install the toolchains:

- [go](https://golang.org)
- [dep](https://github.com/golang/dep)
- [msitools](https://wiki.gnome.org/msitools)
- [dpkg](https://wiki.debian.org/Teams/Dpkg)

> You can install them via homebrew

Clone this repo to your `${GOPATH}/src/github.com/leancloud/lean-cli`, then have a look at `Makefile`, Then run `make all` to build releases.

Ensure all codes is formated by [gofmt](https://golang.org/cmd/gofmt/). Commit message should write in [gitmoji](https://gitmoji.carloscuesta.me/).

## Release

Tag the current commit with version name, and create a [release](https://github.com/leancloud/lean-cli/releases) with this tag. run `$ make all` and attach the build result (under `./_build` folder) to the release.

The homebrew guys will update the home brew [formula](https://github.com/Homebrew/homebrew-core/blob/master/Formula/lean-cli.rb). If not, or we are in a hurry, just make a pull request to them.

[Releases](https://releases.leanapp.cn) will fetch from GitHub automatically. If not, or we are in a hurry, just execute cloud function `updateRepo` with argument `{"repo": "leancloud/lean-cli"}` to update.
