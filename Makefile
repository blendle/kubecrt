BINARY=kubecrt

TAG=$(shell git for-each-ref refs/tags --sort=-taggerdate --format='%(refname:short)' --count=1)
MAJOR=`echo $(TAG) | awk -F[v.] '{print $$2}'`
MINOR=`echo $(TAG) | awk -F[v.] '{print $$3}'`
PATCH=`echo $(TAG) | awk -F[v.] '{print $$4}'`
GIT_COMMIT=`git rev-parse --short @`
LDFLAGS=-X github.com/blendle/kubecrt/config.version=$(VERSION) -X github.com/blendle/kubecrt/config.gitrev=$(GIT_COMMIT)

build:
	mkdir -p bin
	go build -o bin/$(BINARY)

release-prepare:
	@mkdir -p _dist

patch: release-prepare
	@version=v$(MAJOR).$(MINOR).$$(expr $(PATCH) + 1); \
	git tag $$version; \
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w $(LDFLAGS)" -o _dist/$(BINARY)-$$version-linux-amd64; \
	GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w $(LDFLAGS)" -o _dist/$(BINARY)-$$version-darwin-amd64

minor: release-prepare
	@version=v$(MAJOR).$$(expr $(MINOR) + 1).0; \
	git tag $$version; \
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w $(LDFLAGS)" -o _dist/$(BINARY)-$$version-linux-amd64; \
	GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w $(LDFLAGS)" -o _dist/$(BINARY)-$$version-darwin-amd64

major: release-prepare
	@version=v$$(expr $(MAJOR) + 1).0.0; \
	git tag $$version; \
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w $(LDFLAGS)" -o _dist/$(BINARY)-$$version-linux-amd64; \
	GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w $(LDFLAGS)" -o _dist/$(BINARY)-$$version-darwin-amd64

push:
	git push --tags
