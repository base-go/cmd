# Build variables
BINARY_NAME=base
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null | sed 's/^v//' || echo "dev")
COMMIT_HASH=$(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
GO_VERSION=$(shell go version | cut -d ' ' -f 3)

# LDFLAGS for version information
LDFLAGS=-ldflags "-X main.Version=${VERSION} \
                  -X main.CommitHash=${COMMIT_HASH} \
                  -X main.BuildDate=${BUILD_DATE} \
                  -X main.GoVersion=${GO_VERSION}"

.PHONY: all build clean install test dev

all: clean build

build:
	@echo "Building Base CLI..."
	@echo "Version: ${VERSION}"
	@echo "Commit: ${COMMIT_HASH}"
	@echo "Build Date: ${BUILD_DATE}"
	@echo "Go Version: ${GO_VERSION}"
	@echo "Build command: go build ${LDFLAGS} -o ${BINARY_NAME}"
	go build -v ${LDFLAGS} -o ${BINARY_NAME}

clean:
	@echo "Cleaning..."
	rm -f ${BINARY_NAME}

install: build
	@echo "Installing..."
	mkdir -p ${HOME}/.base
	mv ${BINARY_NAME} ${HOME}/.base/${BINARY_NAME}
	echo "darude" | sudo -S ln -sf ${HOME}/.base/${BINARY_NAME} /usr/local/bin/${BINARY_NAME}

dev: clean
	@echo "Building for development..."
	go build -o ${BINARY_NAME}
	./${BINARY_NAME} version

test:
	@echo "Running tests..."
	go test -v ./...

# Release targets
.PHONY: release release-patch release-minor release-major

CURRENT_VERSION=$(shell git describe --tags --abbrev=0 2>/dev/null | sed 's/^v//' || echo "0.0.0")

release-patch:
	@echo "Creating patch release..."
	$(eval NEW_VERSION=$(shell echo ${CURRENT_VERSION} | awk -F. '{$$NF = $$NF + 1;} 1' | sed 's/ /./g'))
	@git tag -a v${NEW_VERSION} -m "Release v${NEW_VERSION}"
	@git push origin v${NEW_VERSION}
	@echo "Released v${NEW_VERSION}"

release-minor:
	@echo "Creating minor release..."
	$(eval NEW_VERSION=$(shell echo ${CURRENT_VERSION} | awk -F. '{$$(NF-1) = $$(NF-1) + 1; $$NF = 0;} 1' | sed 's/ /./g'))
	@git tag -a v${NEW_VERSION} -m "Release v${NEW_VERSION}"
	@git push origin v${NEW_VERSION}
	@echo "Released v${NEW_VERSION}"

release-major:
	@echo "Creating major release..."
	$(eval NEW_VERSION=$(shell echo ${CURRENT_VERSION} | awk -F. '{$$1 = substr($$1,2) + 1; $$(NF-1) = 0; $$NF = 0;} 1' | sed 's/ /./g'))
	@git tag -a v${NEW_VERSION} -m "Release v${NEW_VERSION}"
	@git push origin v${NEW_VERSION}
	@echo "Released v${NEW_VERSION}"

# Development helpers
.PHONY: fmt lint

fmt:
	@echo "Formatting code..."
	go fmt ./...
	find . -name "*.go" -exec goimports -w {} \;

lint:
	@echo "Linting code..."
	go vet ./...
	golint ./...
