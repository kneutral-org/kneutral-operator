package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	monitoringv1alpha1 "github.com/kneutral-org/kneutral-operator/api/v1alpha1"
)

// AlertRuleReconciler reconciles a AlertRule object
type AlertRuleReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

// +kubebuilder:rbac:groups=monitoring.kneutral.io,resources=alertrules,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=monitoring.kneutral.io,resources=alertrules/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=monitoring.kneutral.io,resources=alertrules/finalizers,verbs=update
// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=prometheusrules,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop
func (r *AlertRuleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the AlertRule instance
	alertRule := &monitoringv1alpha1.AlertRule{}
	err := r.Get(ctx, req.NamespacedName, alertRule)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Return and don't requeue
			log.Info("AlertRule resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get AlertRule")
		return ctrl.Result{}, err
	}

	// Check if the AlertRule is being deleted
	if !alertRule.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is being deleted
		if controllerutil.ContainsFinalizer(alertRule, "alertrule.kneutral.io/finalizer") {
			// Delete the associated PrometheusRule
			if err := r.deletePrometheusRule(ctx, alertRule); err != nil {
				log.Error(err, "Failed to delete PrometheusRule")
				return ctrl.Result{}, err
			}

			// Remove finalizer
			controllerutil.RemoveFinalizer(alertRule, "alertrule.kneutral.io/finalizer")
			if err := r.Update(ctx, alertRule); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(alertRule, "alertrule.kneutral.io/finalizer") {
		controllerutil.AddFinalizer(alertRule, "alertrule.kneutral.io/finalizer")
		if err := r.Update(ctx, alertRule); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Generate PrometheusRule from AlertRule
	prometheusRule := r.generatePrometheusRule(alertRule)

	// Set AlertRule as the owner of the PrometheusRule
	if err := controllerutil.SetControllerReference(alertRule, prometheusRule, r.Scheme); err != nil {
		log.Error(err, "Failed to set owner reference")
		return ctrl.Result{}, err
	}

	// Check if PrometheusRule already exists
	found := &monitoringv1.PrometheusRule{}
	err = r.Get(ctx, types.NamespacedName{Name: prometheusRule.Name, Namespace: prometheusRule.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new PrometheusRule", "PrometheusRule.Namespace", prometheusRule.Namespace, "PrometheusRule.Name", prometheusRule.Name)
		err = r.Create(ctx, prometheusRule)
		if err != nil {
			log.Error(err, "Failed to create new PrometheusRule", "PrometheusRule.Namespace", prometheusRule.Namespace, "PrometheusRule.Name", prometheusRule.Name)
			return ctrl.Result{}, err
		}
		// PrometheusRule created successfully - update status
		return r.updateStatus(ctx, alertRule, prometheusRule.Name, "Active")
	} else if err != nil {
		log.Error(err, "Failed to get PrometheusRule")
		return ctrl.Result{}, err
	}

	// PrometheusRule already exists - update it
	found.Spec = prometheusRule.Spec
	found.Labels = prometheusRule.Labels
	log.Info("Updating existing PrometheusRule", "PrometheusRule.Namespace", found.Namespace, "PrometheusRule.Name", found.Name)
	err = r.Update(ctx, found)
	if err != nil {
		log.Error(err, "Failed to update PrometheusRule", "PrometheusRule.Namespace", found.Namespace, "PrometheusRule.Name", found.Name)
		return ctrl.Result{}, err
	}

	// Update status
	return r.updateStatus(ctx, alertRule, found.Name, "Active")
}

// generatePrometheusRule creates a PrometheusRule from an AlertRule
func (r *AlertRuleReconciler) generatePrometheusRule(alertRule *monitoringv1alpha1.AlertRule) *monitoringv1.PrometheusRule {
	labels := map[string]string{
		"app.kubernetes.io/managed-by": "kneutral-operator",
		"app.kubernetes.io/instance":   "kneutral",
		"app.kubernetes.io/name":       alertRule.Name,
	}

	// Merge user-provided labels
	for k, v := range alertRule.Spec.Labels {
		labels[k] = v
	}

	prometheusRule := &monitoringv1.PrometheusRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("kneutral-%s", alertRule.Name),
			Namespace: alertRule.Namespace,
			Labels:    labels,
		},
		Spec: monitoringv1.PrometheusRuleSpec{
			Groups: []monitoringv1.RuleGroup{},
		},
	}

	// Convert AlertGroups to RuleGroups
	for _, group := range alertRule.Spec.Groups {
		ruleGroup := monitoringv1.RuleGroup{
			Name: group.Name,
		}

		if group.Interval != "" {
			interval := monitoringv1.Duration(group.Interval)
			ruleGroup.Interval = &interval
		}

		// Convert Rules
		for _, rule := range group.Rules {
			promRule := monitoringv1.Rule{
				Alert:       rule.Alert,
				Expr:        intstr.FromString(rule.Expr),
				Labels:      rule.Labels,
				Annotations: rule.Annotations,
			}

			if rule.For != "" {
				forDuration := monitoringv1.Duration(rule.For)
				promRule.For = &forDuration
			}

			ruleGroup.Rules = append(ruleGroup.Rules, promRule)
		}

		prometheusRule.Spec.Groups = append(prometheusRule.Spec.Groups, ruleGroup)
	}

	return prometheusRule
}

// deletePrometheusRule deletes the PrometheusRule associated with an AlertRule
func (r *AlertRuleReconciler) deletePrometheusRule(ctx context.Context, alertRule *monitoringv1alpha1.AlertRule) error {
	prometheusRule := &monitoringv1.PrometheusRule{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      fmt.Sprintf("kneutral-%s", alertRule.Name),
		Namespace: alertRule.Namespace,
	}, prometheusRule)
	if err != nil {
		if errors.IsNotFound(err) {
			// PrometheusRule not found, nothing to delete
			return nil
		}
		return err
	}

	return r.Delete(ctx, prometheusRule)
}

// updateStatus updates the AlertRule status
func (r *AlertRuleReconciler) updateStatus(ctx context.Context, alertRule *monitoringv1alpha1.AlertRule, prometheusRuleName, state string) (ctrl.Result, error) {
	now := metav1.Now()
	alertRule.Status.LastReconcileTime = &now
	alertRule.Status.PrometheusRuleName = prometheusRuleName
	alertRule.Status.State = state

	// Update conditions
	condition := metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionTrue,
		ObservedGeneration: alertRule.Generation,
		LastTransitionTime: now,
		Reason:             "ReconcileSuccess",
		Message:            fmt.Sprintf("PrometheusRule %s created/updated successfully", prometheusRuleName),
	}

	// Update or append the condition
	found := false
	for i, c := range alertRule.Status.Conditions {
		if c.Type == condition.Type {
			alertRule.Status.Conditions[i] = condition
			found = true
			break
		}
	}
	if !found {
		alertRule.Status.Conditions = append(alertRule.Status.Conditions, condition)
	}

	err := r.Status().Update(ctx, alertRule)
	if err != nil {
		log.FromContext(ctx).Error(err, "Failed to update AlertRule status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AlertRuleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&monitoringv1alpha1.AlertRule{}).
		Owns(&monitoringv1.PrometheusRule{}).
		Complete(r)
}
