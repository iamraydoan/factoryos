# ADR-0008: Split Auth-Free OpenAPI Contract from Gateway Auth via Overlay Spec 1.0

* **Status:** Accepted
* **Date:** 2026-09-30
* **Authors:** FactoryOS Architecture Team

---

## 1. Context

The telemetry REST contract (`api/contracts/openapi/telemetry/v1/`) mixed the
transport-neutral API shape with gateway authentication concerns: an absolute
`servers` URL (`http://localhost/api/v1`), a global `security` requirement, a
`components/securitySchemes` reference to `common/v1/security/oidc.yaml`, and
per-operation `security: [{zitadel: [<scope>]}]` blocks. This coupled the
versioned contract to one environment's identity provider and leaked gateway
auth into `oapi-codegen` output (generated clients/servers embedding Zitadel
scopes). Health probes carried `security: []` overrides, further entangling
liveness semantics with auth configuration.

## 2. Decision

Adopt a two-artifact pipeline per [Overlay Specification 1.0](https://spec.openapis.org/overlay/v1.0.0.html):

1. **Pure contract (source of truth, for `oapi-codegen` only):** versioned specs under
   `api/contracts/openapi/<domain>/v1/` are 100% auth-free — relative `servers: [{url: /api/v1}]`,
   no global `security`, no `components/securitySchemes`, no per-operation `security` blocks.
   Root `redocly.yaml` disables the `security-defined` rule since auth-free sources are intentional.
2. **Gateway overlay (Swagger UI only, never codegen):** `api/contracts/openapi/overlays/<env>.gateway.overlay.yaml`
   (`overlay: 1.0.0`) re-applies the environment host, the Zitadel `openIdConnect` scheme, and
   per-operation scopes post-bundle. New `make openapi-gateway` target validates
   (`overlay validate`) and applies (`overlay apply`) each overlay to every bundled spec,
   emitting `dist/openapi.gateway.<env>.yaml`. Tool: `go install github.com/speakeasy-api/openapi/cmd/openapi@latest`,
   invoked via explicit `$(go env GOPATH)/bin/openapi` (`OPENAPI_OVERLAY_BIN`) to avoid the
   colliding npm `openapi` alias.
3. **Codegen consumes pure only:** `make openapi-gen` reads `openapi.bundled.yaml`, never
   `openapi.gateway.*.yaml`. `make docs-up` depends on `openapi-gateway`; the Compose
   `swagger-ui` service defaults to the gateway artifact with `SWAGGER_SPEC_FILE` env override.
4. The now-unreferenced `common/v1/security/oidc.yaml` fragment is deleted; the overlay owns
   the `openIdConnectUrl`.

## 3. Consequences

### Positive Impacts
* **Reusable SDK:** generated Go clients/servers carry zero auth refs; one contract serves all environments.
* **Env-specific auth:** staging/prod overlays can be added later without touching versioned specs.
* **Explicit failures:** overlay validate/apply errors fail the build loudly (no silent fallbacks).

### Negative Impacts & Trade-offs
* **Two artifacts to publish:** consumers must pick pure (codegen) vs gateway (docs) — mitigated by
  documenting the pipeline in `docs/07-api/README.md`.
* **Overlay tool drift:** `@latest` pin for the Speakeasy CLI may shift flag syntax — mitigated by
  the explicit binary path and fail-fast validate step.
