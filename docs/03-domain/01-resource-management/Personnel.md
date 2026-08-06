# Resource Management — Personnel

> **ISA-95 Reference:** Part 2, Section 6 — Personnel Object Model

## 1. Overview
Defines the human resource model: roles, skill certifications, and shift assignments. Personnel data is consumed by Production Operations (dispatch) and Maintenance Operations (technician assignment).

---

## 2. Entity Definitions

| Entity | ISA-95 Term | Description |
|---|---|---|
| **PersonClass** | Personnel Class | A role category defining required skills (e.g., "Certified CNC Operator") |
| **Person** | Person | An individual employee with assigned roles and certifications |
| **QualificationRecord** | Qualification | A certification linking a Person to a PersonClass for a specific Work Center |
| **Shift** | — | A named working period (Day / Evening / Night) with defined start/end times |
| **ShiftAssignment** | — | Assignment of a Person to a Shift and Work Unit for a given date |

---

## 3. Qualification Model
A Person may only execute an Operation at a Work Unit if they hold a valid **QualificationRecord** for the relevant PersonClass. This enforces:
- ISO / IATF compliance (only trained operators run certified processes).
- Automatic blocking of unqualified dispatch in the Production Operations module.

---

## 4. Key Events

| Event | Description |
|---|---|
| `resource.person.qualified` | A Person receives a new qualification for a PersonClass |
| `resource.person.qualification_expired` | A qualification reaches its expiry date |
| `resource.shift_assignment.created` | A Person is assigned to a Shift and Work Unit |
