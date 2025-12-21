# AI Usage Boundaries

This document defines the normative boundaries for the use of Artificial Intelligence (AI) within OpenTrusty. As a security-critical piece of infrastructure, OpenTrusty maintains a strict "Logic-First" philosophy to ensure determinism, absolute auditability, and the integrity of the trust chain.

## 1. Absolute Prohibitions: The "Red Line"

Artificial Intelligence MUST NOT participate in any deterministic security decision or artifact generation.

- **❌ No Auth Decision-Making**: AI shall not decide whether to `Permit` or `Deny` a request. All authorization logic must be encoded in deterministic, auditable code.
- **❌ No Token Generation**: AI shall not generate, sign, or manage the lifecycle of Access Tokens, ID Tokens, or Refresh Tokens.
- **❌ No Request Blocking**: AI shall not act as an automated blocking agent in the core request path. Probabilistic "blocking" is a fundamental defect in auth infrastructure.

**Rationale**: Security infrastructure requires absolute determinism and an explainable causal chain. AI models are non-deterministic black boxes prone to hallucinations and probabilistic failures, making them unsuitable for the "Red Line" of identity management.

## 2. Allowed & Encouraged Use Cases (Advisory)

AI is positioned as an **Observability and Configuration Assistant**—providing intelligent advice and automating manual toil without holding the "Keys to the Kingdom."

### 2.1 Auth Debug Copilot
AI may be used to analyze complex protocol flows and state transitions for debugging and audit purposes:
- **Login Pattern Analysis**: Identifying unusual login modes or source reputation.
- **Token Misuse Detection**: Highlighting anomalies in token usage patterns (e.g., unexpected IP/User-Agent shifts).
- **Cross-Tenant Behavior**: Detecting patterns that might suggest cross-tenant probing or lateral movement.
- **👉 Rule**: AI provides **alerts and suggestions** to human operators; it does NOT block the traffic automatically.

### 2.2 Compliance & Audit Automation
AI acts as a **Security Reviewer** to verify system integrity:
- **Protocol Verification**: "Does this implementation strictly follow OIDC Core §3.1?"
- **Isolation Auditing**: "Does this proposed code change or configuration risk breaking tenant isolation?"
- **Audit Log Narrative**: Generating human-readable summaries of complex multi-step auth events.

### 2.3 Intelligent Configuration Generation
AI may be used to bootstrap complex configurations by generating:
- **Tenant & Client Profiles**: E.g., "Generate a configuration for a multi-tenant SaaS IdP."
- **Scope & Redirect URI Definitions**: Proposing least-privilege scope sets and valid redirect patterns.
- **👉 Mandatory Requirement**: All AI-generated configurations must be accompanied by an **explanation of intent** and require **explicit human approval** before being applied to a production environment.

## 3. The Explainability Requirement

For any allowed AI use case (e.g., Anomaly Detection), the output MUST include a **verifiable causal chain**. The system must be able to state:
1. What data points were analyzed?
2. What pattern was detected?
3. Why does this pattern warrant a "Risk Score" shift or a specific configuration suggestion?

## 4. Long-Term Commitment

Any proposal to integrate AI into the core "Red Line" paths will be rejected. OpenTrusty is committed to a future where security is proven by logic, while operational complexity is managed with AI assistance.
