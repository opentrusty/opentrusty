# OpenTrusty Threat Model (STRIDE)

This document outlines the threat model for OpenTrusty using the STRIDE methodology.

## 1. Attack Surface Overview

- **Identity & Credentials**: User login, password storage, session cookies.
- **Protocol Flows**: OAuth2 Authorization Code Flow, PKCE, Refresh Tokens, OIDC ID Tokens.
- **Multi-tenancy**: Tenant isolation in the database and execution context.
- **Transport**: HTTP/HTTPS communication, Discovery metadata.

## 2. STRIDE Analysis

| Threat | Attack Vector | Mitigation in OpenTrusty | Residual Risk |
| :--- | :--- | :--- | :--- |
| **Spoofing** | Adversary imitates a valid client to steal codes/tokens. | Strict `redirect_uri` exact match; Client authentication (secret hashing). | Compomised client secret; open redirect in client application. |
| **Tampering** | Modification of Authorize request or tokens in transit. | PKCE (S256) for public clients; `at_hash` in OIDC; signed JWTs (RS256). | Root CA compromise; memory corruption on server. |
| **Repudiation** | User or Client denies performing a security action. | Centralized audit logging of all login/token events (who, when, what). | Log injection if transport is insecure; log deletion by administrator. |
| **Information Disclosure** | Leakage of PII or cross-tenant data. | Argon2id for passwords; Tenant-scoped repositories; HTTP-only/Secure cookies. | Side-channel attacks; OS-level access to the database. |
| **Denial of Service** | Flooding token or login endpoints. | Rate limiting middleware; short-lived authorization codes. | Distributed DoS (DDoS) outside application layer. |
| **Elevation of Privilege** | Tenant A user access Tenant B resources. | Fail-closed Tenant Middleware; strict tenant-aware `sub` derivation. | Internal logic bugs in RBAC enforcement. |

## 3. Core Mitigations

### 3.1 Session Security
OpenTrusty uses **Database-backed Sessions** instead of JWTs for primary browser sessions. This allows for immediate revocation and prevents the common "stateless logout" problem.

### 3.2 Password Hashing
We use **Argon2id** (the winner of the Password Hashing Competition) with security-evaluated parameters to ensure maximum resistance against GPU/ASIC brute-force attacks.

### 3.3 Protocol Compliance
Strict adherence to **RFC 6749**, **RFC 7636**, and **OIDC Core** prevents common protocol-level vulnerabilities like code injection or state mismatch.
