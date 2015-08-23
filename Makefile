CWD:=$(shell pwd)
GOPATH:=$(CWD)/vendor
GO:=$(shell which go)
VERSION:=0.1
HARDWARE=$(shell uname -m)

build: vendor
	GOPATH=$(GOPATH) GOOS=linux $(GO) build -ldflags="-X main.version=$(VERSION)"
	docker build -t dev/dock:latest .

dev: vendor
	GOPATH=$(GOPATH) $(GO) build -ldflags="-X main.version=$(VERSION)"

release:

clean:
	rm -rf ./dock ./release ./vendor/pkg

vendor:
	mkdir -p ./vendor/src/github.com/robinmonjo
	ln -s $(CWD) ./vendor/src/github.com/robinmonjo/
	GOPATH=$(GOPATH) sh vendor.sh
