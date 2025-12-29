# Beta Build Artifacts

This document describes the build artifacts produced by OpenTrusty for the Beta release.

---

## 1. Core Backend (`opentrusty`)

### Binary
| Artifact | Path | Description |
|----------|------|-------------|
| `opentrusty` | `bin/opentrusty` | Single static Go binary, no external dependencies. |

### Build Command
```bash
make build
```

### Versioning
Versioning is determined at build time. Currently, Git commit SHA is embedded:
```bash
go build -ldflags "-X main.version=$(git describe --tags --always)"
```

---

## 2. Control Panel UI (`opentrusty-control-panel`)

### Static Assets
| Artifact | Path | Description |
|----------|------|-------------|
| HTML/JS/CSS bundle | `dist/` | Vite-built static assets. |

### Build Command
```bash
make build
```

### Serving
The `dist/` folder contains static files that can be served by any HTTP server (Nginx, Caddy, etc.). The backend does **not** serve the UI; they are separate processes.

### Versioning
Currently determined by `package.json` version field.

---

## 3. Test Artifacts

### UI E2E Test Reports
| Artifact | Path | Description |
|----------|------|-------------|
| HTML Report | `artifacts/tests/ui/report/index.html` | Playwright interactive report. |
| Screenshots | `artifacts/tests/ui/results/` | Failure screenshots (on failure only). |
| Videos | `artifacts/tests/ui/results/` | Test run videos (on failure only). |
| Traces | `artifacts/tests/ui/results/` | Playwright traces for debugging. |

### Backend Test Reports
| Artifact | Path | Description |
|----------|------|-------------|
| Unit Test Report | `artifacts/tests/ut-report.md` | Markdown summary of unit tests. |
| System Test Report | `artifacts/tests/st-report.md` | Markdown summary of system tests. |

---

## 4. Deployment Package (Future)

For GA release, a combined deployment package may include:
- Pre-built Go binary
- Pre-built UI assets
- Example configuration files
- Systemd unit files

This is **not** provided for Beta. Users must build from source.
