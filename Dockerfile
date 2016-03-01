FROM gliderlabs/alpine:3.3

RUN apk add --no-cache bash make go git gcc musl-dev python

# create workspace
ENV GOPATH=/go
RUN mkdir -p /go/src/github.com/robinmonjo/dock
ADD . /go/src/github.com/robinmonjo/dock
WORKDIR /go/src/github.com/robinmonjo/dock

# build the app
RUN IN_CONTAINER=true make && mv dock /usr/local/bin