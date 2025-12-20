# Protocol Error Model: OAuth2 / OIDC

This document defines the strict requirements for error handling, classification, and propagation within the OAuth2 and OpenID Connect (OIDC) protocol modules of OpenTrusty.

## 1. Protocol Error Definition

A **Protocol Error** is a specialized domain object that encapsulates a failure in a delegated authorization or identity exchange according to the semantics of a specific protocol specification (e.g., RFC 6749).

- **vs. Domain Error**: A domain error (e.g., `identity.ErrUserNotFound`) describes *what* happened in the business logic. A Protocol Error (e.g., `invalid_grant`) describes *how* that failure affects the protocol state.
- **vs. Transport Error**: A transport error (e.g., `401 Unauthorized`) is an implementation detail of the delivery mechanism. A Protocol Error exists independently of whether it is delivered via JSON, redirect parameters, or SOAP.

## 2. Canonical Error Categories

The system MUST categorize failures using the following standard identifiers:

### OAuth2 Standard Errors (RFC 6749)
- `invalid_request`: The request is missing a parameter, includes an unsupported parameter, or is otherwise malformed.
- `invalid_client`: Client authentication failed (e.g., unknown client, no authentication included, unsupported method).
- `invalid_grant`: The provided authorization grant (code, refresh token) or password is invalid, expired, revoked, or used for the wrong client/tenant.
- `unauthorized_client`: The authenticated client is not authorized to use this authorization grant type.
- `unsupported_grant_type`: The authorization grant type is not supported by the authorization server.
- `invalid_scope`: The requested scope is invalid, unknown, malformed, or exceeds the scope granted by the resource owner.
- `server_error`: The authorization server encountered an unexpected condition that prevented it from fulfilling the request.
- `temporarily_unavailable`: The server is currently unable to handle the request due to maintenance or overload.

### OIDC Specific Errors (OIDC Core)
- `interaction_required`: The Authorization Server requires user interaction of some form.
- `login_required`: The Authorization Server requires user authentication.
- `consent_required`: The Authorization Server requires user consent.
- `account_selection_required`: The User Agent is required to select a user account.

## 3. Mandatory Error Fields

Every protocol error response MUST contain or support the following fields:

- **`error`**: (Required) A single ASCII error code from the categories above.
- **`error_description`**: (Optional) Human-readable ASCII text providing additional information. This field MUST NOT contain stack traces, database IDs, or PII.
- **`error_uri`**: (Optional) A URI identifying a human-readable web page with information about the error.
- **`state`**: (Required if present in request) If the request included a `state` parameter, the error response MUST include the same value.

## 4. Error Propagation Rules

- **Translation Boundary**: Domain-specific errors MUST be converted into Protocol Errors at the **Protocol Service** boundary.
- **Privacy Policy**: Internal details (SQL errors, specific validation failures, internal service names) MUST NOT be exposed in `error_description`.
- **Classification Invariant**: All unhandled domain errors MUST be mapped to `server_error` to prevent information leakage.
- **Security Check**: Errors involving `client_secret` or `redirect_uri` mismatches MUST BE handled with caution to prevent enumeration or redirection attacks.

## 5. Mapping to Transport (HTTP)

The Transport Layer is responsible ONLY for the serialization and delivery of Protocol Errors:

### Status Code Rules
- `invalid_client` involving `Basic Auth`: MUST return `401 Unauthorized` with appropriate `WWW-Authenticate` header.
- Other Protocol Errors in Token/UserInfo endpoints: MUST return `400 Bad Request`.
- `server_error`: MUST return `500 Internal Server Error`.

### Delivery Mechanism
- **Redirect-based**: Errors during the `authorize` flow MUST be appended as query parameters to the validated `redirect_uri`.
- **JSON-based**: Errors during `token`, `userinfo`, or `discovery` exchanges MUST be returned in a JSON body with an `application/json` Content-Type.
- **Forbidden**: Protocol errors MUST NEVER be returned as raw text or HTML unless no other delivery mechanism is viable (e.g., `redirect_uri` is invalid).

---
**Rule Citation**: This model satisfies **Rule 4.2 (Compliance)** and **Rule 2 (Domain-Driven Architecture)** by ensuring that the protocol remains the "source of truth" for its own error semantics.
