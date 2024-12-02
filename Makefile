.DEFAULT_GOAL := help

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' Makefile | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: tidy
tidy: ## Run go mod tidy
	go mod tidy

.PHONY: run-app-etcd
run-app-etcd: ## Run etcd
	@echo "Running etcd"
	@docker compose -f example/docker-compose.yml run --rm app-etcd

.PHONY: run-app-consul
run-app-consul: ## Run consul
	@echo "Running consul"
	@docker compose -f example/docker-compose.yml run --rm app-consul

.PHONY: run-app-zk
run-app-zk: ## Run zookeeper
	@echo "Running zookeeper"
	@docker compose -f example/docker-compose.yml run --rm app-zookeeper
