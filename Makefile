CWD:=$(shell pwd)
GOPATH:=$(CWD)/vendor
GO:=$(shell which go)
VERSION:=0.1
HARDWARE=$(shell uname -m)
IMAGE_NAME=robinmonjo/alpine-dock:dev

build: vendor
ifeq ($(IN_CONTAINER), true)
	GOPATH=$(GOPATH) $(GO) build -ldflags="-X main.version=$(VERSION)"
else
	docker build -t $(IMAGE_NAME) .
endif
	
test: build
ifeq ($(IN_CONTAINER), true)
	GOPATH=$(GOPATH) bash -c 'cd port && $(GO) test'
	#GOPATH=$(GOPATH) bash -c 'cd logrotate && $(GO) test'
	GOPATH=$(GOPATH) bash -c 'cd iowire && $(GO) test'
else
	docker run -it -w "/dock" -e IN_CONTAINER=true $(IMAGE_NAME) bash -c 'make test'
endif

integration: build
	GOPATH=$(GOPATH) TEST_IMAGE=$(IMAGE_NAME) bash -c 'cd integration && $(GO) test'

clean:
	rm -rf ./dock ./release ./vendor/pkg

vendor:
	mkdir -p ./vendor/src/github.com/robinmonjo
	ln -s $(CWD) ./vendor/src/github.com/robinmonjo/
	GOPATH=$(GOPATH) sh vendor.sh
