[← Index](README.md)

# Part 0 — Before you start

### Who this is for

You are comfortable with Java — classes, generics, records, exceptions, streams. You are
**new to Spring Boot and Maven**, or new enough that you would rather have things explained
than assumed. Every Spring and Maven concept you need is introduced where it first appears.

### What you will build

Two shared library modules, and the wiring that makes them active in `production-service`:

| Module | Responsibility |
|---|---|
| `libs/factoryos-common-error` | The error taxonomy and the domain error base. **No Spring, no gRPC.** |
| `libs/factoryos-common-error-adapters` | Two adapters that project the taxonomy onto HTTP and gRPC. |

Then you will convert real, existing failures in `production-service` from "HTTP 500 with a
stack trace" into properly classified client errors — the ones ADR-0007 was written to fix.

You implement the code. This tutorial is the guide.

### Prerequisites

- **JDK 21** — check with `java -version`
- **Maven 3.9+** — check with `mvn -v`

Every command in this tutorial runs **offline** (`-o`). The Maven cache in this development
environment is already complete, so nothing will try to reach the network. If you are on a
different machine and `-o` fails, drop the flag and Maven will download what it needs.

### The four defects this fixes

This is not a refactor for its own sake. ADR-0007 was written because of concrete,
observed bugs. Verbatim from [ADR-0007 §1](../../05-adr/0007-global-error-handling-standard.md):

> Consequences observed in the codebase:
>
> * A malformed path or query parameter escaped as an HTTP 500 with a default framework body.
> * An out-of-range page size produced a bare 400 carrying no machine-readable code.
> * A bad pagination cursor produced an HTTP 500, because nothing classified it as a client error.
>
> Each is a *user error reported as a server error*, and none carries a code a client can branch on.

Read that middle line again: **each is a user error reported as a server error.** A client
that sends `?limit=500` is told the *server* broke. It will retry, because that is what you
do with a 500. It will retry forever, because `limit=500` will never become valid.

By the end of this tutorial, all four are fixed, and you will have tests that prove it.

### How to build and test

```bash
cd services && mvn -o -B test        # all Java services
cd services && mvn -o -B validate    # reactor sanity only — fast, no compilation
```

> **Two ways to run the Java suite.** `make test` from the repository root runs everything —
> Go, the SDK, and the Java reactor (via `test-java`, which is `mvn -B verify` with the
> JaCoCo gate). For the Java side alone, offline, skip `make` and run `mvn` from `services/`
> as shown above.

---


---

[Index](README.md) · [Part 1 — The core idea →](01-the-core-idea.md)
