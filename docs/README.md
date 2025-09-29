# Kneutral Operator Documentation

Welcome to the comprehensive documentation for the Kneutral Operator API! This operator simplifies the management of Prometheus AlertRules by providing both Kubernetes Custom Resources and a REST API interface.

## 📚 Documentation Structure

### 🌐 Interactive API Documentation
- **[Swagger UI](swagger-ui/index.html)** - Interactive API explorer with live testing capabilities
- **[OpenAPI Specification](api/openapi.yaml)** - Complete API schema in OpenAPI 3.0 format

### 📖 Usage Guides
- **[API Usage Guide](API_USAGE.md)** - Comprehensive guide with examples and best practices
- **[Quick Start Guide](#quick-start)** - Get up and running in minutes

### 🧪 Examples & Testing
- **[Example AlertRules](examples/)** - Ready-to-use AlertRule configurations
- **[API Test Script](examples/test-api.sh)** - Automated testing and demo script

## 🚀 Quick Start

### 1. Check API Health
```bash
curl http://kneutral-operator-api.kneutral-system:8090/health
```

### 2. List Existing AlertRules
```bash
curl http://kneutral-operator-api.kneutral-system:8090/api/v1/alertrules
```

### 3. Create Your First AlertRule
```bash
curl -X POST http://kneutral-operator-api.kneutral-system:8090/api/v1/namespaces/monitoring/alertrules \
  -H 'Content-Type: application/json' \
  -d '{
    "metadata": {
      "name": "my-first-alert"
    },
    "spec": {
      "groups": [{
        "name": "basic.rules",
        "rules": [{
          "alert": "ServiceDown",
          "expr": "up == 0",
          "for": "5m",
          "labels": {"severity": "critical"},
          "annotations": {
            "summary": "Service is down",
            "description": "{{ $labels.instance }} has been down for more than 5 minutes"
          }
        }]
      }]
    }
  }'
```

## 📊 API Overview

The Kneutral Operator provides RESTful endpoints for complete AlertRule lifecycle management:

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | Health check |
| `GET` | `/api/v1/alertrules` | List all AlertRules |
| `GET` | `/api/v1/namespaces/{ns}/alertrules` | List AlertRules in namespace |
| `POST` | `/api/v1/namespaces/{ns}/alertrules` | Create AlertRule |
| `GET` | `/api/v1/namespaces/{ns}/alertrules/{name}` | Get specific AlertRule |
| `PUT` | `/api/v1/namespaces/{ns}/alertrules/{name}` | Update AlertRule |
| `DELETE` | `/api/v1/namespaces/{ns}/alertrules/{name}` | Delete AlertRule |

## 🎯 Common Use Cases

### Infrastructure Monitoring
Monitor CPU, memory, disk, and network metrics across your infrastructure.

**Example:** [Basic CPU Alert](examples/basic-cpu-alert.json)

### Application Performance
Track response times, error rates, and throughput for your applications.

**Example:** [Application Performance](examples/application-performance.json)

### Network Monitoring
Monitor bandwidth usage, interface status, and network errors.

**Example:** [Network Monitoring](examples/network-monitoring.json)

### Kubernetes Cluster Health
Keep track of node status, pod health, and resource utilization.

**Example:** [Kubernetes Cluster](examples/kubernetes-cluster.json)

## 🧪 Testing & Examples

### Interactive Testing
Use our comprehensive test script to explore the API:

```bash
# Run all tests
./examples/test-api.sh

# Interactive demo
./examples/test-api.sh -i

# Verbose output
./examples/test-api.sh -v

# Custom API URL
KNEUTRAL_API_URL=http://my-cluster:8090 ./examples/test-api.sh
```

### Example AlertRules

We provide production-ready examples for common monitoring scenarios:

- **[basic-cpu-alert.json](examples/basic-cpu-alert.json)** - CPU usage monitoring
- **[network-monitoring.json](examples/network-monitoring.json)** - Network traffic and interface monitoring
- **[application-performance.json](examples/application-performance.json)** - Application performance metrics
- **[kubernetes-cluster.json](examples/kubernetes-cluster.json)** - Kubernetes cluster monitoring

## 🔧 Advanced Usage

### Bulk Operations
```bash
# Deploy multiple AlertRules from files
for file in examples/*.json; do
  name=$(basename "$file" .json)
  curl -X POST \
    "http://kneutral-operator-api.kneutral-system:8090/api/v1/namespaces/monitoring/alertrules" \
    -H 'Content-Type: application/json' \
    -d "@$file"
done
```

### CI/CD Integration
```yaml
# GitHub Actions example
- name: Deploy AlertRules
  run: |
    for alert in alerts/*.json; do
      curl -X POST \
        "${{ secrets.KNEUTRAL_API_URL }}/api/v1/namespaces/monitoring/alertrules" \
        -H 'Content-Type: application/json' \
        -d "@$alert"
    done
```

### Monitoring the Operator
```bash
# Create alerts for the operator itself
curl -X POST http://kneutral-operator-api.kneutral-system:8090/api/v1/namespaces/monitoring/alertrules \
  -H 'Content-Type: application/json' \
  -d '{
    "metadata": {"name": "operator-health"},
    "spec": {
      "groups": [{
        "name": "kneutral.operator",
        "rules": [{
          "alert": "KneutralOperatorDown",
          "expr": "up{job=\"kneutral-operator\"} == 0",
          "for": "5m",
          "labels": {"severity": "critical"},
          "annotations": {
            "summary": "Kneutral Operator is down",
            "description": "The Kneutral Operator has been unreachable for more than 5 minutes"
          }
        }]
      }]
    }
  }'
```

## 🔐 Security Considerations

### Network Security
- The API currently runs without authentication
- Ensure proper network policies restrict access
- Consider using a service mesh or API gateway for additional security

### RBAC
The operator requires specific Kubernetes permissions:
- Read/write access to `AlertRule` custom resources
- Read/write access to `PrometheusRule` resources
- Event creation permissions

### Best Practices
- Use namespace-scoped deployments when possible
- Implement resource quotas to prevent abuse
- Monitor API usage and set up alerting
- Regularly audit AlertRule configurations

## 🐛 Troubleshooting

### Common Issues

1. **AlertRule not creating PrometheusRule**
   - Check operator logs: `kubectl logs -n kneutral-system deployment/kneutral-operator`
   - Verify RBAC permissions
   - Check AlertRule status: `kubectl get alertrule -A`

2. **API returning 500 errors**
   - Verify operator is running
   - Check Kubernetes API server connectivity
   - Review operator logs for detailed error messages

3. **PromQL expressions not working**
   - Validate expressions in Prometheus UI first
   - Check for correct metric names and labels
   - Verify time ranges and functions

### Debug Commands
```bash
# Check operator status
kubectl get pods -n kneutral-system

# View operator logs
kubectl logs -n kneutral-system deployment/kneutral-operator

# Check AlertRule status
kubectl describe alertrule my-alert -n monitoring

# Verify generated PrometheusRule
kubectl get prometheusrules -n monitoring
```

## 🤝 Contributing

We welcome contributions! Please see the main README.md for contribution guidelines.

### API Documentation Updates
When updating the API:
1. Update the OpenAPI specification in `docs/api/openapi.yaml`
2. Add examples to `docs/examples/`
3. Update this documentation
4. Test with the provided test script

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/kneutral/kneutral-operator/issues)
- **Discussions**: [GitHub Discussions](https://github.com/kneutral/kneutral-operator/discussions)
- **Documentation**: This documentation site

## 📄 License

This project is licensed under the Apache License 2.0. See the LICENSE file for details.