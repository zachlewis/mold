
NAME = mold

BRANCH = $(shell git rev-parse --abbrev-ref HEAD)
COMMIT = $(shell git rev-parse --short HEAD)
BUILDTIME = $(shell date +%Y-%m-%dT%T%z)

clean:
	rm -f ${NAME}
	rm -rf vendor/*

test:
	go test -v -cover ./...

deps:
	go get github.com/tools/godep
	go get golang.org/x/tools/cmd/cover
	godep restore -v

${NAME}:
	CGO_ENABLED=0 go build -a -tags netgo -installsuffix netgo -ldflags="-X main.branch=${BRANCH} -X main.commit=${COMMIT} -X main.buildtime=${BUILDTIME} -w" -o ${NAME} .

all: clean ${NAME}
