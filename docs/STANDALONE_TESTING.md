# Standalone Testing Guide

This guide shows you how to test the Kneutral Operator API without needing a full Kubernetes cluster.

## 🚀 Quick Start (No Kubernetes Required)

### Option 1: Standalone Mode (Recommended)

**Build and run the standalone server:**

```bash
# Build the standalone server
go build -o bin/standalone ./cmd/standalone/

# Run with mock data
./bin/standalone --mock-data=true --api-bind-address=:8090
```

The server will start with pre-loaded example data and provide:
- 🌐 **API Documentation**: http://localhost:8090/docs
- 📊 **Health Check**: http://localhost:8090/health
- 🔍 **API Endpoints**: http://localhost:8090/api/v1/alertrules

### Option 2: Docker Container

**Build and run in Docker:**

```bash
# Build Docker image for standalone mode
docker build -f Dockerfile.standalone -t kneutral-operator-standalone .

# Run container
docker run -p 8090:8090 kneutral-operator-standalone
```

### Option 3: Pre-built Testing

**Use the test script with standalone mode:**

```bash
# Start standalone server in background
./bin/standalone &
SERVER_PID=$!

# Wait for server to start
sleep 2

# Run API tests
cd docs/examples && ./test-api.sh

# Stop server
kill $SERVER_PID
```

## 🧪 Testing Scenarios

### 1. Basic API Operations

```bash
# Health check
curl http://localhost:8090/health

# List all AlertRules (shows pre-loaded examples)
curl http://localhost:8090/api/v1/alertrules

# Get specific AlertRule
curl http://localhost:8090/api/v1/namespaces/monitoring/alertrules/cpu-monitoring
```

### 2. Create New AlertRule

```bash
curl -X POST http://localhost:8090/api/v1/namespaces/test/alertrules \
  -H 'Content-Type: application/json' \
  -d '{
    "metadata": {
      "name": "my-test-alert"
    },
    "spec": {
      "groups": [{
        "name": "test.rules",
        "rules": [{
          "alert": "TestServiceDown",
          "expr": "up{job=\"my-service\"} == 0",
          "for": "5m",
          "labels": {
            "severity": "critical",
            "team": "backend"
          },
          "annotations": {
            "summary": "My service is down",
            "description": "Service {{ $labels.job }} has been down for more than 5 minutes"
          }
        }]
      }]
    }
  }'
```

### 3. Update AlertRule

```bash
curl -X PUT http://localhost:8090/api/v1/namespaces/test/alertrules/my-test-alert \
  -H 'Content-Type: application/json' \
  -d '{
    "groups": [{
      "name": "test.rules.updated",
      "interval": "60s",
      "rules": [{
        "alert": "TestServiceDownUpdated",
        "expr": "up{job=\"my-service\"} == 0",
        "for": "10m",
        "labels": {
          "severity": "warning",
          "team": "backend",
          "updated": "true"
        },
        "annotations": {
          "summary": "My service is down (updated)",
          "description": "Service has been down for more than 10 minutes"
        }
      }]
    }]
  }'
```

### 4. Delete AlertRule

```bash
curl -X DELETE http://localhost:8090/api/v1/namespaces/test/alertrules/my-test-alert
```

## 🔧 Advanced Testing

### Load Testing with Multiple Requests

```bash
#!/bin/bash
# Create multiple AlertRules for load testing

for i in {1..10}; do
  curl -X POST http://localhost:8090/api/v1/namespaces/load-test/alertrules \
    -H 'Content-Type: application/json' \
    -d "{
      \"metadata\": {
        \"name\": \"load-test-alert-$i\"
      },
      \"spec\": {
        \"groups\": [{
          \"name\": \"load-test.rules\",
          \"rules\": [{
            \"alert\": \"LoadTestAlert$i\",
            \"expr\": \"test_metric > $i\",
            \"labels\": {\"severity\": \"info\"},
            \"annotations\": {\"summary\": \"Load test alert $i\"}
          }]
        }]
      }
    }"
  echo "Created alert $i"
done
```

### Error Handling Testing

```bash
# Test 404 error
curl -v http://localhost:8090/api/v1/namespaces/test/alertrules/non-existent

# Test 400 error (invalid data)
curl -X POST http://localhost:8090/api/v1/namespaces/test/alertrules \
  -H 'Content-Type: application/json' \
  -d '{"metadata": {"name": "invalid"}, "spec": {}}'

# Test 409 error (already exists)
curl -X POST http://localhost:8090/api/v1/namespaces/monitoring/alertrules \
  -H 'Content-Type: application/json' \
  -d '{"metadata": {"name": "cpu-monitoring"}, "spec": {"groups": []}}'
```

## 📊 Pre-loaded Example Data

The standalone mode includes these example AlertRules:

### 1. CPU Monitoring (`monitoring/cpu-monitoring`)
- High CPU usage warning (>80%)
- Critical CPU usage alert (>95%)

### 2. Application Performance (`production/app-performance`)
- High response time alerts
- Error rate monitoring

### 3. Network Monitoring (`network/arista-dom-monitoring`)
- DOM RX power monitoring (based on your Arista example)
- Optical interface health checks

## 🐛 Debugging

### Check Server Logs
The standalone server outputs detailed logs:

```bash
./bin/standalone --mock-data=true 2>&1 | tee server.log
```

### Validate JSON Payloads
Use `jq` to validate your JSON:

```bash
echo '{...}' | jq .
```

### Monitor API Calls
Use verbose curl for debugging:

```bash
curl -v -X GET http://localhost:8090/api/v1/alertrules
```

## 🚢 Alternative Testing Methods

### Option 1: Kind (Kubernetes in Docker)

If you want to test with a real Kubernetes cluster:

```bash
# Install kind
go install sigs.k8s.io/kind@latest

# Create cluster
kind create cluster --name kneutral-test

# Deploy the operator
make helm-install

# Test against real cluster
KNEUTRAL_API_URL=http://localhost:8090 ./docs/examples/test-api.sh
```

### Option 2: Minikube

```bash
# Start minikube
minikube start

# Deploy operator
helm install kneutral-operator ./helm/kneutral-operator

# Forward port
kubectl port-forward svc/kneutral-operator-api 8090:8090

# Test
./docs/examples/test-api.sh
```

### Option 3: Docker Compose (Future Enhancement)

```yaml
# docker-compose.test.yml
version: '3.8'
services:
  kneutral-api:
    build:
      context: .
      dockerfile: Dockerfile.standalone
    ports:
      - "8090:8090"
    environment:
      - MOCK_DATA=true
```

## 📝 Creating Custom Test Data

You can modify `internal/mock/data.go` to add your own test AlertRules:

```go
func PopulateCustomData(client *MockClient) {
    customAlert := &monitoringv1alpha1.AlertRule{
        // ... your custom AlertRule definition
    }
    client.Create(context.Background(), customAlert)
}
```

## 🎯 Makefile Integration

Add these convenient targets to test without Kubernetes:

```bash
# Start standalone server
make standalone

# Test API without Kubernetes
make test-standalone

# Run interactive demo
make demo-standalone
```

## ✅ What Gets Tested

The standalone mode validates:

- ✅ **API Endpoints**: All CRUD operations
- ✅ **Request Validation**: JSON schema validation
- ✅ **Response Formats**: Correct JSON responses
- ✅ **Error Handling**: 400, 404, 409, 500 errors
- ✅ **CORS Headers**: Cross-origin requests
- ✅ **Content Types**: JSON content handling
- ✅ **Status Codes**: HTTP status code compliance

## ❌ What's NOT Tested

- ❌ **Kubernetes Integration**: No real K8s resources
- ❌ **RBAC**: No Kubernetes permissions
- ❌ **PrometheusRule Creation**: No actual Prometheus integration
- ❌ **Webhooks**: No admission controllers
- ❌ **Network Policies**: No K8s networking

For full integration testing, use a real Kubernetes cluster with the complete operator deployment.