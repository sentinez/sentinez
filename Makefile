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

# Repository root; every included target runs from here. Must stay above
# the includes, since each one appends to MAKEFILE_LIST.
SZ_ROOT := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))

.PHONY: default default.print

default: default.print 	\
	sz.edge.build			\
	sz.dataplane.build		\
	sz.apiserver.build 		\
	sz.realtime.build      	\
	sz.greeter.build 		\
	sz.centraldata.build

default.print:
	@echo "[BUILD] senz: build sentinez and services"

# Repo-wide targets, one file per concern.
include ./hack/make/test.mk
include ./hack/make/proto.mk
include ./hack/make/lint.mk
include ./hack/make/docker.mk
include ./hack/make/ee.mk

include ./hack/net/dev/Makefile
include ./api/Makefile

# Service targets follow cmd/sz<name>[/v1], one Makefile per service:
#   sz.<name>.build | sz.<name>.run | sz.<name>.image.build
include ./cmd/szapiserver/Makefile
include ./cmd/szrealtime/Makefile
include ./cmd/szgreeter/Makefile
include ./cmd/szcentraldata/Makefile
include ./cmd/szedge/Makefile
include ./cmd/szdataplane/Makefile
