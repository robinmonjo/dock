CWD:=$(shell pwd)
GO:=GO15VENDOREXPERIMENT=1 go
VERSION:=0.4
IMAGE_NAME=robinmonjo/alpine-dock:dev

build:
ifeq ($(IN_CONTAINER), true)
	$(GO) build -ldflags="-X main.version=$(VERSION)"
else
	docker build -t $(IMAGE_NAME) .
endif

binary:
	$(GO) build -ldflags="-X main.version=$(VERSION)"
	
test: build
ifeq ($(IN_CONTAINER), true)
	bash -c 'cd port && $(GO) test'
	bash -c 'cd logrotate && $(GO) test'
	bash -c 'cd iowire && $(GO) test'
	bash -c 'cd procfs && $(GO) test'
else
	docker run -it -w "/go/src/github.com/robinmonjo/dock" -e IN_CONTAINER=true $(IMAGE_NAME) bash -c 'make test'
endif

integration: build
	TEST_IMAGE=$(IMAGE_NAME) bash -c 'cd integration && $(GO) test'
	
release:
	mkdir -p release
	GOOS=linux $(GO) build -ldflags="-X main.version=$(VERSION)" -o release/dock
	cd release && tar -zcf dock-v$(VERSION).tgz dock
	rm release/dock
	
vendor:
	bash vendor.sh