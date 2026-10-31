# Enterprise features; included by the root Makefile, run from $(SZ_ROOT).
# The private scale repository is checked out in scale/, inside the root
# module, and compiled with the enterprise build tag when SCALE=1, e.g.
#   make scale.checkout && make sz.edge.build SCALE=1
#
# Its URL is never committed: set SCALE_REPO in the environment or in the
# git-ignored hack/make/scale.local.mk (SCALE_REPO := <url>).
-include $(SZ_ROOT)/hack/make/scale.local.mk

SCALE_BRANCH ?= main
SCALE_DIR := $(SZ_ROOT)/scale

# Build tags of the go builds; the *_oss.go stubs of scale/ build without.
SZ_GOTAGS ?= $(if $(SCALE),enterprise)

.PHONY: scale.checkout scale.pull scale.track

# The public stubs stay in place: scale/ is not empty, so no git clone.
scale.checkout:
	@test ! -d $(SCALE_DIR)/.git || \
		{ echo "[SKIP]  scale: already checked out"; exit 0; }; \
	test -n "$(SCALE_REPO)" || \
		{ echo "[FAIL]  scale: SCALE_REPO is not set"; exit 1; }; \
	git -C $(SCALE_DIR) init -q -b $(SCALE_BRANCH) && \
	git -C $(SCALE_DIR) remote add origin "$(SCALE_REPO)" && \
	git -C $(SCALE_DIR) fetch -q origin $(SCALE_BRANCH) && \
	git -C $(SCALE_DIR) checkout -q -f -B $(SCALE_BRANCH) \
		origin/$(SCALE_BRANCH) && \
	echo "[DONE]  scale: checked out in scale/"

scale.pull:
	@git -C $(SCALE_DIR) pull -q --ff-only origin $(SCALE_BRANCH)
	@echo "[DONE]  scale: scale/ up to date"

# git skips the files of scale/ once it holds a nested repository: track
# a new public stub through the index, then git add works on it as usual.
scale.track:
	@cd $(SZ_ROOT) && \
	for f in scale/pkg_import.go scale/*_oss.go; do \
		git update-index --add \
			--cacheinfo 100644,$$(git hash-object -w $$f),$$f; \
	done
	@echo "[DONE]  scale: public stubs tracked"
