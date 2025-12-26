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

## 5. System Planes

This repository represents **TWO** logical planes ONLY:

### 5.1. Authentication Plane (`auth.*`)
-   OIDC/OAuth2 protocol endpoints
-   Server-rendered login, consent, and error pages
-   Session cookie management
-   AI MUST treat login pages as **protocol surfaces**, NOT as UI components

### 5.2. Management API Plane (`api.*`)
-   REST APIs for tenant, user, client, and policy management
-   Consumed by Control Panel UI, CLI tools, and automation
-   AI MUST treat all consumers as **untrusted external clients**

### 5.3. Control Panel UI (`console.*`) — NOT IN THIS REPO
-   Lives in `opentrusty-control-panel` repository
-   AI MUST NOT add any Control Panel UI code to this repository
-   AI MUST interact with UI ONLY via Management API references

## 6. Cross-Repo Contract Awareness

**AI MUST** be aware of the `docs/_ai/integration-contract.md` which binds this repository to `opentrusty-control-panel`.

-   **Constraint**: Changes to the Management API (`api.*`) MUST be backward compatible or coordinated with Control Panel updates.
-   **Constraint**: If you modify `auth.*` session logic, you MUST check if it breaks Control Panel's cookie assumption.

## 7. Documentation Update Triggers

**AI MUST update relevant `docs/_ai/` files when a change affects:**

| Change Type | Files to Update |
|-------------|-----------------|
| Domain boundaries | `architecture-map.md` |
| Tenant resolution logic | `invariants.md`, `authority-model.md` |
| Protocol surface (endpoints, flows) | `protocol-scope.md` |
| RBAC roles or scopes | `authority-model.md` |
| New API endpoints | `update-matrix.md` |

> **Status**: ACTIVE
> **Last Updated**: 2025-12-26
