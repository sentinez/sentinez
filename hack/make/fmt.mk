# Go formatting; included by the root Makefile, run from $(SZ_ROOT).
# Covers every tracked or new (not ignored) .go file of all modules, git
# submodules excluded; files deleted from the worktree are skipped.
SZ_GOFILES = $(shell cd $(SZ_ROOT) && git ls-files --cached --others \
	--exclude-standard '*.go' ':!:_submodules/**' | \
	while read -r f; do [ -f "$$f" ] && echo "$$f"; done)

.PHONY: fmt fmt.go fmt.go.check

fmt: fmt.go fmt.proto

fmt.go:
	@echo "[FMT] go files are formatting ..."
	@cd $(SZ_ROOT) && gofmt -w $(SZ_GOFILES)

# Fails, listing the files, when any Go file is not gofmt-formatted.
fmt.go.check:
	@cd $(SZ_ROOT) && files="$$(gofmt -l $(SZ_GOFILES))"; \
	if [ -n "$$files" ]; then \
		echo "[FMT] run 'make fmt.go', unformatted files:"; \
		echo "$$files"; exit 1; \
	fi
