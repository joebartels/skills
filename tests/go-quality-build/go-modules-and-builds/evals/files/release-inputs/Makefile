GO ?= go

.PHONY: test build preview source

test:
	$(GO) test ./...

build:
	mkdir -p dist
	$(GO) build -o dist/codes ./cmd/codes

preview:
	$(GO) generate ./...

# Export the working source tree, including new files awaiting review.
source:
	mkdir -p dist
	git ls-files --cached --others --exclude-standard -z | tar --null -T - -czf dist/codes-src.tar.gz
