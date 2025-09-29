# Kneutral Operator API Usage Guide

This guide provides comprehensive examples for using the Kneutral Operator REST API to manage AlertRule resources.

## Table of Contents
- [Base URL and Authentication](#base-url-and-authentication)
- [Health Check](#health-check)
- [AlertRule Management](#alertrule-management)
- [Common Use Cases](#common-use-cases)
- [Error Handling](#error-handling)
- [Best Practices](#best-practices)

## Base URL and Authentication

### Base URL
```
http://kneutral-operator-api.kneutral-system:8090
```

For local development:
```
http://localhost:8090
```

### Authentication
Currently, the API does not require authentication. Ensure your network policies and RBAC are properly configured to secure access.

## Health Check

### Check API Health
```bash
curl -X GET http://kneutral-operator-api.kneutral-system:8090/health
```

**Response:**
```json
{
  "status": "healthy"
}
```

## AlertRule Management

### 1. List All AlertRules (Cluster-wide)

```bash
curl -X GET \
  http://kneutral-operator-api.kneutral-system:8090/api/v1/alertrules \
  -H 'Content-Type: application/json'
```

**Response:**
```json
{
  "apiVersion": "monitoring.kneutral.io/v1alpha1",
  "kind": "AlertRuleList",
  "items": [
    {
      "metadata": {
        "name": "example-alerts",
        "namespace": "monitoring",
        "creationTimestamp": "2023-12-01T10:00:00Z"
      },
      "spec": {
        "groups": [...]
      },
      "status": {
        "state": "Active",
        "prometheusRuleName": "kneutral-example-alerts"
      }
    }
  ]
}
```

### 2. List AlertRules in a Namespace

```bash
curl -X GET \
  http://kneutral-operator-api.kneutral-system:8090/api/v1/namespaces/monitoring/alertrules \
  -H 'Content-Type: application/json'
```

### 3. Get a Specific AlertRule

```bash
curl -X GET \
  http://kneutral-operator-api.kneutral-system:8090/api/v1/namespaces/monitoring/alertrules/example-alerts \
  -H 'Content-Type: application/json'
```

**Response:**
```json
{
  "apiVersion": "monitoring.kneutral.io/v1alpha1",
  "kind": "AlertRule",
  "metadata": {
    "name": "example-alerts",
    "namespace": "monitoring",
    "labels": {
      "app": "kneutral"
    }
  },
  "spec": {
    "groups": [
      {
        "name": "example.rules",
        "interval": "30s",
        "rules": [
          {
            "alert": "HighCPUUsage",
            "expr": "cpu_usage_percent > 80",
            "for": "5m",
            "labels": {
              "severity": "warning"
            },
            "annotations": {
              "summary": "High CPU usage detected",
              "description": "CPU usage is {{ $value }}% on {{ $labels.instance }}"
            }
          }
        ]
      }
    ]
  },
  "status": {
    "state": "Active",
    "prometheusRuleName": "kneutral-example-alerts",
    "lastReconcileTime": "2023-12-01T10:05:00Z",
    "conditions": [
      {
        "type": "Ready",
        "status": "True",
        "reason": "ReconcileSuccess",
        "message": "PrometheusRule kneutral-example-alerts created/updated successfully"
      }
    ]
  }
}
```

### 4. Create a New AlertRule

```bash
curl -X POST \
  http://kneutral-operator-api.kneutral-system:8090/api/v1/namespaces/monitoring/alertrules \
  -H 'Content-Type: application/json' \
  -d '{
    "metadata": {
      "name": "cpu-alerts"
    },
    "spec": {
      "groups": [
        {
          "name": "cpu.rules",
          "interval": "30s",
          "rules": [
            {
              "alert": "HighCPUUsage",
              "expr": "cpu_usage_percent > 80",
              "for": "5m",
              "labels": {
                "severity": "warning",
                "team": "infrastructure"
              },
              "annotations": {
                "summary": "High CPU usage on {{ $labels.instance }}",
                "description": "CPU usage is {{ $value }}% for more than 5 minutes"
              }
            },
            {
              "alert": "CriticalCPUUsage",
              "expr": "cpu_usage_percent > 95",
              "for": "2m",
              "labels": {
                "severity": "critical",
                "team": "infrastructure"
              },
              "annotations": {
                "summary": "Critical CPU usage on {{ $labels.instance }}",
                "description": "CPU usage is {{ $value }}% for more than 2 minutes"
              }
            }
          ]
        }
      ]
    }
  }'
```

**Response (201 Created):**
```json
{
  "apiVersion": "monitoring.kneutral.io/v1alpha1",
  "kind": "AlertRule",
  "metadata": {
    "name": "cpu-alerts",
    "namespace": "monitoring",
    "creationTimestamp": "2023-12-01T10:10:00Z"
  },
  "spec": {
    "groups": [...]
  }
}
```

### 5. Update an AlertRule

```bash
curl -X PUT \
  http://kneutral-operator-api.kneutral-system:8090/api/v1/namespaces/monitoring/alertrules/cpu-alerts \
  -H 'Content-Type: application/json' \
  -d '{
    "groups": [
      {
        "name": "cpu.rules",
        "interval": "60s",
        "rules": [
          {
            "alert": "HighCPUUsage",
            "expr": "cpu_usage_percent > 85",
            "for": "10m",
            "labels": {
              "severity": "warning",
              "team": "infrastructure"
            },
            "annotations": {
              "summary": "High CPU usage on {{ $labels.instance }}",
              "description": "CPU usage is {{ $value }}% for more than 10 minutes",
              "runbook_url": "https://runbooks.company.com/cpu-high"
            }
          }
        ]
      }
    ]
  }'
```

### 6. Delete an AlertRule

```bash
curl -X DELETE \
  http://kneutral-operator-api.kneutral-system:8090/api/v1/namespaces/monitoring/alertrules/cpu-alerts
```

**Response: 204 No Content**

## Common Use Cases

### Network Monitoring Alerts

```bash
curl -X POST \
  http://kneutral-operator-api.kneutral-system:8090/api/v1/namespaces/monitoring/alertrules \
  -H 'Content-Type: application/json' \
  -d '{
    "metadata": {
      "name": "network-alerts",
      "labels": {
        "category": "network",
        "team": "network-ops"
      }
    },
    "spec": {
      "groups": [
        {
          "name": "network.bandwidth",
          "interval": "30s",
          "rules": [
            {
              "alert": "HighBandwidthUsage",
              "expr": "rate(network_bytes_total[5m]) > 1000000000",
              "for": "2m",
              "labels": {
                "severity": "warning"
              },
              "annotations": {
                "summary": "High bandwidth usage on {{ $labels.interface }}",
                "description": "Bandwidth usage is {{ $value | humanize }}B/s"
              }
            },
            {
              "alert": "InterfaceDown",
              "expr": "network_interface_up == 0",
              "for": "1m",
              "labels": {
                "severity": "critical"
              },
              "annotations": {
                "summary": "Network interface {{ $labels.interface }} is down",
                "description": "Interface {{ $labels.interface }} on {{ $labels.instance }} has been down for more than 1 minute"
              }
            }
          ]
        }
      ]
    }
  }'
```

### Application Performance Monitoring

```bash
curl -X POST \
  http://kneutral-operator-api.kneutral-system:8090/api/v1/namespaces/monitoring/alertrules \
  -H 'Content-Type: application/json' \
  -d '{
    "metadata": {
      "name": "app-performance-alerts"
    },
    "spec": {
      "groups": [
        {
          "name": "app.performance",
          "interval": "15s",
          "rules": [
            {
              "alert": "HighResponseTime",
              "expr": "http_request_duration_seconds{quantile=\"0.95\"} > 2",
              "for": "5m",
              "labels": {
                "severity": "warning",
                "service": "{{ $labels.service }}"
              },
              "annotations": {
                "summary": "High response time for {{ $labels.service }}",
                "description": "95th percentile response time is {{ $value }}s",
                "grafana_url": "https://grafana.company.com/d/app-dashboard"
              }
            },
            {
              "alert": "HighErrorRate",
              "expr": "rate(http_requests_total{status=~\"5..\"}[5m]) / rate(http_requests_total[5m]) > 0.1",
              "for": "3m",
              "labels": {
                "severity": "critical",
                "service": "{{ $labels.service }}"
              },
              "annotations": {
                "summary": "High error rate for {{ $labels.service }}",
                "description": "Error rate is {{ $value | humanizePercentage }}"
              }
            }
          ]
        }
      ]
    }
  }'
```

### Infrastructure Monitoring (Based on your Arista DOM example)

```bash
curl -X POST \
  http://kneutral-operator-api.kneutral-system:8090/api/v1/namespaces/monitoring/alertrules \
  -H 'Content-Type: application/json' \
  -d '{
    "metadata": {
      "name": "arista-dom-alerts",
      "labels": {
        "vendor": "arista",
        "category": "optical"
      }
    },
    "spec": {
      "groups": [
        {
          "name": "kneutral.arista.dom",
          "rules": [
            {
              "alert": "LowDOMRXPowerCritical",
              "expr": "(\n  10 * log10(arista_smnp_entSensorValue{entPhysicalDescr=~\"DOM RX Power.*\"} / 1000)\n  < on(desc, entPhysicalDescr) group_left\n  10 * log10(arista_smnp_aristaSensorThresholdLowCritical{entPhysicalDescr=~\"DOM RX Power.*\"} / 1000)\n)\nand\n(\n  10 * log10(arista_smnp_entSensorValue{entPhysicalDescr=~\"DOM RX Power.*\"} / 1000) != -30\n)",
              "for": "5m",
              "labels": {
                "severity": "critical",
                "source": "kneutral"
              },
              "annotations": {
                "summary": "Critical: Low DOM RX Power on {{ $labels.entPhysicalDescr }} at {{ $labels.desc }}",
                "description": "DOM RX Power is below low critical threshold\\nDevice: {{ $labels.desc }}\\nSite: {{ $labels.site }}\\nRole: {{ $labels.role }}\\nLocation: {{ $labels.location }}\\nInterface: {{ $labels.entPhysicalDescr }}\\nCurrent Power: {{ $value | printf \"%.2f\" }} dBm",
                "grafanaUrl": "https://mon.monitor.driveuc.com/d/arista-interfaces/arista-network-interfaces"
              }
            }
          ]
        }
      ]
    }
  }'
```

## Error Handling

### Common Error Responses

#### 400 Bad Request
```json
{
  "error": "Invalid request body: missing required field 'groups'"
}
```

#### 404 Not Found
```json
{
  "error": "AlertRule not found"
}
```

#### 409 Conflict
```json
{
  "error": "AlertRule already exists"
}
```

#### 500 Internal Server Error
```json
{
  "error": "Failed to create AlertRule",
  "details": "admission webhook validation failed"
}
```

## Best Practices

### 1. Naming Conventions
- Use kebab-case for AlertRule names: `cpu-alerts`, `network-monitoring`
- Use descriptive alert names: `HighCPUUsage`, `DatabaseConnectionLost`
- Group related alerts together: `app.performance`, `infra.disk`

### 2. Label Management
```json
{
  "labels": {
    "severity": "warning|critical",
    "team": "infrastructure|application|network",
    "service": "api|database|cache",
    "environment": "production|staging|development"
  }
}
```

### 3. Annotation Best Practices
```json
{
  "annotations": {
    "summary": "Brief description with template variables",
    "description": "Detailed description with context and values",
    "runbook_url": "Link to troubleshooting documentation",
    "dashboard_url": "Link to relevant monitoring dashboard"
  }
}
```

### 4. Expression Guidelines
- Use appropriate time ranges: `[5m]` for short-term, `[1h]` for trends
- Include meaningful thresholds based on SLAs
- Test expressions in Prometheus before deploying
- Use `rate()` for counters, `increase()` for discrete events

### 5. Timing Configuration
- `for`: How long condition must be true before alerting
  - `1m` for critical infrastructure issues
  - `5m` for performance degradation
  - `15m` for capacity planning alerts
- `interval`: How often to evaluate rules
  - `15s-30s` for critical alerts
  - `1m-5m` for standard monitoring

### 6. Error Recovery
Always verify your AlertRule was processed correctly:

```bash
# Check status after creation/update
curl -X GET \
  http://kneutral-operator-api.kneutral-system:8090/api/v1/namespaces/monitoring/alertrules/my-alerts

# Look for status.state = "Active"
# Check that status.prometheusRuleName is populated
```

### 7. Bulk Operations
For multiple AlertRules, consider scripting:

```bash
#!/bin/bash
NAMESPACE="monitoring"
API_BASE="http://kneutral-operator-api.kneutral-system:8090/api/v1"

# Create multiple alert rules
for rule_file in alerts/*.json; do
  rule_name=$(basename "$rule_file" .json)
  echo "Creating AlertRule: $rule_name"

  curl -X POST \
    "$API_BASE/namespaces/$NAMESPACE/alertrules" \
    -H 'Content-Type: application/json' \
    -d "@$rule_file" \
    --fail --silent --show-error

  if [ $? -eq 0 ]; then
    echo "✓ Successfully created $rule_name"
  else
    echo "✗ Failed to create $rule_name"
  fi
done
```

## Integration Examples

### CI/CD Pipeline Integration

```yaml
# .github/workflows/deploy-alerts.yml
name: Deploy AlertRules
on:
  push:
    paths:
      - 'alerts/**'
    branches:
      - main

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Deploy AlertRules
        run: |
          for alert in alerts/*.json; do
            curl -X POST \
              "${{ secrets.KNEUTRAL_API_URL }}/api/v1/namespaces/monitoring/alertrules" \
              -H 'Content-Type: application/json' \
              -d "@$alert"
          done
```

### Monitoring and Alerting on AlertRules

You can also monitor the operator itself:

```json
{
  "metadata": {
    "name": "operator-monitoring"
  },
  "spec": {
    "groups": [
      {
        "name": "kneutral.operator",
        "rules": [
          {
            "alert": "KneutralOperatorDown",
            "expr": "up{job=\"kneutral-operator\"} == 0",
            "for": "5m",
            "labels": {
              "severity": "critical"
            },
            "annotations": {
              "summary": "Kneutral Operator is down",
              "description": "The Kneutral Operator has been down for more than 5 minutes"
            }
          }
        ]
      }
    ]
  }
}
```