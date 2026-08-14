package processor

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"
)

// OEEConfig holds tuning parameters for the streaming OEE aggregator.
type OEEConfig struct {
	// WindowDuration is the rolling window for OEE computation (default: 1 hour).
	WindowDuration time.Duration
	// SnapshotInterval is how often OEE snapshots are computed and emitted (default: 15s).
	SnapshotInterval time.Duration
	// DefaultIdealCycleTime is the fallback cycle time when no per-asset override is set.
	DefaultIdealCycleTime time.Duration
}

// DefaultOEEConfig returns sensible defaults for OEE aggregation.
func DefaultOEEConfig() OEEConfig {
	return OEEConfig{
		WindowDuration:        1 * time.Hour,
		SnapshotInterval:      15 * time.Second,
		DefaultIdealCycleTime: 30 * time.Second,
	}
}

// OEESnapshot is a point-in-time OEE reading for a single Physical Asset.
type OEESnapshot struct {
	PhysicalAssetID string        `json:"physical_asset_id"`
	Timestamp       time.Time     `json:"timestamp"`
	WindowStart     time.Time     `json:"window_start"`
	WindowDuration  time.Duration `json:"window_duration"`
	Availability    float64       `json:"availability"`
	Performance     float64       `json:"performance"`
	Quality         float64       `json:"quality"`
	OEE             float64       `json:"oee"`
	RunTime         time.Duration `json:"run_time"`
	PlannedTime     time.Duration `json:"planned_time"`
	TotalOutput     int64         `json:"total_output"`
	GoodOutput      int64         `json:"good_output"`
}

// assetOEEState tracks the raw counters for a single Physical Asset within the rolling window.
type assetOEEState struct {
	mu              sync.Mutex
	windowStart     time.Time
	totalTicks      int64 // total sensor readings received (each = 1 tick interval)
	runningTicks    int64 // ticks where machine was "running"
	actualOutput    int64 // total parts produced
	defectiveOutput int64 // parts flagged as defective
	lastReadingTime time.Time
}

// OEEHandlerFunc is a callback invoked each time an OEE snapshot is computed.
type OEEHandlerFunc func(snapshot OEESnapshot)

// OEEAggregator maintains per-asset in-memory state and computes rolling OEE
// from streaming sensor readings. It follows the same pipeline pattern as
// AlertEvaluator — called inline from the consumer for each reading.
type OEEAggregator struct {
	mu              sync.RWMutex
	states          map[string]*assetOEEState
	idealCycleTime  map[string]time.Duration
	config          OEEConfig
	handler         OEEHandlerFunc // set via constructor or SetHandler; never nil

	done chan struct{}
	wg   sync.WaitGroup
}

// NewOEEAggregator creates a new aggregator with the given config and optional callback.
func NewOEEAggregator(cfg OEEConfig, handler OEEHandlerFunc) *OEEAggregator {
	if cfg.WindowDuration <= 0 {
		cfg.WindowDuration = 1 * time.Hour
	}
	if cfg.SnapshotInterval <= 0 {
		cfg.SnapshotInterval = 15 * time.Second
	}
	if cfg.DefaultIdealCycleTime <= 0 {
		cfg.DefaultIdealCycleTime = 30 * time.Second
	}
	if handler == nil {
		handler = defaultOEEHandler
	}
	return &OEEAggregator{
		states:         make(map[string]*assetOEEState),
		idealCycleTime: make(map[string]time.Duration),
		config:         cfg,
		handler:        handler,
		done:           make(chan struct{}),
	}
}

func defaultOEEHandler(snapshot OEESnapshot) {
	log.Printf("[OEE] Asset: %s | OEE: %.1f%% (A=%.1f%% P=%.1f%% Q=%.1f%%) | Window: %v",
		snapshot.PhysicalAssetID,
		snapshot.OEE*100,
		snapshot.Availability*100,
		snapshot.Performance*100,
		snapshot.Quality*100,
		snapshot.WindowDuration)
}

// SetIdealCycleTime overrides the ideal cycle time for a specific asset.
// This is used for performance calculation: Performance = (Output * IdealCycleTime) / RunTime.
func (a *OEEAggregator) SetIdealCycleTime(assetID string, d time.Duration) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.idealCycleTime[assetID] = d
}

// SetHandler replaces the snapshot callback handler.
// Pass nil to restore the default log-based handler.
func (a *OEEAggregator) SetHandler(handler OEEHandlerFunc) {
	if handler == nil {
		handler = defaultOEEHandler
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.handler = handler
}

// Start launches a background goroutine that periodically computes and emits OEE snapshots.
func (a *OEEAggregator) Start(ctx context.Context) {
	a.wg.Add(1)
	go a.snapshotLoop(ctx)
	log.Println("[OEE] Streaming OEE aggregator started.")
}

func (a *OEEAggregator) snapshotLoop(ctx context.Context) {
	defer a.wg.Done()

	ticker := time.NewTicker(a.config.SnapshotInterval)
	defer ticker.Stop()

	for {
		select {
		case <-a.done:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			a.computeAndEmitAll()
		}
	}
}

func (a *OEEAggregator) computeAndEmitAll() {
	// Collect snapshot of all asset IDs under read lock.
	a.mu.RLock()
	assetIDs := make([]string, 0, len(a.states))
	for id := range a.states {
		assetIDs = append(assetIDs, id)
	}
	a.mu.RUnlock()

	for _, id := range assetIDs {
		snapshot := a.GetSnapshotForAsset(id)
		if snapshot != nil && a.handler != nil {
			a.handler(*snapshot)
		}
	}
}

// Stop signals the background goroutine and waits for it to exit.
func (a *OEEAggregator) Stop() {
	close(a.done)
	a.wg.Wait()
	log.Println("[OEE] OEE aggregator stopped.")
}

// getOrCreateState returns the state for an asset, creating it if needed.
func (a *OEEAggregator) getOrCreateState(assetID string, ts time.Time) *assetOEEState {
	a.mu.Lock()
	defer a.mu.Unlock()

	state, exists := a.states[assetID]
	if !exists {
		state = &assetOEEState{
			windowStart: ts,
		}
		a.states[assetID] = state
	}
	return state
}

// resetIfStale resets the state if the window has elapsed.
//
// Design note: Fixed window with hard reset (O(1) per reading) rather than a sliding
// window. See ADR-0006 for the full rationale.
func (a *OEEAggregator) resetIfStale(state *assetOEEState, ts time.Time) {
	state.mu.Lock()
	defer state.mu.Unlock()

	if ts.Sub(state.windowStart) > a.config.WindowDuration {
		state.windowStart = ts
		state.totalTicks = 0
		state.runningTicks = 0
		state.actualOutput = 0
		state.defectiveOutput = 0
	}
}

// metricClass defines the OEE component a sensor reading maps to.
type metricClass int

const (
	metricUnclassified metricClass = iota // not matched to any OEE component
	metricAvailability                     // machine running state indicators
	metricPerformance                      // production output / cycle counters
	metricQualityGood                      // good output counters (non-defective)
	metricQualityDefective                 // defective output counters (reject, scrap, defect)
)

// classifyMetric determines which OEE component a metric name belongs to.
//
// Priority order (highest to lowest):
//  1. Availability — status/running/operational/state/uptime/downtime/machine_status
//  2. Quality (defective) — reject/scrap/defect
//  3. Quality (good) — good/quality (but NOT reject/scrap/defect)
//  4. Performance — cycle/output/parts/produced/count
//  5. Uncategorized — everything else (counts as output, quality from sensor flag)
//
// This explicit ordering prevents ambiguity when a metric name matches multiple
// patterns (e.g., "reject_count" matches both performance and quality).
func classifyMetric(name string) metricClass {
	// 1. Availability has highest priority — machine state is foundational.
	if isAvailabilityMetric(name) {
		return metricAvailability
	}

	// 2. Quality defective: "reject", "scrap", "defect" — checked before general quality
	//    and before performance, since "reject_count" would otherwise match "count".
	if isDefectiveMetric(name) {
		return metricQualityDefective
	}

	// 3. Quality good: "good", "quality" — non-defective quality indicators.
	if isQualityGoodMetric(name) {
		return metricQualityGood
	}

	// 4. Performance: "cycle", "output", "parts", "produced", "count".
	if isPerformanceMetric(name) {
		return metricPerformance
	}

	// 5. Uncategorized: not matched by any pattern above.
	return metricUnclassified
}

// ProcessReading classifies a single sensor reading and updates the per-asset OEE state.
// This is called inline from the consumer for each sensor reading, same as AlertEvaluator.
func (a *OEEAggregator) ProcessReading(assetID, metricName string, value float64, quality string, ts time.Time) {
	if assetID == "" {
		return
	}

	state := a.getOrCreateState(assetID, ts)
	a.resetIfStale(state, ts)

	state.mu.Lock()
	defer state.mu.Unlock()

	metricLower := strings.ToLower(metricName)
	class := classifyMetric(metricLower)

	switch class {
	case metricAvailability:
		state.totalTicks++
		if isRunningValue(metricLower, value) {
			state.runningTicks++
		}

	case metricPerformance, metricQualityGood:
		state.actualOutput += int64(value)

	case metricQualityDefective:
		state.actualOutput += int64(value)
		state.defectiveOutput += int64(value)

	case metricUnclassified:
		// Uncategorized sensor reading — count as one unit of output.
		// The sensor quality flag determines whether it is good or defective.
		state.actualOutput++
		if strings.EqualFold(quality, "BAD") {
			state.defectiveOutput++
		}
	}

	state.lastReadingTime = ts
}

// isAvailabilityMetric checks if the metric relates to machine running state.
func isAvailabilityMetric(name string) bool {
	patterns := []string{"status", "running", "operational", "state", "uptime", "downtime", "machine_status"}
	for _, p := range patterns {
		if strings.Contains(name, p) {
			return true
		}
	}
	return false
}

// isRunningValue determines if the metric value indicates the machine is running.
// For numeric values: > 0 means running. For string-based status encoded as numbers:
// 1 = running, 0 = stopped. For binary flags: 1.0 = running.
func isRunningValue(metricName string, value float64) bool {
	// For "downtime" metrics, the semantics are inverted: value > 0 means NOT running.
	if strings.Contains(metricName, "downtime") {
		return value <= 0
	}
	return value > 0
}

// isPerformanceMetric checks if the metric relates to production output.
func isPerformanceMetric(name string) bool {
	patterns := []string{"cycle", "output", "parts", "produced", "count"}
	for _, p := range patterns {
		if strings.Contains(name, p) {
			return true
		}
	}
	return false
}

// isQualityGoodMetric checks if the metric represents good (non-defective) output.
func isQualityGoodMetric(name string) bool {
	patterns := []string{"good", "quality"}
	for _, p := range patterns {
		if strings.Contains(name, p) {
			return true
		}
	}
	return false
}

// isDefectiveMetric checks if the quality metric represents defective output.
func isDefectiveMetric(name string) bool {
	patterns := []string{"reject", "scrap", "defect"}
	for _, p := range patterns {
		if strings.Contains(name, p) {
			return true
		}
	}
	return false
}

// GetSnapshotForAsset computes and returns the current OEE snapshot for a single asset.
func (a *OEEAggregator) GetSnapshotForAsset(assetID string) *OEESnapshot {
	a.mu.RLock()
	state, exists := a.states[assetID]
	a.mu.RUnlock()

	if !exists {
		return nil
	}

	return a.computeOEE(assetID, state)
}

// GetSnapshots returns the latest OEE snapshot for all tracked assets.
func (a *OEEAggregator) GetSnapshots() []OEESnapshot {
	a.mu.RLock()
	assetIDs := make([]string, 0, len(a.states))
	for id := range a.states {
		assetIDs = append(assetIDs, id)
	}
	a.mu.RUnlock()

	snapshots := make([]OEESnapshot, 0, len(assetIDs))
	for _, id := range assetIDs {
		if s := a.GetSnapshotForAsset(id); s != nil {
			snapshots = append(snapshots, *s)
		}
	}
	return snapshots
}

// computeOEE performs the OEE calculation for a single asset.
// OEE = Availability × Performance × Quality
//
// Lock ordering: a.mu (read-only) is acquired FIRST, then state.mu.
// This matches ProcessReading which acquires a.mu (via getOrCreateState) then state.mu.
// Both paths acquire a.mu before state.mu, preventing deadlock.
func (a *OEEAggregator) computeOEE(assetID string, state *assetOEEState) *OEESnapshot {
	// Read ideal cycle time BEFORE acquiring state.mu to maintain consistent lock order.
	a.mu.RLock()
	idealCycle, hasIdeal := a.idealCycleTime[assetID]
	if !hasIdeal {
		idealCycle = a.config.DefaultIdealCycleTime
	}
	windowDuration := a.config.WindowDuration
	a.mu.RUnlock()

	state.mu.Lock()
	defer state.mu.Unlock()

	now := time.Now()

	snapshot := &OEESnapshot{
		PhysicalAssetID: assetID,
		Timestamp:       now,
		WindowStart:     state.windowStart,
		WindowDuration:  windowDuration,
		PlannedTime:     windowDuration,
		TotalOutput:     state.actualOutput,
	}

	// --- Availability ---
	// Run Time = (runningTicks / totalTicks) × Planned Production Time
	if state.totalTicks > 0 {
		runFraction := float64(state.runningTicks) / float64(state.totalTicks)
		snapshot.RunTime = time.Duration(runFraction * float64(windowDuration))
		snapshot.Availability = runFraction
	} else {
		// No data → default to 1.0 (no negative bias; degrade only with evidence)
		snapshot.Availability = 1.0
		snapshot.RunTime = windowDuration
	}

	// --- Performance ---
	// Performance = (ActualOutput × IdealCycleTime) / RunTime

	if snapshot.RunTime > 0 && state.actualOutput > 0 {
		idealOutput := float64(snapshot.RunTime) / float64(idealCycle)
		if idealOutput > 0 {
			snapshot.Performance = float64(state.actualOutput) / idealOutput
			if snapshot.Performance > 1.0 {
				snapshot.Performance = 1.0 // Cap at 100%
			}
		} else {
			snapshot.Performance = 1.0
		}
	} else {
		snapshot.Performance = 1.0
	}

	// --- Quality ---
	// Quality = GoodUnits / TotalUnits
	goodOutput := state.actualOutput - state.defectiveOutput
	if goodOutput < 0 {
		goodOutput = 0
	}
	snapshot.GoodOutput = goodOutput

	if state.actualOutput > 0 {
		snapshot.Quality = float64(goodOutput) / float64(state.actualOutput)
	} else {
		snapshot.Quality = 1.0
	}

	// --- OEE ---
	snapshot.OEE = snapshot.Availability * snapshot.Performance * snapshot.Quality

	return snapshot
}
