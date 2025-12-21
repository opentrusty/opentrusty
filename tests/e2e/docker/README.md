# Docker E2E Test Suite

This directory contains the orchestration files for running OpenTrusty's End-to-End (E2E) tests in a clean, containerized environment.

## Architecture

The test suite uses `docker-compose` to spin up:
- **opentrusty_test**: A fresh build of the OpenTrusty server.
- **postgres_test**: A dedicated PostgreSQL instance for isolation.

The tests are executed by the `run-e2e.sh` script, which manages the lifecycle of these containers.

## Running Locally

To run the complete E2E suite:

```bash
make test-e2e
```

Alternatively, from this directory:

```bash
./run-e2e.sh
```

## Determinism

The suite is designed to be fully deterministic:
- It uses a fresh database on every run (`docker-compose down -v`).
- It waits for the service to be healthy before starting tests.
- It returns a non-zero exit code if any test fails.
