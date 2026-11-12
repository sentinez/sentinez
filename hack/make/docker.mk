# Docker image and compose targets; included by the root Makefile, run
# from $(SZ_ROOT).
SZ_COMPOSE := $(SZ_ROOT)/deploy/docker/docker-compose.yaml

# proxy.yaml the apiserver imports from Consul at startup, under the key
# of its --proxy_config flag.
CONSUL_ADDR ?= http://localhost:8500
CONSUL_PROXY_KEY ?= sentinez/proxy.yaml
CONSUL_PROXY_FILE ?= $(SZ_ROOT)/cmd/szedge/v1/proxy.yaml

.PHONY: image.clear compose.up compose.down consul.proxy.put consul.env.put

image.clear:
	@docker rmi hashicorp/consul
	@docker rmi sentinez/apiserver
	@docker rmi sentinez/edge

compose.up:
	@cd $(SZ_ROOT) && docker compose -f $(SZ_COMPOSE) up -d

compose.down:
	@cd $(SZ_ROOT) && docker compose -f $(SZ_COMPOSE) down

consul.proxy.put:
	@curl -sSf -X PUT --data-binary @$(CONSUL_PROXY_FILE) \
		$(CONSUL_ADDR)/v1/kv/$(CONSUL_PROXY_KEY) >/dev/null
	@echo "put $(CONSUL_PROXY_FILE) at $(CONSUL_PROXY_KEY)"

# Puts the SENZ_* lines of ENV_FILE in the Consul KV: by default the env shared
# by all the services (deploy/consul/shared.env, copied from its .example),
# or the file of one service when SERVICE is its service_key. See
# hack/consul-env-put.sh.
#   make consul.env.put
#   make consul.env.put ENV_FILE=cmd/szedge/v1/.env SERVICE=sentinez.dmz.edge.v1
ENV_FILE ?= $(SZ_ROOT)/deploy/consul/shared.env

consul.env.put:
	@test -f "$(ENV_FILE)" || { echo "no $(ENV_FILE): cp deploy/consul/shared.env.example deploy/consul/shared.env, then edit it"; exit 1; }
	@CONSUL_ADDR=$(CONSUL_ADDR) $(SZ_ROOT)/hack/consul-env-put.sh "$(ENV_FILE)" "$(SERVICE)"
