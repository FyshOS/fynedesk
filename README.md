<p align="center">
  <a href="https://godoc.org/github.com/fyne-io/desktop" title="GoDoc Reference" rel="nofollow"><img src="https://img.shields.io/badge/go-documentation-blue.svg?style=flat" alt="GoDoc Reference"></a>
  <a href="https://goreportcard.com/report/github.com/fyne-io/desktop"><img src="https://goreportcard.com/badge/github.com/fyne-io/desktop" alt="Code Status" /></a>
  <a href="https://travis-ci.org/fyne-io/desktop"><img src="https://travis-ci.org/fyne-io/desktop.svg" alt="Build Status" /></a>
  <a href='https://coveralls.io/github/fyne-io/desktop?branch=develop'><img src='https://coveralls.io/repos/github/fyne-io/desktop/badge.svg?branch=develop' alt='Coverage Status' /></a>
</p>

# About

This project is very early stages and is not yet usable as a desktop environment.
However it is available for anyone to test, play with or contribute to.

# Prerequisites

Before you can use the Fyne tools you need to have a stable copy of EFL installed. This is being automated by our
[bootstrap](https://github.com/fyne-io/bootstrap/) scripts, but for now you can follow the
[setup instructions](https://github.com/fyne-io/bootstrap/blob/master/README.md).

# Getting Started

Using standard go tools you can install Fyne's desktop using:

    go get github.com/fyne-io/desktop

And you can run that simply as:

    cd $GOPATH/src/github.com/fyne-io/desktop
    go run cmd/fynedesk/main.go

It should look like this:

<p align="center" markdown="1">
  <img src="desktop-dark-current.png" alt="Fyne Desktop - Dark" />
</p>
