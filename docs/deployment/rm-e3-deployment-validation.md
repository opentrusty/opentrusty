# Deployment Validation Report (RM-E3)

**Date**: 2025-12-25
**Scope**: Systemd Single-Node Deployment
**Status**: VALIDATED

## 1. Artifact Verification

| Artifact | Verification | Status |
| :--- | :--- | :--- |
| `deploy/systemd/opentrusty.service` | Checked logic, user isolation, and restart policies. | ✅ PASS |
| `deploy/systemd/opentrusty.env.example` | Verified configuration keys match `internal/config`. | ✅ PASS |
| `deploy/systemd/README.md` | Verified instructions match `opentrusty.service` logic. | ✅ PASS |

## 2. Security Hardening Check

The systemd unit file employs the following hardening directives:

-   `User=opentrusty`: **Present**. Prevents root execution.
-   `NoNewPrivileges=true`: **Present**. Prevents privilege escalation.
-   `ProtectSystem=full`: **Present**. Protects `/usr` and `/boot`.
-   `ProtectHome=true`: **Present**. Hides `/home`, `/root`.
-   `PrivateTmp=true`: **Present**. Isolates temporary files.

## 3. Installation Flow Validation

1.  **Binary Placement**: `/usr/local/bin/opentrusty` (Standard convention).
2.  **Config Placement**: `/etc/opentrusty/opentrusty.env` with `600` permissions.
3.  **User Creation**: `useradd -r -s /bin/false opentrusty` (System user, no shell).

## 4. Recovery & Resilience

-   **Restart Policy**: `Restart=on-failure`. Service will auto-restart if it crashes.
-   **Restart Delay**: `RestartSec=5s`. Prevents rapid restart loops.

## 5. Limitations

-   **Single Node**: This deployment model is not HA.
-   **Database Dependency**: Requires external PostgreSQL; no automatic dependency handling in unit file (relies on `After=network.target`).

## 6. Conclusion

The systemd deployment configuration is mature, secure by default, and ready for Beta distribution.
