# Error Handling — Implementation Tutorial

> A hands-on guide to implementing [ERROR_HANDLING.md](../ERROR_HANDLING.md) in the
> FactoryOS Java services. The standard defines the *contract*; this tutorial walks you
> through building it, file by file.
>
> **Status:** companion to [ADR-0007](../../05-adr/0007-global-error-handling-standard.md).

**Who this is for:** you are comfortable with Java, and **new to Spring Boot and Maven**.
Every framework concept you need is explained where it first appears.

**What you build:** two shared library modules, and the wiring that makes them active in
`production-service`.

**You implement the code.** This tutorial is the guide — it contains every file you need to
create, and the command to verify each step.

---

## Read in order

Each part ends with a **"Check yourself"** block — a command to run and the output that means
it worked. Do not skip those; they catch mistakes at the step that caused them.

| # | Part | What you do |
|---|---|---|
| 0 | [Before you start](00-before-you-start.md) | Prerequisites, the four defects this fixes, how to build |
| 1 | [The core idea](01-the-core-idea.md) | Why a domain error carries no HTTP status |
| 2 | [Maven ground floor](02-maven-ground-floor.md) | Modules, parent POMs, devcontainer profiles |
| 3 | [The taxonomy module](03-taxonomy-module.md) | Six classes, and why they import no transport type |
| 4 | [The HTTP adapter](04-http-adapter.md) | RFC 9457 Problem Details, the HTTP projection, traceId |
| 5 | [The gRPC adapter](05-grpc-adapter.md) | `google.rpc.Status`, the gRPC projection, and the two traps |
| 6a | [Adopting it — domain side](06a-domain-adoption.md) | Reclassify the real failures in `production-service` |
| 6b | [Adopting it — transport side](06b-transport-adoption.md) | Register the adapters; prove it live |
| 7 | [Proving the taxonomy](07-proving-the-taxonomy.md) | The test suites the standard requires |
| 8 | [Rolling it out](08-rolling-it-out.md) | The other services, EPIC-001, CHANGELOG |

**Parts 3–5 build the libraries.** Nothing there affects `production-service` until Part 6b.

---

## The one rule everything follows

From [ERROR_HANDLING.md §1](../ERROR_HANDLING.md):

> **The domain layer classifies an error by its *category*. A transport adapter maps category → transport code.**

An error is classified **once**, semantically, by the layer that understands it. It is then
*projected* onto HTTP and gRPC by adapters. Nothing in the domain layer knows which protocol
delivered the request.

---

## Reference

- [ERROR_HANDLING.md](../ERROR_HANDLING.md) — the standard being implemented. **The source of truth for every rule.**
- [ADR-0007](../../05-adr/0007-global-error-handling-standard.md) — the decision and its rationale
- [PAGINATION_DESIGN.md](../PAGINATION_DESIGN.md) §5.5 — the cursor error codes
- [`api/contracts/openapi/common/v1/schemas/errors.yaml`](../../../api/contracts/openapi/common/v1/schemas/errors.yaml) — the REST error schema

---

## Commands you will use

```bash
cd services && mvn -o -B test                                    # all Java services
cd services && mvn -o -B clean test                              # after changing dependencies
cd services && mvn -o -B test -pl ../libs/factoryos-common-error -am
```

> **`make test` from the repository root runs both sides.** It fans out to `test-go` and
> `test-java`, and `test-java` is `cd services && mvn -B verify` — so it also enforces the
> JaCoCo coverage gate. Use it when you want everything at once; use `mvn -o -B test` from
> `services/` when you want the Java suite alone, offline, without the coverage gate.

---

## A note on assumptions

Everything here was verified against this repository on **2026-09-17** — the code compiles, the
tests pass, and the captured HTTP and gRPC responses are real output from a running service.
Where a command needs network access, it is called out. Where something cannot be verified in
this environment, it says so rather than guessing.

If a step does not behave as described, that is a documentation bug worth reporting, not
something you did wrong.
