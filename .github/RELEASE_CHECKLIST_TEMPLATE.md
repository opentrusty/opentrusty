# Release Checklist: {VERSION}

**Maturity Level**: {LEVEL}  
**Release Date**: {DATE}  
**Release Manager**: {MANAGER}

---

## Automated Checks (Verified by CI)

- [ ] Unit Tests Pass
- [ ] API Documentation Freshness
- [ ] Integration Tests Pass (Beta+)
- [ ] Docker E2E Tests Pass (Beta+)
- [ ] Systemd Smoke Test Pass (RC+)
- [ ] Security Scan Pass (RC+)

---

## Documentation Requirements

### Alpha
- [ ] API documentation generated and published
- [ ] Known issues documented in release notes
- [ ] Migration guide (if schema changes introduced)

### Beta (includes Alpha requirements)
- [ ] Complete API documentation with examples
- [ ] Migration guide for breaking changes
- [ ] Known limitations and workarounds documented
- [ ] Security assumptions documented

### RC (includes Beta requirements)
- [ ] Security assumptions and threat model reviewed
- [ ] Deployment guide with systemd instructions
- [ ] Rollback procedures documented
- [ ] Upgrade path from previous version tested

### GA (includes RC requirements)
- [ ] Production deployment best practices
- [ ] Monitoring and observability guide
- [ ] Incident response procedures
- [ ] Support and maintenance policy

---

## Manual Verification Requirements

### Alpha
- [ ] Core features functionally complete
- [ ] Release notes created with known issues

### Beta
- [ ] At least 2 alpha releases published previously
- [ ] No known critical bugs in core flows (OAuth2/OIDC)
- [ ] Community testing feedback incorporated
- [ ] Migration path tested from previous beta

### RC
- [ ] At least 1 beta release with >2 weeks of community testing
- [ ] No known security vulnerabilities
- [ ] API surface stable (no further changes planned)
- [ ] Performance benchmarks compared to baseline
- [ ] Security audit completed or waived (2+ maintainer consensus)

### GA
- [ ] At least 1 RC release with >4 weeks of production pilot testing
- [ ] Zero critical or high-severity bugs
- [ ] Full documentation suite complete
- [ ] Support and maintenance plan defined
- [ ] Upgrade path validated from previous GA version
- [ ] Performance regression tests pass

---

## Governance Approvals

- [ ] Release proposal reviewed by maintainers
- [ ] Breaking changes (if any) documented and justified
- [ ] License compliance verified (all new files have headers)
- [ ] Tag naming convention followed: `v{MAJOR}.{MINOR}.{PATCH}[_{MATURITY}{NUMBER}]`

---

## Sign-off

**Maintainer 1**: _________________ Date: _______  
**Maintainer 2** (for RC/GA): _________________ Date: _______

---

## Notes

{Additional notes about this release}

---

**Instructions**:
1. Copy this template to `.github/releases/{VERSION}-checklist.md`
2. Fill in the version, level, date, and manager information
3. Check off items as they are completed
4. Ensure all required items for the maturity level are checked
5. Obtain maintainer sign-offs
6. Commit the checklist before pushing the release tag
