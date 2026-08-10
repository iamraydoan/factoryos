package processor

import (
	"testing"
	"time"
)

func TestAlertEvaluator_DefaultRules(t *testing.T) {
	var capturedAlerts []TelemetryAlert
	evaluator := NewAlertEvaluator(nil, func(alert TelemetryAlert) {
		capturedAlerts = append(capturedAlerts, alert)
	})

	now := time.Now()

	// Case 1: Normal temperature (70.0°C) -> No alert
	alert := evaluator.EvaluateReading("asset-001", "temperature_celsius", 70.0, "GOOD", now)
	if alert != nil {
		t.Fatalf("expected nil alert for normal temp, got: %+v", alert)
	}

	// Case 2: Warning temperature (85.0°C)
	alert = evaluator.EvaluateReading("asset-001", "temperature_celsius", 85.0, "GOOD", now)
	if alert == nil || alert.Severity != SeverityWarning {
		t.Fatalf("expected WARNING alert, got: %+v", alert)
	}

	// Case 3: Critical overheat (102.5°C)
	alert = evaluator.EvaluateReading("asset-001", "temperature_celsius", 102.5, "GOOD", now)
	if alert == nil || alert.Severity != SeverityCritical {
		t.Fatalf("expected CRITICAL alert, got: %+v", alert)
	}

	if len(capturedAlerts) != 2 {
		t.Fatalf("expected 2 captured alerts in callback, got: %d", len(capturedAlerts))
	}
}

func TestAlertEvaluator_CustomDynamicRules(t *testing.T) {
	customWarn := 50.0
	customCrit := 100.0
	customRules := []AlertRule{
		{
			ID:                "CUSTOM-TORQUE-01",
			MetricNamePattern: "torque",
			WarningThreshold:  &customWarn,
			CriticalThreshold: &customCrit,
			Unit:              "Nm",
			Description:       "Motor Torque Output",
		},
	}

	evaluator := NewAlertEvaluator(customRules, nil)
	now := time.Now()

	// Normal torque
	alert := evaluator.EvaluateReading("cnc-01", "motor_torque_nm", 30.0, "GOOD", now)
	if alert != nil {
		t.Fatalf("expected nil alert for normal torque, got: %+v", alert)
	}

	// Critical torque
	alert = evaluator.EvaluateReading("cnc-01", "motor_torque_nm", 120.0, "GOOD", now)
	if alert == nil || alert.Severity != SeverityCritical {
		t.Fatalf("expected CRITICAL alert for high torque, got: %+v", alert)
	}

	// Dynamic injection: Add custom flow rate rule
	flowWarn := 10.0
	evaluator.AddRule(AlertRule{
		ID:                "CUSTOM-FLOW-01",
		MetricNamePattern: "flow",
		WarningThreshold:  &flowWarn,
		Unit:              "L/min",
		Description:       "Coolant Flow Rate",
	})

	alert = evaluator.EvaluateReading("cnc-01", "coolant_flow_rate", 15.0, "GOOD", now)
	if alert == nil || alert.Severity != SeverityWarning {
		t.Fatalf("expected WARNING alert for flow rate, got: %+v", alert)
	}

	// Test SetRules
	evaluator.SetRules(DefaultAlertRules())
	alert = evaluator.EvaluateReading("cnc-01", "coolant_flow_rate", 15.0, "GOOD", now)
	if alert != nil {
		t.Fatalf("expected nil alert after reverting rules, got: %+v", alert)
	}
}

func TestAlertEvaluator_BadQuality(t *testing.T) {
	evaluator := NewAlertEvaluator(nil, nil)
	now := time.Now()

	alert := evaluator.EvaluateReading("asset-003", "flow_rate", 12.0, "BAD", now)
	if alert == nil || alert.Severity != SeverityWarning {
		t.Fatalf("expected WARNING alert for BAD quality, got: %+v", alert)
	}
}
