# Security Assumptions & Trusted Computing Base (TCB)

OpenTrusty is designed to be a highly secure Identity Provider, but its security depends on several foundational assumptions about its operating environment.

## 1. Environment Assumptions

### 1.1 Secure Network Transport
OpenTrusty assumes it is deployed behind a trusted **Reverse Proxy** (e.g., Nginx, Envoy, Cloudflare) that terminates TLS.
- **Assumption**: All traffic between the end-user and the reverse proxy is encrypted using modern TLS (1.2+).
- **Responsibility**: The operator must ensure that the proxy correctly handles certificate rotation and secure cipher suites.

### 1.2 Underlying Host Integrity
OpenTrusty assumes that the host OS and container runtime are secure.
- **Assumption**: An attacker does not have root-level access to the server.
- **Responsibility**: The operator must apply OS security patches and use minimal container images (e.g., distroless).

### 1.3 Database Security
The database (PostgreSQL) is the source of truth for all sessions, users, and tokens.
- **Assumption**: The database is only accessible by the OpenTrusty process via a secure network link.
- **Responsibility**: Use IAM roles or encrypted connections for DB access. Never expose the DB port to the public internet.

## 2. What OpenTrusty Protects Against

- **Online Brute Force**: Via rate limiting and Argon2id computational cost.
- **Token Replay (OIDC Code Flow)**: Via one-time use codes and mandatory `nonce`.
- **Cross-Tenant Leakage**: Via strict logical isolation at the repository layer.
- **Stateless Session Hijacking**: Via database-backed session invalidation.

## 3. What OpenTrusty Does NOT Protect Against

- **Compromised End-User Device**: If the user's browser is compromised by malware, cookies and sessions can be stolen.
- **Social Engineering**: OpenTrusty provides facts (Identity); it cannot prevent a user from sharing their password.
- **Compromised Developer Machine**: If a developer machine with Git push access is compromised, the integrity of the binary cannot be guaranteed.
- **Denial of Service (L3/L4)**: Application-level rate limiting cannot stop massive volumetric DDoS attacks.
