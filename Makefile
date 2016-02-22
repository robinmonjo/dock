CWD:=$(shell pwd)
GO:=GO15VENDOREXPERIMENT=1 go
VERSION:=0.1
HARDWARE=$(shell uname -m)
IMAGE_NAME=robinmonjo/alpine-dock:dev

build:
ifeq ($(IN_CONTAINER), true)
	$(GO) build -ldflags="-X main.version=$(VERSION)"
else
	docker build -t $(IMAGE_NAME) .
endif
	
test: build
ifeq ($(IN_CONTAINER), true)
	bash -c 'cd port && $(GO) test'
	bash -c 'cd logrotate && $(GO) test'
	bash -c 'cd iowire && $(GO) test'
else
	docker run -it -w "/go/src/github.com/robinmonjo/dock" -e IN_CONTAINER=true $(IMAGE_NAME) bash -c 'make test'
endif

integration: build
	TEST_IMAGE=$(IMAGE_NAME) bash -c 'cd integration && $(GO) test'