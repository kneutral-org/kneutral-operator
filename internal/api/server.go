package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	monitoringv1alpha1 "github.com/kneutral-org/kneutral-operator/api/v1alpha1"
)

// Server represents the API server
type Server struct {
	client  client.Client
	address string
	log     logr.Logger
}

// NewServer creates a new API server
func NewServer(client client.Client, address string) *Server {
	return &Server{
		client:  client,
		address: address,
		log:     ctrl.Log.WithName("api-server"),
	}
}

// Start starts the API server
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Health endpoint
	mux.HandleFunc("/health", s.handleHealth)

	// AlertRule CRUD endpoints
	mux.HandleFunc("/api/v1/alertrules", s.handleAlertRules)
	mux.HandleFunc("/api/v1/namespaces/", s.handleNamespacedAlertRules)

	// Serve OpenAPI spec
	mux.HandleFunc("/openapi/v2", s.handleOpenAPISpec)

	// Serve documentation (for standalone mode)
	mux.HandleFunc("/docs", s.handleDocs)
	mux.HandleFunc("/docs/", s.handleDocs)

	s.log.Info("API server listening", "address", s.address)
	return http.ListenAndServe(s.address, s.corsMiddleware(mux))
}

// corsMiddleware adds CORS headers
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// handleHealth handles health check requests
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

// handleAlertRules handles AlertRule operations (list all)
func (s *Server) handleAlertRules(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listAlertRules(w, r, "")
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleNamespacedAlertRules handles namespaced AlertRule operations
func (s *Server) handleNamespacedAlertRules(w http.ResponseWriter, r *http.Request) {
	// Parse namespace and name from URL
	// Expected format: /api/v1/namespaces/{namespace}/alertrules/{name}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/namespaces/"), "/")
	if len(parts) < 2 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	namespace := parts[0]

	// Check if this is a collection operation or a specific resource
	if len(parts) == 2 && parts[1] == "alertrules" {
		// Collection operations
		switch r.Method {
		case http.MethodGet:
			s.listAlertRules(w, r, namespace)
		case http.MethodPost:
			s.createAlertRule(w, r, namespace)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	} else if len(parts) == 3 && parts[1] == "alertrules" {
		// Specific resource operations
		name := parts[2]
		switch r.Method {
		case http.MethodGet:
			s.getAlertRule(w, r, namespace, name)
		case http.MethodPut:
			s.updateAlertRule(w, r, namespace, name)
		case http.MethodDelete:
			s.deleteAlertRule(w, r, namespace, name)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	} else {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
	}
}

// listAlertRules lists all AlertRules
func (s *Server) listAlertRules(w http.ResponseWriter, r *http.Request, namespace string) {
	ctx := context.Background()
	alertRuleList := &monitoringv1alpha1.AlertRuleList{}

	opts := []client.ListOption{}
	if namespace != "" {
		opts = append(opts, client.InNamespace(namespace))
	}

	if err := s.client.List(ctx, alertRuleList, opts...); err != nil {
		s.log.Error(err, "Failed to list AlertRules")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(alertRuleList); err != nil {
		s.log.Error(err, "Failed to encode response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// getAlertRule gets a specific AlertRule
func (s *Server) getAlertRule(w http.ResponseWriter, r *http.Request, namespace, name string) {
	ctx := context.Background()
	alertRule := &monitoringv1alpha1.AlertRule{}

	if err := s.client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, alertRule); err != nil {
		if errors.IsNotFound(err) {
			http.Error(w, "AlertRule not found", http.StatusNotFound)
			return
		}
		s.log.Error(err, "Failed to get AlertRule")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(alertRule); err != nil {
		s.log.Error(err, "Failed to encode response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// createAlertRule creates a new AlertRule
func (s *Server) createAlertRule(w http.ResponseWriter, r *http.Request, namespace string) {
	ctx := context.Background()

	var alertRule monitoringv1alpha1.AlertRule
	if err := json.NewDecoder(r.Body).Decode(&alertRule); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Set namespace from URL
	alertRule.Namespace = namespace

	// Set TypeMeta if not provided
	if alertRule.TypeMeta.Kind == "" {
		alertRule.TypeMeta = metav1.TypeMeta{
			APIVersion: "monitoring.kneutral.io/v1alpha1",
			Kind:       "AlertRule",
		}
	}

	// Validate required fields
	if alertRule.Name == "" {
		http.Error(w, "AlertRule name is required", http.StatusBadRequest)
		return
	}

	if len(alertRule.Spec.Groups) == 0 {
		http.Error(w, "At least one alert group is required", http.StatusBadRequest)
		return
	}

	if err := s.client.Create(ctx, &alertRule); err != nil {
		if errors.IsAlreadyExists(err) {
			http.Error(w, "AlertRule already exists", http.StatusConflict)
			return
		}
		s.log.Error(err, "Failed to create AlertRule")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(&alertRule); err != nil {
		s.log.Error(err, "Failed to encode response")
	}
}

// updateAlertRule updates an existing AlertRule
func (s *Server) updateAlertRule(w http.ResponseWriter, r *http.Request, namespace, name string) {
	ctx := context.Background()

	// Get existing AlertRule
	existing := &monitoringv1alpha1.AlertRule{}
	if err := s.client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, existing); err != nil {
		if errors.IsNotFound(err) {
			http.Error(w, "AlertRule not found", http.StatusNotFound)
			return
		}
		s.log.Error(err, "Failed to get AlertRule")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Decode update from request body
	var update monitoringv1alpha1.AlertRule
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Update the spec
	existing.Spec = update.Spec

	if err := s.client.Update(ctx, existing); err != nil {
		s.log.Error(err, "Failed to update AlertRule")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(existing); err != nil {
		s.log.Error(err, "Failed to encode response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// deleteAlertRule deletes an AlertRule
func (s *Server) deleteAlertRule(w http.ResponseWriter, r *http.Request, namespace, name string) {
	ctx := context.Background()

	alertRule := &monitoringv1alpha1.AlertRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	if err := s.client.Delete(ctx, alertRule); err != nil {
		if errors.IsNotFound(err) {
			http.Error(w, "AlertRule not found", http.StatusNotFound)
			return
		}
		s.log.Error(err, "Failed to delete AlertRule")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleOpenAPISpec serves the OpenAPI specification
func (s *Server) handleOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	spec := getOpenAPISpec()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(spec); err != nil {
		s.log.Error(err, "Failed to encode OpenAPI spec")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// handleDocs serves basic API documentation
func (s *Server) handleDocs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Kneutral Operator API Documentation</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; line-height: 1.6; }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 20px; border-radius: 8px; margin-bottom: 20px; }
        .endpoint { background: #f8f9fa; padding: 15px; margin: 10px 0; border-radius: 5px; border-left: 4px solid #007bff; }
        .method { font-weight: bold; color: #007bff; }
        pre { background: #f1f1f1; padding: 10px; border-radius: 4px; overflow-x: auto; }
        .example { background: #e8f5e8; padding: 10px; border-radius: 4px; margin: 10px 0; }
    </style>
</head>
<body>
    <div class="header">
        <h1>🚀 Kneutral Operator API</h1>
        <p>REST API for managing Prometheus AlertRules</p>
    </div>

    <h2>Quick Start</h2>
    <div class="example">
        <h3>1. Health Check</h3>
        <pre>curl ` + s.address + `/health</pre>

        <h3>2. List AlertRules</h3>
        <pre>curl ` + s.address + `/api/v1/alertrules</pre>

        <h3>3. Create AlertRule</h3>
        <pre>curl -X POST ` + s.address + `/api/v1/namespaces/monitoring/alertrules \
  -H 'Content-Type: application/json' \
  -d '{
    "metadata": {"name": "test-alert"},
    "spec": {
      "groups": [{
        "name": "test.rules",
        "rules": [{
          "alert": "TestAlert",
          "expr": "up == 0",
          "labels": {"severity": "warning"},
          "annotations": {"summary": "Test alert"}
        }]
      }]
    }
  }'</pre>
    </div>

    <h2>API Endpoints</h2>

    <div class="endpoint">
        <span class="method">GET</span> /health<br>
        <small>Check API health status</small>
    </div>

    <div class="endpoint">
        <span class="method">GET</span> /api/v1/alertrules<br>
        <small>List all AlertRules across all namespaces</small>
    </div>

    <div class="endpoint">
        <span class="method">GET</span> /api/v1/namespaces/{namespace}/alertrules<br>
        <small>List AlertRules in a specific namespace</small>
    </div>

    <div class="endpoint">
        <span class="method">POST</span> /api/v1/namespaces/{namespace}/alertrules<br>
        <small>Create a new AlertRule</small>
    </div>

    <div class="endpoint">
        <span class="method">GET</span> /api/v1/namespaces/{namespace}/alertrules/{name}<br>
        <small>Get a specific AlertRule</small>
    </div>

    <div class="endpoint">
        <span class="method">PUT</span> /api/v1/namespaces/{namespace}/alertrules/{name}<br>
        <small>Update an existing AlertRule</small>
    </div>

    <div class="endpoint">
        <span class="method">DELETE</span> /api/v1/namespaces/{namespace}/alertrules/{name}<br>
        <small>Delete an AlertRule</small>
    </div>

    <h2>OpenAPI Specification</h2>
    <p><a href="/openapi/v2">View OpenAPI JSON</a></p>

    <h2>Examples</h2>
    <p>The API is pre-loaded with example data for testing. Try the endpoints above!</p>
</body>
</html>`
	w.Write([]byte(html))
}