# Update Matrix

This table dictates which documentation files **MUST** be updated when specific code changes occur.

| Change Type | Files that MUST be Updated |
| :--- | :--- |
| **Database Schema** | `docs/_ai/invariants.md`, `docs/_ai/authority-model.md` |
| **New API Endpoint** | `docs/api/swagger.json`, `docs/_ai/protocol-scope.md` |
| **New Domain/Package** | `docs/_ai/architecture-map.md` |
| **Auth/Role Changes** | `docs/_ai/authority-model.md`, `docs/_ai/invariants.md` |
| **OAuth2/OIDC Flow** | `docs/_ai/protocol-scope.md` |
| **Tenant Logic** | `docs/_ai/invariants.md` |

## Enforcement

-   When planning a change, consulting this matrix is **MANDATORY**.
-   Updates to documentation must happen in the **SAME** Pull Request/Task as the code change.
