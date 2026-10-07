[Index](README.md) · [← Part 6a — Adopting it: the domain side](06a-domain-adoption.md)

# Part 6b — Adopting it: the transport side

The classification exists. Now the adapters have to be *active* in `production-service`.

This part has one subtle step. Skipping it gives you a service that compiles, starts, and
still returns 500s — with no error message to explain why.

### The step everyone misses: component scanning

Spring Boot finds your beans by scanning the package tree **underneath the application
class**. Open it:

```java
// services/production-service/src/main/java/com/factoryos/production/ProductionServiceApplication.java
package com.factoryos.production;

@SpringBootApplication
public class ProductionServiceApplication {
```

So Spring scans `com.factoryos.production.**` — controllers, repositories, your gRPC service.
All found, because they are under that package.

Now look at where the adapters live:

```
com.factoryos.production       ← the app scans here
com.factoryos.common.error     ← the adapters live HERE
```

**Different tree. Not scanned.** `ProblemDetailExceptionHandler`, `GrpcErrorAdvice` and
`TraceIdFilter` are sibling packages, so component scanning never reaches them. Your service
starts perfectly and then returns 500s for everything, because the advice classes were never
registered.

This is not speculation — the framework's own discovery mechanism confirms it. Spring gRPC
finds advice beans by asking the context:

```
GrpcAdviceDiscoverer → applicationContext.getBeansWithAnnotation(GrpcAdvice.class)
```

If the bean is not in the context, it is not found. And there is **no auto-configuration
file** in the adapters module, so nothing registers it for you.

### Register the adapters

Add an auto-configuration class to the adapters module. This is the correct fix for a library
used by several services — each one gets the adapters by adding a dependency, with no
per-service code.

Create
`libs/factoryos-common-error-adapters/src/main/java/com/factoryos/common/error/ErrorHandlingAutoConfiguration.java`:

```java
package com.factoryos.common.error;

import java.util.LinkedHashMap;
import java.util.Map;

import org.springframework.beans.factory.ObjectProvider;
import org.springframework.boot.autoconfigure.AutoConfiguration;
import org.springframework.boot.autoconfigure.condition.ConditionalOnClass;
import org.springframework.boot.autoconfigure.condition.ConditionalOnMissingBean;
import org.springframework.boot.autoconfigure.condition.ConditionalOnWebApplication;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

import io.grpc.Status;
import io.grpc.StatusException;

@AutoConfiguration
public class ErrorHandlingAutoConfiguration {

    @Configuration(proxyBeanMethods = false)
    @ConditionalOnClass(StatusException.class)
    static class GrpcAdapters {

        @Bean
        @ConditionalOnMissingBean
        GrpcStatusFactory grpcStatusFactory(ObjectProvider<GrpcStatusOverrideProvider> providers) {
            Map<String, Status.Code> merged = new LinkedHashMap<>();
            providers.orderedStream().forEach(p -> merged.putAll(p.overrides()));
            return new GrpcStatusFactory(merged);
        }

        @Bean
        @ConditionalOnMissingBean
        GrpcErrorAdvice grpcErrorAdvice(GrpcStatusFactory statusFactory) {
            return new GrpcErrorAdvice(statusFactory);
        }
    }

    @Configuration(proxyBeanMethods = false)
    @ConditionalOnClass(name = "org.springframework.web.servlet.DispatcherServlet")
    @ConditionalOnWebApplication
    static class WebAdapters {

        @Bean
        @ConditionalOnMissingBean
        ProblemDetailExceptionHandler problemDetailExceptionHandler() {
            return new ProblemDetailExceptionHandler();
        }

        @Bean
        @ConditionalOnMissingBean
        TraceIdFilter traceIdFilter() {
            return new TraceIdFilter();
        }
    }
}
```

**There is no override-table bean.** The merge happens inside `grpcStatusFactory`, so the table
never becomes part of the service's injectable bean surface — nothing else can collide with a
bare `Map<String, Status.Code>` or accidentally autowire the internal merge table. A service that
defines no provider simply merges nothing and every code falls back to its category default.
That is the fail-safe: no overrides, no surprise statuses.

**What `@AutoConfiguration` is.** It is a specialised `@Configuration` that Spring Boot loads
**automatically** from every jar on the classpath — no component scan required. This is how
Spring Boot itself works internally: `DataSourceAutoConfiguration`, `WebMvcAutoConfiguration`
and hundreds of others are loaded exactly this way.

**Why the conditions sit on nested classes.** `@ConditionalOnClass` means "only create these
beans if this class is present", and the official guidance is to put it on a nested
`@Configuration` rather than on each `@Bean` method. On Boot 4 a method-level condition also
works for these adapters — the class really is absent when the condition says so — but the
nested class is the shape Boot documents, and it means the whole protocol slice is skipped in
one place instead of the condition being restated on every bean in it. So a service with no
gRPC dependency does not get a `GrpcErrorAdvice` bean, and does not fail.
`@ConditionalOnMissingBean` lets a service override an adapter with its own if it ever needs to.

### Tell Spring Boot the class exists

Auto-configuration is discovered through a metadata file. Create
`libs/factoryos-common-error-adapters/src/main/resources/META-INF/spring/org.springframework.boot.autoconfigure.AutoConfiguration.imports`:

```
com.factoryos.common.error.ErrorHandlingAutoConfiguration
```

One fully-qualified class name per line. Without this file, `@AutoConfiguration` is just an
annotation nobody reads.

> **This is a Spring Boot 3+ file.** Older tutorials show `META-INF/spring.factories` with an
> `org.springframework.boot.autoconfigure.EnableAutoConfiguration=` key. That still works in
> some versions but is deprecated; Boot 4 reads the `.imports` file. If your auto-configuration
> "does nothing", check which file you created.

### Add the dependencies to the service

In `services/production-service/pom.xml`, add both libraries — **without** a `<version>`:

```xml
        <!--
          FactoryOS shared error taxonomy + transport adapters. No <version>:
          both are managed by the parent (factoryos-common-error.version).
        -->
        <dependency>
            <groupId>com.factoryos</groupId>
            <artifactId>factoryos-common-error</artifactId>
        </dependency>
        <dependency>
            <groupId>com.factoryos</groupId>
            <artifactId>factoryos-common-error-adapters</artifactId>
        </dependency>
```

That versionless declaration is exactly the `dependencyManagement` behaviour from Part 2:
the parent already knows the version, so the service does not restate it.

### Build

```bash
cd services && mvn -o -B clean test
```

Expect `BUILD SUCCESS` across all eight modules.

> **If `production-service` fails with `Could not find artifact
> com.factoryos:factoryos-common-error:jar:0.1.0-SNAPSHOT`**, the libraries are not in your
> local Maven repository yet. Building the reactor with `-am` compiles them, but the
> **protobuf-maven-plugin resolves dependencies from the local repository**, not the reactor,
> so it needs them installed:
>
> ```bash
> cd services && mvn -B install -N -DskipTests              # the parent POM
> cd services && mvn -B install -pl ../libs/factoryos-common-error,../libs/factoryos-common-error-adapters -DskipTests
> ```
>
> Note these omit `-o`. `mvn install` needs to fetch its own plugin dependencies the first
> time, so it cannot run offline until it has done so once. After that, `-o` works again.

### Verify it live

Start the service against the real database. **In a devcontainer use the docker
`.env`** — see Part 2 for why `localhost` cannot work.

```bash
make setup-production-docker
make run-production
```

Now send the requests that used to return 500. These are **real captured responses**:

```bash
curl -s -w "\nHTTP %{http_code}  %{content_type}\n" \
  "http://localhost:3001/api/v1/work-orders?limit=500"
```

```json
{
  "detail": "pageSize must be between 1 and 100, got 500",
  "instance": "/api/v1/work-orders",
  "status": 422,
  "title": "Page size out of range",
  "type": "https://factoryos.dev/errors/page-size-out-of-range",
  "code": "PAGE_SIZE_OUT_OF_RANGE",
  "retryable": false,
  "traceId": "7ae80c66-86ed-4ff9-b52b-114a2ae9e553",
  "timestamp": "2026-09-17T01:38:19.643097754Z",
  "errors": [
    {
      "field": "limit",
      "message": "must be between 1 and 100",
      "current": "500",
      "allowed": ["1..100"]
    }
  ]
}
```

```
HTTP 422  application/problem+json
```

Note the two different statuses. `limit=500` is a **422** — the query parsed fine, the *value*
is out of range, and `PAGE_SIZE_OUT_OF_RANGE` is a `VALIDATION` code (`ERROR_HANDLING.md` §2).
A `cursor=notavalidcursor` is a **400** — the cursor string could not be parsed at all, so its
code is `MALFORMED_REQUEST`. The distinction is the whole reason those two categories exist.

Run all four and confirm each fails with its own code:

```bash
for u in "limit=500" "page=-1" "cursor=notavalidcursor" "cursor=v2.abc"; do
  printf "%-26s " "$u"
  curl -s -o /tmp/r.json -w "HTTP %{http_code} " "http://localhost:3001/api/v1/work-orders?$u"
  grep -o '"code":"[A-Z_]*"' /tmp/r.json
done
```

| Request | Expected |
|---|---|
| `?limit=500` | `HTTP 422` … `"code":"PAGE_SIZE_OUT_OF_RANGE"` |
| `?page=-1` | `HTTP 422` … `"code":"PAGE_SIZE_OUT_OF_RANGE"` |
| `?cursor=notavalidcursor` | `HTTP 400` … `"code":"UNSUPPORTED_CURSOR_VERSION"` |
| `?cursor=v2.abc` | `HTTP 400` … `"code":"UNSUPPORTED_CURSOR_VERSION"` |

And confirm a valid request still works — the change must not affect the happy path:

```bash
curl -s "http://localhost:3001/api/v1/work-orders?limit=5"
# {"data":[],"pagination":{"page":0,"limit":5,"total":0,"totalPages":0}}
```

### Verify the gRPC side

The gRPC adapter is separately registered, so it needs its own check. Captured output from a
Java client against the running service:

```
REQUEST                gRPC CODE            ErrorInfo.reason          category / retryable
pageSize=500           INVALID_ARGUMENT     PAGE_SIZE_OUT_OF_RANGE    VALIDATION / false
pageSize=-1            INVALID_ARGUMENT     PAGE_SIZE_OUT_OF_RANGE    VALIDATION / false
pageToken=bad          INVALID_ARGUMENT     UNSUPPORTED_CURSOR_VERSION MALFORMED_REQUEST / false
workCenterId=bad       INVALID_ARGUMENT     MALFORMED_REQUEST         MALFORMED_REQUEST / false
```

Read the fourth row carefully. The same code string appears over gRPC as would appear over
REST — `MALFORMED_REQUEST` — which is the entire point of the design: **a client that
switches on `code` behaves identically on both protocols.**

> **If you want to reproduce this**, `grpcurl` is the easy way when it is installed:
>
> ```bash
> grpcurl -plaintext -d '{"page_size": 500}' localhost:4001 production.v1.ProductionService/ListWorkOrders
> ```
>
> The service registers gRPC reflection (visible in the startup log as
> `Registered gRPC service: grpc.reflection.v1.ServerReflection`), so `grpcurl` can discover
> the schema without a proto file. If `grpcurl` is unavailable, write a small Java client —
> the generated stub is `<production-service>/target/generated-sources/protobuf`.

### Prove the two special cases

These are the subtle parts of the standard, so verify them rather than assuming.

**`NOT_FOUND` must be gRPC 5, never 14.** A missing entity is permanent — reporting it as
`UNAVAILABLE` would tell a retrying client to hammer a request that can never succeed.
Asserted in the test suite rather than by hand, because you need a real missing entity to
provoke it.

**The conflict split.** `CONFLICT` maps to `FAILED_PRECONDITION` (9) — *except* a lost
optimistic lock, which maps to `ABORTED` (10):

| Code | gRPC | Retryable |
|---|---|---|
| `WORK_ORDER_INVALID_TRANSITION` | `FAILED_PRECONDITION` (9) | no |
| `WORK_ORDER_ALREADY_EXISTS` | `FAILED_PRECONDITION` (9) | no |
| `WORK_ORDER_CONCURRENT_MODIFICATION` | **`ABORTED` (10)** | **yes** |

All three are the same *category* and the same *exception class* — only the code differs, and
only one of them carries a code-level gRPC override. That is why the adapter projects from
`ex.code()` and not `ex.getClass()`, and why the override table exists: a code-level entry for
`WORK_ORDER_CONCURRENT_MODIFICATION` selects `ABORTED`, while the other two fall back to the
`CONFLICT` default. Part 7 asserts all three by code.

### Check yourself

- `ErrorHandlingAutoConfiguration.java` exists, and the `.imports` file names it
- `production-service/pom.xml` declares both libraries **without** a version
- `mvn -o -B clean test` is green across all eight modules
- `?limit=500` returns **422** with `application/problem+json` and `PAGE_SIZE_OUT_OF_RANGE`
  (a `VALIDATION` code); `?cursor=notavalidcursor` returns **400** (`MALFORMED_REQUEST`)
- A valid `?limit=5` request still returns **200**
- If it still returns 500, the adapters are not registered — re-read the component-scan
  section at the top of this part

---


---

[← Part 6a — Adopting it: the domain side](06a-domain-adoption.md) · [Index](README.md) · [Part 7 — Proving the taxonomy →](07-proving-the-taxonomy.md)
