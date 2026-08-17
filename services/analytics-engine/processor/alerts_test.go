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

func TestAlertEvaluator_DirectionBelow(t *testing.T) {
	warnThreshold := 0.85
	critThreshold := 0.70

	rules := []AlertRule{
		{
			ID:                "OEE-TEST",
			MetricNamePattern: "oee",
			WarningThreshold:  &warnThreshold,
			CriticalThreshold: &critThreshold,
			Direction:         DirectionBelow,
		},
	}

	var captured *TelemetryAlert
	handler := func(alert TelemetryAlert) {
		captured = &alert
	}

	evaluator := NewAlertEvaluator(rules, handler)

	// Above all thresholds — no alert
	captured = nil
	evaluator.EvaluateReading("asset-1", "oee", 0.90, QualityGood, time.Now())
	if captured != nil {
		t.Fatal("expected no alert when value is above thresholds")
	}

	// Below warning — should fire warning
	captured = nil
	evaluator.EvaluateReading("asset-1", "oee", 0.80, QualityGood, time.Now())
	if captured == nil || captured.Severity != SeverityWarning {
		t.Fatalf("expected warning alert when value drops below warning threshold, got: %+v", captured)
	}

	// Below critical — should fire critical
	captured = nil
	evaluator.EvaluateReading("asset-1", "oee", 0.65, QualityGood, time.Now())
	if captured == nil || captured.Severity != SeverityCritical {
		t.Fatalf("expected critical alert when value drops below critical threshold, got: %+v", captured)
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
