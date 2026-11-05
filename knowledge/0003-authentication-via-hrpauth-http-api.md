# 0003 — Authentication via HRPAuth HTTP API

## Status

Accepted

## Date

2026-10-14

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
4. Only `Authorization: Bearer` tokens are accepted.

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

- Auth flow: `controllers/auth.go::AuthenticateUser` (verify then refresh;
  UUID required).
- HTTP endpoints called: `clients/core_client.go::VerifyToken`
  (`POST /internal/verify-token`), `clients/core_client.go::GetUserInfo`
  (`GET /services/users/` + uuid).
- OAuth2 client_credentials wiring: `utils/core_api_client.go::GetCoreAPIClient`
  (TokenURL = base_url + `/oauth/token`; fallback when unconfigured is the
  unauthenticated branch in the same function).
- Default `core_api.base_url` (`http://127.0.0.1`, port `2778`):
  `config/config.go (field: base_url)`.
- Upstream side (repo HRPAuth @ `8dd67593`): routes registered under an
  empty group prefix —
  [xref] `/home/lnb/Desktop/HA/HRPAuth/main.go:L174 @ 8dd67593`
  (`api := r.Group("")`),
  [xref] `/home/lnb/Desktop/HA/HRPAuth/main.go:L197 @ 8dd67593`
  (`POST /internal/verify-token`),
  [xref] `/home/lnb/Desktop/HA/HRPAuth/main.go:L201 @ 8dd67593`
  (`GET /services/users/:uuid`),
  [xref] `/home/lnb/Desktop/HA/HRPAuth/main.go:L180 @ 8dd67593`
  (`POST /oauth/token`). Paths therefore match the client. Verified
  2026-11-04.
- Module path lineage: `main.go (field: github.com/lnb/HRPAuth-Backend-Go)`
  (`github.com/lnb/HRPAuth-Backend-Go/...` imports).
