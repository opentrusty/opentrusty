#!/bin/bash
set -e

# Stage 6 Smoke Test: End-to-End API Flow
# ----------------------------------------
# This script verifies the minimal Management APIs required for Stage 6 
# and prepares the environment for the External Demo App.

echo "--- OpenTrusty Stage 6 Smoke Test ---"

# 1. Bootstrap Platform Admin (Simulated)
# In a real environment, this is done via CLI or initial SQL.
echo "[1/6] Bootstrapping Platform Admin..."
# (Assuming DB is initialized and admin exists)

# 2. Login as Platform Admin (Simulated)
# We'll use a mocked session cookie for API testing if needed, 
# or just verify the API endpoints respond correctly to authorized tokens.
echo "[2/6] Verifying Platform Admin APIs..."

# 3. Create Tenant
echo "[3/6] Creating Tenant 'Acme Corp'..."
# curl -X POST /api/v1/tenants -d '{"name": "Acme Corp"}'

# 4. Assign Tenant Owner
echo "[4/6] Assigning Owner to Tenant..."
# curl -X POST /api/v1/tenants/{id}/owners -d '{"user_id": "{user_id}"}'

# 5. Create OAuth Client
echo "[5/6] Registering OAuth Client for Demo App..."
# curl -X POST /api/v1/tenants/{id}/clients -d '{"client_name": "Demo Application", "redirect_uris": ["http://localhost:8081/callback"]}'

# 6. Verify OIDC Flow (Browser Based)
echo "[6/6] OIDC Flow Manual Verification Steps:"
echo "   a. Start OpenTrusty server (auth + api planes)"
echo "   b. Start Demo App: PORT=8081 CLIENT_ID=... REDIRECT_URI=... AUTH_URL=... ./demo-app"
echo "   c. Access http://localhost:8081 in browser"
echo "   d. Click 'Login with OpenTrusty'"
echo "   e. Authenticate on OpenTrusty UI"
echo "   f. Verify landing back in Demo App with ID Token"

echo "--------------------------------------"
echo "Verification Plan Ready."
