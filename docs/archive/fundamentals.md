# OpenTrusty Constitution: Core Fundamentals

This document defines the four pillars of the OpenTrusty architecture. These definitions serve as the source of truth for all architectural decisions.

## 1. Identity

**Definition**
An **Identity** is a unique, persistent digital representation of a human or machine actor within the system, distinct from the credentials used to verify it.

**What It IS vs. What It Is NOT**
- It **IS** a collection of profile attributes (e.g., ID, email, name) and state (e.g., active, locked).
- It **IS NOT** a set of permissions (Roles).
- It **IS NOT** an authentication method (Password).

**Ownership**
Managed globally by the Identity Provider (OpenTrusty). An identity belongs to a specific **Tenant** but maintains its own lifecycle independent of the applications it accesses.

**Lifecycle**
- **Create**: Provisioned via self-registration or administrative invitation.
- **Update**: mutable profile attributes can be modified by the user or admin.
- **Revoke**: Identities are never hard-deleted to preserve audit trails; they are "soft-deleted" or deactivated.

---

## 2. Tenant

**Definition**
A **Tenant** is the primary logical boundary for data isolation, representing a distinct customer, organization, or environment.

**What It IS vs. What It Is NOT**
- It **IS** a strict security boundary; data from one tenant must never leak to another.
- It **IS NOT** a hierarchical folder or a simple tag; it enforces row-level isolation at the database layer.

**Ownership**
Managed by System Administrators. Tenants are the top-level containers for users, sessions, and configurations.

**Lifecycle**
- **Create**: Established by system administration or automated onboarding flows.
- **Update**: Configuration settings (e.g., security policies) can be adjusted.
- **Revoke**: Can be deactivated, locking all associated identities and resources.

---

## 3. Authentication Protocol

**Definition**
The **Authentication Protocol** is the strictly defined mechanism by which an Identity proves its claim, implemented exclusively via **OAuth2 / OpenID Connect (OIDC)**.

**What It IS vs. What It Is NOT**
- It **IS** a standardized exchange of credentials for time-limited tokens (Access Tokens, ID Tokens).
- It **IS NOT** a custom, proprietary login flow.
- It **IS NOT** authorization; checking credentials does not imply permission to act.

**Ownership**
Owned and enforced by the OpenTrusty Core Service. Client applications do not perform authentication; they delegate it to OpenTrusty.

**Lifecycle**
- **Create (Session)**: Initiated via a successful credential exchange (Login).
- **Issue (Token)**: Short-lived Access Tokens are issued based on an active Session.
- **Revoke**: Sessions can be terminated (Logout), effectively revoking the ability to issue new tokens. Tokens expire automatically.

---

## 4. Authorization

**Definition**
**Authorization** is the evaluation of whether an authenticated Identity is permitted to perform a specific action on a specific resource, implemented via **Role-Based Access Control (RBAC)**.

**What It IS vs. What It Is NOT**
- It **IS** a policy assignment (e.g., "User X has Role Y in Project Z").
- It **IS NOT** global; authorization is always scoped to a specific Tenant or Project context.

**Ownership**
Managed by Tenant Administrators or Resource Owners. The Identity Provider supplies the *facts* (Roles), but the Resource Server (Application) enforces the *decision*.

**Lifecycle**
- **Create**: Permissions are bundled into Roles.
- **Assign**: Roles are granted to Identities for specific scopes.
- **Revoke**: Assignments can be removed instantly, taking effect when the next token is issued or validated.
