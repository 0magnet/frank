git_revision := $(shell git log -1 --pretty="%H")
git_branch := $(shell git log -1 --pretty="%d")
host_info := $(shell uname -s -r -p)
build_date := $(shell date)

ldflags := -X 'main.gitRevision=$(git_revision)' -X 'main.gitBranch=$(git_branch)' -X 'main.buildTime=$(build_date)' -X 'main.hostInfo=$(host_info)'

all: run clean_assets

run: build_assets
	@go run -ldflags "$(ldflags)" . --settings="./settings.json"

build: clean build_assets
	@echo "Building Frank - 🌭"
	@go build -o frank -ldflags "$(ldflags) -s -w" .
	@chmod 755 frank
	@mkdir bin; mv frank bin/

clean:
	@rm -rf bin

test:
	@go test ./...

build_assets:
	@go run -tags=bundleAssets assets/bundler.go

clean_assets:
	@rm assets/icons.go assets/pages.go