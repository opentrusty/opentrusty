# Test Annotation and Documentation Standard

This document defines the standard for documenting Unit Tests (UT) and Service Tests (ST) in OpenTrusty. The goal is to ensure test intent is clear, results can be parsed into structured reports, and reports are suitable for publication.

## Annotation Format

All test functions must be preceded by a comment block using the following fields. Keys are case-sensitive.

```go
// TestPurpose: <Short description of what invariant or guarantee this test protects>
// Scope: <Unit Test | Service Test | Integration Test>
// Security: <Description of security relevance, e.g., "Prevents privilege escalation (CWE-269)"> (Optional if not security-related)
// Permissions: <Comma-separated list of required permissions or roles> (Optional)
// RelatedDocs: <Path to related documentation or RFC> (Optional)
// Expected: <Description of expected outcome, e.g., "Returns HTTP 403 Forbidden">
// Test Case ID: <ID from api_test_cases.md if applicable> (Optional)
func TestComponentName_Scenario_ExpectedBehavior(t *testing.T) {
    ...
}
```

## naming Convention

Test function names should be stable and descriptive, following the pattern:

`Test<Component>_<Scenario>_<ExpectedBehavior>`

*   **Component**: The struct or subsystem being tested (e.g., `TenantService`, `OAuth2Handler`).
*   **Scenario**: The condition or input being tested (e.g., `InvalidRole`, `ExpiredToken`).
*   **ExpectedBehavior**: The outcome (e.g., `ReturnsError`, `RevokesToken`).

### Examples

**Good:**
*   `TestTenantAdmin_AssignPlatformRole_ReturnsForbidden`
*   `TestOAuth2_AuthCodeReplay_RevokesTokens`

**Bad:**
*   `TestCreateUser` (Too generic)
*   `TestError` (Vague)

## Implementation Guide

1.  **Do NOT** change test logic unless required for correctness.
2.  **Do** rename tests to match the convention if they are ambiguous.
3.  **Do** add comments for every test function.

## Parsability

These comments are designed to be parsed by `scripts/generate-test-report.sh` (or future tools) to generate markdown tables for `docs.opentrusty.org`.
