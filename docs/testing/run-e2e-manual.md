# Runbook: Manual E2E UI Review

**Target Audience**: QA Engineers, Product Owners
**Goal**: Visually verify user flows using the automated driver (--headed mode).

## 1. Setup

Start the environment (same as automated):

```bash
cd /Users/mw/workspace/repo/github.com/opentrusty/opentrusty
./tests/local/run-all.sh --no-tests
```

## 2. Run Visual Verification

Execute Playwright in **Headed Mode**. This will launch a measurable Chrome window and perform the actions visible to the user.

```bash
cd /Users/mw/workspace/repo/github.com/opentrusty/opentrusty-control-panel

# Run all tests visually
npx playwright test --headed
```

### Slow Motion (Optional)
To slow down execution for better observation, create a temporary config or use the inspector:

```bash
npx playwright test --debug
```
*Click "Resume" in the Inspector overlay to step through.*

## 3. Verification Checklist

During the headed run, observe:

- [ ] **Login**: Does the login page render correctly (CSS, Layout)?
- [ ] **Tenants**: do new tenants appear instantly in the list (Optimistic UI)?
- [ ] **OIDC**: Does the Demo App look correct? Is the redirect smooth?
- [ ] **Errors**: Do any "Something went wrong" toasts appear unexpectedly?

## 4. Completion

If the script completes successfully and visual observations were clean, the Release Candidate is **Approved**.
