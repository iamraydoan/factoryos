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
- **Local Generation:** Developers run `buf generate` locally to generate Go and Java SDK classes/structs directly into `platform/platform-sdk/` and commit them to Git.
- **CI Validation:** CI pipelines validate backward compatibility and format rules (`buf lint`, `buf breaking`) and execute `buf generate && git diff --exit-code` to ensure committed SDK code is 100% synchronized with `.proto` definitions.
