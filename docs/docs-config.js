const docsConfig = {
    sections: [
        {
            title: "Architecture",
            path: "architecture",
            items: [
                { title: "Architecture Rules", file: "docs/architecture/architecture-rules.md" },
                { title: "Domain Model", file: "docs/architecture/domain-model.md" },
                { title: "Runtime Assumptions", file: "docs/architecture/runtime-assumptions.md" },
                { title: "Docs Publication", file: "docs/architecture/docs-publication.md" }
            ]
        },
        {
            title: "Protocols",
            path: "protocols",
            items: [
                { title: "Identity Model", file: "docs/domain/identity-model.md" },
                { title: "Protocol Boundary", file: "docs/architecture/protocol-boundary.md" },
                { title: "Error Model", file: "docs/architecture/protocol-error-model.md" }
            ]
        },
        {
            title: "Security",
            path: "security",
            items: [
                { title: "Threat Model", file: "docs/security/threat-model.md" },
                { title: "Security Assumptions", file: "docs/security/security-assumptions.md" },
                { title: "System Audit Report", file: "docs/system-audit-report.md" }
            ]
        },
        {
            title: "Deployment",
            path: "deployment",
            items: [
                { title: "Deployment Philosophy", file: "docs/fundamentals/deployment-philosophy.md" },
                { title: "Storage Model", file: "docs/fundamentals/storage-model.md" },
                { title: "Operations Guide", file: "docs/deployment/OPERATIONS.md" },
                { title: "Operations Manual", file: "docs/operations/bootstrap.md" },
                { title: "Systemd Guide", file: "deploy/systemd/README.md" }
            ]
        },
        {
            title: "Roadmap",
            path: "roadmap",
            items: [
                { title: "Evolution Roadmap", file: "docs/roadmap/evolution-roadmap.md" },
                { title: "Tech Debt Strategy", file: "docs/roadmap/tech-debt-strategy.md" }
            ]
        },
        {
            title: "API",
            path: "api",
            items: [
                { title: "Interactive Reference", file: "api/index.html" }
            ]
        },
        {
            title: "Testing & Quality",
            path: "testing",
            items: [
                { title: "Unit Test Report", file: "docs/testing/ut-report.md" },
                { title: "System Test Report", file: "docs/testing/st-report.md" },
                { title: "E2E Test Report", file: "docs/testing/e2e-report.md" },
                { title: "Annotation Standard", file: "docs/testing/test-annotation-standard.md" }
            ]
        }
    ]
};

if (typeof module !== 'undefined') {
    module.exports = docsConfig;
}
