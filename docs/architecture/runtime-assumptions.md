# Runtime & Scaling Assumptions

This document defines the core runtime assumptions and scaling model for OpenTrusty. These assumptions are normative and constrain all future feature designs.

## 1. Stateless Execution Model

The OpenTrusty process is strictly **stateless**.

- **No Local State**: The application must not rely on local file system state, in-memory caches that require synchronization, or sticky sessions for correctness.
- **Restartability**: Any instance of OpenTrusty can be terminated and replaced at any time without data loss or protocol violation, provided an external database is available.
- **Shared-Nothing**: Multiple instances of OpenTrusty do not communicate with each other directly. All coordination happens through the shared persistence layer.

## 2. Horizontal Scaling

OpenTrusty is designed to scale horizontally by adding more identical instances behind a load balancer.

- **Identical Nodes**: Every node in an OpenTrusty cluster is functionally identical and can handle any incoming request.
- **Load Balancing**: Standard Layer 7 (HTTP) load balancing is sufficient. No specialized "session awareness" is required at the load balancer level.
- **Scaling Limit**: The scaling limit of an OpenTrusty deployment is governed solely by the capacity of the external PostgreSQL database.

## 3. Externalization of State

All stateful components of the system are externalized.

- **Primary State (PostgreSQL)**: All identity data, tenant configurations, and session records live in the database.
- **Cryptographic Keys**: Private keys for JWT signing (OIDC) and encryption must be provided as external configuration (environment variables or mounted files) or fetched from a secure vault.
- **Logs & Traces**: Observability data is treated as an outbound stream and should be aggregated by external systems (OTEL, centralized logging).

## 4. Explicit Non-Guarantees

To maintain architectural simplicity and performance, OpenTrusty makes the following explicit non-guarantees:

- **No In-Memory Cache Coherence**: Distributed caches (like Redis) are not currently required or supported for core protocol correctness.
- **No Built-in HA for Database**: High availability of the PostgreSQL layer is an operational requirement, not a feature of the OpenTrusty binary.
- **No Background Job Guarantees**: Any background tasks (if implemented) are best-effort unless backed by the persistent store.

## 5. Design Constraints

Any future feature proposed for OpenTrusty MUST adhere to these assumptions:
1. Can this feature work if there are 100 instances of the app running simultaneously?
2. Does this feature fail if the local instance is killed and replaced?
3. Does this feature require nodes to talk to each other? (If yes, it must be redesigned).
