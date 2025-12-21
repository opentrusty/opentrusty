# Systemd Smoke Test

This directory contains a smoke test suite for verifying OpenTrusty's compatibility with `systemd`.

## Overview

The test uses a specialized Ubuntu-based container (`jrei/systemd-ubuntu`) to provide a functional systemd environment within Docker. This allows us to verify:
- Binary execution under standard systemd paths (`/usr/local/bin`).
- Unit file correctness and service descriptors.
- Hardening features (NoNewPrivileges, PrivateTmp, etc.).
- Logging to `journald`.

## Running the Test

To run the systemd smoke test:

```bash
make test-systemd
```

Or from this directory:

```bash
sudo ./smoke_test.sh
```

> [!NOTE]
> Running systemd inside Docker requires `--privileged` mode and volume mapping of `/sys/fs/cgroup`.

## Verification Flow

1. Builds a static Linux binary.
2. Starts the systemd container.
3. Installs the binary, service unit, and environment file.
4. Starts the service via `systemctl`.
5. Verifies the process is active and responding (or at least reached the initialization stage).
