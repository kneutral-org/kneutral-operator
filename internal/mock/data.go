package mock

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	monitoringv1alpha1 "github.com/kneutral-org/kneutral-operator/api/v1alpha1"
)

// PopulateExampleData adds sample AlertRule data to the mock client
func PopulateExampleData(client *MockClient) {
	ctx := context.Background()

	// Example 1: Basic CPU monitoring
	cpuAlert := &monitoringv1alpha1.AlertRule{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "monitoring.kneutral.io/v1alpha1",
			Kind:       "AlertRule",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cpu-monitoring",
			Namespace: "monitoring",
			Labels: map[string]string{
				"category": "infrastructure",
				"team":     "platform",
			},
		},
		Spec: monitoringv1alpha1.AlertRuleSpec{
			Groups: []monitoringv1alpha1.AlertGroup{
				{
					Name:     "cpu.rules",
					Interval: "30s",
					Rules: []monitoringv1alpha1.Rule{
						{
							Alert: "HighCPUUsage",
							Expr:  "100 - (avg by(instance) (irate(node_cpu_seconds_total{mode=\"idle\"}[5m])) * 100) > 80",
							For:   "5m",
							Labels: map[string]string{
								"severity": "warning",
							},
							Annotations: map[string]string{
								"summary":     "High CPU usage on {{ $labels.instance }}",
								"description": "CPU usage is {{ $value | printf \"%.2f\" }}% for more than 5 minutes",
							},
						},
						{
							Alert: "CriticalCPUUsage",
							Expr:  "100 - (avg by(instance) (irate(node_cpu_seconds_total{mode=\"idle\"}[5m])) * 100) > 95",
							For:   "2m",
							Labels: map[string]string{
								"severity": "critical",
							},
							Annotations: map[string]string{
								"summary":     "Critical CPU usage on {{ $labels.instance }}",
								"description": "CPU usage is {{ $value | printf \"%.2f\" }}% - immediate action required",
							},
						},
					},
				},
			},
		},
	}

	// Example 2: Application performance monitoring
	appAlert := &monitoringv1alpha1.AlertRule{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "monitoring.kneutral.io/v1alpha1",
			Kind:       "AlertRule",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "app-performance",
			Namespace: "production",
			Labels: map[string]string{
				"category": "application",
				"team":     "backend",
			},
		},
		Spec: monitoringv1alpha1.AlertRuleSpec{
			Groups: []monitoringv1alpha1.AlertGroup{
				{
					Name:     "app.response_time",
					Interval: "15s",
					Rules: []monitoringv1alpha1.Rule{
						{
							Alert: "HighResponseTime",
							Expr:  "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 2",
							For:   "5m",
							Labels: map[string]string{
								"severity": "warning",
								"service":  "{{ $labels.service }}",
							},
							Annotations: map[string]string{
								"summary":     "High response time for {{ $labels.service }}",
								"description": "95th percentile response time is {{ $value | printf \"%.3f\" }}s",
								"grafana_url": "https://grafana.company.com/d/app-performance",
							},
						},
					},
				},
				{
					Name:     "app.error_rate",
					Interval: "30s",
					Rules: []monitoringv1alpha1.Rule{
						{
							Alert: "HighErrorRate",
							Expr:  "rate(http_requests_total{status=~\"5..\"}[5m]) / rate(http_requests_total[5m]) > 0.05",
							For:   "3m",
							Labels: map[string]string{
								"severity": "warning",
								"service":  "{{ $labels.service }}",
							},
							Annotations: map[string]string{
								"summary":     "High error rate for {{ $labels.service }}",
								"description": "Error rate is {{ $value | humanizePercentage }}",
							},
						},
					},
				},
			},
		},
	}

	// Example 3: Network monitoring (based on your Arista example)
	networkAlert := &monitoringv1alpha1.AlertRule{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "monitoring.kneutral.io/v1alpha1",
			Kind:       "AlertRule",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "arista-dom-monitoring",
			Namespace: "network",
			Labels: map[string]string{
				"category": "network",
				"vendor":   "arista",
				"type":     "optical",
			},
		},
		Spec: monitoringv1alpha1.AlertRuleSpec{
			Labels: map[string]string{
				"app.kubernetes.io/instance": "kneutral",
			},
			Groups: []monitoringv1alpha1.AlertGroup{
				{
					Name: "kneutral.arista.dom",
					Rules: []monitoringv1alpha1.Rule{
						{
							Alert: "LowDOMRXPowerCritical",
							Expr: `(
  10 * log10(arista_smnp_entSensorValue{entPhysicalDescr=~"DOM RX Power.*"} / 1000)
  < on(desc, entPhysicalDescr) group_left
  10 * log10(arista_smnp_aristaSensorThresholdLowCritical{entPhysicalDescr=~"DOM RX Power.*"} / 1000)
)
and
(
  10 * log10(arista_smnp_entSensorValue{entPhysicalDescr=~"DOM RX Power.*"} / 1000) != -30
)`,
							For: "5m",
							Labels: map[string]string{
								"severity": "critical",
								"source":   "kneutral",
							},
							Annotations: map[string]string{
								"summary": "Critical: Low DOM RX Power on {{ $labels.entPhysicalDescr }} at {{ $labels.desc }}",
								"description": `DOM RX Power is below low critical threshold
Device: {{ $labels.desc }}
Site: {{ $labels.site }}
Role: {{ $labels.role }}
Location: {{ $labels.location }}
Interface: {{ $labels.entPhysicalDescr }}
Current Power: {{ $value | printf "%.2f" }} dBm`,
								"grafanaUrl": "https://mon.monitor.driveuc.com/d/arista-interfaces/arista-network-interfaces",
							},
						},
					},
				},
			},
		},
	}

	// Create the sample AlertRules
	client.Create(ctx, cpuAlert)
	client.Create(ctx, appAlert)
	client.Create(ctx, networkAlert)
}