# Audit Logging & Access Model

## Overview

Audit logs in OpenTrusty provide an immutable, append-only record of all security-sensitive events. This document defines the storage guarantees, visibility rules, and the mandatory access declaration flow for Platform Administrators.

---

## 1. Immutability Guarantee

Audit logs are authoritative for compliance and security forensics.
- **Append-Only**: Logs are written once and never modified.
- **No Suppression**: The system explicitly prohibits any API, CLI, or administrative function for deleting, truncating, or suppressing logs.
- **Uniformity**: Standard standardized redaction rules apply to all logs to ensure no PII or secrets are stored.

---

## 2. Platform Admin: Scoped Access Flow

Platform Administrators do **NOT** have default visibility into tenant-scoped audit logs. To obtain access, they must follow the **Explicit Access Declaration** flow.

### Phase 1: Access Declaration
Access is requested via an intent-based API call or UI form.
- **Endpoint**: `POST /api/platform/audit-queries`
- **Required Fields**:
  - `tenant_id`: The specific UUID of the tenant to inspect.
  - `start_date` / `end_date`: Time range (Maximum 30-day window).
  - `reason`: Mandatory free-text justification.

### Phase 2: Audit-of-Audit
Upon successful validation of the declaration, the system emits a primary audit record:
- **Event Type**: `audit.read`
- **Metadata**:
  - `actor_id`: ID of the Platform Admin.
  - `target_tenant_id`: The tenant being accessed.
  - `reason`: The provided justification.
  - `window_start` / `window_end`: Declared time range.

### Phase 3: Results Retrieval
The Platform Admin can then retrieve the results using the Query ID.
- **Endpoint**: `GET /api/platform/audit-queries/{query_id}/results`

---

## 3. UI Design Requirements

The Control Panel UI enforces the two-phase flow to prevent accidental or implicit cross-tenant visibility.

1. **Audit Declaration Entry**: Platform Admin is presented with the Access Declaration form. No logs are rendered until submission.
2. **Access Banner**: When viewing scoped logs, a prominent banner is displayed:
   > **READ-ONLY SCOPED ACCESS**
   > This view is restricted to Tenant [ID] for the window [Start] to [End]. This access has been recorded in the system audit log.
3. **Isolation**: No "Platform-wide" mixed view exists.

---

## 4. Permission Mapping

| Role | Permission | Logic |
| :--- | :--- | :--- |
| **Platform Admin** | `audit:read` | Scoped via Declaration |
| **Tenant Owner** | `audit:read` | Full access (own tenant) |
| **Tenant Admin** | `audit:read` | Operational access (own tenant) |
| **Tenant Member** | None | No access |
| **All Other Roles** | None | No access |

---

## 5. Security Configuration

To ensure the integrity of administrative access declarations, the system requires a dedicated signing key for stateless query tokens.

- **Variable**: `AUDIT_QUERY_SIGNING_KEY`
- **Requirement**: Must be a high-entropy secret (e.g., a 32-character random string).
- **Enforcement**: OpenTrusty will refuse to start if this variable is empty.
