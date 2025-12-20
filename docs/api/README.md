# OpenTrusty API Documentation

This directory contains the auto-generated OpenAPI 3.1 specifications for the OpenTrusty API.

## Structure
- `swagger.json`: The latest generated specification (Swagger 2.0 / OpenAPI 3.0 compatible).
- `swagger.yaml`: YAML version.

## Usage
To regenerate the documentation:
```bash
make docs-gen
```

## Versioning
Documentation is versioned by git tags.
The CI pipeline publishes `openapi-vX.Y.Z.json` to the documentation site.

## Development
- Annotations are located in `internal/transport/http/*.go`.
- We use [swag](https://github.com/swaggo/swag) for generation.
