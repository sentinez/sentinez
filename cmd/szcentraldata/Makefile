# Targets for cmd/szcentraldata; included by the root Makefile.
# Override the binary name with SENTINEZ_OUT and the image with TAG.
SZCENTRALDATA_DIR := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))

.PHONY: sz.centraldata.build sz.centraldata.run sz.centraldata.image.build
.PHONY: sz.centraldata.cert

sz.centraldata.build: SENTINEZ_OUT ?= centraldata
sz.centraldata.build:
	@go build -ldflags="-s -w" -o $(SZCENTRALDATA_DIR)/v1/bin/$(SENTINEZ_OUT) $(SZCENTRALDATA_DIR)/v1
	@echo "[DONE]  sz: sz.centraldata.v1 ... ok"

sz.centraldata.run: SENTINEZ_OUT ?= centraldata
sz.centraldata.run: sz.centraldata.build
	@$(SZCENTRALDATA_DIR)/v1/bin/$(SENTINEZ_OUT) --env_file=$(SZCENTRALDATA_DIR)/v1/.env

sz.centraldata.image.build: TAG ?= sentinez/centraldata
sz.centraldata.image.build:
	@docker buildx build -f $(SZCENTRALDATA_DIR)/v1/Dockerfile -t $(TAG):latest $(SZCENTRALDATA_DIR)/../..

# Issues $(SZ_CENTRALDATA_CERT_NAME).pem and its key in $(SZ_INTERNAL_CERT_DIR),
# and the CA there first if there is none.
sz.centraldata.cert:
	@$(SZ_ROOT)/hack/mkcert-internal.sh $(SZ_INTERNAL_CERT_DIR) \
		$(SZ_CENTRALDATA_CERT_NAME) $(SZ_CENTRALDATA_SAN)
