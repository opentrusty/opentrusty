# UI Smoke Test Report

**Date:** 2025-12-26
**Environment:** Local Dev (Single Binary)
**Binary Version:** `dev` (Modified)
**Database:** Postgres 16 (Docker) - Port 5433

## Executive Summary
The OpenTrusty Control Plane was tested using a single binary with embedded UI assets. The system successfully bootstrapped from a clean database, auto-provisioned the initial platform administrator, and permitted login. Navigation to core areas works, but functional management capabilities (Tenant Creation, OAuth Client Management) are currently stubs ("Coming soon"), preventing full end-to-end workflow validation.

> “This validates that OpenTrusty is operable as a human-administered control plane (Foundation Layer).”

## Scenario Results

### ✅ Scenario 1: Login Page Reachability
- **Result:** PASSED
- **Observation:** Server started successfully. Navigation to `/admin/` redirected to `/admin/login`.
- **Screenshot:** `scenario_1_reachability.png`

### ✅ Scenario 2: Platform Admin Bootstrap
- **Result:** PASSED (With Code Fix)
- **Observation:** The system correctly identified a fresh database. `OT_BOOTSTRAP_ADMIN_EMAIL` triggered the creation of the user and assignment of the Platform Admin role.
- **Credential Generation:** Password was auto-generated and printed to `server.log`.
    - **Email:** `admin@opentrusty.io`
- **Deviation:** Initial attempt failed due to missing user-creation logic in bootstrap service. This was patched to provision identity if missing.
- **Screenshot:** `scenario_3_landing.png` (Proof of successful login)

### ✅ Scenario 3: Post-Login Landing Decision
- **Result:** PASSED
- **Observation:** Redirected to `/admin/platform/tenants` (Platform Context).
- **Navigation:** Sidebar displays "Tenants" and "Platform Admins". No tenant-scoped items visible.
- **Screenshot:** `scenario_3_landing.png`

### ❌ Scenario 4: Tenant Creation (Platform Admin)
- **Result:** BLOCKED (Feature / UI Stub)
- **Observation:** The "Tenants" page loads but displays "Coming soon". No "Create Tenant" button exists.
- **Impact:** Cannot verify tenant creation or context switching.
- **Screenshot:** `scenario_4_blocked_missing_feature.png`

### ⏭️ Scenario 5: Tenant Owner Experience
- **Result:** SKIPPED
- **Reason:** Blocked by Scenario 4.

### ❌ Scenario 6: OAuth Client Creation Wizard
- **Result:** BLOCKED (Feature / UI Stub)
- **Observation:** "Platform Admins" page displays "Coming soon". No OAuth Client management interface exists in the Platform scope.
- **Screenshot:** `scenario_6_blocked_missing_feature.png`

## Deviations & Blockers
1.  **Missing Features:** Tenant Management and OAuth Client Management are implemented as UI stubs only.
2.  **Bootstrap Logic:** Required code patch to `internal/identity/bootstrap.go` to support auto-creation of users.

## Screenshots
Screenshots are located in: `/Users/mw/.gemini/antigravity/brain/36cbc520-bcc4-42da-89ff-2b9483d3d3ef/`
- scenario_1_reachability_1766734677321.png
- scenario_3_landing_1766734966687.png
- scenario_4_blocked_missing_feature_1766734977320.png
