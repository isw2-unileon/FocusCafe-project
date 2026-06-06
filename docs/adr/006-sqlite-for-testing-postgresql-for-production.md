# ADR-006: SQLite for Testing, PostgreSQL (Supabase) for Production

## Status

Accepted

## Date

2026-04-22

## Context

The application requires a relational database for both production and automated testing. In production, the system connects to a managed Supabase PostgreSQL instance. However, requiring every developer and CI runner to spin up a full PostgreSQL server (or the Supabase local stack) just to execute unit and integration tests would slow down the feedback loop and complicate the CI pipeline.

We needed a strategy that:
- Guarantees fast, isolated test execution.
- Avoids external network dependencies during tests.
- Uses the same GORM abstraction so repository code does not change between test and production.
- Remains compatible with the production PostgreSQL schema.

## Decision

We will use **SQLite** (pure-Go driver, no CGO) for all automated tests, while the production deployment continues to target **PostgreSQL via Supabase**.

1. **Production**: `backend/internal/database/database.go` opens a PostgreSQL connection using the `DATABASE_URL` environment variable pointing to Supabase.
2. **Tests**: Each test package creates its own isolated SQLite in-memory database via `gorm.Open(sqlite.Open("file:testName?mode=memory&cache=shared"))`. Migrations (`AutoMigrate`) run on this in-memory instance before tests execute.
3. **Driver**: The `github.com/glebarez/sqlite` driver is used because it is a pure-Go implementation that does not require CGO or a GCC toolchain, making it frictionless on Windows and CI runners.
4. **Compatibility**: All GORM operations (migrations, associations, transactions) are written using GORM's generic API, which translates correctly to both SQLite and PostgreSQL dialects.

## Consequences

### Positive

- **Fast Feedback Loop**: Tests start instantly without waiting for PostgreSQL boot or Docker containers.
- **Zero External Dependencies in CI**: No need to configure a PostgreSQL service in GitHub Actions or local machines.
- **Isolated Databases**: Because each test receives its own named in-memory database (`file:TestName?mode=memory`), tests cannot pollute each other's data.
- **Single Codebase**: Repository and service code remains identical; only the database driver string changes between environments.

### Negative

- **Dialect Differences**: SQLite and PostgreSQL handle certain SQL features differently (e.g. `RANDOM()` vs `RAND()`, `SERIAL` vs `INTEGER PRIMARY KEY AUTOINCREMENT`). We must ensure GORM migrations are dialect-neutral.
- **No Full-Text Search or Advanced Types**: Features such as PostgreSQL `tsvector` or JSONB operators are unavailable in SQLite, so we avoid them in the shared schema.
- **Risk of Schema Divergence**: If a developer adds a PostgreSQL-specific constraint or type without testing it in SQLite, the CI may pass while production breaks. We mitigate this by keeping the schema simple and GORM-driven.
- **Connection Pooling Behaviour**: SQLite in-memory with `cache=shared` behaves differently from PostgreSQL connection pools. We limit `SetMaxOpenConns(1)` in test setup where necessary to prevent "database is locked" errors.

## Alternatives Considered

- **Use PostgreSQL in Docker for tests**: Rejected because it adds ~5-10 seconds to every test run and complicates Windows developer setup.
- **Use a shared SQLite file on disk**: Rejected because file-based SQLite introduces I/O latency and stale state between test suites.
- **Use `testcontainers-go` with PostgreSQL**: Rejected to avoid Docker dependency in the CI pipeline for a university project with limited infrastructure.
