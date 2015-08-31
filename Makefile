GIT_HASH=$(shell git rev-parse HEAD)
DATE_TAG=$(shell date +%m%d%y-%H:%M:%S)
TAG_CMD := git log -n1 --pretty="format:%d" | sed "s/, /\n/g" | grep tag: | sed "s/tag: \|)//g"
VERSION=$(shell $(TAG_CMD))

.PHONY: test clean

build:
	go build -ldflags="-X=main.GitHash=$(GIT_HASH) -X=main.BuildDate=$(DATE_TAG) -X=main.Version=$(VERSION)"
	mv nitro-tftp test/

all:
	gox -ldflags="-X=main.GitHash=$(GIT_HASH) -X=main.BuildDate=$(DATE_TAG) -X=main.Version=$(VERSION)"

test:
	go test -v .

clean:
	rm nitro-tftp_*

release: all
	@echo "Release"
