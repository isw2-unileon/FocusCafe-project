# ADR-004: Supabase as Backend-as-a-Service

## Status

Accepted

## Date

2025-04-20

## Context

The project requires user authentication and a relational PostgreSQL database. Building and maintaining a custom authentication system (password hashing, JWT issuance, token refresh, OAuth flows) and managing a PostgreSQL instance directly would consume significant development time outside the project's scope.

We needed a solution that:
- Provides secure user authentication out of the box.
- Offers a managed PostgreSQL database.
- Supports OAuth providers (specifically Google) without custom OAuth server code.
- Keeps infrastructure overhead minimal for a university project with limited DevOps resources.

## Decision

We will use **Supabase** as our Backend-as-a-Service (BaaS) to handle:

- **Authentication**: Email/password and OAuth (Google) sign-in, with automatic JWT management.
- **Database**: Managed PostgreSQL instance with Row Level Security (RLS) policies.
- **JWT Validation**: Supabase exposes a JWKS endpoint that our Go backend uses to validate tokens via the `github.com/MicahParks/keyfunc` library.

The Go backend connects to the same PostgreSQL instance for application data (users, groups, orders, sessions), while leveraging Supabase Auth for identity management.

## Consequences

### Positive

- **Zero Auth Infrastructure Maintenance**: We do not manage password hashes, refresh tokens, or email verification flows.
- **Built-in OAuth**: Google sign-in works with minimal configuration (Client ID + Secret in Supabase dashboard).
- **Managed PostgreSQL**: Automated backups, connection pooling, and scaling are handled by Supabase.
- **RLS Policies**: Database-level security rules provide an additional safety net beyond application-level checks.

### Negative

- **Vendor Lock-in**: Migrating away from Supabase Auth would require rebuilding the user identity layer.
- **Local Development Overhead**: Developers must run the Supabase CLI (or Docker stack) locally to replicate the auth and database environment.
- **Network Dependency**: Authentication requests go to Supabase's servers; outages or latency affect login/sign-up flows.

## Alternatives Considered

- **Firebase Auth + Cloud Firestore**: Rejected because Firestore is a NoSQL document database; we needed the relational model (groups, orders, users) that PostgreSQL provides.
- **Auth0 + Self-hosted PostgreSQL**: Rejected due to Auth0's free-tier limitations and the extra work of managing our own database server.
- **Custom JWT + bcrypt + Self-hosted PostgreSQL**: Rejected because the time required to implement and secure a custom auth system exceeded the project timeline.
