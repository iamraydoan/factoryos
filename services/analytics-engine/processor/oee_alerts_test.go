package processor

import (
	"testing"
	"time"
)

func TestOEEAlertEvaluator_NoAlertWhenAboveThresholds(t *testing.T) {
	rules := DefaultOEEAlertRules()
	var alerts []OEEAlert
	eval := NewOEEAlertEvaluator(rules, func(a OEEAlert) {
		alerts = append(alerts, a)
	}, 0)

	snapshot := OEESnapshot{
		PhysicalAssetID: "asset-1",
		Timestamp:       time.Now(),
		Availability:    0.95,
		Performance:     0.95,
		Quality:         0.95,
		OEE:             0.857, // above 0.85 warning
	}

	result := eval.EvaluateSnapshot(snapshot)
	if result != nil {
		t.Fatalf("expected no alert when OEE is above thresholds, got: %+v", result)
	}
	if len(alerts) != 0 {
		t.Fatalf("expected 0 handler calls, got: %d", len(alerts))
	}
}

func TestOEEAlertEvaluator_WarningFires(t *testing.T) {
	warn := 0.85
	crit := 0.70
	rules := []OEEAlertRule{
		{
			ID:                "OEE-TEST",
			Component:         ComponentOEE,
			WarningThreshold:  &warn,
			CriticalThreshold: &crit,
		},
	}

	var captured *OEEAlert
	eval := NewOEEAlertEvaluator(rules, func(a OEEAlert) {
		captured = &a
	}, 0)

	snapshot := OEESnapshot{
		PhysicalAssetID: "asset-1",
		Timestamp:       time.Now(),
		OEE:             0.80, // below 0.85 warning
	}

	result := eval.EvaluateSnapshot(snapshot)
	if result == nil || result.Severity != SeverityWarning {
		t.Fatalf("expected warning alert, got: %+v", result)
	}
	if captured == nil || captured.Severity != SeverityWarning {
		t.Fatalf("expected handler to receive warning alert, got: %+v", captured)
	}
}

func TestOEEAlertEvaluator_CriticalFiresBeforeWarning(t *testing.T) {
	warn := 0.85
	crit := 0.70
	rules := []OEEAlertRule{
		{
			ID:                "OEE-TEST",
			Component:         ComponentOEE,
			WarningThreshold:  &warn,
			CriticalThreshold: &crit,
		},
	}

	var captured *OEEAlert
	eval := NewOEEAlertEvaluator(rules, func(a OEEAlert) {
		captured = &a
	}, 0)

	snapshot := OEESnapshot{
		PhysicalAssetID: "asset-1",
		Timestamp:       time.Now(),
		OEE:             0.65, // below 0.70 critical
	}

	result := eval.EvaluateSnapshot(snapshot)
	if result == nil || result.Severity != SeverityCritical {
		t.Fatalf("expected critical alert, got: %+v", result)
	}
	if captured == nil || captured.Severity != SeverityCritical {
		t.Fatalf("expected handler to receive critical alert, got: %+v", captured)
	}
}

func TestOEEAlertEvaluator_AssetSpecificRule(t *testing.T) {
	warn := 0.85
	rules := []OEEAlertRule{
		{
			ID:               "OEE-SPECIFIC",
			Component:        ComponentOEE,
			WarningThreshold: &warn,
			AssetID:          "asset-target",
		},
	}

	eval := NewOEEAlertEvaluator(rules, nil, 0)

	// Different asset — should not fire.
	snapshot := OEESnapshot{
		PhysicalAssetID: "asset-other",
		Timestamp:       time.Now(),
		OEE:             0.50,
	}
	if result := eval.EvaluateSnapshot(snapshot); result != nil {
		t.Fatalf("expected no alert for non-target asset, got: %+v", result)
	}

	// Target asset — should fire.
	snapshot.PhysicalAssetID = "asset-target"
	result := eval.EvaluateSnapshot(snapshot)
	if result == nil {
		t.Fatal("expected warning alert for target asset")
	}
}

func TestOEEAlertEvaluator_ComponentAvailability(t *testing.T) {
	warn := 0.90
	rules := []OEEAlertRule{
		{
			ID:               "AVAIL-TEST",
			Component:        ComponentAvailability,
			WarningThreshold: &warn,
		},
	}

	eval := NewOEEAlertEvaluator(rules, nil, 0)

	snapshot := OEESnapshot{
		PhysicalAssetID: "asset-1",
		Timestamp:       time.Now(),
		Availability:    0.85, // below 0.90
		Performance:     0.99,
		Quality:         0.99,
		OEE:             0.83, // OEE is fine, availability is the issue
	}

	result := eval.EvaluateSnapshot(snapshot)
	if result == nil || result.Component != ComponentAvailability {
		t.Fatalf("expected availability alert, got: %+v", result)
	}
}

func TestOEEAlertEvaluator_CooldownSuppression(t *testing.T) {
	warn := 0.85
	rules := []OEEAlertRule{
		{
			ID:               "OEE-CD",
			Component:        ComponentOEE,
			WarningThreshold: &warn,
		},
	}

	fireCount := 0
	eval := NewOEEAlertEvaluator(rules, func(a OEEAlert) {
		fireCount++
	}, 5*time.Minute)

	now := time.Now()
	snapshot := OEESnapshot{
		PhysicalAssetID: "asset-1",
		Timestamp:       now,
		OEE:             0.80,
	}

	// First call — should fire.
	eval.EvaluateSnapshot(snapshot)
	if fireCount != 1 {
		t.Fatalf("expected 1 fire, got: %d", fireCount)
	}

	// Second call within cooldown — should be suppressed.
	snapshot.Timestamp = now.Add(1 * time.Minute)
	eval.EvaluateSnapshot(snapshot)
	if fireCount != 1 {
		t.Fatalf("expected still 1 fire (cooldown), got: %d", fireCount)
	}

	// Third call after cooldown — should fire again.
	snapshot.Timestamp = now.Add(6 * time.Minute)
	eval.EvaluateSnapshot(snapshot)
	if fireCount != 2 {
		t.Fatalf("expected 2 fires (cooldown expired), got: %d", fireCount)
	}
}

func TestOEEAlertEvaluator_NilHandler(t *testing.T) {
	warn := 0.85
	rules := []OEEAlertRule{
		{
			ID:               "OEE-NIL",
			Component:        ComponentOEE,
			WarningThreshold: &warn,
		},
	}

	eval := NewOEEAlertEvaluator(rules, nil, 0)

	snapshot := OEESnapshot{
		PhysicalAssetID: "asset-1",
		Timestamp:       time.Now(),
		OEE:             0.80,
	}

	// Should not panic with nil handler.
	result := eval.EvaluateSnapshot(snapshot)
	if result == nil {
		t.Fatal("expected alert to be returned even with nil handler")
	}
}

func TestOEEAlertEvaluator_DynamicRuleManagement(t *testing.T) {
	eval := NewOEEAlertEvaluator(nil, nil, 0)

	snapshot := OEESnapshot{
		PhysicalAssetID: "asset-1",
		Timestamp:       time.Now(),
		OEE:             0.80,
	}

	// No rules — no alert.
	if result := eval.EvaluateSnapshot(snapshot); result != nil {
		t.Fatalf("expected no alert with no rules, got: %+v", result)
	}

	// Add a rule.
	warn := 0.85
	eval.AddRule(OEEAlertRule{
		ID:               "DYNAMIC",
		Component:        ComponentOEE,
		WarningThreshold: &warn,
	})

	result := eval.EvaluateSnapshot(snapshot)
	if result == nil {
		t.Fatal("expected alert after adding rule")
	}

	// Replace all rules with empty.
	eval.SetRules(nil)
	if result := eval.EvaluateSnapshot(snapshot); result != nil {
		t.Fatalf("expected no alert after clearing rules, got: %+v", result)
	}
}

func TestDefaultOEEAlertRules(t *testing.T) {
	rules := DefaultOEEAlertRules()
	if len(rules) != 4 {
		t.Fatalf("expected 4 default rules, got: %d", len(rules))
	}

	ids := map[string]bool{}
	for _, r := range rules {
		ids[r.ID] = true
		if r.WarningThreshold == nil {
			t.Fatalf("rule %s: expected non-nil WarningThreshold", r.ID)
		}
		if r.CriticalThreshold == nil {
			t.Fatalf("rule %s: expected non-nil CriticalThreshold", r.ID)
		}
		if *r.WarningThreshold <= *r.CriticalThreshold {
			t.Fatalf("rule %s: warning threshold (%.2f) must be above critical (%.2f)",
				r.ID, *r.WarningThreshold, *r.CriticalThreshold)
		}
	}

	for _, expected := range []string{"OEE-LOW", "AVAIL-LOW", "PERF-LOW", "QUAL-LOW"} {
		if !ids[expected] {
			t.Fatalf("missing expected rule: %s", expected)
		}
	}
}
