BINARY=kubecrt

VERSION=0.1.0
GIT_COMMIT=`git rev-parse --short @`
LDFLAGS=-X github.com/blendle/kubecrt/config.version=$(VERSION) -X github.com/blendle/kubecrt/config.gitrev=$(GIT_COMMIT)

build:
	mkdir -p bin
	go build -o bin/$(BINARY)

release:
	mkdir -p _dist
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w $(LDFLAGS)" -o _dist/$(BINARY)
	GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w $(LDFLAGS)" -o _dist/$(BINARY)_darwin64
