package api

// getOpenAPISpec returns the OpenAPI specification for the API
func getOpenAPISpec() map[string]interface{} {
	return map[string]interface{}{
		"swagger": "2.0",
		"info": map[string]interface{}{
			"title":       "Kneutral Operator API",
			"description": "API for managing AlertRules in Kubernetes",
			"version":     "v1alpha1",
		},
		"basePath": "/api/v1",
		"schemes":  []string{"http", "https"},
		"consumes": []string{"application/json"},
		"produces": []string{"application/json"},
		"paths": map[string]interface{}{
			"/health": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Health check",
					"description": "Check if the API server is healthy",
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "API server is healthy",
						},
					},
				},
			},
			"/alertrules": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "List all AlertRules",
					"description": "List all AlertRules across all namespaces",
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "List of AlertRules",
						},
					},
				},
			},
			"/namespaces/{namespace}/alertrules": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "List AlertRules in namespace",
					"description": "List all AlertRules in a specific namespace",
					"parameters": []map[string]interface{}{
						{
							"name":        "namespace",
							"in":          "path",
							"required":    true,
							"type":        "string",
							"description": "Namespace name",
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "List of AlertRules",
						},
					},
				},
				"post": map[string]interface{}{
					"summary":     "Create AlertRule",
					"description": "Create a new AlertRule in the namespace",
					"parameters": []map[string]interface{}{
						{
							"name":        "namespace",
							"in":          "path",
							"required":    true,
							"type":        "string",
							"description": "Namespace name",
						},
						{
							"name":        "body",
							"in":          "body",
							"required":    true,
							"description": "AlertRule object",
							"schema": map[string]interface{}{
								"$ref": "#/definitions/AlertRule",
							},
						},
					},
					"responses": map[string]interface{}{
						"201": map[string]interface{}{
							"description": "AlertRule created",
						},
						"400": map[string]interface{}{
							"description": "Invalid request",
						},
						"409": map[string]interface{}{
							"description": "AlertRule already exists",
						},
					},
				},
			},
			"/namespaces/{namespace}/alertrules/{name}": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Get AlertRule",
					"description": "Get a specific AlertRule",
					"parameters": []map[string]interface{}{
						{
							"name":        "namespace",
							"in":          "path",
							"required":    true,
							"type":        "string",
							"description": "Namespace name",
						},
						{
							"name":        "name",
							"in":          "path",
							"required":    true,
							"type":        "string",
							"description": "AlertRule name",
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "AlertRule details",
						},
						"404": map[string]interface{}{
							"description": "AlertRule not found",
						},
					},
				},
				"put": map[string]interface{}{
					"summary":     "Update AlertRule",
					"description": "Update an existing AlertRule",
					"parameters": []map[string]interface{}{
						{
							"name":        "namespace",
							"in":          "path",
							"required":    true,
							"type":        "string",
							"description": "Namespace name",
						},
						{
							"name":        "name",
							"in":          "path",
							"required":    true,
							"type":        "string",
							"description": "AlertRule name",
						},
						{
							"name":        "body",
							"in":          "body",
							"required":    true,
							"description": "Updated AlertRule object",
							"schema": map[string]interface{}{
								"$ref": "#/definitions/AlertRule",
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "AlertRule updated",
						},
						"400": map[string]interface{}{
							"description": "Invalid request",
						},
						"404": map[string]interface{}{
							"description": "AlertRule not found",
						},
					},
				},
				"delete": map[string]interface{}{
					"summary":     "Delete AlertRule",
					"description": "Delete an AlertRule",
					"parameters": []map[string]interface{}{
						{
							"name":        "namespace",
							"in":          "path",
							"required":    true,
							"type":        "string",
							"description": "Namespace name",
						},
						{
							"name":        "name",
							"in":          "path",
							"required":    true,
							"type":        "string",
							"description": "AlertRule name",
						},
					},
					"responses": map[string]interface{}{
						"204": map[string]interface{}{
							"description": "AlertRule deleted",
						},
						"404": map[string]interface{}{
							"description": "AlertRule not found",
						},
					},
				},
			},
		},
		"definitions": map[string]interface{}{
			"AlertRule": map[string]interface{}{
				"type":     "object",
				"required": []string{"metadata", "spec"},
				"properties": map[string]interface{}{
					"metadata": map[string]interface{}{
						"type":     "object",
						"required": []string{"name"},
						"properties": map[string]interface{}{
							"name": map[string]interface{}{
								"type":        "string",
								"description": "Name of the AlertRule",
							},
							"namespace": map[string]interface{}{
								"type":        "string",
								"description": "Namespace of the AlertRule",
							},
							"labels": map[string]interface{}{
								"type":        "object",
								"description": "Labels for the AlertRule",
								"additionalProperties": map[string]interface{}{
									"type": "string",
								},
							},
						},
					},
					"spec": map[string]interface{}{
						"type":     "object",
						"required": []string{"groups"},
						"properties": map[string]interface{}{
							"groups": map[string]interface{}{
								"type":        "array",
								"description": "Alert groups",
								"items": map[string]interface{}{
									"$ref": "#/definitions/AlertGroup",
								},
							},
							"labels": map[string]interface{}{
								"type":        "object",
								"description": "Labels to add to the generated PrometheusRule",
								"additionalProperties": map[string]interface{}{
									"type": "string",
								},
							},
						},
					},
				},
			},
			"AlertGroup": map[string]interface{}{
				"type":     "object",
				"required": []string{"name", "rules"},
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Name of the alert group",
					},
					"interval": map[string]interface{}{
						"type":        "string",
						"description": "How often rules in the group are evaluated",
					},
					"rules": map[string]interface{}{
						"type":        "array",
						"description": "Alert rules",
						"items": map[string]interface{}{
							"$ref": "#/definitions/Rule",
						},
					},
				},
			},
			"Rule": map[string]interface{}{
				"type":     "object",
				"required": []string{"alert", "expr"},
				"properties": map[string]interface{}{
					"alert": map[string]interface{}{
						"type":        "string",
						"description": "Alert name",
					},
					"expr": map[string]interface{}{
						"type":        "string",
						"description": "PromQL expression to evaluate",
					},
					"for": map[string]interface{}{
						"type":        "string",
						"description": "How long the alert must be pending before firing",
					},
					"labels": map[string]interface{}{
						"type":        "object",
						"description": "Labels to add or override",
						"additionalProperties": map[string]interface{}{
							"type": "string",
						},
					},
					"annotations": map[string]interface{}{
						"type":        "object",
						"description": "Annotations to add",
						"additionalProperties": map[string]interface{}{
							"type": "string",
						},
					},
				},
			},
		},
	}
}
