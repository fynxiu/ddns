.PONY: all script dep test lint

app = ddns
tag =

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

build-alpine:
	@echo "building $(app) for alpine"
	CGO_ENABLED=0 OS=alpine ARCH=amd64 go build -o script/alpine/$(app)

docker-alpine: override tag=`git describe --tags $(git rev-list --tags --max-count=1)`
docker-alpine: build-alpine
	@echo "building docker image for $(app) alpine version"
	docker build ./script/alpine -t ddns
	rm -f ./script/alpine/ddns
	docker tag ddns fynxiu/ddns:$(tag)-alpine
	docker push fynxiu/ddns:$(tag)-alpine
