# Test targets; included by the root Makefile, run from $(SZ_ROOT).

.PHONY: test.cover

# test.cover runs every Go module, as CI does.
test.cover:
	@cd $(SZ_ROOT) && bash ./hack/gotest.sh -cover
