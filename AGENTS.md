# FactoryOS — Agent Guide

Event-driven manufacturing operations platform (IIoT / MES / MOM). Shop-floor systems that must
survive network partitions and ingest high-frequency signals.

**Polyglot monorepo:** a Go workspace (`go.work`) *and* a separate Maven reactor (`services/pom.xml`).
They build, test, and version independently — check which side you are on before running anything.

---

## Golden lifecycle — HARD GATE

Do **not** write feature code until steps 1–5 are satisfied. The schema-first mandate, not a suggestion.

1. **Roadmap** — confirm the active milestone in `docs/08-roadmap/MILESTONES.md`.
2. **RFC** — feature spec in `docs/06-rfc/` (template `0000-rfc-template.md`).
3. **ADR** — if architectural, in `docs/05-adr/` (template `0000-adr-template.md`).
4. **Epic** — actionable task in `docs/09-epics/`.
5. **Contract first** — schema in `api/contracts/`. No endpoint, consumer, or message type exists
   before its contract.
6. **Implement + test**, then tick `[x]` in the Epic checklist.

A bug fix or a refactor does not need a new RFC. Anything that adds or changes behaviour does.

---

## Commands

**Prefer `make`.** The targets below are the interface; the raw `go` / `mvn` commands are their
implementations. Run `make help` for the full list.

### Test & build

```bash
make test              # EVERYTHING — Go modules + the Java/Maven reactor
make test-all          # alias of the above
make test-go           # Go workspace modules only
make test-java         # Java reactor only
make test-java-service SERVICE=production-service   # one Java service
make test-coverage     # Go coverage reports

make build-all         # Go binaries -> bin/
make run-production    # sources .env, then spring-boot:run
```

### Go

Workspace members: `platform/{edge-runtime,platform-sdk}`, `services/{analytics-engine,
ingestion-service,resource-service}`, `examples/mock-plc-simulator`.

```bash
cd services/resource-service && go test -race ./...          # one module, direct
# analytics-engine is scoped, not ./...:
cd services/analytics-engine && go test -race ./processor/... ./db/... ./consumer/... ./config/...
```

### Java

Reactor members: `services/{production,warehouse,quality,maintenance,planning}-service`, plus the
shared `libs/factoryos-common-error*` modules.

```bash
cd services && mvn -pl production-service test      # direct equivalent
```

### Contracts & codegen

```bash
make proto-lint   proto-gen                       # buf (api/contracts/)
make openapi-lint openapi-bundle openapi-gen
make install-tools                                # buf, protoc-gen-go(-grpc), oapi-codegen, redocly
```

Contracts are the source of truth. **Never hand-edit generated files**
(`platform/platform-sdk/go/gen/**`) — change the contract and regenerate.

### Environment & infrastructure

Pick the profile matching *where the process runs*: `host` (DB on localhost), `docker` (container
names), `devcontainer`. Nothing starts without the matching `.env`.

```bash
make setup-env-host | setup-env-docker | setup-env-devcontainer   # required FIRST
make infra-up | infra-down | infra-ps | infra-logs                # Postgres/Kafka/MQTT/Traefik
make services-up | services-down | docs-up | docs-down
make help                                                          # lists everything
```

---

## Hard rules

- **Schema-first.** Never write an endpoint, consumer, or message without its contract in `api/contracts/`.
- **UUIDv7** for every entity primary key (time-ordered).
- **Event naming** — exactly `<domain>.<entity>.<past_tense_action>` (`production.work_order.completed`,
  `quality.ncr.raised`).
- **Bounded-context isolation.** No direct cross-service database queries — use events or gRPC.
- **Never swallow errors.** No silent fallbacks; log context, propagate status codes.
- **Edge stays offline-capable.** Code in `platform/edge-runtime` must work disconnected via its local
  SQLite store-and-forward buffer.
- **Errors** — RFC 9457 `application/problem+json` with a stable `code`. `400` = body cannot be parsed,
  `422` = parses but is semantically invalid.
- **Wire conventions** — camelCase; collections use the `{ data, meta }` envelope.
- **Never declare `spring-boot-starter-parent` in a service** — the parent pom pins one Spring Boot for
  the whole reactor; a service declaring its own resolves a different Boot major than its siblings.
- **Commits** — Conventional Commits (`feat(production): …`). Branches `feature/<domain>-<desc>`,
  `fix/<domain>-<desc>`.
- **Coverage** — 80% (Go, CI-gated) and 60% (Java, JaCoCo `mvn verify`, generated protobuf excluded).
  Both are floors to raise, not ceilings.

---

## Gotchas

- **Spring Boot 4.1.1 / Java 21 is bleeding-edge.** Artifacts renamed and split:
  `spring-boot-starter-webmvc` (not `-web`), per-slice `*-test` starters, Spring gRPC (not `net.devh`).
  **Do not copy Spring Boot 3 patterns** — verify against the actual pom.
- **Java tests need the `test` profile.** `@SpringBootTest` classes must add `@ActiveProfiles("test")`
  or they dial real Postgres (`localhost:5432`) and fail. `src/test/resources/application-test.yml`
  supplies H2 + gRPC port 0.
- **Go versions drift.** `go.work` says `1.26.5`; CI pins `1.22`; modules range `1.22`–`1.26.5`.
- **Codegen instructions conflict.** `CONTRIBUTING.md` says to commit regenerated SDK code, but
  `.gitignore` ignores `platform/platform-sdk/go/gen/` (0 files tracked). Either way, don't hand-edit.
- **Formatting is implicit** — no `.editorconfig`, `golangci-lint`, `checkstyle`, `spotless`, or
  `prettier`. Match surrounding style; `gofmt` for Go.
- **The bundled OpenAPI file is what codegen consumes**; the gateway-overlay variant is Swagger UI only
  (ADR-0008).

---

## Architecture map

```
platform/    edge-runtime (Go, OPC-UA/MQTT + offline buffer) · platform-sdk (Go/Java/TS + generated)
             workflow-engine (durable state machine)
services/    Go:   analytics-engine, ingestion-service, resource-service
             Java: production, warehouse, quality, maintenance, planning
libs/        factoryos-common-error + -adapters (shared error taxonomy, Maven modules)
api/         contracts/ (Protobuf, OpenAPI, AsyncAPI) · architecture/ (C4, threat models)
examples/    mock-plc-simulator (Go)
deploy/      traefik/ · mosquitto/ · terraform/ · helm/
docs/        00-governance … 10-developer-guide (numbered hierarchy)
```

Traefik fronts REST (`/api/v1/...`) and gRPC. Storage: TimescaleDB/PostgreSQL. Messaging: Kafka
(events, transactional outbox) + Mosquitto MQTT (edge).

---

## Read these, don't guess

| Topic | File |
|---|---|
| **Governance & principles (primary)** | `docs/00-governance/PROJECT_BIBLE.md` |
| Contribution workflow, commits, PR gates | `docs/10-developer-guide/CONTRIBUTING.md` |
| Local setup | `docs/10-developer-guide/Local-Environment-Setup.md` |
| Error handling | `docs/07-api/ERROR_HANDLING.md` |
| Pagination / query / events | `docs/07-api/` (`PAGINATION_DESIGN`, `QUERY_CONVENTIONS`, `ASYNC_EVENT_CATALOG`) |
| Full docs index | `docs/README.md` |
| Ideas / parking lot | `INBOX.md` |
| House REST conventions | `.agents/skills/api-design/SKILL.md` |

Nothing above is authoritative if it contradicts `PROJECT_BIBLE.md`.
