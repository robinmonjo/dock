CWD:=$(shell pwd)
GOPATH:=$(CWD)/vendor
GO:=$(shell which go)
VERSION:=0.1
HARDWARE=$(shell uname -m)
IMAGE_NAME=robinmonjo/ubuntu-dock:dev

build: vendor
	GOPATH=$(GOPATH) GOOS=linux $(GO) build -ldflags="-X main.version=$(VERSION)"
	docker build -t $(IMAGE_NAME) .

#build on mac os
mac: vendor
	GOPATH=$(GOPATH) $(GO) build -ldflags="-X main.version=$(VERSION)"

test:
	GOPATH=$(GOPATH) TEST_IMAGE=$(IMAGE_NAME) bash -c 'cd integration && $(GO) test'

release:

clean:
	rm -rf ./dock ./release ./vendor/pkg

vendor:
	mkdir -p ./vendor/src/github.com/robinmonjo
	ln -s $(CWD) ./vendor/src/github.com/robinmonjo/
	GOPATH=$(GOPATH) sh vendor.sh
