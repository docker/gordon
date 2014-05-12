## Pulls

Pulls is a small cli application name to help you manage pull requests for your repository.
It was created by Michael Crosby to improve the productivity of the [Docker](https://docker.io) maintainers.

Gordon assumes that the git `origin` is the upstream repository where the Issues and Pull Requests are managed.
This is _not_ the workflow described in the [GitHub fork a repository](https://help.github.com/articles/fork-a-repo)
documentation.

Quick installation instructions:

* Install Go 1.2+ from http://golang.org/
* Install with `go get -u github.com/dotcloud/gordon/{pulls,issues}`
* Make sure your `$PATH` includes *x*/bin where *x* is each directory in your `$GOPATH` environment variable.
* Call `pulls --help` and `issues --help`
* Add your github token with `pulls auth <UserName> --add <token>`

Dockerfile container build:

If you don't have Go set up and want to try out Gordon, you can use the Dockerfile to build it, and then
can either copy the 2 executables to your local Linux host:

- Build: `docker build -t gordon .`
- Copy: `docker run --name gore gordon true && docker cp gore:/go/bin/pulls . && docker cp gore:/go/bin/issues . && docker rm gore``

You could also run from inside the container:
- Setup an alias: `pulls() { docker run --rm -it -v $PWD:/src --workdir /src -e HOME=/src gordon pulls $@; }`
- Set the GitHub API token: `pulls auth SvenDowideit --add 1373a7583d30623abcb2b233fe45090fe2e4a3e1a2`
- List open PR's: `pulls`
