# lean(1)

[![Build Status](https://travis-ci.org/leancloud/lean-cli.svg?branch=master)](https://travis-ci.org/leancloud/lean-cli) [![GoDoc](https://godoc.org/github.com/leancloud/lean-cli?status.svg)](https://godoc.org/github.com/leancloud/lean-cli)

Command line tool to develop and manage [LeanCloud](https://leancloud.cn) apps.

## Install

- via `go get`: `$ go get github.com/leancloud/lean-cli/lean`
- via `homebrew`: `$ brew install lean-cli`

## Develop

Install the dependences first:

- [go](https://golang.org)
- [glide](https://glide.sh)

Clone this repo to your `${GOPATH}/src/github.com/leancloud/lean-cli`, then have a look at `Makefile`.

## Packaging

Install this dependences:

- [msitool](https://wiki.gnome.org/msitools)
- [dpkg](https://wiki.debian.org/Teams/Dpkg)

> You can install them via homebrew

and

```bash
$ make all
```
