# 0004 — Dedicated database and standalone migrations

## Status

Accepted

## Date

2026-10-14 (commit `5f469c4d`)

## Context

HASkinLib originally shared migration bookkeeping with HRPAuth: a single
`schema_migrations` table with a `service` column tracked several services'
migrations in one database ("shared era"). That coupling made independent
releases risky: one service's migration state could block or corrupt
another's.

## Decision

Option 1 was adopted: **HASkinLib runs against a dedicated database with a
standard, standalone migration chain.**

1. Migrations are embedded SQL files executed by golang-migrate against a
   plain `schema_migrations (version, dirty)` table owned by HASkinLib.
2. At startup, if an existing `schema_migrations` table has the legacy
   `service` column, startup **fails hard** with an explicit message asking
   for a dedicated database — it will not silently continue on a shared
   legacy database.
3. Migration SQL is table-prefix-rewritten at load time so
   `texture_list_skin` / `texture_list_cape` become
   `<table_prefix>texture_list_skin|cape` (default prefix `hasl_`).
4. The baseline migration (000006) creates the two tables and migrates rows
   from a legacy single `texture_list` table if present, then drops it.
   Follow-ups 000007/000008 cap skin/cape `name` lengths.

## Consequences

- Benefit: HASkinLib's schema evolves independently; dirty-state detection
  and versioning are standard golang-migrate behavior.
- Cost: operators must provision a dedicated database (or at least a
  database without the legacy shared `schema_migrations` shape).
- Tension worth knowing: built-in *defaults* still read `db_name: hrpa`
  with comment "defaults follow HRPAuth to ease sharing the same database
  server". This is connection-level convenience only — the migration guard
  (point 2) is what enforces database-level independence. Deployments are
  expected to set `database.db_name` to a dedicated database; the deployer
  template injects `${DB_NAME}` rather than hardcoding `hrpa`.
- Constraint: GORM models and migration SQL must stay aligned manually
  (models declare the same columns/indexes as the baseline SQL, including
  the odd lowercase column name `previewfile`).
- The `uid` column references HRPAuth's `users.uid` by convention only —
  there is no foreign key across databases.

## Evidence

- Decision comment ("方案 1 明确要求独立数据库"): `controllers/startup_controller.go::EnsureMigrations`.
- Legacy shared-table rejection: `controllers/startup_controller.go::ensureStandardMigrationsTable`
  (error text: "legacy shared schema_migrations table
  detected ... option 1 requires a dedicated HASkinLib database").
- Prefix rewriting of embedded migrations: `controllers/startup_controller.go::buildMigrationFS`.
- Baseline schema + legacy `texture_list` migration:
  `database/migrations/000006_merged_baseline.up.sql::texture_list_skin`,
  `database/migrations/000006_merged_baseline.up.sql::texture_list_cape`.
- Name-length migrations: `database/migrations/000007_limit_skin_name.up.sql`,
  `database/migrations/000008_limit_cape_name.up.sql`.
- Default DB values and comment: `config/config.go (field: db_name)`
  (default `db_name: hrpa`), `config/config.go (field: table_prefix)`
  (`table_prefix: hasl_`).
- Model/table naming: `models/models.go::TextureListSkinBase`,
  `models/models.go::TextureListCape` (unique key `(uid, hash)`,
  column `previewfile`).
- Deployer template uses `${DB_NAME}`: repo HA-Services-Deployer
  (@ `657bcb88`),
  [xref] `/home/lnb/Desktop/HA/HA-Services-Deployer/deploy/config/haskinlib.yaml.tmpl:L3-L9 @ 657bcb88`.
- Decision commit: `5f469c4d864294b84d46bd6e33993abf1e79da58`
  ("refactor: switch HASkinLib to standard standalone migrations",
  2026-10-14). Earlier multi-service migration support (the shared era) was
  introduced by `8f73ca959a7a3eb8f49ae3b22c18c73f6a5e240c` (2026-08-02) and
  replaced by this decision; the merged baseline replaced incremental
  migrations in `ad8606846a92349f040b8a1a86ce7843c6125075` (2026-08-02).
