HAS_DOCKER := $(shell command -v docker;)
HAS_DOCKER_COMPOSE := $(shell command -v docker-compose;)
HAS_KUBECTL := $(shell command -v kubectl;)
PWD := $(shell pwd;)

.PHONY: check-docker
check-docker:
ifndef HAS_DOCKER
	@echo "Missing docker"
	@exit 1
endif
	@true

.PHONY: check-kubectl
check-kubectl:
ifndef HAS_KUBECTL
	@echo "Missing kubectl"
	@exit 1
endif
	@true

.PHONY: check-docker-compose
check-docker-compose:
ifndef HAS_DOCKER_COMPOSE
	@echo "Missing docker-compose"
	@exit 1
endif
	@true

.PHONY: launch-kubernetes
launch-kubernetes: check-docker check-docker-compose check-kubectl
	@echo " --> Creating local kubernetes cluster"
	@docker-compose down >/dev/null 2>&1|| true 
	docker-compose up -d
	@echo " --> Waiting for Kubernetes (60s) "
	@sleep 60
	@echo " --> Setting up Kubernetes"
	@$(PWD)/hack/get-kube-config.sh > ./kubeconfig.yaml
	@echo " --> Kubernetes is UP!"
	@echo "To use this cluster, export the following:"
	@echo "   export KUBCONFIG=$(pwd)/kubeconfig.yaml"