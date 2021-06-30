VERSION := $(shell git describe --always --long --dirty)
all: install

fetch:
	go mod download

test: fetch
	@echo Running tests
	go test -v

coverage: fetch
	@echo Running Test with Coverage export
	go test -v -coverprofile=cover.out
	go test -json > report.json
	go tool cover -html=cover.out -o cover.html

coverall: coverage
	@echo Running Test with Coverall
	goveralls -coverprofile=cover.out -service="github" -repotoken=${COVERALLS_TOKEN}

install: fetch
	@echo Installing package
	go install