# FactoryOS - Developer Inbox & Idea Backlog

Welcome to the **FactoryOS Inbox**. This file serves as a light-weight "Parking Lot" / GTD Inbox to quickly capture raw ideas, unsorted tasks, technical debt, and future enhancements before they are triaged into formal specifications (RFCs, ADRs, or Epics).

---

## 📥 Quick Capture (Unsorted Ideas & Tasks)

> *Add new items here as soon as you think of them. No strict structure required!*

- [ ] **[YYYY-MM-DD]** Example: Research using NATS JetStream key-value store for edge state synchronization.
- [ ] 

---

## 💡 Feature & Product Ideas

- [ ] **[2026-08-07]** **[Edge Platform]** **Cloud-Based Edge Fleet Management & Device Authentication/Anti-Spoofing Engine**
  - **Fleet Monitoring & Heartbeat:** Centralized Cloud dashboard/control plane to track all active Edge Runtimes in real-time (how many Edge nodes are active, online/offline status, runtime version, SQLite buffer depth, CPU/RAM stats).
  - **Edge Authentication & Anti-Fake Security:** Prevent unauthorized/fake Edge Runtimes (anyone downloading the binary and spoofing telemetry). Implement 1-time Activation Key provisioning, mTLS (X.509 device certificate issuance), or TPM 2.0 hardware binding so only authorized Edge IPCs can register and publish telemetry to Cloud Kafka.

---

## 🛠️ Technical Debt & Refactoring

- [ ] **[YYYY-MM-DD]** **[Service/Platform]** Title/Description

---

## 🔬 Spikes & Research Topics

- [ ] **[YYYY-MM-DD]** Title/Description

---

## 🔄 Triage & Lifecycle Guide

When an item in this Inbox is ready to be acted upon, follow the [FactoryOS Development Lifecycle](AGENTS.md):

1. **Idea / Raw Task:** Logged in `INBOX.md`.
2. **Architecture / Design Decision:** If it involves major technical changes, create an ADR in `docs/05-adr/`.
3. **Feature Specification:** If it's a new feature, create an RFC in `docs/06-rfc/`.
4. **Actionable Implementation:** Breakdown into an Epic checklist under `docs/09-epics/`.
5. **Clean Up:** Mark completed or promoted items in `INBOX.md` as `[x]` or move them to archive.
