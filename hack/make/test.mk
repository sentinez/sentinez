# Test targets; included by the root Makefile, run from $(SZ_ROOT).

.PHONY: test.cover bench

# test.cover runs every Go module, as CI does.
test.cover:
	@cd $(SZ_ROOT) && bash ./hack/gotest.sh -cover

# bench runs the benchmarks of the edge request path and writes their
# medians into the README (BENCH_COUNT runs each, 6 by default).
bench:
	@cd $(SZ_ROOT) && bash ./hack/bench.sh $(BENCH_COUNT)
