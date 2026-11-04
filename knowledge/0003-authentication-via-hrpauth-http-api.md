# 0003 — Authentication via HRPAuth HTTP API

## Status

Accepted

## Date

2026-10-14 (commit `dca3325a`; user refresh + table prefix in
`b59ee173a1852bfdc376c9a15e29f64b63f07b2d`, same-day follow-up)

## Context

HASkinLib needs to know *which HA user* uploads or deletes a texture. The
user database, OAuth2 server, and token lifecycle all live in HRPAuth. Two
options existed: import HRPAuth's backend code as a shared Go library
(shared models, shared DB access), or call HRPAuth over HTTP as an external
service.

Note the code still carries the fork lineage: the Go module path is
`github.com/lnb/HRPAuth-Backend-Go`. This is historical naming from being
spun out of the HRPAuth backend; **no code is shared at build time** — the
service compiles its own copies of config/database/controllers.

## Decision

Authenticate through HRPAuth's **HTTP API**, not a shared library:

1. Every authenticated route extracts the `Authorization: Bearer` token and
   calls `POST {core_api.base_url}/internal/verify-token` to resolve it to a
   user identity (`uid`, `uuid`, `email`, `username`).
2. The service then refreshes the identity with
   `GET {core_api.base_url}/services/users/:uuid` and uses the refreshed
   record (so upstream renames/state changes are authoritative).
3. Service-to-service authentication uses OAuth2 **client_credentials**
   against `{core_api.base_url}/oauth/token` when `client_id`/`client_secret`
   are configured; without credentials it falls back to unauthenticated
   calls (logged as such).
4. Legacy form-field auth (`remember_token`, `rt`) was removed; only Bearer
   tokens are accepted.

## Consequences

- Benefit: no shared Go module with HRPAuth; the two can evolve and deploy
  independently. The only contract is the HTTP surface.
- Cost: every authenticated request costs two extra HTTP roundtrips to
  HRPAuth (verify + refresh); HRPAuth availability is a hard dependency for
  upload/delete/profile routes (public gallery/pull/preview are unaffected).
- Cost: the response shape of `/internal/verify-token` is duplicated as a
  local struct (`models.CoreUserIdentity`); upstream shape changes require a
  code change here.
- Constraint: `core_api.base_url` must point at the HRPAuth **root** service
  (the endpoints are registered at root level, not under an API prefix).
- Failure mapping: verify failure → `401 oauth_invalid_grant`; missing
  token → `401 oauth_login_required`; missing base URL → `500`.

## Evidence

- Auth flow: `controllers/auth.go:14-45` (verify then refresh; UUID
  required).
- HTTP endpoints called: `clients/core_client.go:71-73`
  (`POST /internal/verify-token`), `clients/core_client.go:94`
  (`GET /services/users/` + uuid).
- OAuth2 client_credentials wiring: `utils/core_api_client.go:36-42`
  (TokenURL = base_url + `/oauth/token`), fallback when unconfigured
  `utils/core_api_client.go:25-34`.
- Default `core_api.base_url` `http://127.0.0.1:2778`: `config/config.go:71-75`.
- Upstream side (repo HRPAuth @ `8dd67593`): routes registered under an
  empty group prefix — `main.go:174` (`api := r.Group("")`),
  `main.go:197` (`POST /internal/verify-token`), `main.go:201`
  (`GET /services/users/:uuid`), `main.go:180` (`POST /oauth/token`).
  Paths therefore match the client. Verified 2026-11-04.
- Module path lineage: `main.go:7-9` (`github.com/lnb/HRPAuth-Backend-Go/...`).
- Decision commit: `dca3325a5c36c5e9077bd90037cb6cba48020e96`
  ("refactor: verify skinlib users through HRPAuth core", 2026-10-14).
- Legacy remember-token removal: commit
  `3423290f3d5546b1b5c782c7643864fe93aa37dd` ("chore: remove remember token
  code and add build script", 2026-10-03); deprecation note in
  `docs/skinlib-api.md:31`.
