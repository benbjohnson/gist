gist
====

This project is a mirror for hosting GitHub gist repositories. It's used as
the server for [http://gist.exposed][http://gist.exposed].


## Getting Started

Simply `go get` the gist project:

```sh
$ go get github.com/benbjohnson/gist/...
```

You can run the server by using the `gistd` program and specifying a data
directory and GitHub API keys:

```sh
$ gistd -d ~/gist -key $GITHUB_API_TOKEN -secret $GITHUB_API_SECRET
```

[git2go]: https://github.com/libgit2/git2go

