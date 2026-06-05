# ADR-003: Replacing Axios with Fetch API and Centralizing the HTTP Client

## Status

Accepted

## Date

2026-05-10

## Context

Initially the frontend was intended to use Axios for all HTTP requests. However, modern browsers ship a native Fetch API that covers the same use cases without adding a third-party dependency. In addition, the configuration logic (base URL, JWT injection, and error handling) was duplicated across multiple service files, creating a maintenance burden and making the codebase harder to scale.

## Decision

We have decided to:

1. **Remove Axios** as a project dependency to reduce the overall bundle size.
2. **Implement a custom wrapper** in `src/services/api_client.ts` using the native Fetch API.
3. **Use function declarations** (`export function`) instead of arrow functions to ensure consistency, improve readability in stack traces, and take advantage of function hoisting.
4. **Centralise JWT token injection and 401 Unauthorized handling** (interceptor logic) inside this single entry point.
5. **Abstract the service layer** so that each service (e.g. `user_service.ts`) only defines its specific path and endpoints, delegating the network call to the shared client.

## Consequences

### Positive

- **Reduced Bundle Size**: Eliminating Axios removes kilobytes of JavaScript the client no longer needs to download.
- **DRY Principle**: Changes to the base URL or security logic are made in a single file rather than across every service.
- **Strong Typing**: TypeScript generics `<T>` in the core fetcher guarantee that all service responses are strictly typed from the source.
- **Decoupling**: The UI and services no longer care about the underlying network implementation; they only interact with the data layer.

### Negative / Risks

- **Manual Error Handling**: Unlike Axios, Fetch does not automatically reject promises on HTTP error status codes (4xx/5xx). We had to manually implement the logic to parse error messages and throw exceptions.
- **Missing Features**: Native Fetch lacks out-of-the-box features such as request timeouts or upload progress tracking, which would need to be added manually if required in the future.

## Alternatives Considered

- **Keep Axios**: Rejected because the bundle-size cost was unjustified for our relatively simple request patterns.
- **Use a lighter wrapper such as ky or redaxios**: Rejected to avoid introducing yet another dependency when the native API is sufficient.
