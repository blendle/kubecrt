BINARY=kubecrt

TAG=$(shell git for-each-ref refs/tags --sort=-creatordate --format='%(refname:short)' --count=1)
MAJOR=`echo $(TAG) | awk -F[v.] '{print $$2}'`
MINOR=`echo $(TAG) | awk -F[v.] '{print $$3}'`
PATCH=`echo $(TAG) | awk -F[v.] '{print $$4}'`
GIT_COMMIT=`git rev-parse --short @`
LDFLAGS=-X github.com/blendle/kubecrt/config.version=$(TAG) -X github.com/blendle/kubecrt/config.gitrev=$(GIT_COMMIT)

build:
	mkdir -p bin
	go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY)

prep:
	@mkdir -p _dist

dist: prep
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w $(LDFLAGS)" -o _dist/$(BINARY)_$(TAG)_linux_amd64
	GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w $(LDFLAGS)" -o _dist/$(BINARY)_$(TAG)_darwin_amd64

patch: prep
	@version=v$(MAJOR).$(MINOR).$$(expr $(PATCH) + 1); \
	git tag $$version

minor: prep
	@version=v$(MAJOR).$$(expr $(MINOR) + 1).0; \
	git tag $$version

major: prep
	@version=v$$(expr $(MAJOR) + 1).0.0; \
	git tag $$version

push:
	git push --tags
