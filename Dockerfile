FROM debian:wheezy
MAINTAINER Sven Dowideit <SvenDowideit@home.org.au> (@SvenDowideit)
#
# Run gordon in a container
#    `pulls() { docker run --rm -it -v $PWD:/src --workdir /src -e HOME=/src gordon pulls $@; }`
#

# Packaged dependencies
RUN apt-get update && apt-get install -yq --no-install-recommends build-essential ca-certificates curl git mercurial vim-tiny

# Install Go from binary release
RUN curl -s https://go.googlecode.com/files/go1.2.1.linux-amd64.tar.gz | tar -v -C /usr/local -xz
ENV PATH /go/bin:/usr/local/go/bin:$PATH
ENV GOPATH  /go

RUN go get -u github.com/dotcloud/gordon/pulls
RUN go get -u github.com/dotcloud/gordon/issues
