gist [![Build Status](https://drone.io/github.com/benbjohnson/gist/status.png)](https://drone.io/github.com/benbjohnson/gist/latest) [![Coverage Status](https://coveralls.io/repos/benbjohnson/gist/badge.png?branch=master)](https://coveralls.io/r/benbjohnson/gist?branch=master) [![GoDoc](https://godoc.org/github.com/benbjohnson/gist?status.png)](https://godoc.org/github.com/benbjohnson/gist) ![Project status](http://img.shields.io/status/experimental.png?color=red)
====

This project is a mirror for hosting GitHub gist repositories.


## Getting Started

### Prerequisites

Before you start using the Gist application you'll need to create a new GitHub
application. You can do this by going to your GitHub settings page, clicking
on "Applications", and then creating a new application. You'll need to copy the
token and secret that are generated and pass those into your application later.


### Installing and Running

Simply `go get` the gist project:

```sh
$ go get github.com/benbjohnson/gist/...
```

Then run the `gistd` binary:

```sh
$ gistd -d ~/gist -key $GITHUB_API_TOKEN -secret $GITHUB_API_SECRET
Listening on http://localhost:40000
```

You can now visit [http://localhost:40000](http://localhost:40000) to view
the application.

