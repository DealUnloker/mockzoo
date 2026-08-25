# mockzoo

Mockzoo is a small mock/sandbox pet API. One domain (`pet`), one transport (REST/Fiber), no auth,
Postgres storage, hand-written OpenAPI 3.1 spec served by the app. It exists to be a deliberately
boring, stable backend for the owner's frontend template and demo projects, replacing a dependency
on the public Swagger Petstore sandbox. Module path: `github.com/DealUnloker/mockzoo`. The Go
version is declared in `go.mod`.

This is a trimmed fork of [evrone/go-clean-template](https://github.com/evrone/go-clean-template):
gRPC, AMQP/RabbitMQ, NATS, JWT auth, the `user`/`task`/`translation` domains, OpenTelemetry tracing,
Prometheus metrics and swag codegen were all removed. See the "Trimmed from go-clean-template"
section in `README.md` for the full list and why. Do not re-add any of it without discussing first
— the whole point of this fork is to be smaller than the template it came from.

## Commands — drive everything through the Makefile

The `Makefile` loads `.env` (falling back to `.env.example`) and exports it into every target, and
several targets carry flags that matter. Running the underlying tool directly means running it with
different settings than CI. Use the target, not the tool.

| Task | Use this | Not this |
| --- | --- | --- |
| Install tool binaries (`mockgen`, `migrate`, linters) | `make bin-deps` | `go install ...` |
| Unit tests | `make test` | `go test ./...` — the target adds `-race -covermode atomic` and scopes to `./internal/... ./pkg/...` |
| Integration tests | `make compose-up-integration-test` | `make integration-test` — see below |
| Lint | `make linter-golangci` | `golangci-lint run` |
| Format | `make format` | `gofmt` — the target runs `go fix`, `gofumpt`, and `gci` with the repo's import grouping |
| Regenerate mocks | `make mock` | `mockgen ...` |
| Tidy / verify modules | `make deps` | `go mod tidy` |
| Vulnerability scan | `make deps-audit` | `govulncheck ./...` |
| Start dependencies (Postgres) | `make compose-up` | `docker compose up` |
| Start the whole stack including the app | `make compose-up-all` | `docker compose up` |
| Tear down | `make compose-down` | `docker compose down` |
| Run the app locally | `make run` | `go run ./cmd/app` — the target builds with `-tags migrate` so migrations run on start |
| Create a migration | `make migrate-create <name>` | `migrate create ...` |
| Apply migrations | `make migrate-up` | `migrate -path ... up` |
| Full check before pushing | `make pre-commit` | running the steps by hand |

`make help` lists every target.

Two traps in these targets:

- **`make integration-test` is not the one you want.** It runs `go test ./integration-test/...` on
  the host, where the suite cannot resolve the `app` container hostname it needs, so it always
  fails. `make compose-up-integration-test` is the real entry point.
- **`make migrate-create <name>` prints an error after it succeeds.** The target reads the name via
  `$(word 2,$(MAKECMDGOALS))`, so `make` then tries to build `<name>` as a target and reports
  `No rule to make target`. The migration files are already created; ignore that line.

Never claim a change is done without `make format`, `make linter-golangci` and `make test` passing.

## Dependency rule

```
cmd/app → internal/app → internal/controller/restapi ─┐
                       → internal/repo/persistent/pet ─┤→ internal/usecase (interfaces)
                                                        └→ internal/entity
```

- `internal/entity` — domain types and sentinel errors (`Pet`, `PetStatus`, `PetSpecies`,
  `ErrPetNotFound`, `ErrValidation`). Imports nothing from this module.
- `internal/usecase/contracts.go` — the `Pet` interface the controller calls.
  `internal/usecase/pet/` implements it.
- `internal/repo/contracts.go` — the `PetRepo` interface the use case calls.
  `internal/repo/persistent/pet/` (Postgres) implements it.
- `internal/controller/restapi/v1/` — HTTP handlers plus `request/` and `response/` DTOs.
- `pkg/` — transport-agnostic infrastructure (httpserver, logger, postgres). Must not import
  `internal/`.

Inner layers never import outer ones. Wiring happens exactly once, in `internal/app/app.go`.

## Where code goes

Decide by asking what the code knows about:

| The code knows about… | It belongs in | Shape |
| --- | --- | --- |
| nothing but the domain | `internal/entity/pet.go` | struct + methods (`PetStatus.Valid`, `PetSpecies.Valid`) |
| a domain rule (defaults, clamping, validation) | `internal/usecase/pet/pet.go` | method on `UseCase` |
| SQL, the `pets` table, a driver error code | `internal/repo/persistent/pet/pet.go` | method on `Repo` |
| an HTTP status or the error envelope | `internal/controller/restapi/v1/pet.go` | handler |
| a server, pool or middleware with no domain knowledge | `pkg/<name>/` | reusable package |

Layout rules:

- Constructors are always `New`, return the *interface* from `contracts.go`, and take their
  dependencies as arguments. No package-level state — `gochecknoglobals` and `gochecknoinits` are on.
- Unexported package constants are prefixed with an underscore (`_defaultLimit`, `_seedIDStart`).
  `mnd` rejects bare numeric literals, so add a constant rather than inlining one.
- `internal/repo/persistent/pet/seed.go` is the single source of truth for the three canonical
  seed pets; `Reset` uses it directly and the `seed_pets` migration mirrors the same values by
  hand. Keep both in sync — there is no codegen linking them.

## The handler contract

Every controller method does the same five things in the same order. Copy the shape from a
neighbouring handler in `internal/controller/restapi/v1/pet.go`.

1. **Parse path/query params** into their typed form (`petId` → `int64` via `parsePetID`,
   `status`/`species` query params → `*entity.PetStatus` / `*entity.PetSpecies`). A parse failure
   is a 400, not a 500.
2. **Decode into a transport-local DTO** from `v1/request/` for bodies (`request.CreatePet`,
   `request.UpdatePet`). Never decode into an `entity` type, and never pass a `request.*` type into
   the use case — map it into a `usecase.CreatePetInput` / `usecase.UpdatePetInput` first.
3. **Validate the DTO** with the injected `*validator.Validate` (`r.v.Struct(req)`). Structural
   validation (required, max length, oneof) lives here. Domain validation (enum validity given the
   full picture, "name cannot be empty on update") lives on the entity or the use case and returns
   `entity.ErrValidation` — don't duplicate it in the controller.
4. **Call the use case** with `ctx.UserContext()` (not `ctx.Context()`).
5. **Map the error and encode the response.** `errors.Is` against `entity.ErrPetNotFound` /
   `entity.ErrValidation` → the matching HTTP status and `{"error":{"code","message"}}` envelope
   (`_codePetNotFound` / `_codeValidationFailed` / `_codeInternal`). On success, return an `entity`
   type directly when its JSON shape is already the API shape (`entity.Pet` is tagged for this);
   otherwise use a `response.*` wrapper (`response.PetList`, `response.ResetResult`).

## The hard API-compatibility rule

`docs/openapi.json` is hand-written and is the single source of truth for the `/v1` surface — there
is no codegen step generating it from annotations. The owner's frontend template consumes this spec
with `@hey-api/openapi-ts` and Zod response validation, generating a typed client against it. Three
names in that spec must never be renamed, because the frontend's old Swagger-Petstore-based client
already used them and the frontend is not touched when this spec changes:

- the `getPetById` operationId
- the `petId` path parameter name
- the `Pet` schema name

Breaking any of these breaks the generated frontend client silently (wrong operation name, wrong
generated function signature) rather than with a compile error. If you rename them anyway, say so
loudly and separately — don't bundle it with an unrelated change.

**Update `docs/openapi.json` in the same commit as any handler change** under
`internal/controller/restapi`. Route paths, request/response shapes, status codes and error codes
in the spec must exactly match runtime behavior — that is the entire point of hand-maintaining it
instead of trusting annotations to stay in sync.

## The seed-data / reset contract

Three pets are always present after a fresh migration or a call to `POST /v1/admin/reset`:

| id | name | species | status |
| --- | --- | --- | --- |
| 1 | Barsik | cat | available |
| 2 | Rex | dog | pending |
| 3 | Kesha | bird | sold |

These ids are stable and must never change — demos and the frontend's own fixtures rely on
`GET /v1/pets/1` always being Barsik. The `pets_id_seq` sequence is set to 1000 after seeding, so
every pet created through `POST /v1/pets` gets id 1001 or higher and can never collide with a seed
row. `Reset` (`internal/repo/persistent/pet/pet.go`) does this in one transaction: truncate, reinsert
the three rows from `seedPets()`, `setval`. If you change the seed data, update `seedPets()` in
`internal/repo/persistent/pet/seed.go`, the `seed_pets` migration, and the seed table in `README.md`
together.

## Conventions that will bite you

**Mocks are generated, not written.** `internal/usecase/mocks_*_test.go` come from `make mock`
(runs `mockgen` against the two `contracts.go` files — see the `mock` target in `Makefile`). After
any contract change run `make mock` before running tests.

**Error style.** Domain failures are sentinel errors in `internal/entity/errors.go`
(`ErrPetNotFound`, `ErrValidation`). Wrap `ErrValidation` with `fmt.Errorf("%w: <detail>", ...)` to
attach a specific reason rather than adding a new sentinel. Every other error is wrapped with the
layer path: `fmt.Errorf("PetUseCase - Create - uc.repo.Create: %w", err)`. The `err113` linter
rejects `errors.New`/`fmt.Errorf` for new dynamic errors — add or reuse a sentinel in `entity`
instead. Controllers are the only place that translates a sentinel into a transport status.

**Config is environment-only.** `config/config.go` parses env vars with `caarlos0/env`. Every field
except `PG_URL` carries an `envDefault`, so the container only needs `PG_URL` set. A new setting
needs the struct field, a default (unless it must be required), an entry in `.env.example`, and —
if the container needs it — the `environment` block in `docker-compose.yml`.

**The linter is strict and non-obvious.** `.golangci.yml` enables ~45 linters. The ones that reject
otherwise-fine code: `wsl_v5` and `nlreturn` (blank line required before `return` and around
blocks), `funlen` (65 lines / 40 statements), `gocyclo` (10), `gocognit` (15), `mnd`,
`gochecknoglobals`, `exhaustive` (switches over enum types need every case or a `default`), `dupl`
(threshold 100).

## Adding a use case method

1. Add the method to `internal/usecase/contracts.go` (`Pet` interface) and, if it needs new
   persistence, to `internal/repo/contracts.go` (`PetRepo` interface).
2. Implement it in `internal/usecase/pet/pet.go` and, if needed, `internal/repo/persistent/pet/pet.go`.
3. `make mock`, then extend `internal/usecase/pet_test.go` (table-driven, gomock, `t.Parallel()` —
   `paralleltest` enforces it).
4. Add the handler in `internal/controller/restapi/v1/pet.go` + route in
   `internal/controller/restapi/v1/router.go` + update `docs/openapi.json` in the same commit.
5. Extend `integration-test/pet_test.go` if the change is observable end-to-end.

## Database

`migrations/` holds golang-migrate pairs (`<timestamp>_<name>.up.sql` / `.down.sql`). Create with
`make migrate-create <name>`, apply with `make migrate-up`. The app applies them itself at startup
only when built with the `migrate` build tag (`internal/app/migrate.go`) — that is what `make run`
and the Dockerfile do. Queries are built with Squirrel via the embedded `*postgres.Postgres`; there
is no ORM and no raw string concatenation.

## Tests

- Unit: `internal/...` and `pkg/...`, run by `make test` with `-race`. Use-case tests use the
  generated gomock mocks, are table-driven, and call `t.Parallel()` (`paralleltest` enforces it).
- Integration: `integration-test/` runs *inside* the docker network — it resolves the service as
  host `app`, so it fails on the host machine. Always run it through
  `make compose-up-integration-test`.
