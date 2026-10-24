# Test targets; included by the root Makefile, run from $(SZ_ROOT).

.PHONY: test.cover

test.cover:
	@cd $(SZ_ROOT) && go test ./... -cover
