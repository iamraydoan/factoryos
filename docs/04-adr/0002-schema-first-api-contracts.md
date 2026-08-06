# ADR 0002: Schema-First API Contracts Governance

* **Status:** Accepted
* **Deciders:** Founding CTO, Principal Software Architect
* **Date:** 2026-08-05

## Context & Problem Statement
With multiple engineering teams building microservices across cloud and edge runtimes, uncoordinated API changes lead to runtime serialization errors, subtle data corruption, and broken integrations.

## Decision Outcome
Enforce **Schema-First API Contracts** managed in a dedicated repository (`api-contracts`).

* **gRPC / Protocol Buffers** for synchronous cross-service RPCs and internal control plane APIs.
* **AsyncAPI / JSON Schema** for event streaming catalog payloads.
* **OpenAPI 3.0** for public-facing developer REST APIs.

### Code Generation & CI Enforcement
- CI pipelines automatically validate backward compatibility on pull requests using linter rules (`buf breaking`, `spectral`).
- Platform SDKs in Go, TypeScript, and Python are generated directly from versioned contract releases.
