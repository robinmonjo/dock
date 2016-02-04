FROM gliderlabs/alpine:3.3

RUN apk add --no-cache bash

ADD dock /usr/local/bin/dock
ADD integration/assets /assets
