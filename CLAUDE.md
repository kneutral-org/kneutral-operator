# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

The kneutral-operator is a Kubernetes operator that integrates with the kneutral-api service to manage and orchestrate monitoring and alerting resources in a Kubernetes environment. This operator is part of the larger kneutral ecosystem which includes the API service for alert management and time-series data handling.

## Development Commands

### Initial Setup
```bash
# Initialize Go module (if not already done)
go mod init github.com/kneutral/kneutral-operator

# Install operator-sdk (if not installed)
brew install operator-sdk

# Install kubebuilder (alternative to operator-sdk)
curl -L https://github.com/kubernetes-sigs/kubebuilder/releases/latest/download/kubebuilder_linux_amd64 -o kubebuilder
chmod +x kubebuilder && sudo mv kubebuilder /usr/local/bin/
```

### Building and Running
```bash
# Build the operator binary
go build -o bin/manager main.go

# Run the operator locally against the cluster
make run

# Build and push Docker image
make docker-build docker-push IMG=<registry>/kneutral-operator:tag

# Deploy to cluster
make deploy IMG=<registry>/kneutral-operator:tag
```

### Testing
```bash
# Run unit tests
go test ./...

# Run unit tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run integration tests (requires test cluster)
make test

# Run specific test
go test -run TestControllerReconcile ./controllers/
```

### Code Quality
```bash
# Run go fmt
go fmt ./...

# Run go vet
go vet ./...

# Run golangci-lint (if configured)
golangci-lint run

# Tidy dependencies
go mod tidy
```

### Kubernetes Operations
```bash
# Install CRDs into cluster
make install

# Uninstall CRDs from cluster
make uninstall

# Generate manifests (CRDs, RBAC, etc.)
make manifests

# Generate code (deepcopy, client, etc.)
make generate
```

## Architecture

### Expected Project Structure

```
kneutral-operator/
├── api/               # Custom Resource Definitions
│   └── v1alpha1/      # API version
├── controllers/       # Reconciliation logic
├── config/           # Kubernetes manifests
│   ├── crd/          # CustomResourceDefinitions
│   ├── manager/      # Deployment manifests
│   ├── rbac/         # RBAC permissions
│   └── samples/      # Example CRs
├── hack/             # Build and setup scripts
├── internal/         # Internal packages
│   ├── client/       # kneutral-api client
│   └── utils/        # Utility functions
└── main.go           # Operator entry point
```

### Integration Points

1. **kneutral-api**: The operator should communicate with kneutral-api service for:
   - Alert management and notifications
   - Time-series metrics querying
   - WebSocket connections for real-time updates

2. **Monitoring Stack**: Integration with:
   - Mimir/Prometheus for metrics storage
   - Grafana for visualization
   - GoAlert for alert routing

### Key Design Considerations

1. **Controller Pattern**: Implement reconciliation loops that watch for changes to custom resources
2. **API Client**: Create a robust client for kneutral-api with retry logic and circuit breakers
3. **Status Management**: Properly update CR status to reflect the actual state
4. **Error Handling**: Implement proper error handling with exponential backoff for retries
5. **Observability**: Add metrics, logging, and tracing for operator operations

### Custom Resources (Expected)

Based on the kneutral ecosystem, consider implementing CRDs for:
- Alert configurations and rules
- Monitoring targets and scrape configs
- Notification channels and webhooks
- TSDB query configurations

### Environment Configuration

Expected environment variables:
- `KNEUTRAL_API_URL`: URL of the kneutral-api service
- `KNEUTRAL_API_TOKEN`: Authentication token for API access
- `NAMESPACE`: Namespace to watch (empty for all namespaces)
- `METRICS_ADDR`: Address for metrics endpoint
- `ENABLE_LEADER_ELECTION`: Enable leader election for HA