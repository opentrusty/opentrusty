# Playwright Debugging Guide

This guide explains how to debug the E2E UI test suite using the generated artifacts (traces, screenshots, videos).

## 1. Artifact Policies

The test suite is configured to generate debugging assets **only on failure** to reduce noise.

| Artifact | Policy | Location |
| :--- | :--- | :--- |
| **Screenshot** | `only-on-failure` | `test-results/<test-name>/test-failed-1.png` |
| **Video** | `retain-on-failure` | `test-results/<test-name>/video.webm` |
| **Trace** | `retain-on-failure` | `test-results/<test-name>/trace.zip` |

## 2. Viewing Traces

Traces are the most powerful debugging tool. They capture DOM snapshots, console logs, and network activity for every step.

### If a test fails locally:
Playwright automatically outputs a command to view the report (which embeds traces).

```bash
npx playwright show-report
```

### To view a specific trace file:
```bash
npx playwright show-trace test-results/path/to/trace.zip
```

## 3. Interactive Debugging (Inspector)

To run tests step-by-step with the Playwright Inspector:

```bash
# Run all tests in debug mode
npx playwright test --debug

# Run a specific test file
npx playwright test tests/e2e/04-oidc-flow.spec.ts --debug
```

## 4. Headed Mode (Visual Verification)

To watch the browser execution without pausing (unless a `page.pause()` is hit):

```bash
npx playwright test --headed
```

## 5. UI Mode (Best for Local Dev)

Playwright UI Mode provides a time-travel debugger and watch mode.

```bash
npx playwright test --ui
```
