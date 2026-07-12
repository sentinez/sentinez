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
	dmz.edge.build			\
	dmz.dataplane.build		\
	acz.apiserver.build 	\
	acz.realtime.build      \
	mesh.greeter.build 		

default.print:
	@echo "[BUILD] senz: build sentinez and services"

test.cover:
	@go test ./... -cover

fmt.proto:
	@cd ./api && buf format -w

#####################################################################
# Go linting tool                                              
#####################################################################
lint: lint.go lint.proto lint.core lint.core lint.shared lint.modules lint.contrib.httphz

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

lint.modules:
	@echo "[LINT] modules is linting ..."
	@cd ./staging/src/github.com/sentinez/modules && golangci-lint run

lint.contrib.httphz:
	@echo "[LINT] httphz is linting ..."
	@cd ./staging/src/github.com/sentinez/contrib/httphz && golangci-lint run

#####################################################################
#####################################################################

acz.realtime.run: SENTINEZ_OUT ?= realtime
acz.realtime.run:
	@go build -ldflags="-s -w" -o ./cmd/acz-realtime/bin/$(SENTINEZ_OUT) ./cmd/acz-realtime && \
 	./cmd/acz-realtime/bin/$(SENTINEZ_OUT)

acz.realtime.build: SENTINEZ_OUT ?= realtime
acz.realtime.build:
	@go build -ldflags="-s -w" -o ./cmd/acz-realtime/bin/$(SENTINEZ_OUT) ./cmd/acz-realtime
	@echo "[DONE]  senz: acz.realtime ... ok"

acz.apiserver.build: SENTINEZ_OUT ?= apiserver
acz.apiserver.build:
	@go build -ldflags="-s -w" -o ./cmd/acz-apiserver/bin/$(SENTINEZ_OUT) ./cmd/acz-apiserver
	@echo "[DONE]  senz: acz.apiserver ... ok"

acz.apiserver.run: SENTINEZ_OUT ?= apiserver
acz.apiserver.run:
	@go build -ldflags="-s -w" -o ./cmd/acz-apiserver/bin/$(SENTINEZ_OUT) ./cmd/acz-apiserver && \
 	./cmd/acz-apiserver/bin/$(SENTINEZ_OUT) 

acz.apiserver.image.build: TAG ?= sentinez/acz-apiserver
acz.apiserver.image.build:
	docker buildx build -f ./cmd/acz-apiserver/Dockerfile -t $(TAG):latest .

mesh.greeter.build: SENTINEZ_OUT ?= greeter
mesh.greeter.build:
	@go build -ldflags="-s -w" -o ./cmd/mesh-greeter/v1/bin/$(SENTINEZ_OUT) ./cmd/mesh-greeter/v1
	@echo "[DONE]  senz: mesh.greeter.v1 ... ok"

mesh.greeter.run: SENTINEZ_OUT ?= greeter
mesh.greeter.run:
	@go build -ldflags="-s -w" -o ./cmd/mesh-greeter/v1/bin/$(SENTINEZ_OUT) ./cmd/mesh-greeter/v1 && \
	./cmd/mesh-greeter/v1/bin/$(SENTINEZ_OUT) --env_file=./cmd/mesh-greeter/v1/.env

mesh.greeter.image.build: TAG ?= sentinez/greeter
mesh.greeter.image.build:
	docker buildx build -f ./cmd/mesh-greeter/v1/Dockerfile -t $(TAG):latest .

dmz.edge.run: SENTINEZ_OUT ?= edge
dmz.edge.run:
	@go build -ldflags="-s -w" -o ./cmd/dmz-edge/v1/bin/$(SENTINEZ_OUT) ./cmd/dmz-edge/v1 && \
	./cmd/dmz-edge/v1/bin/$(SENTINEZ_OUT) \
		--cert_file=cmd/dmz-edge/v1/is.s6z.io.vn.cert \
		--cert_key_file=cmd/dmz-edge/v1/is.s6z.io.vn.key \
		--rule_path=./deploy/ruleroot/v4-16-0 \
		--proxy_config=./cmd/dmz-edge/v1/proxy.yaml \
		--env_file=./cmd/dmz-edge/v1/.env

dmz.dataplane.build: SENTINEZ_OUT ?= dataplane
dmz.dataplane.build:
	@go build -ldflags="-s -w" -o ./cmd/dmz-dataplane/v1/bin/$(SENTINEZ_OUT) ./cmd/dmz-dataplane/v1 
	@echo "[DONE]  senz: dmz.dataplane ... ok"


dmz.dataplane.run: SENTINEZ_OUT ?= dataplane
dmz.dataplane.run:
	sudo ip netns exec gateway ./cmd/dmz-dataplane/v1/bin/$(SENTINEZ_OUT) \
		--env_file=./cmd/dmz-dataplane/v1/.env

dmz.edge.build: SENTINEZ_OUT ?= edge
dmz.edge.build:
	@go build -ldflags="-s -w" -o ./cmd/dmz-edge/v1/bin/$(SENTINEZ_OUT) ./cmd/dmz-edge/v1
	@echo "[DONE]  senz: dmz.edge.v1 ... ok"

dmz.edge.image.build: TAG ?= sentinez/edge
dmz.edge.image.build:
	@docker buildx build -f ./cmd/dmz-edge/v1/Dockerfile -t $(TAG):latest .

image.clear:
	@docker rmi hashicorp/consul
	@docker rmi sentinez/sentinez_api
	@docker rmi sentinez/sentinez_edge

compose.up:
	@docker compose -f deploy/docker/docker-compose.yaml up -d

compose.down:
	@docker compose -f deploy/docker/docker-compose.yaml down

include ./hack/net/dev/Makefile
