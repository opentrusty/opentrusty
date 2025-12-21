# Deployment Philosophy

This document defines the architectural intent and normative constraints for deploying OpenTrusty. It serves as a lock on design decisions regarding distribution and infrastructure requirements.

## 1. Binary-First Philosophy

OpenTrusty is designed as a **statically-linked, standalone Linux binary**. 

- **Primary Artifact**: The single binary produced by `make build` is the primary, production-ready artifact of the project.
- **Minimal Surface**: We prioritize minimal external dependencies. OpenTrusty does not require a runtime environment (like Node.js, Python, or a JVM) or complex auxiliary services beyond its primary database.
- **Direct Execution**: The binary is intended to run directly on the host OS, making it suitable for bare-metal servers, virtual machines, and edge devices.

## 2. Optional Containerization

While Docker is widely used, it is **not a requirement** for running OpenTrusty.

- **Convenience, Not Coupling**: Docker-related artifacts (Dockerfile, Compose) are provided strictly for convenience in development, testing, and container-native workflows.
- **Secondary Status**: The core architecture must never depend on container-specific features or abstractions. If a feature only works inside Docker, it is considered a defect.
- **Path Isolation**: To emphasize this separation, all Docker-related files live under `/deploy/docker/`, signalling that they are orchestration artifacts rather than core application logic.

## 3. Supported Deployment Modes

OpenTrusty officially supports and encourages the following deployment patterns:

### Bare Metal / VM (+ systemd)
The recommended production mode. The binary is managed as a `systemd` service. This provides the lowest latency, highest transparency, and matches the project's minimalist goal.

### Docker / Docker Compose
Supported for users who prefer container orchestration. We provide a multi-stage, non-root Dockerfile that mirrors the binary-first approach by simply wrapping the binary in a minimal Alpine runtime.

## 4. Non-Requirements

OpenTrusty explicitly does **NOT** require the following:

- **Docker/Podman**: You can run OpenTrusty without ever installing a container engine.
- **Kubernetes**: OpenTrusty is designed to be easy to manage without the overhead of heavy orchestration.
- **Baked-in Configuration**: The system is configured via environment variables, allowing it to fit into any deployment secret management system (vault, cloud-init, etc.) without modification.

## 5. Summary
Architecture is intent. OpenTrusty's intent is to be a fast, light, and independent piece of software that respects the host system and the operator's choice of environment.
