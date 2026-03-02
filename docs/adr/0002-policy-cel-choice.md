# ADR 0002: Policy Engine with CEL

## Status

Accepted

## Context

Agent autonomy requires runtime policy enforcement. Policies must be:
- Declarative (not code)
- Evaluable without execution
- Auditable
- Version-controllable
- Language-agnostic

## Decision

We use **Common Expression Language (CEL)** for policy expressions.

### Why CEL?

1. **Sandboxed**: No arbitrary code execution
2. **Fast**: Compiles to bytecode, microsecond evaluation
3. **Type-safe**: Static type checking at compile time
4. **Standard**: Used by Kubernetes, Google Cloud, Firebase
5. **Simple**: C-like syntax, easy to learn

### Policy Structure

```yaml
version: "1"
rules:
  - name: block-prod-writes
    match:
      tool: db_write
      environment: production
    action: deny
    message: "Production writes blocked"
    
  - name: high-risk-approval
    match:
      condition: "risk_score >= 0.8 && environment == 'production'"
    action: require_approval
    approvers: [admin]
```

### Available Variables

- `tool` (string) - Tool name
- `environment` (string) - Deployment environment
- `risk_score` (double) - 0.0 to 1.0
- `input` (map) - Tool input parameters
- `input.to` (string) - Special accessor for nested fields

## Consequences

### Positive

- **Security**: No code injection risks
- **Performance**: Sub-millisecond evaluation
- **Auditability**: Policies are YAML, easy to review
- **Testability**: Can test expressions without runtime
- **Portability**: CEL implementations exist for many languages

### Negative

- Learning curve for CEL syntax
- Limited to expression evaluation (no loops, no side effects)
- Complex conditions can be hard to debug

## Alternatives Considered

### 1. Rego (Open Policy Agent)
- Pros: More powerful, Prolog-like logic programming
- Cons: Steeper learning curve, overkill for agent policies

### 2. JavaScript/Lua Embedded
- Pros: Full programming language
- Cons: Security risks, harder to audit, performance overhead

### 3. Custom DSL
- Pros: Perfect fit for our use case
- Cons: Maintenance burden, no ecosystem, users learn custom syntax

### 4. Python (via exec)
- Pros: Familiar to ML/AI developers
- Cons: Security nightmare, not sandboxed

## Implementation

- Policy engine: `pkg/policy/engine.go`
- CEL evaluation: `pkg/policy/cel.go`
- Rule matching: `pkg/policy/rules.go`

## Related

- ADR 0001: State Machine (policy checks at ToolSelect state)
- Google CEL Spec: https://github.com/google/cel-spec
