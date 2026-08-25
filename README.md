# Mockzoo

[![License](https://img.shields.io/badge/License-MIT-success)](LICENSE)
[![Web Framework](https://img.shields.io/badge/Fiber-Web%20Framework-blue)](https://github.com/gofiber/fiber)
[![Query Builder](https://img.shields.io/badge/Squirrel-SQL%20Query%20Builder-blue)](https://github.com/Masterminds/squirrel)
[![Database Migrations](https://img.shields.io/badge/Migrations-Seamless%20Schema%20Updates-blue)](https://github.com/golang-migrate/migrate)
[![Logging](https://img.shields.io/badge/ZeroLog-Structured%20Logging-blue)](https://github.com/rs/zerolog)
[![Testing](https://img.shields.io/badge/Testify-Testing%20Framework-blue)](https://github.com/stretchr/testify)
[![Mocking](https://img.shields.io/badge/Mock-Mocking%20Library-blue)](https://go.uber.org/mock)

Mockzoo is a small, deliberately boring, stable mock/sandbox backend for pet data. It exists to give
the owner's frontend template and other demo/pet projects a backend they control instead of
depending on the public [Swagger Petstore](https://petstore.swagger.io) sandbox — a service whose
records are shared, mutable by anyone, and occasionally go missing.

Mockzoo serves the same shape of thing (pets, with a status and a species) but with three
guarantees the public Petstore doesn't make:

- Seed pets keep **stable ids** (1, 2, 3) that never change or disappear.
- Anyone can restore the sandbox to its seed state at any time with `POST /v1/admin/reset`.
- The API surface is hand-maintained and versioned in this repo — no surprises.

It is a single-transport (REST), single-domain (pets) clean-architecture Go service, built as a
trimmed-down fork of [evrone/go-clean-template](https://github.com/evrone/go-clean-template)
(MIT licensed — see [`LICENSE`](LICENSE) and the [Trimmed from go-clean-template](#trimmed-from-go-clean-template)
section below for what was removed and why).

## Content

- [API overview](#api-overview)
- [The Pet schema](#the-pet-schema)
- [Seed data](#seed-data)
- [Quick start](#quick-start)
- [Local development](#local-development)
- [Configuration](#configuration)
- [OpenAPI](#openapi)
- [Architecture](#architecture)
- [Trimmed from go-clean-template](#trimmed-from-go-clean-template)

## API overview

All routes are served under `/v1`. Every failing response returns
`{"error": {"code": "...", "message": "..."}}`, with `code` one of `pet_not_found`,
`validation_failed`, `not_found` (unmatched route), `method_not_allowed`, or `internal`.

| Method | Path             | operationId    | Description                                            |
| ------ | ---------------- | -------------- | -------------------------------------------------------|
| GET    | `/pets`          | `listPets`     | List pets, filterable by `status`/`species`, paginated |
| POST   | `/pets`          | `createPet`    | Create a pet (id ≥ 1001)                                |
| GET    | `/pets/{petId}`  | `getPetById`   | Get a pet by id                                        |
| PATCH  | `/pets/{petId}`  | `updatePet`    | Partially update a pet                                 |
| DELETE | `/pets/{petId}`  | `deletePet`    | Delete a pet                                            |
| POST   | `/admin/reset`   | `resetSandbox` | Restore the three seed pets, discard everything else   |
| GET    | `/openapi.json`  | —              | The hand-maintained OpenAPI 3.1 spec, served verbatim   |

Outside `/v1`: `GET /healthz` (liveness probe) and `GET /docs` (Scalar API reference UI).

## The Pet schema

```jsonc
{
  "id": 1,
  "name": "Barsik",
  "status": "available",   // available | pending | sold
  "species": "cat",         // dog | cat | bird | fish | other
  "breed": "domestic shorthair", // optional
  "photoUrl": "https://...",     // optional
  "tags": ["fluffy", "demo"],    // always present, empty array if none
  "createdAt": "2026-01-01T00:00:00Z"
}
```

## Seed data

| id  | name   | species | status    | breed               | tags               |
| --- | ------ | ------- | --------- | ------------------- | ------------------ |
| 1   | Barsik | cat     | available | domestic shorthair  | fluffy, demo        |
| 2   | Rex    | dog     | pending   | german shepherd     | good-boy, demo      |
| 3   | Kesha  | bird    | sold      | budgerigar          | talks, demo         |

Ids 1–3 are stable and reserved for these seed pets. The id sequence starts at 1000, so every
pet you create gets an id ≥ 1001 and never collides with a seed row. `POST /v1/admin/reset`
truncates the table and restores exactly this seed data — safe to call as often as you like, e.g.
between test runs or demo sessions.

## Quick start

```bash
git clone <this-repo> mockzoo && cd mockzoo
docker compose up --build -d
curl http://localhost:8080/healthz
curl http://localhost:8080/v1/pets
```

This starts Postgres and the app (`docker-compose.yml`: services `db` and `app`). Migrations run
automatically on startup — the app is built with the `migrate` Go build tag.

## Local development

The `Makefile` loads `.env` (falling back to `.env.example`) and exports it into every target.

| Task | Command |
| --- | --- |
| Run the app locally (needs a local Postgres, migrations applied on start) | `make run` |
| Start dependencies only (Postgres) | `make compose-up` |
| Start the whole stack including the app | `make compose-up-all` |
| Tear down | `make compose-down` |
| Unit tests | `make test` |
| Integration tests (in Docker, against a real app + Postgres) | `make compose-up-integration-test` |
| Lint | `make linter-golangci` |
| Format | `make format` |
| Regenerate mocks (after a `contracts.go` change) | `make mock` |
| Tidy / verify modules | `make deps` |
| Vulnerability scan | `make deps-audit` |
| Create a migration | `make migrate-create <name>` |
| Apply migrations | `make migrate-up` |
| Full check before pushing | `make pre-commit` |

`make help` lists every target.

## Configuration

Every setting is an environment variable (parsed with `caarlos0/env`); only `PG_URL` is required —
a container runs with just that one set.

| Variable                | Default                | Description                              |
| ------------------------ | ----------------------- | ----------------------------------------- |
| `APP_NAME`               | `mockzoo`                | Service name                              |
| `APP_VERSION`            | `1.0.0`                  | Service version                           |
| `HTTP_PORT`              | `8080`                   | HTTP listen port                          |
| `HTTP_USE_PREFORK_MODE`  | `false`                  | Fiber prefork mode                        |
| `LOG_LEVEL`              | `info`                   | zerolog level                             |
| `PG_POOL_MAX`            | `2`                      | Postgres pool size                        |
| `PG_URL`                 | *(required)*             | Postgres connection string                |

## OpenAPI

The OpenAPI 3.1 specification is hand-written and lives at [`docs/openapi.json`](docs/openapi.json)
— it is the single source of truth for the `/v1` surface, embedded into the binary via `go:embed`
and served verbatim at `GET /v1/openapi.json`. A browsable reference UI
([Scalar](https://github.com/scalar/scalar), loaded from a CDN) is served at `GET /docs`.

The spec is consumed by [`@hey-api/openapi-ts`](https://heyapi.dev/) with Zod response validation
on the frontend side, so it must exactly match runtime behavior. `getPetById` / the `petId`
parameter / the `Pet` schema name are a hard compatibility contract — see
[`AGENTS.md`](AGENTS.md) for why they must never be renamed. There is no codegen step: any handler
change must update `docs/openapi.json` by hand, in the same commit.

## Architecture

Mockzoo keeps the clean-architecture layering of its upstream: `entity` → `usecase` ← `repo` /
`controller`, with `pkg/` as transport-agnostic infrastructure that never imports `internal/`. See
[`AGENTS.md`](AGENTS.md) for the full dependency rule, the handler contract, and how to add a use
case or a new domain.

## Trimmed from go-clean-template

Mockzoo started as a copy of
[evrone/go-clean-template](https://github.com/evrone/go-clean-template) (MIT licensed, see
[`LICENSE`](LICENSE)) and cut everything not needed for a single-domain REST sandbox:

| Removed                                    | Why                                                                         |
| ------------------------------------------- | ---------------------------------------------------------------------------- |
| gRPC transport                              | One transport (REST) is enough for a sandbox API                            |
| AMQP RPC (RabbitMQ)                         | No async messaging use case here                                            |
| NATS RPC                                    | No async messaging use case here                                            |
| JWT auth + `user` domain                    | Mockzoo has no user accounts; everything is public and unauthenticated      |
| `task` domain                               | Not needed — pets are the only domain                                       |
| `translation` domain + Google Translate webapi | Not needed — pets are the only domain                                    |
| OpenTelemetry / Jaeger tracing              | Unnecessary observability overhead for a mock API                           |
| Prometheus metrics                          | Unnecessary observability overhead for a mock API                           |
| `swag` Swagger codegen                      | Replaced by a hand-written OpenAPI 3.1 spec (needed for exact frontend compatibility) |
| `nginx` reverse proxy                       | Nothing to route between with a single transport                            |
| `README_RU.md` / `README_CN.md`             | Single-maintainer sandbox project; one README is enough                     |
