
NAME = mold
VERSION = $(shell grep 'const VERSION' version.go | cut -d'"' -f 2)

BRANCH = $(shell git rev-parse --abbrev-ref HEAD)
COMMIT = $(shell git rev-parse --short HEAD)
BUILDTIME = $(shell date +%Y-%m-%dT%T%z)

LD_OPTS = -ldflags="-X main.branch=${BRANCH} -X main.commit=${COMMIT} -X main.buildtime=${BUILDTIME} -w"
BUILD_CMD = CGO_ENABLED=0 go build -a -tags netgo -installsuffix netgo $(LD_OPTS)
GOOS = $(shell go env GOOS)

clean:
	rm -rf $(NAME) vendor/* dist coverage.out
	go clean -i ./...

test:
	go test -cover $(shell go list ./... | grep -v /vendor/)

deps:
	go get -u github.com/golang/dep/cmd/dep
	dep ensure

${NAME}:
	go build $(LD_OPTS) -o $(NAME) .

dist:
	rm -rf ./dist
	mkdir ./dist

	GOOS=darwin $(BUILD_CMD) -o ./dist/$(NAME) .
	cd ./dist && tar -czf $(NAME)-darwin-$(VERSION).tgz $(NAME); rm -f $(NAME)

	GOOS=linux $(BUILD_CMD) -o ./dist/$(NAME) .
	cd ./dist && tar -czf $(NAME)-linux-$(VERSION).tgz $(NAME); rm -f $(NAME)

	GOOS=windows $(BUILD_CMD) -o ./dist/$(NAME).exe .
	cd ./dist && zip $(NAME)-windows-$(VERSION).zip $(NAME).exe; rm -f $(NAME).exe

all: clean ${NAME}
