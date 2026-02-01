DISTDIR=${CURDIR}/dist

$(DISTDIR):
	mkdir -p $@

.PHONY: cli
cli: $(DISTDIR)/sdsheet

go.mod: $(shell find ./ -type f -name '*.go')
	go mod tidy

$(DISTDIR)/sdsheetscraper: $(DISTDIR) go.mod $(shell find ./ -type f -name '*.go')
	CGO_ENABLED=0 go build -C cmd/cli/sdsheetscraper -o $@  .

.PHONY: build
build: $(DISTDIR)/sdaservice
	@echo "built stuff"

$(DISTDIR)/sdaservice: go.mod $(shell find ./ -type f -name '*.go') | $(DISTDIR)
	CGO_ENABLED=0 go build -o $@  .
