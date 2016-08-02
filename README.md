# lean

Command line tool to develop and manage [LeanCloud](https://leancloud.cn) apps.

## Install

- via `go get`: `$ go get github.com/leancloud/lean-cli/lean`
- via `homebrew`: `$ brew install leancloud/leancloud/lean-cli`

## Develop

Instatall the dependences first:

- [go](https://golang.org)
- [glide](https://glide.sh)


Clone this repo to your `${GOPATH}/src/github.com/leancloud/lean-cli`

Goto `lean` directory, and run `glide install` to get the third party dependences, then run `go build` to get the `lean` binary.
