# Engineering Contributing Guide

Welcome to the FactoryOS engineering repository. This document outlines development standards and workflows for all software engineers.

---

## 1. Core Principles

1. **Schema-First:** Never write a microservice endpoint without defining its contract in `api-contracts` first.
2. **Event-Driven First:** Prefer asynchronous events over synchronous gRPC RPCs for cross-domain side effects.
3. **Fail-Safe Edge Design:** Code targeting the shop floor or `edge-runtime` must operate seamlessly offline.
4. **Zero Code Duplication:** Shared context, logging, tracing, and event wrappers belong in `platform-sdk`.

---

## 2. Commit & Branch Conventions

* Branch Naming: `feature/<domain>-<short-description>`, `fix/<domain>-<short-description>`
* Commit Messages: Follow Conventional Commits specification:
  * `feat(production): add station hold trigger event`
  * `fix(edge): resolve OPC-UA reconnection backoff loop`
  * `docs(adr): add ADR 0003 for durable workflow engine`

---

## 3. Pull Request Requirements

* All PRs must pass automated linting (`buf lint`, linter checks).
* Architectural changes must include an updated or new ADR file under `docs/03-adr/`.
* Public API changes must demonstrate backward compatibility verification.
