# Docker image and compose targets; included by the root Makefile, run
# from $(SZ_ROOT).
SZ_COMPOSE := $(SZ_ROOT)/deploy/docker/docker-compose.yaml

# proxy.yaml the apiserver imports from Consul at startup, under the key
# of its --proxy_config flag.
CONSUL_ADDR ?= http://sentinez.consul:8500
CONSUL_PROXY_KEY ?= sentinez/proxy.yaml
CONSUL_PROXY_FILE ?= $(SZ_ROOT)/cmd/szedge/v1/proxy.yaml
# the Consul KV of the edges: the lock of a zone is leader/<zone>
SZ_EDGE_KV = $(CONSUL_ADDR)/v1/kv/sentinez/dmz/edge

.PHONY: image.clear compose.up compose.down consul.proxy.put consul.env.put
.PHONY: compose.edge.up compose.edge.down compose.edge.leader
.PHONY: compose.edge.consul compose.edge.roles

image.clear:
	@docker rmi hashicorp/consul
	@docker rmi sentinez/apiserver
	@docker rmi sentinez/edge

compose.up:
	@cd $(SZ_ROOT) && docker compose -f $(SZ_COMPOSE) up -d

compose.down:
	@cd $(SZ_ROOT) && docker compose -f $(SZ_COMPOSE) down

# A cluster of 3 edges electing the leader of their zone through Consul
# (deploy/docker/cluster). Starts the Consul and the Kafka of $(SZ_COMPOSE)
# first, and its centraldata, which publishes to the edges, last; the
# images come from make sz.edge.image.build sz.centraldata.image.build.
SZ_EDGE_COMPOSE := $(SZ_ROOT)/deploy/docker/cluster/docker-compose.yaml
SZ_EDGE_ZONE ?= han

compose.edge.up:
	@cd $(SZ_ROOT) && docker compose -f $(SZ_COMPOSE) up -d \
		sentinez.consul sentinez.kafka-init
	@cd $(SZ_ROOT) && docker compose -f $(SZ_EDGE_COMPOSE) up -d
	@cd $(SZ_ROOT) && docker compose -f $(SZ_COMPOSE) up -d \
		sentinez.apps.centraldata

compose.edge.down:
	@cd $(SZ_ROOT) && docker compose -f $(SZ_EDGE_COMPOSE) down

# The node holding the lock of the zone, and the status it last wrote.
compose.edge.leader:
	@echo "lock:   $$(curl -sSf \
		'$(SZ_EDGE_KV)/leader/$(SZ_EDGE_ZONE)?raw')"
	@echo "status: $$(curl -sSf \
		'$(SZ_EDGE_KV)/zone/$(SZ_EDGE_ZONE)/status?raw')"

# The role of every running edge: leader when it holds the lock of its
# zone (SZ_ZONE, else $(SZ_EDGE_ZONE)), whose node is its hostname and a
# suffix; else follower.
compose.edge.roles:
	@printf '%-24s %-6s %s\n' EDGE ZONE ROLE
	@for c in $$(docker ps --filter name=sentinez.dmz.edge \
		--format '{{.Names}}' | sort); do \
		set -- $$(docker inspect \
			-f '{{.Config.Hostname}}{{range .Config.Env}} {{.}}{{end}}' $$c); \
		host=$$1; \
		zone=$$(printf '%s\n' "$$@" | sed -n 's/^SZ_ZONE=//p' | \
			tr A-Z a-z); \
		zone=$${zone:-$(SZ_EDGE_ZONE)}; \
		lock=$$(curl -sSf "$(SZ_EDGE_KV)/leader/$$zone?raw" 2>/dev/null); \
		case "$$lock" in "$$host"-*) role=leader;; *) role=follower;; esac; \
		printf '%-24s %-6s %s\n' $$c $$zone $$role; \
	done

# The Raft peers of the Consul servers, and which of them leads.
compose.edge.consul:
	@docker exec sentinez.consul consul operator raft list-peers

consul.proxy.put:
	@curl -sSf -X PUT --data-binary @$(CONSUL_PROXY_FILE) \
		$(CONSUL_ADDR)/v1/kv/$(CONSUL_PROXY_KEY) >/dev/null
	@echo "put $(CONSUL_PROXY_FILE) at $(CONSUL_PROXY_KEY)"

# Puts the SZ_* lines of ENV_FILE in the Consul KV: by default the env shared
# by all the services (deploy/consul/shared.env, copied from its .example),
# or the file of one service when SERVICE is its service_key. See
# hack/consul-env-put.sh.
#   make consul.env.put
#   make consul.env.put ENV_FILE=cmd/szedge/v1/.env SERVICE=sentinez.dmz.edge.v1
ENV_FILE ?= $(SZ_ROOT)/deploy/consul/shared.env

consul.env.put:
	@test -f "$(ENV_FILE)" || { echo "no $(ENV_FILE): cp deploy/consul/shared.env.example deploy/consul/shared.env, then edit it"; exit 1; }
	@CONSUL_ADDR=$(CONSUL_ADDR) $(SZ_ROOT)/hack/consul-env-put.sh "$(ENV_FILE)" "$(SERVICE)"
