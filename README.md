# Kneutral Operator

A Kubernetes operator that manages Prometheus AlertRules by creating and maintaining PrometheusRule resources. It provides both a Kubernetes CRD interface and a REST API for managing alert configurations.

## Features

- **AlertRule CRD**: Custom Resource Definition for defining alert rules
- **PrometheusRule Generation**: Automatically creates and manages PrometheusRule resources
- **REST API**: Web API for CRUD operations on alert rules
- **ROSA Compatible**: Designed to work on Red Hat OpenShift Service on AWS
- **Helm Chart**: Easy deployment using Helm

## Architecture

The operator watches for `AlertRule` custom resources and automatically creates corresponding `PrometheusRule` resources that can be consumed by Prometheus Operator.

```
AlertRule (CRD) → Kneutral Operator → PrometheusRule → Prometheus
```

## Installation

### Prerequisites

- Kubernetes cluster (1.19+) or OpenShift/ROSA
- Prometheus Operator installed
- Helm 3 (optional, for Helm installation)

### Install with Helm

```bash
# Add the Helm repository (when published)
# helm repo add kneutral https://charts.kneutral.io
# helm repo update

# Install the operator
helm install kneutral-operator ./helm/kneutral-operator \
  --namespace kneutral-system \
  --create-namespace

# For ROSA/OpenShift
helm install kneutral-operator ./helm/kneutral-operator \
  --namespace kneutral-system \
  --create-namespace \
  --set openshift.enabled=true
```

### Install with kubectl

```bash
# Install CRDs
kubectl apply -f config/crd/

# Install RBAC
kubectl apply -f config/rbac/

# Deploy the operator
kubectl apply -f config/manager/
```

## Usage

### Creating an AlertRule

```yaml
apiVersion: monitoring.kneutral.io/v1alpha1
kind: AlertRule
metadata:
  name: example-alerts
  namespace: monitoring
spec:
  groups:
    - name: example.rules
      interval: 30s
      rules:
        - alert: HighRequestLatency
          expr: |
            http_request_duration_seconds{quantile="0.5"} > 1
          for: 10m
          labels:
            severity: warning
          annotations:
            summary: High request latency
            description: "Request latency is above 1 second (current value: {{ $value }}s)"
```

Apply the AlertRule:

```bash
kubectl apply -f alertrule.yaml
```

The operator will automatically create a corresponding PrometheusRule.

### Using the REST API

The operator exposes a REST API on port 8090 by default.

#### List all AlertRules

```bash
curl http://kneutral-operator-api.kneutral-system:8090/api/v1/alertrules
```

#### Get a specific AlertRule

```bash
curl http://kneutral-operator-api.kneutral-system:8090/api/v1/namespaces/monitoring/alertrules/example-alerts
```

#### Create an AlertRule via API

```bash
curl -X POST http://kneutral-operator-api.kneutral-system:8090/api/v1/namespaces/monitoring/alertrules \
  -H "Content-Type: application/json" \
  -d '{
    "metadata": {
      "name": "api-created-alerts"
    },
    "spec": {
      "groups": [{
        "name": "api.rules",
        "rules": [{
          "alert": "TestAlert",
          "expr": "up == 0",
          "for": "5m",
          "labels": {"severity": "critical"},
          "annotations": {"summary": "Instance is down"}
        }]
      }]
    }
  }'
```

#### Update an AlertRule

```bash
curl -X PUT http://kneutral-operator-api.kneutral-system:8090/api/v1/namespaces/monitoring/alertrules/api-created-alerts \
  -H "Content-Type: application/json" \
  -d '{"spec": {...}}'
```

#### Delete an AlertRule

```bash
curl -X DELETE http://kneutral-operator-api.kneutral-system:8090/api/v1/namespaces/monitoring/alertrules/api-created-alerts
```

## API Documentation

### Interactive Documentation
The operator provides comprehensive API documentation:

- **Interactive Swagger UI**: `docs/swagger-ui/index.html`
- **Complete API Guide**: `docs/API_USAGE.md`
- **OpenAPI Specification**: `docs/api/openapi.yaml`

### Local Documentation Server
```bash
# Serve documentation locally
make docs

# Access at: http://localhost:8080
```

### API Endpoint
Live OpenAPI spec from the running operator:
```
http://kneutral-operator-api.kneutral-system:8090/openapi/v2
```

## Configuration

### Helm Values

Key configuration options in `values.yaml`:

```yaml
operator:
  replicaCount: 1
  watchNamespace: ""  # Empty for all namespaces
  leaderElection:
    enabled: true

api:
  enabled: true
  port: 8090
  ingress:
    enabled: false

openshift:
  enabled: true  # Enable for ROSA/OpenShift
```

### Environment Variables

- `WATCH_NAMESPACE`: Namespace to watch (empty for all)
- `ENABLE_LEADER_ELECTION`: Enable leader election for HA

## Development

### Prerequisites

- Go 1.21+
- Docker
- kubectl configured with a test cluster

### Building

```bash
# Build the binary
make build

# Build Docker image
make docker-build IMG=myrepo/kneutral-operator:dev

# Push Docker image
make docker-push IMG=myrepo/kneutral-operator:dev
```

### Running Locally

```bash
# Install CRDs
make install

# Run the operator locally
make run
```

### Testing

```bash
# Run unit tests
make test

# Apply example AlertRules
make apply-example

# Test the API
make test-api

# Run interactive API demo
make test-api-demo
```

## Examples

See the `config/samples/` directory for example AlertRule configurations, including:
- `alertrule-arista-dom.yaml`: Arista DOM (Digital Optical Monitoring) alerts

## Troubleshooting

### Check operator logs

```bash
kubectl logs -n kneutral-system deployment/kneutral-operator
```

### Check AlertRule status

```bash
kubectl get alertrules -A
kubectl describe alertrule <name> -n <namespace>
```

### Verify PrometheusRule creation

```bash
kubectl get prometheusrules -A
```

## License

Apache License 2.0

## Support

- **Issues**: [GitHub Issues](https://github.com/kneutral-org/kneutral-operator/issues)
- **Discussions**: [GitHub Discussions](https://github.com/kneutral-org/kneutral-operator/discussions)
- **Documentation**: Complete API docs in `/docs` folder