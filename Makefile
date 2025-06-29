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

default: default.print \
	apiserver.build \
	greeter.build \
	edge.build

default.print:
	@echo "[BUILD] SENTINEZ: build sentinez and services"

test.cover:
	@go test ./... -cover

#####################################################################
# Go linting tool                                              
#####################################################################
lint: lint.go lint.proto

lint.go:
	@golangci-lint run

lint.proto:
	@cd ./api && buf lint

#####################################################################
#####################################################################

websocket.run: SENTINEZ_OUT ?= websocket
websocket.run:
	@go build -ldflags="-s -w" -o ./cmd/websocket/bin/$(SENTINEZ_OUT) ./cmd/websocket && \
 	./cmd/websocket/bin/$(SENTINEZ_OUT)

apiserver.build: SENTINEZ_OUT ?= apiserver
apiserver.build:
	@go build -ldflags="-s -w" -o ./cmd/apiserver/bin/$(SENTINEZ_OUT) ./cmd/apiserver
	@echo "[DONE]  SENTINEZ: gateway.apiserver ... ok"

apiserver.run: SENTINEZ_OUT ?= apiserver
apiserver.run:
	@go build -ldflags="-s -w" -o ./cmd/apiserver/bin/$(SENTINEZ_OUT) ./cmd/apiserver && \
 	./cmd/apiserver/bin/$(SENTINEZ_OUT)

apiserver.build.image: TAG ?= sentinez/sentinez
apiserver.build.image:
	docker buildx build -f ./cmd/apiserver/Dockerfile -t $(TAG):latest .


greeter.build: SENTINEZ_OUT ?= greeter
greeter.build:
	@go build -ldflags="-s -w" -o ./cmd/greeter/v1/bin/$(SENTINEZ_OUT) ./cmd/greeter/v1
	@echo "[DONE]  SENTINEZ: core.greeter.v1 ... ok"

greeter.run: SENTINEZ_OUT ?= greeter
greeter.run:
	@go build -ldflags="-s -w" -o ./cmd/greeter/v1/bin/$(SENTINEZ_OUT) ./cmd/greeter/v1 && \
	./cmd/greeter/v1/bin/$(SENTINEZ_OUT) -a=127.0.0.1:8001

greeter.build.image: TAG ?= sentinez/sentinez.core.greeter
greeter.build.image:
	docker buildx build -f ./cmd/greeter/v1/Dockerfile -t $(TAG):latest .

edge.run: SENTINEZ_OUT ?= edge
edge.run:
	@go build -ldflags="-s -w" -o ./cmd/edge/v1/bin/$(SENTINEZ_OUT) ./cmd/edge/v1 && \
	./cmd/edge/v1/bin/$(SENTINEZ_OUT)

edge.build: SENTINEZ_OUT ?= edge
edge.build:
	@go build -ldflags="-s -w" -o ./cmd/edge/v1/bin/$(SENTINEZ_OUT) ./cmd/edge/v1
	@echo "[DONE]  SENTINEZ: gateway.edge.v1 ... ok"
