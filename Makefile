
NAME = mold
VERSION = $(shell grep 'const VERSION' version.go | cut -d'"' -f 2)

BRANCH = $(shell git rev-parse --abbrev-ref HEAD)
COMMIT = $(shell git rev-parse --short HEAD)
BUILDTIME = $(shell date +%Y-%m-%dT%T%z)

BUILD_CMD = CGO_ENABLED=0 go build -a -tags netgo -installsuffix netgo \
	-ldflags="-X main.branch=${BRANCH} -X main.commit=${COMMIT} -X main.buildtime=${BUILDTIME} -w"
GOOS = $(shell go env GOOS)

clean:
	rm -rf $(NAME) vendor/* build coverage.out
	go clean -i ./...

test:
	rm -f coverage.out
	go test -v -coverprofile=coverage.out ./...

deps:
	go get -d .

${NAME}:
	$(BUILD_CMD) -o $(NAME) .

dist: clean
	mkdir ./build

	$(BUILD_CMD) -o ./build/$(NAME) .
	cd ./build && tar -czf $(NAME)-$(GOOS)-$(VERSION).tgz $(NAME); rm -f $(NAME)

	GOOS=linux $(BUILD_CMD) -o ./build/$(NAME) .
	cd ./build && tar -czf $(NAME)-linux-$(VERSION).tgz $(NAME); rm -f $(NAME)

	GOOS=windows $(BUILD_CMD) -o ./build/$(NAME).exe .
	cd ./build && zip $(NAME)-windows-$(VERSION).zip $(NAME).exe; rm -f $(NAME).exe

all: clean ${NAME}
