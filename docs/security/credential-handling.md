# Credential Handling Policy

## 1. Password Hashing
- **Algorithm**: Argon2id
- **Parameters**:
  - Memory: 64MB (65536 KB)
  - Iterations: 3
  - Parallelism: 4
  - Salt Length: 16 bytes
  - Key Length: 32 bytes
- **Enforcement**: Mandatory for all user passwords.

## 2. OAuth2 Secrets & Tokens
- **Client Secrets**: Hashed using SHA-256 (`ClientSecretHash`). Never stored in plain text.
- **Access Tokens**: Hashed using SHA-256 (`TokenHash`) in the database.
- **Refresh Tokens**: Hashed using SHA-256 (`TokenHash`) in the database.
- **Exposure**:
  - `client_secret` is only returned ONCE during creation or regeneration.
  - `access_token` and `refresh_token` are returned only to the authenticated client.

## 3. Audit & Logs
- **No PII/Secrets in Logs**:
  - `client_secret` MUST NOT be logged.
  - `password` MUST NOT be logged.
  - `session_id` is masked in logs (first 8 chars only).
- **Audit Coverage**: All state-changing operations (Create Client, Provision User, Update Profile) are recorded in the Audit Sink.
