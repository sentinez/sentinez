# Docker image and compose targets; included by the root Makefile, run
# from $(SZ_ROOT).
SZ_COMPOSE := $(SZ_ROOT)/deploy/docker/docker-compose.yaml

.PHONY: image.clear compose.up compose.down

image.clear:
	@docker rmi hashicorp/consul
	@docker rmi sentinez/apiserver
	@docker rmi sentinez/edge

compose.up:
	@cd $(SZ_ROOT) && docker compose -f $(SZ_COMPOSE) up -d

compose.down:
	@cd $(SZ_ROOT) && docker compose -f $(SZ_COMPOSE) down
