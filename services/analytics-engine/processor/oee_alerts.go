package processor

import (
	"fmt"
	"sync"
	"time"
)

// Default OEE alert evaluator tuning parameters.
const (
	DefaultOEEAlertCooldown = 5 * time.Minute
)

// OEEComponent identifies which OEE metric to monitor.
type OEEComponent string

const (
	ComponentOEE          OEEComponent = "oee"
	ComponentAvailability OEEComponent = "availability"
	ComponentPerformance  OEEComponent = "performance"
	ComponentQuality      OEEComponent = "quality"
)

// OEEAlertRule defines a threshold for an OEE component.
type OEEAlertRule struct {
	ID                string
	Component         OEEComponent
	WarningThreshold  *float64 // e.g., 0.85 means 85%
	CriticalThreshold *float64 // e.g., 0.70 means 70%
	AssetID           string   // empty string = applies to all assets
	Description       string
}

// OEEAlert is the output when an OEE threshold is breached.
type OEEAlert struct {
	RuleID    string
	AssetID   string
	Component OEEComponent
	Value     float64
	Threshold float64
	Severity  AlertSeverity
	Message   string
	Timestamp time.Time
}

// OEEAlertHandlerFunc is the callback invoked when an OEE alert fires.
type OEEAlertHandlerFunc func(alert OEEAlert)

// OEEAlertEvaluator checks OEE snapshots against configured thresholds.
type OEEAlertEvaluator struct {
	mu        sync.RWMutex
	rules     []OEEAlertRule
	handler   OEEAlertHandlerFunc
	lastFired map[string]time.Time // key: "ruleID:assetID" → last fire time
	cooldown  time.Duration
}

// NewOEEAlertEvaluator creates an evaluator with the given rules, handler, and cooldown.
// If cooldown is 0, defaults to 5 minutes.
func NewOEEAlertEvaluator(rules []OEEAlertRule, handler OEEAlertHandlerFunc, cooldown time.Duration) *OEEAlertEvaluator {
	if cooldown == 0 {
		cooldown = DefaultOEEAlertCooldown
	}
	return &OEEAlertEvaluator{
		rules:     rules,
		handler:   handler,
		lastFired: make(map[string]time.Time),
		cooldown:  cooldown,
	}
}

// AddRule adds a rule at runtime.
func (e *OEEAlertEvaluator) AddRule(rule OEEAlertRule) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.rules = append(e.rules, rule)
}

// SetRules replaces all rules at runtime.
func (e *OEEAlertEvaluator) SetRules(rules []OEEAlertRule) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.rules = rules
}

// EvaluateSnapshot checks an OEE snapshot against all rules and fires the handler if a threshold is breached.
// Returns the alert if one fired, nil otherwise.
func (e *OEEAlertEvaluator) EvaluateSnapshot(snapshot OEESnapshot) *OEEAlert {
	e.mu.RLock()
	defer e.mu.RUnlock()

	for _, rule := range e.rules {
		// Skip if rule is asset-specific and doesn't match.
		if rule.AssetID != "" && rule.AssetID != snapshot.PhysicalAssetID {
			continue
		}

		value := e.getComponentValue(snapshot, rule.Component)

		// Check critical first (more severe).
		if rule.CriticalThreshold != nil && value < *rule.CriticalThreshold {
			return e.fireIfAllowed(rule, snapshot, value, *rule.CriticalThreshold, SeverityCritical)
		}

		// Then warning.
		if rule.WarningThreshold != nil && value < *rule.WarningThreshold {
			return e.fireIfAllowed(rule, snapshot, value, *rule.WarningThreshold, SeverityWarning)
		}
	}
	return nil
}

// getComponentValue extracts the right metric from a snapshot.
func (e *OEEAlertEvaluator) getComponentValue(s OEESnapshot, c OEEComponent) float64 {
	switch c {
	case ComponentAvailability:
		return s.Availability
	case ComponentPerformance:
		return s.Performance
	case ComponentQuality:
		return s.Quality
	default:
		return s.OEE
	}
}

// fireIfAllowed checks cooldown and fires the alert if enough time has passed since the last one.
func (e *OEEAlertEvaluator) fireIfAllowed(
	rule OEEAlertRule,
	snapshot OEESnapshot,
	value float64,
	threshold float64,
	severity AlertSeverity,
) *OEEAlert {
	key := rule.ID + ":" + snapshot.PhysicalAssetID
	now := snapshot.Timestamp

	// Check cooldown.
	if last, ok := e.lastFired[key]; ok && now.Sub(last) < e.cooldown {
		return nil // suppressed — too soon since last alert for this rule+asset
	}

	// Record fire time and build alert.
	e.lastFired[key] = now

	alert := OEEAlert{
		RuleID:    rule.ID,
		AssetID:   snapshot.PhysicalAssetID,
		Component: rule.Component,
		Value:     value,
		Threshold: threshold,
		Severity:  severity,
		Message: fmt.Sprintf(
			"%s at %.1f%% (threshold: %.1f%%) on asset %s",
			rule.Component, value*100, threshold*100, snapshot.PhysicalAssetID,
		),
		Timestamp: now,
	}

	if e.handler != nil {
		e.handler(alert)
	}
	return &alert
}

// DefaultOEEAlertRules returns sensible default thresholds for OEE monitoring.
func DefaultOEEAlertRules() []OEEAlertRule {
	warnOEE := 0.85
	critOEE := 0.70
	warnComp := 0.90
	critComp := 0.75

	return []OEEAlertRule{
		{
			ID:                "OEE-LOW",
			Component:         ComponentOEE,
			WarningThreshold:  &warnOEE,
			CriticalThreshold: &critOEE,
			Description:       "Composite OEE dropped below threshold",
		},
		{
			ID:                "AVAIL-LOW",
			Component:         ComponentAvailability,
			WarningThreshold:  &warnComp,
			CriticalThreshold: &critComp,
			Description:       "Availability dropped below threshold",
		},
		{
			ID:                "PERF-LOW",
			Component:         ComponentPerformance,
			WarningThreshold:  &warnComp,
			CriticalThreshold: &critComp,
			Description:       "Performance dropped below threshold",
		},
		{
			ID:                "QUAL-LOW",
			Component:         ComponentQuality,
			WarningThreshold:  &warnComp,
			CriticalThreshold: &critComp,
			Description:       "Quality dropped below threshold",
		},
	}
}
