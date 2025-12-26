# Control Plane UI Capabilities

**Status:** Normative  
**Version:** 1.0  
**Last Updated:** 2025-12-26  
**Technology Stack:** React + TypeScript + Tailwind CSS

## Purpose

This document defines the **minimal but complete** feature set required for the OpenTrusty control plane UI to support real-world tenant onboarding and operation. This is the authoritative specification for UI development.

## Critical Context

**OpenTrusty UI is a control plane for operators, not an end-user portal.**

Only these roles access the UI:
- **Platform Admins**: Manage tenants and platform-level configuration
- **Tenant Owners/Admins**: Manage their tenant's users, OAuth clients, and settings
- **Tenant Members**: ❌ Never access this UI (they authenticate via OAuth2 to external applications)

## Post-Login Routing

After successful `/auth/login`, the UI **MUST** route based on user role:

```typescript
if (user.roles.includes('platform_admin')) {
  navigate('/platform/tenants');
} else if (user.roles.includes('tenant_owner') || user.roles.includes('tenant_admin')) {
  navigate(`/tenant/${user.tenant_id}/overview`);
} else {
  // Edge case: user has no admin role (should not happen if backend validates)
  // Show error: "Access denied: admin role required"
}
```

**Rationale:** Platform admins have no single tenant context; tenant admins operate within their assigned tenant.

---

## Platform Admin UI

### Left Navigation

```
OpenTrusty Platform
├── Overview
├── Tenants
│   ├── List Tenants
│   ├── Create Tenant
│   └── (Selected Tenant) → Assign Owner
├── Platform Admins
│   ├── List Admins
│   └── Invite Admin
├── Audit Logs
└── System Settings (read-only in MVP)
```

### Feature Specifications

#### 1. Overview (`/platform/overview`)
- **Total Tenants**: Count of active tenants
- **Total Users**: Aggregate across all tenants
- **Total OAuth Clients**: Aggregate count
- **Recent Activity**: Last 10 audit events (platform-scoped)

#### 2. Tenants (`/platform/tenants`)

**List View:**
| Tenant Name | Tenant ID | Owner Email | Created | Status | Actions |
|-------------|-----------|-------------|---------|--------|---------|
| Acme Corp   | uuid-123  | admin@acme  | 2025-01 | Active | View    |

**Create Tenant:**
- Form fields:
  - Tenant Name (required)
  - Owner Email (required, must not exist)
  - Owner Password (required for new user)
- On submit:
  - Create tenant
  - Create owner user with `tenant_owner` role
  - Assign owner to tenant
  - Send email (if email service configured)

**Assign Owner:**
- Select existing user OR create new user
- Assign `tenant_owner` role
- Only one owner per tenant (enforce or allow multiple?)

#### 3. Platform Admins (`/platform/admins`)

**List View:**
- Email
- Created Date
- Status
- Actions: Remove (with confirmation)

**Invite Admin:**
- Email (required)
- Password (required)
- On submit: Create user with `platform_admin` role

#### 4. Audit Logs (`/platform/audit`)
- Platform-scoped events only
- Filters: Event Type, Date Range, Actor
- Columns: Timestamp, Event Type, Actor, Tenant, Resource, Metadata

#### 5. System Settings (`/platform/settings`)
- **MVP**: Read-only view of:
  - Issuer URL
  - Discovery URL
  - JWKS URL
  - Session Lifetime (non-editable)
- **Future**: Editable configuration

---

## Tenant Owner/Admin UI

### Left Navigation

```
{Tenant Name}
├── Overview
├── Users
│   ├── List Users
│   ├── Invite Admin
│   └── Assign Roles
├── OAuth Clients
│   ├── Client List
│   ├── Create Client (Wizard)
│   └── (Selected Client) → Manage Secrets
├── Sessions (Phase 2)
├── Audit Logs
└── Branding (placeholder / disabled)
```

### Feature Specifications

#### 1. Overview (`/tenant/{tenant_id}/overview`)

**First-Time UX (No OAuth Clients):**
```
Welcome to {Tenant Name}!

Your Tenant Information:
- Tenant ID: {tenant_id}
- Issuer: https://auth.example.com
- Discovery: https://auth.example.com/.well-known/openid-configuration

[Create Your First OAuth Client] (CTA Button)
```

**Standard View:**
- OAuth Clients: Count + Quick List
- Users: Count + Recent Additions
- Recent Activity: Last 10 tenant-scoped events

#### 2. Users (`/tenant/{tenant_id}/users`)

**List View:**
| Email | Roles | Created | Status | Actions |
|-------|-------|---------|--------|---------|
| admin@acme | tenant_admin | 2025-01 | Active | Edit Roles |

**Invite Admin:**
- Email (required)
- Role: `tenant_admin` or `tenant_owner`
- Password (required for new user)
- On submit: Create user + assign role within tenant scope

**Assign Roles:**
- Select user
- Add/Remove roles: `tenant_admin`, `tenant_owner`
- Save

#### 3. OAuth Clients (`/tenant/{tenant_id}/clients`)

**List View:**
| Client Name | Client ID | Type | Created | Status | Actions |
|-------------|-----------|------|---------|--------|---------|
| My Web App  | abc123    | Web  | 2025-01 | Active | View, Regenerate Secret |

**Create Client (Wizard)** - See detailed specification below

#### 4. Audit Logs (`/tenant/{tenant_id}/audit`)
- Tenant-scoped events only
- Filters: Event Type, Date Range, Actor
- Columns: Timestamp, Event Type, Actor, Resource, Metadata

#### 5. Branding (`/tenant/{tenant_id}/branding`)
- **MVP**: Placeholder with message:
  > "Custom branding will be available in a future release."
- **Future**: Logo upload, color scheme, email templates

---

## OAuth Client Creation Wizard (Critical)

This is the **most important** UX flow for tenant admins.

### Step 1: Client Name

```
┌─────────────────────────────────────────┐
│ Create OAuth Client                    │
│                                         │
│ Client Name                             │
│ ┌─────────────────────────────────────┐ │
│ │ My Application                      │ │
│ └─────────────────────────────────────┘ │
│                                         │
│ [Cancel]              [Next: Client Type →]│
└─────────────────────────────────────────┘
```

### Step 2: Client Type

```
Select Application Type:

○ Web Application
  - Server-side app with backend
  - Can securely store client_secret
  - Requires Authorization Code flow

○ Single Page Application (SPA)
  - Frontend-only (React, Vue, Angular)
  - Cannot store secrets securely
  - Requires PKCE

○ Backend Service (Machine-to-Machine)
  - No user interaction
  - Client Credentials grant
  - API-to-API communication

[← Back]              [Next: Redirect URIs →]
```

### Step 3: Redirect URIs

```
Redirect URIs
(Where should users be sent after login?)

┌─────────────────────────────────────┐
│ https://app.example.com/callback    │
└─────────────────────────────────────┘
+ Add Another URI

⚠ For local development: http://localhost:3000/callback

[← Back]              [Next: Scopes →]
```

### Step 4: Scopes Selection

```
Select Scopes:

☑ openid (required for OIDC)
☑ profile (name, given_name, family_name)
☑ email (email, email_verified)
☐ offline_access (refresh tokens)

[← Back]              [Next: Confirmation →]
```

### Step 5: Confirmation

```
Review Client Configuration:

Client Name: My Application
Client Type: Web Application
Redirect URIs:
  - https://app.example.com/callback
Scopes: openid, profile, email

[← Back]              [Create Client]
```

### Step 6: Result Screen (CRITICAL)

```
✅ Client Created Successfully!

⚠ Save these credentials now - client_secret will not be shown again.

Client ID:
┌─────────────────────────────────────┐
│ abc123def456                        │ [Copy]
└─────────────────────────────────────┘

Client Secret:
┌─────────────────────────────────────┐
│ secret_xyz789                       │ [Copy]
└─────────────────────────────────────┘

Issuer:
https://auth.example.com

Discovery Endpoint:
https://auth.example.com/.well-known/openid-configuration

Example Authorization URL:
https://auth.example.com/oauth2/authorize?
  response_type=code&
  client_id=abc123def456&
  redirect_uri=https://app.example.com/callback&
  scope=openid%20profile%20email&
  state=random_state_value

[Download as JSON]    [Go to Client List]
```

**Download JSON Format:**
```json
{
  "client_id": "abc123def456",
  "client_secret": "secret_xyz789",
  "client_name": "My Application",
  "client_type": "web_application",
  "redirect_uris": ["https://app.example.com/callback"],
  "scopes": ["openid", "profile", "email"],
  "issuer": "https://auth.example.com",
  "discovery_url": "https://auth.example.com/.well-known/openid-configuration",
  "token_endpoint": "https://auth.example.com/oauth2/token",
  "authorization_endpoint": "https://auth.example.com/oauth2/authorize"
}
```

---

## Tenant Bootstrap UX

**Scenario:** First-time user creates account.

### Flow

1. **Registration** (if enabled):
   - User submits `/auth/register` with email + password + name
   - Backend creates:
     - User record
     - New tenant (auto-named: "{FirstName}'s Tenant" or email domain)
     - Assigns `tenant_owner` role

2. **Auto-Login**:
   - Redirect to `/auth/login` with success message: "Account created! Please log in."

3. **Post-Login**:
   - Route to `/tenant/{tenant_id}/overview`
   - Show first-time UX (see Overview section)

4. **First Action**:
   - User clicks "Create Your First OAuth Client"
   - Enter OAuth Client Wizard

---

## Explicit Non-Goals (MVP)

The following features are **explicitly excluded** from the MVP:

### 1. End-User Flows
- ❌ No tenant member login
- ❌ No password reset UI for members
- ❌ No user profile management for members
- **Rationale:** Members authenticate via OAuth2; password resets handled by integrating applications

### 2. OAuth Consent Screens
- ❌ No custom consent UI
- ❌ No consent skip/auto-approve settings
- **Rationale:** Phase 2 feature; default consent screen OK for MVP

### 3. Advanced RBAC
- ❌ No custom role creation
- ❌ No fine-grained permission assignment
- **Rationale:** Predefined roles (`platform_admin`, `tenant_owner`, `tenant_admin`) sufficient for MVP

### 4. Multi-Factor Authentication (MFA)
- ❌ No MFA enrollment
- ❌ No TOTP/SMS support
- **Rationale:** Phase 2 security enhancement

### 5. Branding Customization
- ❌ No logo upload
- ❌ No color scheme editor
- ❌ No custom email templates
- **Rationale:** Placeholder exists; implementation deferred

### 6. Session Management UI
- ❌ No active session list
- ❌ No remote session revocation
- **Rationale:** Audit logs provide visibility; management deferred to Phase 2

---

## Design Principles

### 1. Progressive Disclosure
- Show essential information first
- Hide advanced options behind "Advanced Settings" toggles
- Example: OAuth Client wizard starts simple, allows advanced redirect URI patterns later

### 2. Copy-Paste Friendly
- All UUIDs, URLs, secrets: one-click copy buttons
- Example integrations show copyable code snippets

### 3. Destructive Actions Require Confirmation
- Deleting tenant → Modal: "Type tenant name to confirm"
- Regenerating client_secret → Warning: "Old secret will be invalidated"

### 4. Contextual Help
- ℹ️ Info icons next to complex fields
- Inline examples (e.g., redirect URI patterns)
- Link to docs for detailed guides

### 5. Responsive Design
- Mobile-friendly navigation (collapsible sidebar)
- Tables paginated with mobile stack view

---

## Implementation Notes

### Technology Stack
- **Frontend**: React 18+ with TypeScript
- **Styling**: Tailwind CSS (existing)
- **State Management**: React Context or Zustand (lightweight)
- **Forms**: React Hook Form + Zod validation
- **HTTP Client**: OpenAPI-generated client (existing)

### API Integration
- All UI actions map to existing backend API endpoints
- Use OpenAPI types for request/response contracts
- Handle 400/401/403 gracefully with user-friendly messages

### Navigation Structure
```
/login                          → LoginPage
/platform/overview              → PlatformOverview
/platform/tenants               → TenantList
/platform/tenants/create        → CreateTenant
/platform/admins                → PlatformAdminList
/platform/audit                 → PlatformAudit
/tenant/:tenantId/overview      → TenantOverview
/tenant/:tenantId/users         → UserList
/tenant/:tenantId/clients       → ClientList
/tenant/:tenantId/clients/new   → ClientWizard
/tenant/:tenantId/audit         → TenantAudit
```

---

## Acceptance Criteria

The MVP is **complete** when:

✅ Platform admin can create a tenant and assign an owner  
✅ Tenant owner can log in and see their tenant overview  
✅ Tenant owner can create an OAuth client via wizard  
✅ OAuth client result screen shows all required credentials  
✅ Tenant owner can invite a tenant admin  
✅ Audit logs show all administrative actions  
✅ All navigation paths work without 404s  
✅ Mobile layout is functional (collapsible navigation)  

---

## References

- **Authentication Model**: `docs/architecture/tenant-context-resolution.md`
- **API Capabilities**: `docs/fundamentals/capabilities.md`
- **RBAC Roles**: `internal/rbac/constants.go`
- **OpenAPI Spec**: `docs/api/openapi.yaml` (when generated)
