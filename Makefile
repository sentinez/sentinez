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
	apiserver.build 	\
	greeter.build 		\
	edge.build			\
	realtime.build

default.print:
	@echo "[BUILD] senz: build sentinez and services"

test.cover:
	@go test ./... -cover

#####################################################################
# Go linting tool                                              
#####################################################################
lint: lint.go lint.proto lint.core

lint.go:
	@echo "[LINT] sentinez is linting ..."
	@golangci-lint run

lint.proto:
	@echo "[LINT] api is linting ..."
	@cd ./api && buf lint
	@cd ./api && golangci-lint run

lint.tools:
	@cd ./tools && golangci-lint run

lint.core:
	@echo "[LINT] core is linting ..."
	@cd ./core && golangci-lint run

#####################################################################
#####################################################################

realtime.run: SENTINEZ_OUT ?= realtime
realtime.run:
	@go build -ldflags="-s -w" -o ./cmd/realtime/bin/$(SENTINEZ_OUT) ./cmd/realtime && \
 	./cmd/realtime/bin/$(SENTINEZ_OUT)

realtime.build: SENTINEZ_OUT ?= realtime
realtime.build:
	@go build -ldflags="-s -w" -o ./cmd/realtime/bin/$(SENTINEZ_OUT) ./cmd/realtime
	@echo "[DONE]  senz: gateway.realtime ... ok"

apiserver.build: SENTINEZ_OUT ?= apiserver
apiserver.build:
	@go build -ldflags="-s -w" -o ./cmd/apiserver/bin/$(SENTINEZ_OUT) ./cmd/apiserver
	@echo "[DONE]  senz: gateway.apiserver ... ok"

apiserver.run: SENTINEZ_OUT ?= apiserver
apiserver.run:
	@go build -ldflags="-s -w" -o ./cmd/apiserver/bin/$(SENTINEZ_OUT) ./cmd/apiserver && \
 	./cmd/apiserver/bin/$(SENTINEZ_OUT)

apiserver.image.build: TAG ?= sentinez/apiserver
apiserver.image.build:
	docker buildx build -f ./cmd/apiserver/Dockerfile -t $(TAG):latest .

greeter.build: SENTINEZ_OUT ?= greeter
greeter.build:
	@go build -ldflags="-s -w" -o ./cmd/greeter/v1/bin/$(SENTINEZ_OUT) ./cmd/greeter/v1
	@echo "[DONE]  senz: core.greeter.v1 ... ok"

greeter.run: SENTINEZ_OUT ?= greeter
greeter.run:
	@go build -ldflags="-s -w" -o ./cmd/greeter/v1/bin/$(SENTINEZ_OUT) ./cmd/greeter/v1 && \
	./cmd/greeter/v1/bin/$(SENTINEZ_OUT)

greeter.image.build: TAG ?= sentinez/greeter
greeter.image.build:
	docker buildx build -f ./cmd/greeter/v1/Dockerfile -t $(TAG):latest .

edge.run: SENTINEZ_OUT ?= edge
edge.run:
	@go build -ldflags="-s -w" -o ./cmd/edge/v1/bin/$(SENTINEZ_OUT) ./cmd/edge/v1 && \
	./cmd/edge/v1/bin/$(SENTINEZ_OUT) \
		--certificate_file=cmd/edge/v1/is.s6z.io.vn.cert \
		--cert_key_file=cmd/edge/v1/is.s6z.io.vn.key \
		--rule_path=./data/crs/v4-16-0 \
		--proxy_config=./cmd/edge/v1/proxy.yaml \
		--env_file=./cmd/edge/v1/.env

edge.build: SENTINEZ_OUT ?= edge
edge.build:
	@go build -ldflags="-s -w" -o ./cmd/edge/v1/bin/$(SENTINEZ_OUT) ./cmd/edge/v1
	@echo "[DONE]  senz: gateway.edge.v1 ... ok"

edge.image.build: TAG ?= sentinez/edge
edge.image.build:
	@docker buildx build -f ./cmd/edge/v1/Dockerfile -t $(TAG):latest .

image.clear:
	@docker rmi hashicorp/consul
	@docker rmi sentinez/sentinez_api
	@docker rmi sentinez/sentinez_edge

compose.up:
	@docker compose -f deploy/docker/docker-compose.yaml up -d

compose.down:
	@docker compose -f deploy/docker/docker-compose.yaml down
