# Threat Model Lite (OpenTrusty)

## 1. Identity & Access Management
| Threat | Mitigation | Status |
|--------|------------|--------|

## 2. API & Integration
| Threat | Mitigation | Status |
|--------|------------|--------|
| CSRF (SPA) | Custom `X-CSRF-Token` header requirement | Mitigated |
| Tenant Spoofing | Rejecting `X-Tenant-ID` on authenticated routes | Mitigated |
| CORS evasion | Strict origin matching for Admin Plane | Mitigated |

## 3. OIDC Protocol
| Threat | Mitigation | Status |
|--------|------------|--------|
| Code Reuse | One-time use enforcement | Mitigated |
| Redirect URI redirection | Exact string matching | Mitigated |
| Authorization Code Interception | PKCE (S256) | Mitigated |

## 4. Secrets & Data
| Threat | Mitigation | Status |
|--------|------------|--------|
| Database leak (Secrets) | SHA-256 hashing of tokens and client secrets | Mitigated |
| Database leak (Passwords)| Argon2id hashing | Mitigated |
| Debug log leak | Masking of sensitive fields in `slog` | Mitigated |
