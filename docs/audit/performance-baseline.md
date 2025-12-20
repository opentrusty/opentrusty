# OpenTrusty Performance Baseline

**Date**: 2025-12-20
**Version**: v0.1.0-alpha
**Environment**: Apple M5 (Dev Baseline)

This document establishes the minimal performance baseline for critical OpenTrusty operations.
These metrics are CPU-bound baselines measured via `go test -bench` without network or database I/O latency.

## 1. Security Cryptography (Argon2id)
Cost of hashing passwords used for user authentication.

| Operation | Time/Op | Memory/Op | Allocations/Op |
| :--- | :--- | :--- | :--- |
| **Hash Password** | ~11.4 ms | 64 MB | 65 |
| **Verify Password** | ~12.1 ms | 64 MB | 70 |

> **Configuration**: RFC 9106 Recommended (Memory: 64MB, Iterations: 1, Parallelism: 4).
> **Note**: This is the dominant cost for the `POST /auth/login` endpoint.

## 2. Token Issuance (OAuth2/OIDC)
Cost of generating tokens and signing JWTs (excluding persistence).

| Operation | Time/Op | Memory/Op | Allocations/Op | Notes |
| :--- | :--- | :--- | :--- | :--- |
| **OIDC ID Token Signing (RS256)** | ~666 µs | 5.5 KB | 56 | 2048-bit RSA Key |
| **OAuth2 Token Exchange Logic** | ~72 µs | 2.1 KB | 30 | Validation & Generation logic only |

## 3. Database Latency
*Not measured in this baseline (Integration environment required).*

**Estimated Targets (PostgreSQL)**:
- **Read User**: < 1 ms (Indexed)
- **Persist Token**: < 5 ms (Write + WAL)

## 4. End-to-End Latency Model
Estimated latency for `POST /oauth2/token` (Authorization Code Grant):

$$ T_{total} = T_{logic} + T_{sign} + T_{db\_read(code)} + T_{db\_write(tokens)} $$

$$ T_{total} \approx 72\mu s + 666\mu s + 1ms + 5ms \approx \mathbf{6.7ms} $$

Estimated latency for `POST /auth/login` (Password Auth):

$$ T_{total} = T_{hash} + T_{db\_read} + T_{db\_write(session)} $$

$$ T_{total} \approx 12ms + 1ms + 5ms \approx \mathbf{18ms} $$

## Conclusion
OpenTrusty's performance profile is dominated by **Argon2id password hashing**, which is intentional for security. Token issuance overhead is negligible (< 1ms). Production throughput will be I/O bound by the database layer.
