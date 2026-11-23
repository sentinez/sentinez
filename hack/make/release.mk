# Release targets; included by the root Makefile, run from $(SZ_ROOT).

.PHONY: release.check

# release.check runs the automated gates of a release (docs/14-test-plan.md,
# A-01): gofmt, golangci-lint and go test in every module of
# hack/gomodules.sh, buf lint, then builds every service. It stops at the
# first failure.
release.check: fmt.go.check
	@cd $(SZ_ROOT) && bash ./hack/golint.sh
	@cd $(SZ_ROOT) && bash ./hack/protolint.sh
	@cd $(SZ_ROOT) && bash ./hack/gotest.sh
	@$(MAKE) --no-print-directory -C $(SZ_ROOT) default
	@echo "[DONE]  sz: release.check ... ok"
