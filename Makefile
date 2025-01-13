# Build variables
BINARY_NAME=base
VERSION=$(shell git describe --tags --always --dirty)
COMMIT_HASH=$(shell git rev-parse --short HEAD)
BUILD_DATE=$(shell date -u '+%Y-%m-%d %H:%M:%S')
GO_VERSION=$(shell go version | cut -d ' ' -f 3)

# LDFLAGS for version information
LDFLAGS=-ldflags "-X github.com/flakerimi/base/cmd/version.Version=${VERSION} \
                  -X github.com/flakerimi/base/cmd/version.CommitHash=${COMMIT_HASH} \
                  -X github.com/flakerimi/base/cmd/version.BuildDate=${BUILD_DATE} \
                  -X github.com/flakerimi/base/cmd/version.GoVersion=${GO_VERSION}"

.PHONY: all build clean install test

all: clean build

build:
	@echo "Building Base CLI..."
	go build ${LDFLAGS} -o ${BINARY_NAME}

clean:
	@echo "Cleaning..."
	rm -f ${BINARY_NAME}

install: build
	@echo "Installing..."
	mv ${BINARY_NAME} ${GOPATH}/bin/${BINARY_NAME}

test:
	@echo "Running tests..."
	go test -v ./...

# Release targets
.PHONY: release release-patch release-minor release-major

CURRENT_VERSION=$(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")

release-patch:
	@echo "Creating patch release..."
	$(eval NEW_VERSION=$(shell echo ${CURRENT_VERSION} | awk -F. '{$$NF = $$NF + 1;} 1' | sed 's/ /./g'))
	@echo "New version: ${NEW_VERSION}"
	git tag -a ${NEW_VERSION} -m "Release ${NEW_VERSION}"
	git push origin ${NEW_VERSION}

release-minor:
	@echo "Creating minor release..."
	$(eval NEW_VERSION=$(shell echo ${CURRENT_VERSION} | awk -F. '{$$(NF-1) = $$(NF-1) + 1; $$NF = 0;} 1' | sed 's/ /./g'))
	@echo "New version: ${NEW_VERSION}"
	git tag -a ${NEW_VERSION} -m "Release ${NEW_VERSION}"
	git push origin ${NEW_VERSION}

release-major:
	@echo "Creating major release..."
	$(eval NEW_VERSION=$(shell echo ${CURRENT_VERSION} | awk -F. '{$$1 = substr($$1,2) + 1; $$(NF-1) = 0; $$NF = 0;} 1' | sed 's/ /./g' | sed 's/^/v/'))
	@echo "New version: ${NEW_VERSION}"
	git tag -a ${NEW_VERSION} -m "Release ${NEW_VERSION}"
	git push origin ${NEW_VERSION}
