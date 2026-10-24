# Directory of this Makefile, so its targets work both from api/ and when
# included by the root Makefile. Must stay above any other include.
API_DIR := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))

.PHONY: bufgen
bufgen:
	cd $(API_DIR) && buf generate
	cd $(API_DIR)/../ui/packages/proto && pnpm run build
