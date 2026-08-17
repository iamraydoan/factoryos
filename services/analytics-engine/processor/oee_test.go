package processor

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestOEEAggregator_FullOEE(t *testing.T) {
	cfg := OEEConfig{
		WindowDuration:        1 * time.Hour,
		SnapshotInterval:      1 * time.Hour, // disable auto-snapshots
		DefaultIdealCycleTime: 1 * time.Minute,
	}
	agg := NewOEEAggregator(cfg, nil)

	now := time.Now()
	asset := "asset-full-oee"

	// Availability: 8 running out of 10 ticks = 80%
	for i := 0; i < 8; i++ {
		agg.ProcessReading(asset, "machine_status", 1.0, "GOOD", now.Add(time.Duration(i)*time.Minute))
	}
	for i := 0; i < 2; i++ {
		agg.ProcessReading(asset, "machine_status", 0.0, "GOOD", now.Add(time.Duration(8+i)*time.Minute))
	}

	// Performance: 30 parts produced in ~60 min window.
	// Ideal cycle = 1 min → ideal output in 48 min run time = 48 parts.
	// Performance = 30/48 = 62.5%
	for i := 0; i < 30; i++ {
		agg.ProcessReading(asset, "cycle_count", 1.0, "GOOD", now.Add(time.Duration(i)*time.Minute))
	}

	// Quality: 3 defective out of 30 total → 90%
	for i := 0; i < 3; i++ {
		agg.ProcessReading(asset, "reject_count", 1.0, "GOOD", now.Add(time.Duration(i)*time.Minute))
	}

	snap := agg.GetSnapshotForAsset(asset)
	if snap == nil {
		t.Fatal("expected non-nil snapshot")
	}

	// Availability = 8/10 = 0.8
	if snap.Availability < 0.79 || snap.Availability > 0.81 {
		t.Fatalf("expected availability ~0.8, got: %.4f", snap.Availability)
	}

	// Quality = (30-3)/30 = 0.9
	if snap.Quality < 0.89 || snap.Quality > 0.91 {
		t.Fatalf("expected quality ~0.9, got: %.4f", snap.Quality)
	}

	// OEE = A * P * Q
	expectedOEE := snap.Availability * snap.Performance * snap.Quality
	if snap.OEE < expectedOEE-0.01 || snap.OEE > expectedOEE+0.01 {
		t.Fatalf("expected OEE ~%.4f, got: %.4f", expectedOEE, snap.OEE)
	}
}

func TestOEEAggregator_AvailabilityOnly(t *testing.T) {
	cfg := DefaultOEEConfig()
	cfg.SnapshotInterval = 1 * time.Hour
	agg := NewOEEAggregator(cfg, nil)

	now := time.Now()
	asset := "asset-avail"

	// All 10 ticks running → 100% availability
	for i := 0; i < 10; i++ {
		agg.ProcessReading(asset, "machine_status", 1.0, "GOOD", now.Add(time.Duration(i)*time.Minute))
	}

	snap := agg.GetSnapshotForAsset(asset)
	if snap == nil {
		t.Fatal("expected non-nil snapshot")
	}
	if snap.Availability != 1.0 {
		t.Fatalf("expected availability 1.0, got: %.4f", snap.Availability)
	}
	if snap.Performance != 1.0 {
		t.Fatalf("expected performance 1.0 (no output data), got: %.4f", snap.Performance)
	}
	if snap.Quality != 1.0 {
		t.Fatalf("expected quality 1.0 (no output data), got: %.4f", snap.Quality)
	}
}

func TestOEEAggregator_Downtime(t *testing.T) {
	cfg := DefaultOEEConfig()
	cfg.SnapshotInterval = 1 * time.Hour
	agg := NewOEEAggregator(cfg, nil)

	now := time.Now()
	asset := "asset-downtime"

	// 5 running, 5 stopped → 50% availability
	for i := 0; i < 5; i++ {
		agg.ProcessReading(asset, "machine_status", 1.0, "GOOD", now.Add(time.Duration(i)*time.Minute))
	}
	for i := 0; i < 5; i++ {
		agg.ProcessReading(asset, "machine_status", 0.0, "GOOD", now.Add(time.Duration(5+i)*time.Minute))
	}

	snap := agg.GetSnapshotForAsset(asset)
	if snap == nil {
		t.Fatal("expected non-nil snapshot")
	}
	if snap.Availability < 0.49 || snap.Availability > 0.51 {
		t.Fatalf("expected availability ~0.5, got: %.4f", snap.Availability)
	}
}

func TestOEEAggregator_DowntimeMetric(t *testing.T) {
	cfg := DefaultOEEConfig()
	cfg.SnapshotInterval = 1 * time.Hour
	agg := NewOEEAggregator(cfg, nil)

	now := time.Now()
	asset := "asset-downtime-metric"

	// "downtime" metric: value > 0 means NOT running (inverted semantics)
	agg.ProcessReading(asset, "downtime", 0.0, "GOOD", now)
	agg.ProcessReading(asset, "downtime", 0.0, "GOOD", now.Add(1*time.Minute))
	agg.ProcessReading(asset, "downtime", 30.0, "GOOD", now.Add(2*time.Minute)) // 30s downtime
	agg.ProcessReading(asset, "downtime", 0.0, "GOOD", now.Add(3*time.Minute))

	snap := agg.GetSnapshotForAsset(asset)
	if snap == nil {
		t.Fatal("expected non-nil snapshot")
	}
	// 3 running (downtime=0) out of 4 → 75%
	if snap.Availability < 0.74 || snap.Availability > 0.76 {
		t.Fatalf("expected availability ~0.75, got: %.4f", snap.Availability)
	}
}

func TestOEEAggregator_QualityFromFlag(t *testing.T) {
	cfg := DefaultOEEConfig()
	cfg.SnapshotInterval = 1 * time.Hour
	agg := NewOEEAggregator(cfg, nil)

	now := time.Now()
	asset := "asset-quality-flag"

	// 8 GOOD quality, 2 BAD quality → quality = 8/10 = 80%
	for i := 0; i < 8; i++ {
		agg.ProcessReading(asset, "sensor_temp", 50.0, "GOOD", now.Add(time.Duration(i)*time.Minute))
	}
	for i := 0; i < 2; i++ {
		agg.ProcessReading(asset, "sensor_temp", 50.0, "BAD", now.Add(time.Duration(8+i)*time.Minute))
	}

	snap := agg.GetSnapshotForAsset(asset)
	if snap == nil {
		t.Fatal("expected non-nil snapshot")
	}
	if snap.Quality < 0.79 || snap.Quality > 0.81 {
		t.Fatalf("expected quality ~0.8, got: %.4f", snap.Quality)
	}
	if snap.TotalOutput != 10 {
		t.Fatalf("expected total output 10, got: %d", snap.TotalOutput)
	}
	if snap.GoodOutput != 8 {
		t.Fatalf("expected good output 8, got: %d", snap.GoodOutput)
	}
}

func TestOEEAggregator_PerformanceWithIdealCycleTime(t *testing.T) {
	cfg := OEEConfig{
		WindowDuration:        1 * time.Hour,
		SnapshotInterval:      1 * time.Hour,
		DefaultIdealCycleTime: 2 * time.Minute, // 1 part every 2 minutes
	}
	agg := NewOEEAggregator(cfg, nil)

	now := time.Now()
	asset := "asset-perf"

	// All 60 ticks running → 100% availability → 60 min run time
	for i := 0; i < 60; i++ {
		agg.ProcessReading(asset, "machine_status", 1.0, "GOOD", now.Add(time.Duration(i)*time.Minute))
	}

	// Produce 20 parts. Ideal cycle = 2 min → ideal output in 60 min = 30 parts.
	// Performance = 20/30 = 66.7%
	for i := 0; i < 20; i++ {
		agg.ProcessReading(asset, "parts_count", 1.0, "GOOD", now.Add(time.Duration(i)*3*time.Minute))
	}

	snap := agg.GetSnapshotForAsset(asset)
	if snap == nil {
		t.Fatal("expected non-nil snapshot")
	}
	if snap.Performance < 0.65 || snap.Performance > 0.68 {
		t.Fatalf("expected performance ~0.667, got: %.4f", snap.Performance)
	}

	// Override ideal cycle time to 1 minute → ideal output = 60 → perf = 20/60 = 33.3%
	agg.SetIdealCycleTime(asset, 1*time.Minute)
	snap = agg.GetSnapshotForAsset(asset)
	if snap.Performance < 0.32 || snap.Performance > 0.34 {
		t.Fatalf("expected performance ~0.333 after cycle time override, got: %.4f", snap.Performance)
	}
}

func TestOEEAggregator_MultipleAssets(t *testing.T) {
	cfg := DefaultOEEConfig()
	cfg.SnapshotInterval = 1 * time.Hour
	agg := NewOEEAggregator(cfg, nil)

	now := time.Now()

	// Asset A: 100% running
	for i := 0; i < 10; i++ {
		agg.ProcessReading("asset-A", "machine_status", 1.0, "GOOD", now.Add(time.Duration(i)*time.Minute))
	}
	// Asset B: 50% running
	for i := 0; i < 5; i++ {
		agg.ProcessReading("asset-B", "machine_status", 1.0, "GOOD", now.Add(time.Duration(i)*time.Minute))
	}
	for i := 0; i < 5; i++ {
		agg.ProcessReading("asset-B", "machine_status", 0.0, "GOOD", now.Add(time.Duration(5+i)*time.Minute))
	}

	snapA := agg.GetSnapshotForAsset("asset-A")
	snapB := agg.GetSnapshotForAsset("asset-B")
	if snapA == nil || snapB == nil {
		t.Fatal("expected non-nil snapshots for both assets")
	}
	if snapA.Availability != 1.0 {
		t.Fatalf("asset-A: expected availability 1.0, got: %.4f", snapA.Availability)
	}
	if snapB.Availability < 0.49 || snapB.Availability > 0.51 {
		t.Fatalf("asset-B: expected availability ~0.5, got: %.4f", snapB.Availability)
	}

	all := agg.GetSnapshots()
	if len(all) != 2 {
		t.Fatalf("expected 2 snapshots, got: %d", len(all))
	}
}

func TestOEEAggregator_SnapshotCallback(t *testing.T) {
	cfg := DefaultOEEConfig()
	cfg.SnapshotInterval = 50 * time.Millisecond // fast for testing
	agg := NewOEEAggregator(cfg, nil)

	var mu sync.Mutex
	var received []OEESnapshot
	agg.SetHandler(func(s OEESnapshot) {
		mu.Lock()
		received = append(received, s)
		mu.Unlock()
	})

	now := time.Now()
	// Add some data so there's something to snapshot
	agg.ProcessReading("asset-cb", "machine_status", 1.0, "GOOD", now)
	agg.ProcessReading("asset-cb", "cycle_count", 5.0, "GOOD", now)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	agg.Start(ctx)
	defer agg.Stop()

	// Wait for at least one snapshot cycle
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	count := len(received)
	mu.Unlock()

	if count == 0 {
		t.Fatal("expected at least 1 snapshot callback, got 0")
	}
}

func TestOEEAggregator_ZeroDivision(t *testing.T) {
	cfg := DefaultOEEConfig()
	cfg.SnapshotInterval = 1 * time.Hour
	agg := NewOEEAggregator(cfg, nil)

	// No data at all
	snap := agg.GetSnapshotForAsset("nonexistent")
	if snap != nil {
		t.Fatal("expected nil snapshot for unknown asset")
	}

	// Asset with no readings
	agg.states["empty-asset"] = &assetOEEState{windowStart: time.Now()}
	snap = agg.GetSnapshotForAsset("empty-asset")
	if snap == nil {
		t.Fatal("expected non-nil snapshot for empty asset")
	}
	// Should default to 1.0 for all components (no negative bias)
	if snap.Availability != 1.0 {
		t.Fatalf("expected availability 1.0 for empty asset, got: %.4f", snap.Availability)
	}
	if snap.Performance != 1.0 {
		t.Fatalf("expected performance 1.0 for empty asset, got: %.4f", snap.Performance)
	}
	if snap.Quality != 1.0 {
		t.Fatalf("expected quality 1.0 for empty asset, got: %.4f", snap.Quality)
	}
	if snap.OEE != 1.0 {
		t.Fatalf("expected OEE 1.0 for empty asset, got: %.4f", snap.OEE)
	}
}

func TestOEEAggregator_WindowReset(t *testing.T) {
	cfg := OEEConfig{
		WindowDuration:        10 * time.Minute,
		SnapshotInterval:      1 * time.Hour,
		DefaultIdealCycleTime: 1 * time.Minute,
	}
	agg := NewOEEAggregator(cfg, nil)

	now := time.Now()
	asset := "asset-window"

	// Old readings outside the window
	for i := 0; i < 5; i++ {
		agg.ProcessReading(asset, "machine_status", 0.0, "GOOD", now.Add(-20*time.Minute))
	}

	// New readings inside the window
	for i := 0; i < 10; i++ {
		agg.ProcessReading(asset, "machine_status", 1.0, "GOOD", now.Add(time.Duration(i)*time.Minute))
	}

	snap := agg.GetSnapshotForAsset(asset)
	if snap == nil {
		t.Fatal("expected non-nil snapshot")
	}

	// After window reset, only new readings should count: 10/10 = 100%
	if snap.Availability != 1.0 {
		t.Fatalf("expected availability 1.0 after window reset, got: %.4f", snap.Availability)
	}
}

func TestOEEAggregator_ScrapMetrics(t *testing.T) {
	cfg := DefaultOEEConfig()
	cfg.SnapshotInterval = 1 * time.Hour
	agg := NewOEEAggregator(cfg, nil)

	now := time.Now()
	asset := "asset-scrap"

	// 10 parts produced (via cycle_count)
	for i := 0; i < 10; i++ {
		agg.ProcessReading(asset, "cycle_count", 1.0, "GOOD", now.Add(time.Duration(i)*time.Minute))
	}
	// 2 scrapped
	agg.ProcessReading(asset, "scrap_count", 2.0, "GOOD", now.Add(11*time.Minute))

	snap := agg.GetSnapshotForAsset(asset)
	if snap == nil {
		t.Fatal("expected non-nil snapshot")
	}

	// cycle_count: 10 actualOutput (performance metric)
	// scrap_count: 2 actualOutput + 2 defectiveOutput (quality defective metric)
	// Total actualOutput = 12, defective = 2, good = 10, quality = 10/12 = 83.3%
	if snap.TotalOutput != 12 {
		t.Fatalf("expected total output 12, got: %d", snap.TotalOutput)
	}
	if snap.GoodOutput != 10 {
		t.Fatalf("expected good output 10, got: %d", snap.GoodOutput)
	}
	if snap.Quality < 0.82 || snap.Quality > 0.84 {
		t.Fatalf("expected quality ~0.833, got: %.4f", snap.Quality)
	}
}

func TestOEEAggregator_ProcessReadingEmptyAssetID(t *testing.T) {
	cfg := DefaultOEEConfig()
	cfg.SnapshotInterval = 1 * time.Hour
	agg := NewOEEAggregator(cfg, nil)

	// Should not panic or create state for empty asset ID
	agg.ProcessReading("", "machine_status", 1.0, "GOOD", time.Now())

	if len(agg.states) != 0 {
		t.Fatalf("expected no states for empty asset ID, got: %d", len(agg.states))
	}
}

func TestOEEAggregator_DefaultConfig(t *testing.T) {
	// Pass zero-value config to exercise the default-filling branches in NewOEEAggregator.
	agg := NewOEEAggregator(OEEConfig{}, nil)

	if agg.config.WindowDuration != 1*time.Hour {
		t.Fatalf("expected default WindowDuration 1h, got: %v", agg.config.WindowDuration)
	}
	if agg.config.SnapshotInterval != 15*time.Second {
		t.Fatalf("expected default SnapshotInterval 15s, got: %v", agg.config.SnapshotInterval)
	}
	if agg.config.DefaultIdealCycleTime != 30*time.Second {
		t.Fatalf("expected default DefaultIdealCycleTime 30s, got: %v", agg.config.DefaultIdealCycleTime)
	}

	// Verify it works end-to-end with defaults
	now := time.Now()
	agg.ProcessReading("asset-default", "machine_status", 1.0, "GOOD", now)
	snap := agg.GetSnapshotForAsset("asset-default")
	if snap == nil {
		t.Fatal("expected non-nil snapshot with default config")
	}
	if snap.Availability != 1.0 {
		t.Fatalf("expected availability 1.0, got: %.4f", snap.Availability)
	}
}

func TestOEEAggregator_DefaultHandler(t *testing.T) {
	// Pass nil handler to exercise the defaultOEEHandler branch.
	cfg := OEEConfig{
		WindowDuration:        1 * time.Hour,
		SnapshotInterval:      50 * time.Millisecond,
		DefaultIdealCycleTime: 30 * time.Second,
	}
	agg := NewOEEAggregator(cfg, nil) // nil handler → defaultOEEHandler

	now := time.Now()
	agg.ProcessReading("asset-handler", "machine_status", 1.0, "GOOD", now)
	agg.ProcessReading("asset-handler", "cycle_count", 5.0, "GOOD", now)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	agg.Start(ctx)
	defer agg.Stop()

	// Wait for at least one snapshot cycle to trigger defaultOEEHandler
	time.Sleep(200 * time.Millisecond)
}

func TestOEEAggregator_DefectPattern(t *testing.T) {
	cfg := DefaultOEEConfig()
	cfg.SnapshotInterval = 1 * time.Hour
	agg := NewOEEAggregator(cfg, nil)

	now := time.Now()
	asset := "asset-defect"

	// "defect_count" should match isDefectiveMetric via "defect" pattern
	agg.ProcessReading(asset, "cycle_count", 10.0, "GOOD", now)
	agg.ProcessReading(asset, "defect_count", 3.0, "GOOD", now.Add(time.Minute))

	snap := agg.GetSnapshotForAsset(asset)
	if snap == nil {
		t.Fatal("expected non-nil snapshot")
	}

	// cycle_count: 10 actualOutput. defect_count: 3 actualOutput + 3 defectiveOutput.
	// Total: 13 actual, 3 defective, 10 good, quality = 10/13 = 76.9%
	if snap.TotalOutput != 13 {
		t.Fatalf("expected total output 13, got: %d", snap.TotalOutput)
	}
	if snap.GoodOutput != 10 {
		t.Fatalf("expected good output 10, got: %d", snap.GoodOutput)
	}
	if snap.Quality < 0.76 || snap.Quality > 0.78 {
		t.Fatalf("expected quality ~0.769, got: %.4f", snap.Quality)
	}
}

func TestOEEAggregator_OEEAlertIntegration(t *testing.T) {
	var mu sync.Mutex
	var alerts []OEEAlert

	alertEval := NewOEEAlertEvaluator(DefaultOEEAlertRules(), func(alert OEEAlert) {
		mu.Lock()
		alerts = append(alerts, alert)
		mu.Unlock()
	}, 0) // no cooldown for test

	cfg := OEEConfig{
		WindowDuration:        5 * time.Minute,
		SnapshotInterval:      100 * time.Millisecond,
		DefaultIdealCycleTime: 1 * time.Second,
	}

	handler := func(snapshot OEESnapshot) {} // no-op, we care about alerts
	agg := NewOEEAggregator(cfg, handler)
	agg.SetOEEAlertEvaluator(alertEval)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	agg.Start(ctx)

	// Simulate a degraded asset: lots of downtime (value=0 means not running).
	now := time.Now()
	for i := 0; i < 10; i++ {
		agg.ProcessReading("asset-bad", "machine_status", 0, "GOOD", now.Add(time.Duration(i)*time.Second))
	}

	// Wait for at least one snapshot cycle to fire.
	time.Sleep(300 * time.Millisecond)

	mu.Lock()
	alertCount := len(alerts)
	mu.Unlock()

	if alertCount == 0 {
		t.Fatal("expected at least one OEE alert for degraded asset")
	}

	agg.Stop()
}
