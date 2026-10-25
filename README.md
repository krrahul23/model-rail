# ModelRail

ModelRail is a multi-tenant, OpenAI-compatible LLM gateway written in Go. It is
designed to sit in front of external providers (DeepSeek, Google Gemini) and
provide one API, deterministic routing, provider reliability, metering and
budget accounting. A separate Python evaluation engine (planned) will compare
models on quality, latency and cost and feed scorecards back into routing.

> **Status: early development.** The unary gateway path is implemented and runs
> end to end against the built-in mock provider. Streaming, authentication,
> budgets and the evaluation engine are planned — see [Roadmap](#roadmap).

---

## What works today

| Capability | Status |
|---|---|
| YAML configuration with defaults and validation | ✅ |
| Structured logging with zap (JSON/console, request IDs) | ✅ |
| `POST /v1/chat/completions` (non-streaming) | ✅ |
| OpenAI-compatible request/response mapping | ✅ |
| Request validation and typed error envelopes | ✅ |
| Provider abstraction (`ports.Provider`) | ✅ |
| Deterministic mock provider with injectable faults | ✅ |
| Graceful shutdown on SIGINT/SIGTERM | ✅ |
| Streaming (SSE), commit point, disconnect handling | ⏳ Phase 4 |
| API keys, rate limits, quotas, model registry, routing | ⏳ Phase 5 |
| PostgreSQL ledger, budget reservation and settlement | ⏳ Phase 6 |
| DeepSeek and Gemini adapters | ⏳ planned |
| Guardrails and policy enforcement | ⏳ planned |
| Next.js dashboard and Python evaluation engine | ⏳ planned |

Only the mock provider is wired at runtime today. The DeepSeek and Gemini
entries in `config.yml` are configuration placeholders for upcoming adapters.

---

## Architecture

Clean Architecture + Hexagonal (ports and adapters) + pragmatic DDD:

```
controller → application → domain
infrastructure → ports
app → everything (wiring only)
```

- `controller` owns HTTP concerns only: decoding, wire mapping, status and
  error-envelope mapping.
- `application` validates business requests and dispatches use cases.
- `domain` orchestrates the workflow and owns rules; it depends on `ports`.
- `infrastructure` implements `ports`.
- `app` is the composition root.

Each layer is split into feature slices, and each feature owns its contract,
implementation, mock and tests:

```
internal/application/inference/
├── inference_interface.go
├── inference_service.go
├── inference_mock.go
└── inference_service_test.go
```

---

## Tech stack

- **Go 1.27** (`go 1.27.0`, toolchain `go1.27.3`)
- Standard library `net/http` with Go 1.22+ method routing
- `go.uber.org/zap` for structured logging
- `gopkg.in/yaml.v3` for configuration
- **Planned:** PostgreSQL, Redis, Prometheus, Grafana, OpenTelemetry,
  Next.js/TypeScript/Tailwind dashboard, Python evaluation engine

The target deployment is a single small VPS with a shared reverse proxy. No
Kubernetes, Kafka or message broker is used.

---

## Project structure

```
cmd/modelrail/                 # process entry point
internal/
├── app/                       # composition root, lifecycle
│   └── config/                # YAML loading, defaults, validation
├── application/
│   └── inference/             # validation + command dispatch
├── constant/                  # scalar types and constants
├── controller/
│   └── http/
│       ├── router.go          # /healthz, middleware wiring
│       └── v1/
│           ├── routes.go
│           ├── shared/        # v1 error envelope and JSON helpers
│           └── chat/          # handler, request, response, tests
├── domain/
│   └── inference/             # workflow orchestration, error mapping
├── dto/
│   ├── request/               # canonical request structs
│   └── response/              # canonical response structs and errors
├── infrastructure/
│   └── provider/mock/         # deterministic provider adapter
├── logger/                    # zap wrapper (leaf package)
├── middleware/http/           # request ID and request logging
└── ports/                     # interfaces + MockProvider test double
```

---

## Getting started

**Requirements**

- Go 1.27 or newer (the module pins the toolchain)
- No cloud credentials are needed for the mock provider

**Run**

```bash
go run ./cmd/modelrail -config config.yml
```

The gateway listens on `:8080` by default.

**Health**

```bash
curl -sS http://localhost:8080/healthz
# ok
```

**Chat completion**

```bash
curl -sS -X POST http://localhost:8080/v1/chat/completions \
  -H 'Content-Type: application/json' \
  -d '{
    "model": "mock-model",
    "messages": [
      {"role": "user", "content": "Hello"}
    ]
  }'
```

Example response (mock provider):

```json
{
  "id": "…",
  "object": "chat.completion",
  "created": 1792921770,
  "model": "mock-model",
  "choices": [
    {
      "index": 0,
      "message": {"role": "assistant", "content": "Hello from the ModelRail mock provider."},
      "finish_reason": "stop"
    }
  ],
  "usage": {"prompt_tokens": 0, "completion_tokens": 39, "total_tokens": 39}
}
```

Streaming requests (`"stream": true`) are rejected with a typed
`request_invalid` error until Phase 4.

---

## Configuration

`config.yml` at the repository root is the runtime configuration. It is loaded,
defaulted and validated at startup. **No secrets are stored in the file**;
provider credentials are referenced indirectly through environment-variable
names (`api_key_env`).

| Section | Purpose |
|---|---|
| `app` | Service name and environment (`dev`, `test`, `prod`) |
| `logging` | `level` (`debug`/`info`/`warn`/`error`) and `encoding` (`json`/`console`) |
| `server` | Public listen address, reserved admin address, header timeout, shutdown grace |
| `providers.mock` | Response text and simulated pacing |
| `providers.deepseek` | Base URL, model, and the env var name holding the API key |
| `providers.gemini` | Base URL, model, and the env var name holding the API key |

The admin listener address is reserved for future phases; only the public
listener starts today.

---

## Development

```bash
gofmt -w .
go build ./...
go vet ./...
go test ./...
go test -race ./...
```

The project follows strict layering, feature-based packages, typed errors,
structured logging, and constants instead of magic literals. Code is documented
at package, type, field and method level.

---

## API notes

- Success responses follow the OpenAI `chat.completion` shape.
- Errors use the OpenAI-style envelope:
  `{"error":{"message":"…","type":"invalid_request_error","code":"request_invalid"}}`.
- Every response includes `X-Request-Id`.
- The optional `modelrail-task` header (`general`, `coding`, `reasoning`,
  `structured_extraction`) is accepted and passed into routing metadata;
  routing itself arrives in Phase 5.
- Request bodies are limited to 1 MiB.

---

## Design intent

The architecture documents define the following decisions. They are design
targets, not all implemented yet:

- **Reserve → serve → settle:** budget is reserved before inference and the
  actual cost is settled atomically before the client observes completion.
- **Streaming commit point:** no provider switch after the first content chunk;
  failures after commit are reported as SSE errors.
- **No transaction spans a provider call.**
- **Opt-in fallback:** explicit model requests are never silently substituted;
  fallback only along tenant-defined groups.
- **Deterministic routing:** no LLM call in the routing path.
- **Prompt and response content is not stored or logged by default.**

See [`docs/hld.md`](docs/hld.md) for the complete design.

---

## Roadmap

| Phase | Scope | Status |
|---|---|---|
| 1 | Contracts: constants, DTOs, ports | ✅ |
| 2 | Mock provider adapter + contract tests | ✅ |
| 3 | Domain workflow, application validation, unary HTTP API, app wiring | ✅ |
| 4 | Streaming pump, SSE framing, commit point, disconnects | Next |
| 5 | Auth/keys, in-process limits, model registry and routing | Planned |
| 6 | Persistence/ledger, reservation and settlement | Planned |

Later milestones cover guardrails, real provider adapters, the Next.js
dashboard, the Python evaluation engine, and load/fault testing.

---

## Documentation

| Document | Purpose |
|---|---|
| [`docs/hld.md`](docs/hld.md) | High-level design and architecture decisions |
| [`docs/hld-review.md`](docs/hld-review.md) | Historical principal-level review |
| [`docs/lld-01.md`](docs/lld-01.md) | Gateway request lifecycle and provider abstraction |
| [`coding_skills.md`](coding_skills.md) | Reusable structure playbook |
| [`AGENTS.md`](AGENTS.md) | Repository guidance for contributors and AI agents |

---

## Non-goals

- Hosting or training models locally; all inference uses external APIs.
- Kubernetes, Kafka, a message broker or additional databases.
- Paid observability or APM services.
- Storing prompt or response content by default.

## License

No license has been added yet.
