# Tenant Resolution Contract

This document defines the strict architectural boundaries and responsibilities for resolving a **Tenant Identity** within OpenTrusty, specifically addressing blocker **B-TENANT-01**.

## 1. Resolution Boundary: The Transport Edge

Tenant resolution MUST be performed exclusively at the **Transport Edge** (Entry Point) of the system.

- **Layer**: Middleware or dedicated Transport-level Interceptors.
- **Responsibility**: Inspect incoming request metadata (Headers, Hostnames, or Path segments) to identify the target Tenant.
- **Outcome**: The resolved `TenantID` MUST be injected into the request-scoped `Context`.

## 2. Forbidden Resolution Locations

To protect the core domain from transport contamination, Tenant resolution MUST NOT be performed in:

- **Service Layer**: Services MUST assume the `TenantID` is already present in the `Context`. They MUST NOT attempt to look it up from headers or request objects.
- **Repository Layer**: Repositories are consumers of the `TenantID` for row-level isolation. They MUST NOT have logic to "default" or "discover" a tenant.
- **Domain Logic**: Business rules should be tenant-agnostic or rely strictly on the provided context.

## 3. Failure Behavior: Fail-Fast (Pessimistic)

If a target Tenant cannot be strictly resolved from the request:

- **Prohibition**: The system MUST NOT "fail open" to a default tenant (e.g., `"default"`).
- **Required Action**: The request MUST be rejected immediately at the Transport Edge.
- **Response**: Return `HTTP 400 Bad Request` (if parameters are missing) or `HTTP 401 Unauthorized` (if the identifier is present but invalid/unauthorized).

## 4. Exceptions

Strict resolution may be bypassed ONLY for:
- Global health check endpoints (e.g., `/health`).
- Global administrative endpoints (e.g., Platform-level metrics).
- Static assets shared across the platform.

---
**Rule Citation**: This contract satisfies **Rule 2.1 (Isolation)** and **Rule 2.3 (No Cross-Talk)** by ensuring tenant context is a hard requirement for all tenant-scoped business logic.
