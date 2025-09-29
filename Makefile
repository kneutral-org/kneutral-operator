# Image URL to use all building/pushing image targets
IMG ?= kneutral/kneutral-operator:latest
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true,preserveUnknownFields=false"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: fmt vet ## Run tests.
	go test ./... -coverprofile cover.out

##@ Build

.PHONY: build
build: fmt vet ## Build manager binary.
	go build -o bin/manager main.go

.PHONY: run
run: fmt vet ## Run a controller from your host.
	go run ./main.go

.PHONY: docker-build
docker-build: ## Build docker image with the manager.
	docker build -t ${IMG} .

.PHONY: docker-push
docker-push: ## Push docker image with the manager.
	docker push ${IMG}

##@ Deployment

.PHONY: install
install: ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	kubectl apply -f config/crd/

.PHONY: uninstall
uninstall: ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config.
	kubectl delete -f config/crd/

.PHONY: deploy
deploy: ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	kubectl apply -f config/crd/
	kubectl apply -f config/rbac/
	kubectl apply -f config/manager/

.PHONY: undeploy
undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config.
	kubectl delete -f config/manager/
	kubectl delete -f config/rbac/
	kubectl delete -f config/crd/

##@ Helm

.PHONY: helm-install
helm-install: ## Install the operator using Helm
	helm install kneutral-operator helm/kneutral-operator -n kneutral-system --create-namespace

.PHONY: helm-upgrade
helm-upgrade: ## Upgrade the operator using Helm
	helm upgrade kneutral-operator helm/kneutral-operator -n kneutral-system

.PHONY: helm-uninstall
helm-uninstall: ## Uninstall the operator using Helm
	helm uninstall kneutral-operator -n kneutral-system

.PHONY: helm-package
helm-package: ## Package the Helm chart
	helm package helm/kneutral-operator

##@ Examples

.PHONY: apply-example
apply-example: ## Apply example AlertRule
	kubectl apply -f config/samples/

.PHONY: delete-example
delete-example: ## Delete example AlertRule
	kubectl delete -f config/samples/

##@ Documentation

.PHONY: docs
docs: ## Serve API documentation locally
	@echo "Starting documentation server..."
	@cd docs && ./serve-docs.sh

.PHONY: test-api
test-api: ## Run API tests
	@echo "Running API tests..."
	@cd docs/examples && ./test-api.sh

.PHONY: test-api-demo
test-api-demo: ## Run interactive API demo
	@echo "Starting interactive API demo..."
	@cd docs/examples && ./test-api.sh -i

##@ Standalone Testing

.PHONY: build-standalone
build-standalone: ## Build standalone API server (no Kubernetes required)
	go build -o bin/standalone ./cmd/standalone/

.PHONY: standalone
standalone: build-standalone ## Run standalone API server with mock data
	@echo "🚀 Starting standalone API server..."
	@echo "📍 API: http://localhost:8090"
	@echo "📖 Docs: http://localhost:8090/docs"
	@echo "💡 Press Ctrl+C to stop"
	./bin/standalone --mock-data=true --api-bind-address=:8090

.PHONY: test-standalone
test-standalone: build-standalone ## Test API without Kubernetes
	@echo "🧪 Testing API in standalone mode..."
	@./bin/standalone --mock-data=true --api-bind-address=:8090 & \
	SERVER_PID=$$!; \
	sleep 3; \
	cd docs/examples && KNEUTRAL_API_URL=http://localhost:8090 ./test-api.sh; \
	kill $$SERVER_PID

.PHONY: demo-standalone
demo-standalone: build-standalone ## Run interactive demo without Kubernetes
	@echo "🎭 Starting interactive demo in standalone mode..."
	@./bin/standalone --mock-data=true --api-bind-address=:8090 & \
	SERVER_PID=$$!; \
	sleep 3; \
	cd docs/examples && KNEUTRAL_API_URL=http://localhost:8090 ./test-api.sh -i; \
	kill $$SERVER_PID