NO_COLOR=\033[0m
OK_COLOR=\033[32;01m
ERROR_COLOR=\033[31;01m
WARN_COLOR=\033[33;01m

# The import path is the unique absolute name of your repository.
# All subpackages should always be imported as relative to it.
# If you change this, run `make clean`.
IMPORT_PATH := github.com/hellofresh/github-cli
PKG_SRC := $(IMPORT_PATH)

GO_PROJECT_PACKAGES=`go list ./... | grep -v /vendor/`

.PHONY: all clean deps build

all: clean deps build

deps:
	@echo "$(OK_COLOR)==> Installing dependencies$(NO_COLOR)"
	@go get -u github.com/golang/dep/cmd/dep
	@dep ensure

# Builds the project
build: install

# Installs our project: copies binaries
install:
	@echo "$(OK_COLOR)==> Building... $(NO_COLOR)"
	/bin/sh -c "PKG_SRC=$(PKG_SRC) VERSION=${VERSION} ./build/build.sh"

test:
	go test ${GO_PROJECT_PACKAGES} -v
	
# Cleans our project: deletes binaries
clean:
	@echo "$(OK_COLOR)==> Cleaning project$(NO_COLOR)"
	go clean
