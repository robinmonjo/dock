FROM gliderlabs/alpine:3.3

RUN apk add --no-cache bash make go git gcc musl-dev

# add assets for integration testing
ADD integration/assets /assets

RUN mkdir /dock
ADD . /dock/
WORKDIR /dock

# recreate the symlink + launch tests
RUN rm -rf vendor/src/github.com/robinmonjo/dock && ln -s /dock vendor/src/github.com/robinmonjo && IN_CONTAINER=true make && mv dock /usr/local/bin
