# AI CONTRACT

This file defines the strict contract between the AI and the OpenTrusty project.
**ALL AI AGENTS MUST READ AND ADHERE TO THIS CONTRACT.**

## 1. Agency & Limits

### Allowed Actions
-   **Bug Fixes**: You may fix bugs that violate the `invariants.md`.
-   **Tests**: You may add tests to verify invariants.
-   **Refactoring**: You may refactor code strictly within the domain boundaries defined in `architecture-map.md`.

### Requires Explicit User Instruction
-   **Schema Changes**: Modifying `internal/store/postgres/migrations`.
-   **New Features**: Adding capabilities not listed in `protocol-scope.md`.
-   **Unlocking**: Modifying THIS file or any file in `docs/_ai/`.

## 2. Documentation Obligations

-   **Read First**: You **MUST** read `docs/_ai/README.md` and all linked files before modifying code.
-   **Update on Change**: You **MUST** follow the `docs/_ai/update-matrix.md`.
    -   *Example*: If you add a column to `users` table, you **MUST** check `invariants.md` and `authority-model.md` and update them if the change affects the defined logic.

## 3. The STOP Condition

**YOU MUST STOP AND ASK THE USER IF:**

1.  A user request asks you to violate an invariant in `docs/_ai/invariants.md`.
2.  You encounter ambiguity in the `authority-model.md`.
3.  You find that the current code contradicts `docs/_ai/` documentation.

## 4. Repository Scope

This repository is the **Core Identity Provider** (headless backend).

### Allowed
-   OAuth2 / OIDC protocol implementation
-   Session management (server-side only)
-   Admin API endpoints consumed by external clients
-   Database migrations and domain logic

### FORBIDDEN (AI Agents MUST NOT)
-   Add UI components, templates, or frontend assets
-   Add static file serving or frontend routing
-   Import React, Vue, Tailwind, or any frontend framework
-   Assume any client (including Control Panel) is trusted
-   Embed UI code or assets into the core binary

### External Consumer Note
The **Control Panel UI** lives in a separate repository (`opentrusty-control-panel`).
It is an **untrusted API client**. All enforcement happens server-side.

> **Status**: ACTIVE
> **Last Updated**: 2025-12-26
