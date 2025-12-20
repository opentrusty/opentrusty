# Compatibility Matrix

This document lists the client libraries, databases, and environments that are officially validated against OpenTrusty.

## Client Libraries
The following libraries are verified to work with OpenTrusty's OIDC implementation.

| Library | Platform | Version Tested | Verification Method |
| :--- | :--- | :--- | :--- |
| `go-oidc` | Go | v3 | Protocol Compliance Tests |
| `httptest` (Stdlib) | Go | 1.21+ | Internal End-to-End Tests |
| `curl` | CLI | 7.x+ | Protocol validation scripts |

> [!NOTE]
> Since OpenTrusty follows standard RFCs (6749, 7636, OIDC Core), most compliant libraries should work. However, only the above are explicitly tested in our CI pipeline.

## Database Support
OpenTrusty uses standard SQL with minimal dialect-specific features.

| Database | Version | Status |
| :--- | :--- | :--- |
| **PostgreSQL** | 14+ | **Primary / Production** |
| MySQL | 8.0+ | Experimental / Unsupported |
| SQLite | 3.x | Development / Testing Only |

## Environment
| component | Requirement |
| :--- | :--- |
| **Go** | 1.21 or higher |
| **Operating System** | Linux (amd64/arm64), macOS, Windows |
| **Container** | Docker / OCI Compliant Runtime |
