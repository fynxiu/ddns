.PONY: all build dep test lint

app = ddns

all: lint test build

dep:
	@go get -u golang.org/x/lint/golint
	@go mod download

lint:
	golint `go list ./... | grep -v /vendor/`

test:
	go test -v `go list ./... | grep -v /vendor/`

build:
	go build -o $(app)
