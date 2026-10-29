# Enterprise features; included by the root Makefile, run from $(SZ_ROOT).
# The private repository is checked out in pkg/ee, inside the root module,
# and compiled with the enterprise build tag when EE=1, e.g.
#   make ee.checkout && make sz.edge.build EE=1
#
# Its URL is never committed: set EE_REPO in the environment or in the
# git-ignored hack/make/ee.local.mk (EE_REPO := <url>).
-include $(SZ_ROOT)/hack/make/ee.local.mk

EE_BRANCH ?= main
EE_DIR := $(SZ_ROOT)/pkg/ee

# Build tags of the go builds; the *_oss.go stubs of pkg/ee build without.
SZ_GOTAGS ?= $(if $(EE),enterprise)

.PHONY: ee.checkout ee.pull ee.track

# The public stubs stay in place: pkg/ee is not empty, so no git clone.
ee.checkout:
	@test ! -d $(EE_DIR)/.git || { echo "[SKIP]  ee: already checked out"; exit 0; }; \
	test -n "$(EE_REPO)" || { echo "[FAIL]  ee: EE_REPO is not set"; exit 1; }; \
	git -C $(EE_DIR) init -q -b $(EE_BRANCH) && \
	git -C $(EE_DIR) remote add origin "$(EE_REPO)" && \
	git -C $(EE_DIR) fetch -q origin $(EE_BRANCH) && \
	git -C $(EE_DIR) checkout -q -f -B $(EE_BRANCH) origin/$(EE_BRANCH) && \
	echo "[DONE]  ee: checked out in pkg/ee"

ee.pull:
	@git -C $(EE_DIR) pull -q --ff-only origin $(EE_BRANCH)
	@echo "[DONE]  ee: pkg/ee up to date"

# git skips the files of pkg/ee once it holds a nested repository: track a
# new public stub through the index, then git add works on it as usual.
ee.track:
	@cd $(SZ_ROOT) && for f in pkg/ee/pkg_import.go pkg/ee/*_oss.go; do \
		git update-index --add \
			--cacheinfo 100644,$$(git hash-object -w $$f),$$f; \
	done
	@echo "[DONE]  ee: public stubs tracked"
