package mock

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	monitoringv1alpha1 "github.com/kneutral-org/kneutral-operator/api/v1alpha1"
)

// MockClient implements the controller-runtime client.Client interface for testing
type MockClient struct {
	objects map[string]runtime.Object
	mutex   sync.RWMutex
}

// NewMockClient creates a new mock client
func NewMockClient() *MockClient {
	return &MockClient{
		objects: make(map[string]runtime.Object),
	}
}

// objectKey generates a unique key for storing objects
func (m *MockClient) objectKey(obj runtime.Object) (string, error) {
	switch v := obj.(type) {
	case *monitoringv1alpha1.AlertRule:
		return fmt.Sprintf("alertrule/%s/%s", v.Namespace, v.Name), nil
	case *monitoringv1alpha1.AlertRuleList:
		return "alertrulelist", nil
	default:
		return "", fmt.Errorf("unsupported object type: %T", obj)
	}
}

// Get retrieves an object
func (m *MockClient) Get(ctx context.Context, key types.NamespacedName, obj client.Object, opts ...client.GetOption) error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	objKey := fmt.Sprintf("alertrule/%s/%s", key.Namespace, key.Name)
	stored, exists := m.objects[objKey]
	if !exists {
		return errors.NewNotFound(schema.GroupResource{
			Group:    "monitoring.kneutral.io",
			Resource: "alertrules",
		}, key.Name)
	}

	// Copy the stored object to the provided object
	switch v := obj.(type) {
	case *monitoringv1alpha1.AlertRule:
		if alertRule, ok := stored.(*monitoringv1alpha1.AlertRule); ok {
			*v = *alertRule.DeepCopy()
			return nil
		}
	}

	return fmt.Errorf("type mismatch")
}

// List retrieves a list of objects
func (m *MockClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	switch v := list.(type) {
	case *monitoringv1alpha1.AlertRuleList:
		v.Items = []monitoringv1alpha1.AlertRule{}

		// Apply namespace filter if specified
		var namespaceFilter string
		for _, opt := range opts {
			if nsOpt, ok := opt.(client.InNamespace); ok {
				namespaceFilter = string(nsOpt)
				break
			}
		}

		for key, obj := range m.objects {
			if strings.HasPrefix(key, "alertrule/") {
				if alertRule, ok := obj.(*monitoringv1alpha1.AlertRule); ok {
					// Apply namespace filter
					if namespaceFilter == "" || alertRule.Namespace == namespaceFilter {
						v.Items = append(v.Items, *alertRule.DeepCopy())
					}
				}
			}
		}

		v.TypeMeta = metav1.TypeMeta{
			APIVersion: "monitoring.kneutral.io/v1alpha1",
			Kind:       "AlertRuleList",
		}
		return nil
	}

	return fmt.Errorf("unsupported list type: %T", list)
}

// Create creates a new object
func (m *MockClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	key, err := m.objectKey(obj)
	if err != nil {
		return err
	}

	// Check if object already exists
	if _, exists := m.objects[key]; exists {
		return errors.NewAlreadyExists(schema.GroupResource{
			Group:    "monitoring.kneutral.io",
			Resource: "alertrules",
		}, obj.GetName())
	}

	// Set metadata
	obj.SetCreationTimestamp(metav1.NewTime(time.Now()))
	obj.SetUID(types.UID(fmt.Sprintf("mock-uid-%d", time.Now().UnixNano())))
	obj.SetGeneration(1)

	// Set status for AlertRule
	if alertRule, ok := obj.(*monitoringv1alpha1.AlertRule); ok {
		alertRule.Status = monitoringv1alpha1.AlertRuleStatus{
			State:              "Active",
			PrometheusRuleName: fmt.Sprintf("kneutral-%s", alertRule.Name),
			LastReconcileTime:  &metav1.Time{Time: time.Now()},
			Conditions: []metav1.Condition{
				{
					Type:               "Ready",
					Status:             metav1.ConditionTrue,
					ObservedGeneration: 1,
					LastTransitionTime: metav1.NewTime(time.Now()),
					Reason:             "MockReconcileSuccess",
					Message:            fmt.Sprintf("Mock PrometheusRule %s created successfully", fmt.Sprintf("kneutral-%s", alertRule.Name)),
				},
			},
		}
	}

	// Store a deep copy
	m.objects[key] = obj.DeepCopyObject()
	return nil
}

// Update updates an existing object
func (m *MockClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	key, err := m.objectKey(obj)
	if err != nil {
		return err
	}

	// Check if object exists
	existing, exists := m.objects[key]
	if !exists {
		return errors.NewNotFound(schema.GroupResource{
			Group:    "monitoring.kneutral.io",
			Resource: "alertrules",
		}, obj.GetName())
	}

	// Update generation
	if existingObj, ok := existing.(client.Object); ok {
		obj.SetGeneration(existingObj.GetGeneration() + 1)
		obj.SetCreationTimestamp(existingObj.GetCreationTimestamp())
		obj.SetUID(existingObj.GetUID())
	}

	// Update status for AlertRule
	if alertRule, ok := obj.(*monitoringv1alpha1.AlertRule); ok {
		alertRule.Status.LastReconcileTime = &metav1.Time{Time: time.Now()}
		// Keep existing conditions and add update condition
		alertRule.Status.Conditions = append(alertRule.Status.Conditions, metav1.Condition{
			Type:               "Updated",
			Status:             metav1.ConditionTrue,
			ObservedGeneration: obj.GetGeneration(),
			LastTransitionTime: metav1.NewTime(time.Now()),
			Reason:             "MockUpdateSuccess",
			Message:            fmt.Sprintf("Mock AlertRule %s updated successfully", alertRule.Name),
		})
	}

	// Store the updated object
	m.objects[key] = obj.DeepCopyObject()
	return nil
}

// Delete deletes an object
func (m *MockClient) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	key, err := m.objectKey(obj)
	if err != nil {
		return err
	}

	// Check if object exists
	if _, exists := m.objects[key]; !exists {
		return errors.NewNotFound(schema.GroupResource{
			Group:    "monitoring.kneutral.io",
			Resource: "alertrules",
		}, obj.GetName())
	}

	// Delete the object
	delete(m.objects, key)
	return nil
}

// Patch patches an object (simplified implementation)
func (m *MockClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	// For mock purposes, just treat patch as update
	return m.Update(ctx, obj, nil)
}

// DeleteAllOf deletes all objects matching the given options
func (m *MockClient) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	// Not implemented for mock
	return fmt.Errorf("DeleteAllOf not implemented in mock client")
}

// Status returns a status writer
func (m *MockClient) Status() client.StatusWriter {
	return &mockStatusWriter{client: m}
}

// Scheme returns the scheme
func (m *MockClient) Scheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	monitoringv1alpha1.AddToScheme(scheme)
	return scheme
}

// mockStatusWriter implements client.StatusWriter
type mockStatusWriter struct {
	client *MockClient
}

func (m *mockStatusWriter) Update(ctx context.Context, obj client.Object, opts ...client.SubResourceUpdateOption) error {
	// For mock purposes, just update the object
	return m.client.Update(ctx, obj, nil)
}

func (m *mockStatusWriter) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.SubResourcePatchOption) error {
	// For mock purposes, just update the object
	return m.client.Update(ctx, obj, nil)
}

func (m *mockStatusWriter) Create(ctx context.Context, obj client.Object, subResource client.Object, opts ...client.SubResourceCreateOption) error {
	// For mock purposes, not implemented
	return fmt.Errorf("Create not implemented in mock status writer")
}

// GroupVersionKindFor returns the GroupVersionKind for the given object
func (m *MockClient) GroupVersionKindFor(obj runtime.Object) (schema.GroupVersionKind, error) {
	switch obj.(type) {
	case *monitoringv1alpha1.AlertRule:
		return schema.GroupVersionKind{
			Group:   "monitoring.kneutral.io",
			Version: "v1alpha1",
			Kind:    "AlertRule",
		}, nil
	case *monitoringv1alpha1.AlertRuleList:
		return schema.GroupVersionKind{
			Group:   "monitoring.kneutral.io",
			Version: "v1alpha1",
			Kind:    "AlertRuleList",
		}, nil
	default:
		return schema.GroupVersionKind{}, fmt.Errorf("unknown object type: %T", obj)
	}
}

// IsObjectNamespaced returns true if the object is namespaced
func (m *MockClient) IsObjectNamespaced(obj runtime.Object) (bool, error) {
	switch obj.(type) {
	case *monitoringv1alpha1.AlertRule:
		return true, nil
	case *monitoringv1alpha1.AlertRuleList:
		return true, nil
	default:
		return false, fmt.Errorf("unknown object type: %T", obj)
	}
}

// RESTMapper returns the REST mapper (not implemented for mock)
func (m *MockClient) RESTMapper() meta.RESTMapper {
	return nil
}

// SubResource returns a client for a subresource
func (m *MockClient) SubResource(subresource string) client.SubResourceClient {
	return &mockSubResourceClient{client: m, subresource: subresource}
}

// mockSubResourceClient implements client.SubResourceClient
type mockSubResourceClient struct {
	client      *MockClient
	subresource string
}

func (m *mockSubResourceClient) Get(ctx context.Context, obj client.Object, subResource client.Object, opts ...client.SubResourceGetOption) error {
	return fmt.Errorf("SubResource Get not implemented in mock")
}

func (m *mockSubResourceClient) Create(ctx context.Context, obj client.Object, subResource client.Object, opts ...client.SubResourceCreateOption) error {
	return fmt.Errorf("SubResource Create not implemented in mock")
}

func (m *mockSubResourceClient) Update(ctx context.Context, obj client.Object, opts ...client.SubResourceUpdateOption) error {
	return m.client.Update(ctx, obj, nil)
}

func (m *mockSubResourceClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.SubResourcePatchOption) error {
	return m.client.Patch(ctx, obj, patch, nil)
}