---
name: Bug Report
about: Create a report to help us improve
title: '[BUG] '
labels: bug
assignees: ''
---

## Bug Description
A clear and concise description of what the bug is.

## Steps to Reproduce
1. Deploy the operator with...
2. Create AlertRule with...
3. Execute API call...
4. See error

## Expected Behavior
A clear and concise description of what you expected to happen.

## Actual Behavior
What actually happened.

## Environment
- **Operator Version**: (e.g., v1.0.0)
- **Kubernetes Version**: (e.g., 1.28)
- **Platform**: (e.g., ROSA, EKS, GKE, kind)
- **Installation Method**: (e.g., Helm, kubectl)

## API Logs (if applicable)
```
Paste API server logs here
```

## Operator Logs (if applicable)
```
kubectl logs -n kneutral-system deployment/kneutral-operator
```

## AlertRule Configuration (if applicable)
```yaml
# Paste your AlertRule YAML here
```

## Additional Context
Add any other context about the problem here.