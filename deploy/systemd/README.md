# Systemd Deployment Guide

This guide provides instructions for deploying OpenTrusty on Linux virtual machines or bare-metal servers using `systemd`. This is the recommended deployment mode for production environments following our binary-first philosophy.

## 1. Binary Installation

1.  **Build or Download**: Obtain the statically-linked Linux binary. If building from source:
    ```bash
    make build
    ```
2.  **Copy to Host**: Place the binary in a standard execution path:
    ```bash
    # On the target host
    sudo cp opentrusty /usr/local/bin/
    sudo chmod +x /usr/local/bin/opentrusty
    ```

## 2. Environment Configuration

OpenTrusty is configured via environment variables. For `systemd`, we use an environment file.

1.  **Create Directory**:
    ```bash
    sudo mkdir -p /etc/opentrusty
    ```
2.  **Initialize Config**: Use the provided template:
    ```bash
    sudo cp opentrusty.env.example /etc/opentrusty/opentrusty.env
    sudo chmod 600 /etc/opentrusty/opentrusty.env
    ```
3.  **Adjust Variables**: Edit `/etc/opentrusty/opentrusty.env` with your specific settings:
    - `DB_URL`: The connection string for your external PostgreSQL instance.
    - `ISSUER`: The public base URL of your OpenTrusty instance.
    - `PORT`: The port to listen on (default 8080).

## 3. Database Setup

Ensure your external PostgreSQL instance is reachable and a database has been created for OpenTrusty.

```bash
# Example psql command to create the DB if running locally
sudo -u postgres psql -c "CREATE DATABASE opentrusty;"
```

## 4. Service Management

1.  **Install Unit File**:
    ```bash
    sudo cp opentrusty.service /etc/systemd/system/
    sudo systemctl daemon-reload
    ```
2.  **Create Service User**:
    ```bash
    sudo useradd -r -s /bin/false opentrusty
    ```
3.  **Start & Enable**:
    ```bash
    sudo systemctl enable opentrusty
    sudo systemctl start opentrusty
    ```
4.  **Monitoring**:
    - **Status**: `systemctl status opentrusty`
    - **Logs**: `journalctl -u opentrusty -f`

## 5. Security Considerations

- **Non-Root Execution**: The service is configured to run as the `opentrusty` user. Do not change this to root.
- **File Permissions**: Ensure `/etc/opentrusty/opentrusty.env` is only readable by root or the `opentrusty` user (permissions `600` or `640`).
- **Reverse Proxy**: It is strongly recommended to run OpenTrusty behind a reverse proxy (like Nginx or HAProxy) for TLS termination and header hardening.
- **Database Access**: Use a dedicated PostgreSQL user for OpenTrusty with access limited to its own database.
