gist
====

This project is a mirror for hosting GitHub gist repositories. It's used as
the server for [http://gist.exposed][http://gist.exposed].


## Getting Started

To use gist, you'll need to first install [git2go][git2go]. Once that is
installed, simply `go get` the gist project:

```sh
$ go get github.com/benbjohnson/gist/...
```

You can run the server by using the `gistd` program and specifying a data
directory and GitHub API keys:

```sh
$ gistd -d ~/gist -key $GITHUB_KEY -secret $GITHUB_SECRET
```

[git2go]: https://github.com/libgit2/git2go

