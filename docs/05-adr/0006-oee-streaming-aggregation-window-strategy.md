# ADR-0006: OEE Streaming Aggregation — Fixed Window with Hard Reset

* **Status:** Accepted
* **Date:** 2026-08-11
* **Authors:** @raydoan

---

## 1. Context

The real-time OEE (Overall Equipment Effectiveness) streaming aggregator in `services/analytics-engine/processor/oee.go` needs a windowing strategy to compute rolling Availability × Performance × Quality metrics per physical asset.

Key requirements:
- Process high-frequency sensor readings (sub-second to second intervals) inline from the Kafka consumer.
- Maintain per-asset counters (running ticks, actual output, defective output) within a configurable time window.
- Handle gaps in telemetry data gracefully (network disconnection, sensor failure, maintenance windows).

Two primary windowing strategies exist:
1. **Fixed window with hard reset:** A single mutable state per asset. When the window expires, all counters are zeroed and the window restarts from the next reading.
2. **Sliding window (ring buffer):** A bounded buffer of per-tick samples. Old samples are evicted as new ones arrive, maintaining a continuously sliding view.

---

## 2. Decision

Use a **fixed window with hard reset** (O(1) memory and O(1) per-reading processing).

When a reading arrives with a timestamp more than `WindowDuration` after the current window start, the aggregator resets all counters to zero and starts a new window from that reading's timestamp.

---

## 3. Consequences

### Positive Impacts
* **O(1) per reading:** No buffer management, no eviction, no iteration over historical samples. Each `ProcessReading` call is a single counter increment.
* **O(1) memory per asset:** One `assetOEEState` struct per tracked asset, regardless of window size or reading frequency.
* **Simple implementation:** No ring buffer, no circular index, no sample timestamp tracking.
* **Natural gap handling:** A gap longer than the window is indistinguishable from a fresh start, which is the correct semantic — a machine that was offline for an hour should not carry stale running-state data into the next window.

### Negative Impacts & Trade-offs
* **History loss on gap:** A single reading arriving after a gap > `WindowDuration` discards all prior counters. For a 1-hour window, a 61-minute connectivity loss means the next window starts from scratch.
* **No partial window data:** Cannot answer "what was OEE in the last 30 minutes?" — only "what is OEE for the current window?"
* **Sensitivity to gap timing:** Two scenarios with identical total production can yield different OEE if one has a mid-window gap that triggers a reset.

---

## 4. Alternatives Considered

* **Sliding window (ring buffer of per-tick samples):** Would provide smooth, continuously-updating OEE without history loss on gaps. Rejected because: (a) O(n) memory per asset where n = readings per window, which at 1Hz × 1 hour = 3600 samples per asset; (b) adds complexity to an inline consumer hot-path; (c) the manufacturing use case favors simplicity and predictable resource usage over sub-window granularity. Can be revisited if users require partial-window queries.

* **Database-query-based OEE (query TimescaleDB on demand):** Would eliminate in-memory state entirely by computing OEE from raw telemetry rows. Rejected because: (a) latency — each OEE query would scan thousands of rows; (b) puts load on TimescaleDB that scales with number of OEE consumers; (c) defeats the purpose of a real-time streaming aggregator. The in-memory approach provides sub-millisecond `ProcessReading` calls.

---

## 5. References

* `services/analytics-engine/processor/oee.go` — implementation
* `services/analytics-engine/config/config.go` — `OEEConfig.WindowDuration` (default: 1h)
* [EPIC-002: Resource Management & Telemetry](../09-epics/EPIC-002-Resource-Management-Telemetry.md)
