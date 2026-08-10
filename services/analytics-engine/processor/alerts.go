package processor

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// AlertSeverity defines the severity level of a detected anomaly.
type AlertSeverity string

const (
	SeverityWarning  AlertSeverity = "WARNING"
	SeverityCritical AlertSeverity = "CRITICAL"
)

// SensorQuality represents the signal quality reported by the Edge Runtime.
// Using a typed constant prevents magic-string comparisons against Protobuf field values.
type SensorQuality string

const (
	QualityGood      SensorQuality = "GOOD"
	QualityBad       SensorQuality = "BAD"
	QualityUncertain SensorQuality = "UNCERTAIN"
)

// AlertRule defines dynamic threshold boundaries for a specific metric category.
type AlertRule struct {
	ID                string   `json:"id"`
	MetricNamePattern string   `json:"metric_name_pattern"` // Substring matcher (e.g. "temp", "vibrat", "pressure")
	WarningThreshold  *float64 `json:"warning_threshold,omitempty"`
	CriticalThreshold *float64 `json:"critical_threshold,omitempty"`
	Unit              string   `json:"unit"`
	Description       string   `json:"description"`
}

// TelemetryAlert represents an anomaly event evaluated from a sensor reading.
type TelemetryAlert struct {
	RuleID          string        `json:"rule_id,omitempty"`
	PhysicalAssetID string        `json:"physical_asset_id"`
	MetricName      string        `json:"metric_name"`
	Value           float64       `json:"value"`
	Threshold       float64       `json:"threshold"`
	Severity        AlertSeverity `json:"severity"`
	Message         string        `json:"message"`
	Timestamp       time.Time     `json:"timestamp"`
}

// AlertHandlerFunc is a callback for handling generated alerts.
type AlertHandlerFunc func(alert TelemetryAlert)

// AlertEvaluator performs real-time in-memory threshold evaluations based on dynamic rules.
type AlertEvaluator struct {
	mu      sync.RWMutex
	rules   []AlertRule
	handler AlertHandlerFunc
}

// DefaultAlertRules returns standard industrial baseline threshold rules.
func DefaultAlertRules() []AlertRule {
	tempWarn := 80.0
	tempCrit := 95.0
	vibeWarn := 10.0
	vibeCrit := 15.0
	pressWarn := 250.0
	pressCrit := 300.0

	return []AlertRule{
		{
			ID:                "RULE-TEMP-001",
			MetricNamePattern: "temp",
			WarningThreshold:  &tempWarn,
			CriticalThreshold: &tempCrit,
			Unit:              "°C",
			Description:       "Spindle / Motor Operating Temperature",
		},
		{
			ID:                "RULE-VIBE-001",
			MetricNamePattern: "vibrat",
			WarningThreshold:  &vibeWarn,
			CriticalThreshold: &vibeCrit,
			Unit:              "mm/s",
			Description:       "Mechanical Vibration Velocity",
		},
		{
			ID:                "RULE-PRESS-001",
			MetricNamePattern: "pressure",
			WarningThreshold:  &pressWarn,
			CriticalThreshold: &pressCrit,
			Unit:              "bar",
			Description:       "Hydraulic / Pneumatic System Pressure",
		},
	}
}

// NewAlertEvaluator creates a new AlertEvaluator with custom rules or default factory rules.
func NewAlertEvaluator(rules []AlertRule, handler AlertHandlerFunc) *AlertEvaluator {
	if len(rules) == 0 {
		rules = DefaultAlertRules()
	}
	if handler == nil {
		handler = defaultAlertHandler
	}
	return &AlertEvaluator{
		rules:   rules,
		handler: handler,
	}
}

func defaultAlertHandler(alert TelemetryAlert) {
	log.Printf("[ALARM][%s] Asset: %s | Metric: %s | Value: %.2f (Threshold: %.2f) | %s",
		alert.Severity, alert.PhysicalAssetID, alert.MetricName, alert.Value, alert.Threshold, alert.Message)
}

// AddRule allows dynamically injecting a new AlertRule at runtime.
func (e *AlertEvaluator) AddRule(rule AlertRule) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.rules = append(e.rules, rule)
}

// SetRules replaces all active evaluation rules.
func (e *AlertEvaluator) SetRules(rules []AlertRule) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.rules = rules
}

// EvaluateReading checks a single sensor reading against all active dynamic alert rules.
func (e *AlertEvaluator) EvaluateReading(assetID, metricName string, value float64, quality SensorQuality, ts time.Time) *TelemetryAlert {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var alert *TelemetryAlert
	metricLower := strings.ToLower(metricName)

	// 1. Check against dynamic rules
	for _, rule := range e.rules {
		if strings.Contains(metricLower, strings.ToLower(rule.MetricNamePattern)) {
			// Check Critical threshold first
			if rule.CriticalThreshold != nil && value > *rule.CriticalThreshold {
				alert = &TelemetryAlert{
					RuleID:          rule.ID,
					PhysicalAssetID: assetID,
					MetricName:      metricName,
					Value:           value,
					Threshold:       *rule.CriticalThreshold,
					Severity:        SeverityCritical,
					Message: fmt.Sprintf("Critical %s alarm: %.2f %s exceeds safety limit of %.2f %s",
						rule.Description, value, rule.Unit, *rule.CriticalThreshold, rule.Unit),
					Timestamp: ts,
				}
				break
			}

			// Check Warning threshold
			if rule.WarningThreshold != nil && value > *rule.WarningThreshold {
				alert = &TelemetryAlert{
					RuleID:          rule.ID,
					PhysicalAssetID: assetID,
					MetricName:      metricName,
					Value:           value,
					Threshold:       *rule.WarningThreshold,
					Severity:        SeverityWarning,
					Message: fmt.Sprintf("Elevated %s warning: %.2f %s exceeds threshold of %.2f %s",
						rule.Description, value, rule.Unit, *rule.WarningThreshold, rule.Unit),
					Timestamp: ts,
				}
				break
			}
		}
	}

	// 2. Sensor Signal Quality Degraded Check
	if alert == nil && quality == QualityBad {
		alert = &TelemetryAlert{
			RuleID:          "SYS-QUALITY-BAD",
			PhysicalAssetID: assetID,
			MetricName:      metricName,
			Value:           value,
			Threshold:       0.0,
			Severity:        SeverityWarning,
			Message:         "Sensor signal quality reported BAD by Edge Runtime",
			Timestamp:       ts,
		}
	}

	if alert != nil && e.handler != nil {
		e.handler(*alert)
	}

	return alert
}
