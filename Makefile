GIT_HASH=$(shell git rev-parse HEAD)
DATE_TAG=$(shell date +%m%d%y-%H:%M:%S)
VERSION=$(shell git tag)

build:
	go build -ldflags="-X main.GitHash $(GIT_HASH) -X main.BuildDate $(DATE_TAG) -X main.Version $(VERSION)"
	mv nitro-tftp test/

all:
	gox -ldflags="-X main.GitHash $(GIT_HASH) -X main.BuildDate $(DATE_TAG) -X main.Version $(VERSION)"
