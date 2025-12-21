# Project Governance

This document describes how the OpenTrusty project is governed.

## 1. Project Philosophy
OpenTrusty is a community-driven, non-profit Identity Provider project. Our goal is to provide a production-grade, self-hostable IAM solution that is accessible to everyone.

## 2. Maintainer Model
OpenTrusty follows a **Maintainer-led** governance model.

- **Maintainers**: Responsible for the technical direction, code quality, and security of the project. Maintainers have write access to the main repository.
- **Contributors**: Anyone who submits a pull request, opens an issue, or improves documentation.

### Becoming a Maintainer
Potential maintainers are nominated by existing maintainers based on a history of high-quality contributions and demonstrated commitment to the project's values.

## 3. Decision Making
We aim for **lazy consensus** for technical decisions.
- Minor changes require 1 LGTM (Looks Good To Me) from a maintainer.
- Major architectural changes or security-sensitive modifications require a formal proposal and consensus from at least 2 maintainers.

## 4. Code of Conduct
All participants are expected to follow our [Code of Conduct](CODE_OF_CONDUCT.md). We are committed to providing a welcoming and inclusive environment for everyone.

## 5. Non-Profit Commitment
OpenTrusty is under the MIT License. It will always remain free and open-source.

## 6. Release Gate: API Documentation Trustworthiness
- A release of OpenTrusty MUST NOT be published if its public API documentation is incomplete, semantically inaccurate, or out of sync with the released code.
- API documentation is considered a security-critical artifact and is subject to the same rigor as production code.

### Policy
1.  **Code-Driven**: All API specifications must be generated directly from source code annotations. Manual edits to `swagger.json` are prohibited.
2.  **Freshness Guarantee**: The generated specification in the repository must byte-for-byte match the specification generated from the current HEAD commit.
3.  **No Exceptions**: The release pipeline must enforcing this check. If the check fails, the release is blocked.

## 7. API Documentation Publication Model
* GitHub Pages deployment uses the **modern GitHub Actions artifact-based pipeline**.
* The `gh-pages` branch is used **only as a persistent storage bucket** for versioned API documentation history.
* No branch is configured as a Pages source.
* Each release tag appends a new immutable version under `/versions/{tag}/`.
* The root `index.html` is regenerated on each release to reflect the full history.
* Deployment is stateless; persistence is explicit.