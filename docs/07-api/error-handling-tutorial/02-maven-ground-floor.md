[Index](README.md) · [← Part 1 — The core idea](01-the-core-idea.md)

# Part 2 — Maven ground floor

If you already know Maven well, skim this part — but do not skip the last section, which
describes a real trap.

### What a Maven module is

A **module** is a directory containing a `pom.xml` that produces one artifact (here: a
`.jar`). A project with several modules is called a **multi-module build** or a
**reactor** — one command builds all of them, in dependency order.

FactoryOS has eight modules. Seven live under `services/` and `libs/`; the eighth is the
parent that ties them together.

### What a parent POM buys you

A **parent POM** is shared configuration that child modules inherit. The important thing it
carries is the **version of everything**. In this repository, `services/pom.xml` sets:

```xml
<parent>
    <groupId>org.springframework.boot</groupId>
    <artifactId>spring-boot-starter-parent</artifactId>
    <version>4.1.1</version>
</parent>
```

Every service inherits *through* `services/pom.xml` rather than declaring
`spring-boot-starter-parent` itself. Open `services/production-service/pom.xml` and you will
see this comment, which explains exactly why:

```xml
<!--
  Inherited from ../pom.xml, NOT from spring-boot-starter-parent directly.
  Deliberately no <relativePath/>: that would skip the aggregator and let this
  service resolve a different Spring Boot than its siblings.
-->
```

The failure mode being prevented: if one service declared the Boot parent directly with a
different version, that service would run a different Spring Boot major than its siblings.
Two services, one build, incompatible frameworks — and the problem surfaces at runtime, not
at build time. One parent, one Boot version.

### `dependencyManagement` versus `dependencies`

This distinction confuses almost everyone at first, and it is worth two minutes.

- **`<dependencyManagement>`** declares **versions only**. It says "if anyone asks for
  `factoryos-common-error`, give them version `0.1.0-SNAPSHOT`". It adds nothing to the
  build by itself.
- **`<dependencies>`** actually **puts the artifact on the classpath**.

So the pattern is: the parent manages a version once, and each module opts in by declaring
the dependency *without* repeating the version. Here is the real block in
`services/pom.xml`:

```xml
<dependencyManagement>
    <dependencies>
        <!-- Versions only: services opt in by declaring the dependency. -->
        <dependency>
            <groupId>com.factoryos</groupId>
            <artifactId>factoryos-common-error</artifactId>
            <version>${factoryos-common-error.version}</version>
        </dependency>
        <dependency>
            <groupId>com.factoryos</groupId>
            <artifactId>factoryos-common-error-adapters</artifactId>
            <version>${factoryos-common-error.version}</version>
        </dependency>
    </dependencies>
</dependencyManagement>
```

Both of our library modules are **already managed**, under the property
`factoryos-common-error.version`. That means when you add them to a service in Part 6b, you
will write the dependency with no `<version>` tag — the parent supplies it.

### The current state of the build

Both `libs/` modules already exist and are already listed in `<modules>`. **This
tutorial assumes you are starting from empty shells:** the guide below tells you to create
each `.java` file, and if the file already exists, opening it and reading along is the right
move — you are checking your work against the reference, not retyping it.

### See it for yourself

```bash
cd services && mvn -o -B validate
```

This parses every POM and checks the reactor without compiling anything, so it takes under a
second. You should see:

```
[INFO] Reactor Summary:
[INFO]
[INFO] FactoryOS Domain Services Parent 0.1.0-SNAPSHOT .... SUCCESS [  0.001 s]
[INFO] production-service 0.0.1-SNAPSHOT .................. SUCCESS [  0.029 s]
[INFO] FactoryOS Warehouse Service 0.1.0-SNAPSHOT ......... SUCCESS [  0.001 s]
[INFO] FactoryOS Quality Service 0.1.0-SNAPSHOT ........... SUCCESS [  0.000 s]
[INFO] FactoryOS Maintenance Service 0.1.0-SNAPSHOT ....... SUCCESS [  0.000 s]
[INFO] FactoryOS Planning Service 0.1.0-SNAPSHOT .......... SUCCESS [  0.000 s]
[INFO] FactoryOS Common Error Taxonomy 0.1.0-SNAPSHOT ..... SUCCESS [  0.001 s]
[INFO] FactoryOS Common Error Transport Adapters 0.1.0-SNAPSHOT SUCCESS [  0.000 s]
[INFO] ------------------------------------------------------------------------
[INFO] BUILD SUCCESS
```

All eight modules `SUCCESS`. The two we care about are the last two — "Common Error
Taxonomy" and "Common Error Transport Adapters".

### Spring Boot 4 breaks the tutorials you will find online

This project runs **Spring Boot 4.1.1**. Most Spring material on the internet is written for
Boot 3, and several things were renamed. Knowing these four will save you a lot of
confusion:

| What you will read online | What is true in Boot 4 | Why it matters |
|---|---|---|
| `spring-boot-starter-web` | `spring-boot-starter-webmvc` | The old artifact **does not exist** at 4.1.1. The build fails with a missing-artifact error. |
| `spring-boot-starter-test` covers everything | Each slice has its own test starter: `-webmvc-test`, `-grpc-server-test`, `-data-jpa-test`, `-flyway-test` | A slice test fails to compile without its starter. |
| Jackson 2 (`com.fasterxml.jackson`) | **Jackson 3** (`tools.jackson.core`) | Jackson 2 imports will not resolve. You can see Jackson 3 already in use in `Cursor.java`. |
| `org.xolstice.maven.plugins:protobuf-maven-plugin` | `io.github.ascopes:protobuf-maven-plugin` | The old plugin is unmaintained; Boot 4.1.1 manages the replacement. |

> **Rule of thumb:** if a Stack Overflow answer tells you to add a dependency and the build
> says the artifact does not exist, check whether the answer predates Boot 4.

### Connecting to infrastructure from a devcontainer

This bites everyone once, so here it is before you need it.

Connection details live in env, never in a Spring profile. The base
`application.yml` reads `jdbc:postgresql://${DB_HOST:localhost}:5432/factoryos`,
so the same artifact works on the host and in the container — only the env
changes:

| Where the service runs | DB Host source | When to use |
|---|---|---|
| Host machine | `DB_HOST=localhost` (default, or `services/production-service/.env.example`) | Infra on the same host |
| Devcontainer | `DB_HOST=factoryos-db` (`services/production-service/.env.docker.example`) | **Inside the devcontainer** |

### Why `localhost` fails in the devcontainer

`docker-compose.yml` publishes Postgres with `-p 5432:5432`. That mapping exposes the port on
the **Docker host**, not inside your devcontainer. From inside the devcontainer, `localhost`
means *the devcontainer itself*, which has nothing listening on 5432. So you get:

```
Connection to localhost:5432 refused.
```

…while `docker ps` cheerfully shows `factoryos_db  Up 2 hours`. That contradiction is what
makes this confusing — the database clearly *is* running.

**The devcontainer is designed to reach it by name instead.** Look at
`.devcontainer/devcontainer.json`:

```json
"runArgs": [
    "--network=factoryos_net"
]
```

The devcontainer joins the same Docker network as the infrastructure containers. On that
network, Docker's DNS resolves the compose **service name** — so `factoryos-db` reaches the
database:

```bash
getent hosts factoryos-db        # → 172.19.0.6  factoryos-db
```

This is committed configuration, not an accident of one machine. Every developer's
devcontainer behaves the same way.

### Start the service with the docker env

Spring reads OS env, not `.env` files — always run via `make` so the
`.env` is sourced first. Raw `mvn spring-boot:run` ignores `.env` and
falls back to `localhost`.

```bash
make setup-production-docker
make run-production
```

You should see it come up:

```
Tomcat started on port 3001 (http) with context path '/'
Registered gRPC service: production.v1.ProductionService
Registered gRPC service: grpc.reflection.v1.ServerReflection
Registered gRPC service: grpc.health.v1.Health
gRPC Server started, listening on address: [/[0:0:0:0:0:0:0:0]:4001], port: 4001
Started ProductionServiceApplication in 4.521 seconds
```

> The `.env.docker.example` template is the **committed configuration shared by the
> whole team**, rather than a variable each developer must remember. Add
> `run-production-local` if you also want `show-sql` on while you work
> through this tutorial:
>
> ```bash
> make run-production-local
> ```

> **A note on the name.** The compose **service** is `factoryos-db` (hyphen); the
> **container** is `factoryos_db` (underscore). Docker's network aliases happen to resolve
> both, but `.env.docker.example` uses the hyphenated service name. If you ever see an
> `UnknownHostException`, check which one you typed.

We use this env again in Part 6b to prove the error adapters work against a real request.

> **Do not confuse this with the `test` profile.** They are unrelated and serve different
> purposes:
>
> | Profile | Used by | Datasource | Infra needed? |
> |---|---|---|---|
> | `test` | `mvn test` | in-memory H2 | **No** |
> | *(none)* + docker `.env` | running the service | `factoryos-db` | Yes |
>
> `mvn -o -B test` needs no containers at all — H2 lives in memory, so the suite is
> self-contained. That is why the tests pass whether or not your infrastructure is running.

### Cleaning a Java build — a real trap

There are two "clean" commands in this repository and they do **not** do the same thing.

```bash
make clean                                  # removes bin/ and Go test artifacts — NOT Java builds
cd services && mvn -o -B clean              # removes Java target/ directories — the real clean
```

**`make clean` does not touch Java.** Read `Makefile`'s clean target and you will see it
removes `$(BIN_DIR)`, `*.test`, `*.out` and `coverage.html` — nothing else. It never deletes
a `target/` directory.

This is not hypothetical. On 2026-09-16, a reverted `Severity.java` left an orphaned
`Severity.class` sitting in `libs/factoryos-common-error/target/classes/`. `make clean` did
not remove it, because it cannot. The file kept reappearing in the build as a class with no
source. `mvn -o clean` removed it immediately.

**So:** when you revert or delete Java source, run `mvn -o -B clean` before rebuilding.
Otherwise you may be looking at the compiled remains of code that no longer exists.

> **`mvn -o clean` needs `maven-clean-plugin` in the offline cache.** If it fails with
> `A required class was missing ... org/codehaus/plexus/util/Os`, the plugin's transitive
> dependencies are not cached — drop the `-o` flag and Maven will fetch them.

### Check yourself

- `mvn -o -B validate` from `services/` reports all **8 modules** `SUCCESS`
- You can explain the difference between `dependencyManagement` and `dependencies`
- You know that `mvn -o clean` — not `make clean` — is what removes a Java build

---


---

[← Part 1 — The core idea](01-the-core-idea.md) · [Index](README.md) · [Part 3 — The taxonomy module →](03-taxonomy-module.md)
