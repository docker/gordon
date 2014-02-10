## Pulls

Pulls is a small cli application name to help you manage pull requests for your repository.
It was created by Michael Crosby to improve the productivity of the [Docker](https://docker.io) maintainers.


Quick installation instructions:

* Install Go from http://golang.org/
* Install with `go get github.com/dotcloud/gordon/{pulls,issues}`
* Make sure your `$PATH` includes *x*/bin where *x* is each directory in your `$GOPATH` environment variable.
* Call `pulls --help` and `issues --help`
* Add your github token with `pulls auth --add <token>`
