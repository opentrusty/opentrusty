# API Documentation Publication Architecture

This document describes the "Hybrid" architecture used to publish versioned, immutable API documentation for OpenTrusty.

## 1. Architectural Strategy: The Hybrid Model

OpenTrusty uses a hybrid approach to documentation publication, combining the modern **GitHub Actions Pages** deployment with a **Git-based history bucket**.

### Why `deploy-pages`?
We use `actions/deploy-pages` (the modern, beta-less way to handle GitHub Pages) because it:
- Is **stateless**: Every deployment is a fresh bundle, preventing artifact accumulation in the primary deployment branch.
- Is **secure**: It uses GitHub's internal artifact storage and OIDC for deployment permissions.
- Provides **traceability**: Each deployment is directly linked to a specific GitHub Action run.

### Why `gh-pages` as a History Bucket?
Since `deploy-pages` is stateless (it replaces the entire site on every run), it cannot natively maintain a history of past versions (e.g., `v1.0.0`, `v1.1.0`). We use the `gh-pages` branch purely as a **durable storage layer**:
- **Durable History**: It acts as a "hard drive" for all previously published documentation versions.
- **Audit Trail**: Every version addition is a commit on this branch, providing a permanent record of what was published and when.
- **Decoupling**: The branch is never served directly; it is treated as a build dependency for the publication workflow.

## 2. Immutability Rules

To maintain the trustworthiness of our API documentation, we enforce the following immutability rules:

1. **Tag-Linked**: Documentation is only published upon the creation of a Git tag (`v*`).
2. **Read-Only Versions**: Once a version (e.g., `v1.2.0/index.html`) is committed to the `gh-pages` history bucket, it MUST NOT be modified.
3. **Redeploy Restriction**: If a tag is deleted and recreated, the CI will overwrite the existing version folder, but such actions are discouraged and audited.
4. **General Unavailability**: The `gh-pages` branch is protected and should not be manually edited.

## 3. Publication Flow

The end-to-end flow for publishing documentation is as follows:

1. **Trigger**: A maintainer pushes a new version tag (e.g., `git push origin v1.0.0`).
2. **Build Phase**:
    - `swag` generates `swagger.json` from the source code.
    - `scripts/check-docs.sh` verifies that the generated spec matches the committed spec (Release Gate).
    - `redoc-cli` bundles `swagger.json` into a standalone `index.html`.
3. **History Integration Phase**:
    - The workflow checks out the current state of the `gh-pages` branch.
    - The new version is copied into `versions/vX.Y.Z/`.
    - `scripts/generate-docs-index.sh` scans the `versions/` folder and rebuilds the root landing page (`index.html`).
4. **Persistence Phase**:
    - The updated history (including the new version and new index) is committed and pushed back to the `gh-pages` branch.
5. **Deployment Phase**:
    - The entire combined history is uploaded as a Pages artifact.
    - `actions/deploy-pages` is called to fulfill the public release.

---
## 4. Audience Disclaimer
This architecture is designed for maintainers and security reviewers to understand the chain of custody for API documentation. It ensures that the documentation served is a faithful, immutable record of the code at the time of release.
