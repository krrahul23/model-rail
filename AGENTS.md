# AGENTS.md — ModelRail

Guidance for AI agents and contributors working in this repository. Read this
first, then `docs/hld.md` (architecture source of truth) and the relevant LLD.

---

## 1. What this project is

ModelRail is a multi-tenant, OpenAI-compatible LLM gateway written in Go, plus
an independent Python evaluation engine that feeds benchmark evidence back into
model routing.

- Module: `modelrail` (`go 1.27.0`, `toolchain go1.27.3`).
- Providers: DeepSeek (primary), Google Gemini (secondary), Mock (tests only).
  No models run locally; all inference is via external APIs.
- Deployment target: one shared VPS (8 GB RAM / 4 vCPU / 100 GB SSD — **target
  specifications, not purchased infrastructure**), one shared reverse proxy for
  all apps, dedicated PostgreSQL and Redis, static Next.js dashboard, Prometheus,
  Grafana on demand. No Kubernetes, Kafka, message broker or extra database.

## 2. Documents and their authority

| File | Role |
|---|---|
| `docs/hld.md` | **Draft 3 — architecture source of truth.** All Q1–Q21 decisions resolved. |
| `docs/hld-review.md` | Historical principal-level review; not a decision record. |
| `docs/lld-01.md` | Gateway request lifecycle and provider abstraction (current). |
| `coding_skills.md` | Reusable package and transport-structure playbook (pattern, not project decisions). |
| `docs/lld-02…` | Planned: persistence/ledger/settlement, reliability/Redis, HTTP API spec, eval details. |

If code and document disagree, fix one deliberately and say which. Never let
them drift silently. Document changes follow the targeted-edit rule (§8).

## 3. Current implementation state

- **Phase 1 complete:** contracts only:
  - `internal/constant` — scalar types and constants.
  - `internal/dto/request`, `internal/dto/response` — structs only.
  - `internal/ports` — provider and classifier interfaces.
- **Phase 3 complete:** unary end-to-end path — `application` validation,
  `domain` inference workflow (deadline handling, provider-error mapping),
  `controller` OpenAI-compatible `/v1/chat/completions` and `/healthz`, and
  `app` wiring with `cmd/modelrail` plus graceful shutdown. Streaming is
  intentionally rejected with a typed error until Phase 4.
- **Configuration:** `config.yml` at the repository root is the runtime
  configuration — loaded, defaulted and validated by `internal/app/config`.
  Secrets are referenced by environment-variable name (`api_key_env`), never
  stored in YAML.
- **Next phase:** Phase 4 — streaming pump, SSE framing, commit point and
  disconnect handling.
- Implementation proceeds **one phase per turn**, starting simple. Do not jump
  ahead or scaffold future phases.

## 4. Architecture and package layout

Clean Architecture + Hexagonal (ports and adapters) + pragmatic DDD.

```
cmd/modelrail/           # process entry point
internal/
├── app/                 # bootstrap, config, dependency wiring, lifecycle
│   └── config/          # YAML config loading, defaults, validation
├── controller/
│   └── http/
│       ├── router.go     # unversioned routes + middleware wiring
│       └── v1/
│           ├── routes.go # mounts v1 feature handlers
│           ├── shared/   # versioned error envelope + JSON helpers
│           └── chat/     # feature slice: handler, request, response, tests
├── middleware/
│   └── http/             # request ID, logging (transport-specific)
├── application/
│   └── inference/
│       ├── inference_interface.go
│       ├── inference_service.go
│       ├── inference_mock.go
│       └── inference_service_test.go
├── domain/
│   └── inference/
│       ├── inference_interface.go
│       ├── inference_service.go
│       ├── inference_mock.go
│       └── inference_service_test.go
├── infrastructure/
│   └── provider/
│       └── mock/         # adapter + contract tests
├── logger/              # zap wrapper; leaf package, safe to import anywhere
├── constant/            # scalar types and constants (no structs)
├── dto/request/         # request structs only (no constants)
├── dto/response/        # response structs and error value types (no constants)
└── ports/
    ├── provider.go        # interfaces consumed by domain, implemented by infrastructure
    └── provider_mock.go   # MockProvider test double used by domain tests
```

Dependency direction is inward only:

```
controller → application → domain
infrastructure → domain (implements ports)
app → everything (wiring only)
```

Rules:
- `domain/` imports no `application/`, `controller/` or `infrastructure/` package.
- `controller/` holds no business rules; `application/` holds no orchestration.
- Transport code is versioned by feature slice (`controller/http/v1/<feature>`);
  request objects map to canonical DTOs with `ToDTO` (structural only), and
  business validation stays in `application/`. See `coding_skills.md`.
- Each layer is split into feature folders: `<layer>/<feature>/` contains
  `<feature>_interface.go`, `<feature>_service.go`, `<feature>_mock.go` and its
  tests. The mock is named `Mock<Service>` and lives in the same package, so
  outer-layer tests import the feature package directly. Infrastructure groups
  adapters by capability (`provider/<adapter>`, `persistence/<store>`,
  `cache/<cache>`) and implements `ports` instead of defining its own interfaces.
- DTOs never contain constants or scalar type declarations; those live in `constant/`.
- Ports are small (one to three methods) and exist only at real external boundaries.
- Test doubles live next to the contract: feature mocks as `Mock<Service>` in the
  feature package, and the provider port mock as `MockProvider` in
  `ports/provider_mock.go`. Infrastructure adapters remain the real
  implementations used by `app` and their own contract tests.

## 5. Go style (project preferences)

- **Comment everything, outside and inside:** package docs, doc comments on every
  exported type/function/constant, field-level comments inside structs, and
  explanatory inline comments inside method bodies.
- **Named return values and naked returns**, including a named `err` result.
- Prefer the scoped error idiom:
  `if _, err := fn(); err != nil { return }` — returning the named error.
- `gofmt -w .`, `go build ./...`, `go vet ./...` and `go test ./...` must pass
  before a phase is considered complete.
- Keep files small and single-purpose, grouped by concept (see `internal/constant`).
- **No magic literals:** reusable string literals and constant values — log levels
  and encodings, environment names, field names, error codes, provider names —
  live in `internal/constant` and are referenced symbolically in production code.
  Struct tags (YAML/JSON) and test literals are the exception.
- No provider SDK types may cross the adapter boundary; return
  `response.ProviderError` (adapters) or `response.GatewayError` (pipeline).
- Adapters never carry HTTP status codes or gateway-origin error codes.

## 6. Non-negotiable architecture rules

These come from the HLD and must not be weakened without an explicit decision:

1. **No database transaction or row lock is held across a provider call.**
2. **Completion rule:** usage and settlement commit atomically *before* the
   client can observe completion (response body; terminal `[DONE]` for streams).
   Settlement runs on a context detached from the client request.
3. **Streaming commit point:** headers are not flushed and no SSE heartbeat is
   sent before the first content chunk. No provider switch after commit.
4. **Fallback is opt-in:** only along tenant-defined groups, provider allowlist
   and consent required; requests may narrow but never widen a group. Evaluation
   traffic never falls back. Responses always report the served model.
5. **Bounded overspend:** settle true cost, record excess as overage, refuse new
   reservations when available balance is ≤ 0.
6. **Orphaned holds:** release and record platform-absorbed leakage.
7. **Settlement failure:** strict policy — do not deliver completion.
8. **Guardrails:** structural validation (payload size, content type, message
   count) runs before concurrency admission. Output validation is non-streaming
   only.
9. **Auth:** L1 in-process key cache, immediate revocation for new requests, no
   key material or prompt content in logs.
10. **Routing:** deterministic constraint pipeline, no LLM calls; task class
    comes from the `modelrail-task` hint and defaults to `general`.
11. **Content:** prompts and responses are not stored or logged by default.

## 7. Phase roadmap (LLD-01 implementation)

| Phase | Scope |
|---|---|
| 1 ✅ | Contracts: constants, DTOs, ports |
| 2 ✅ | Mock provider adapter + contract tests |
| 3 ✅ | Domain inference workflow, application validation, unary controller endpoint, app wiring |
| 4 | Streaming pump, SSE framing, commit point, disconnect handling |
| 5 | Auth/keys, in-process limits, model registry and routing |
| 6 | LLD-02 persistence/ledger first, then reservation and settlement |

Change the order only with explicit user approval.

## 8. Documentation workflow

- `docs/hld.md` is the architecture source of truth.
- Make **targeted edits** to existing documents; never rewrite, regenerate or
  reformat an entire document.
- After document edits, report the exact sections changed and summarize the
  remaining deferred/open decisions.
- New decisions belong in the HLD decision register (§26) and, when they come
  from a review, in the disposition log (Appendix A).
- Never edit files under `.commandcode/taste/` manually. Record preferences with
  the `taste` tool.

## 9. Commands

```bash
gofmt -w .
go build ./...
go vet ./...
go test ./...
go mod tidy        # only when dependencies change
```

The base Go install may be older than the module requires (`GOTOOLCHAIN=auto`
selects the pinned toolchain, currently go1.27.3).

## 10. Do not

- Add Kafka, Kubernetes, LangChain/LangGraph, a message broker or another database.
- Add paid observability or testing services.
- Introduce scaffolding, files or refactors outside the current phase.
- Leak secrets into the repository or logs.
- Store prompt or response content by default.
- Let provider-specific error shapes or HTTP concepts reach the domain.
- Promise throughput, latency or availability numbers that have not been measured.
