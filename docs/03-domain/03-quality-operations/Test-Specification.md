# Quality Operations — Test Specification

> **ISA-95 Reference:** Part 3, §6 — Quality Test Definition

## 1. Overview
A **Test Specification** defines the quality criteria that must be evaluated at a specific Operation or Work Unit. It is the master template from which Inspection instances are created at runtime.

---

## 2. Entity Definitions

| Entity | Description |
|---|---|
| **TestSpecification** | A named quality checklist assigned to a specific Operation or Work Unit |
| **TestCriteria** | An individual check within a specification (e.g., "Torque value between 45–55 Nm") |
| **AcceptanceCriteria** | The pass/fail threshold for a TestCriteria (numeric range, boolean, or enumerated value) |
| **SamplingPlan** | Defines frequency of testing: 100% inspection, AQL sampling, or skip-lot |

---

## 3. Specification Types

| Type | Description |
|---|---|
| **Inline** | Performed by the Operator at the Station during production |
| **End-of-Line** | Performed by a QC Inspector after all Operations are complete |
| **Receiving** | Performed on incoming material lots before they enter the production floor |
| **First Article** | Full inspection of the first unit produced on a new routing or after a tool change |

---

## 4. Key Events

| Event | Description |
|---|---|
| `quality.test_specification.published` | A new specification version is approved and activated |
| `quality.test_specification.revised` | An existing spec is updated; active inspections reference the new version |
