# Copyright 2025 Duc-Hung Ho.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

.PHONY: default

default: default.print 	\
	sz.edge.build			\
	sz.dataplane.build		\
	sz.apiserver.build 		\
	sz.realtime.build      	\
	sz.greeter.build 		\
	sz.centraldata.build

default.print:
	@echo "[BUILD] senz: build sentinez and services"

test.cover:
	@go test ./... -cover

fmt.proto:
	@cd ./api && buf format -w

#####################################################################
# Go linting tool                                              
#####################################################################
lint: lint.go lint.proto lint.core lint.shared lint.controlplane lint.contrib.httphz lint.tools

lint.go:
	@echo "[LINT] sentinez is linting ..."
	@golangci-lint run

lint.proto:
	@echo "[LINT] api is linting ..."
	@cd ./api && buf lint
	@cd ./api && golangci-lint run

lint.tools:
	@cd ./staging/src/github.com/sentinez/tools && golangci-lint run

lint.core:
	@echo "[LINT] core is linting ..."
	@cd ./staging/src/github.com/sentinez/core && golangci-lint run

lint.shared:
	@echo "[LINT] shared is linting ..."
	@cd ./staging/src/github.com/sentinez/shared && golangci-lint run

lint.controlplane:
	@echo "[LINT] controlplane is linting ..."
	@cd ./staging/src/github.com/sentinez/controlplane && golangci-lint run

lint.contrib.httphz:
	@echo "[LINT] httphz is linting ..."
	@cd ./staging/src/github.com/sentinez/contrib/httphz && golangci-lint run

#####################################################################
#####################################################################

# Service targets follow cmd/sz<name>[/v1]:
#   sz.<name>.build | sz.<name>.run | sz.<name>.image.build
# Override the binary name with SENTINEZ_OUT and the image with TAG.

sz.apiserver.build: SENTINEZ_OUT ?= apiserver
sz.apiserver.build:
	@go build -ldflags="-s -w" -o ./cmd/szapiserver/bin/$(SENTINEZ_OUT) ./cmd/szapiserver
	@echo "[DONE]  senz: sz.apiserver ... ok"

sz.apiserver.run: SENTINEZ_OUT ?= apiserver
sz.apiserver.run: sz.apiserver.build
	@./cmd/szapiserver/bin/$(SENTINEZ_OUT) --env_file=./cmd/szapiserver/.env

sz.apiserver.image.build: TAG ?= sentinez/apiserver
sz.apiserver.image.build:
	@docker buildx build -f ./cmd/szapiserver/Dockerfile -t $(TAG):latest .

sz.realtime.build: SENTINEZ_OUT ?= realtime
sz.realtime.build:
	@go build -ldflags="-s -w" -o ./cmd/szrealtime/bin/$(SENTINEZ_OUT) ./cmd/szrealtime
	@echo "[DONE]  senz: sz.realtime ... ok"

sz.realtime.run: SENTINEZ_OUT ?= realtime
sz.realtime.run: sz.realtime.build
	@./cmd/szrealtime/bin/$(SENTINEZ_OUT) --env_file=./cmd/szrealtime/.env

sz.greeter.build: SENTINEZ_OUT ?= greeter
sz.greeter.build:
	@go build -ldflags="-s -w" -o ./cmd/szgreeter/v1/bin/$(SENTINEZ_OUT) ./cmd/szgreeter/v1
	@echo "[DONE]  senz: sz.greeter.v1 ... ok"

sz.greeter.run: SENTINEZ_OUT ?= greeter
sz.greeter.run: sz.greeter.build
	@./cmd/szgreeter/v1/bin/$(SENTINEZ_OUT) --env_file=./cmd/szgreeter/v1/.env

sz.greeter.image.build: TAG ?= sentinez/greeter
sz.greeter.image.build:
	@docker buildx build -f ./cmd/szgreeter/v1/Dockerfile -t $(TAG):latest .

sz.centraldata.build: SENTINEZ_OUT ?= centraldata
sz.centraldata.build:
	@go build -ldflags="-s -w" -o ./cmd/szcentraldata/v1/bin/$(SENTINEZ_OUT) ./cmd/szcentraldata/v1
	@echo "[DONE]  senz: sz.centraldata.v1 ... ok"

sz.centraldata.image.build: TAG ?= sentinez/centraldata
sz.centraldata.image.build:
	@docker buildx build -f ./cmd/szcentraldata/v1/Dockerfile -t $(TAG):latest .

sz.edge.build: SENTINEZ_OUT ?= edge
sz.edge.build:
	@go build -ldflags="-s -w" -o ./cmd/szedge/v1/bin/$(SENTINEZ_OUT) ./cmd/szedge/v1
	@echo "[DONE]  senz: sz.edge.v1 ... ok"

sz.edge.run: SENTINEZ_OUT ?= edge
sz.edge.run: sz.edge.build
	@./cmd/szedge/v1/bin/$(SENTINEZ_OUT) \
		--cert_file=cmd/szedge/v1/is.s6z.io.vn.cert \
		--cert_key_file=cmd/szedge/v1/is.s6z.io.vn.key \
		--proxy_config=./cmd/szedge/v1/proxy.yaml \
		--env_file=./cmd/szedge/v1/.env \
		--engine=quic

sz.edge.image.build: TAG ?= sentinez/edge
sz.edge.image.build:
	@docker buildx build -f ./cmd/szedge/v1/Dockerfile -t $(TAG):latest .

sz.dataplane.build: SENTINEZ_OUT ?= dataplane
sz.dataplane.build:
	@go build -ldflags="-s -w" -o ./cmd/szdataplane/v1/bin/$(SENTINEZ_OUT) ./cmd/szdataplane/v1
	@echo "[DONE]  senz: sz.dataplane ... ok"

# The dataplane attaches eBPF/XDP programs inside the "gateway" netns.
sz.dataplane.run: SENTINEZ_OUT ?= dataplane
sz.dataplane.run: sz.dataplane.build
	sudo ip netns exec gateway ./cmd/szdataplane/v1/bin/$(SENTINEZ_OUT) \
		--env_file=./cmd/szdataplane/v1/.env

image.clear:
	@docker rmi hashicorp/consul
	@docker rmi sentinez/apiserver
	@docker rmi sentinez/edge

compose.up:
	@docker compose -f deploy/docker/docker-compose.yaml up -d

compose.down:
	@docker compose -f deploy/docker/docker-compose.yaml down

include ./hack/net/dev/Makefile
