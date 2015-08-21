GOPATH:=`pwd`/vendor:$(GOPATH)
GO:=$(shell which go)
VERSION:=0.1
HARDWARE=$(shell uname -m)

build: vendor
	GOPATH=$(GOPATH) go build -ldflags="-X main.version $(VERSION)"

release:


clean:
	rm -rf ./dock ./release ./vendor/pkg

vendor:
	sh vendor.sh
