package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AlertRuleSpec defines the desired state of AlertRule
type AlertRuleSpec struct {
	// Groups is a list of alert groups
	Groups []AlertGroup `json:"groups"`

	// Labels to add to the generated PrometheusRule
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
}

// AlertGroup defines a group of alerts
type AlertGroup struct {
	// Name of the alert group
	Name string `json:"name"`

	// Interval how often rules in the group are evaluated
	// +optional
	Interval string `json:"interval,omitempty"`

	// Rules is a list of alert rules
	Rules []Rule `json:"rules"`
}

// Rule defines a single alert rule
type Rule struct {
	// Alert name
	Alert string `json:"alert"`

	// PromQL expression to evaluate
	Expr string `json:"expr"`

	// For clause - how long the alert must be pending before firing
	// +optional
	For string `json:"for,omitempty"`

	// Labels to add or override
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations to add
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
}

// AlertRuleStatus defines the observed state of AlertRule
type AlertRuleStatus struct {
	// Conditions represent the latest available observations
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// LastReconcileTime is the last time the AlertRule was reconciled
	// +optional
	LastReconcileTime *metav1.Time `json:"lastReconcileTime,omitempty"`

	// PrometheusRuleName is the name of the generated PrometheusRule
	// +optional
	PrometheusRuleName string `json:"prometheusRuleName,omitempty"`

	// State represents the current state of the AlertRule
	// +optional
	State string `json:"state,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.state`
// +kubebuilder:printcolumn:name="PrometheusRule",type=string,JSONPath=`.status.prometheusRuleName`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// AlertRule is the Schema for the alertrules API
type AlertRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AlertRuleSpec   `json:"spec,omitempty"`
	Status AlertRuleStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AlertRuleList contains a list of AlertRule
type AlertRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AlertRule `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AlertRule{}, &AlertRuleList{})
}
