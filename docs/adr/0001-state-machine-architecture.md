# ADR 0001: State Machine Architecture

## Status

Accepted

## Context

Agent execution requires predictable, debuggable behavior. Traditional agent frameworks use implicit control flow, making it hard to reason about agent state, implement replay, or enforce policies at specific execution points.

## Decision

We implement a **finite state machine (FSM)** for agent execution with explicit states:

- `Init` - Agent initialization
- `Planning` - LLM generates next action
- `ToolSelect` - Tool chosen, awaiting validation
- `PolicyCheck` - Policy evaluation in progress
- `WaitingApproval` - Human approval required
- `Executing` - Tool execution in progress
- `Observing` - Recording results
- `Completed` - Goal achieved
- `Failed` - Unrecoverable error
- `Denied` - Policy blocked action
- `Interrupted` - User cancellation

### State Transitions

```
Init → Planning → ToolSelect → PolicyCheck → [WaitingApproval] → Executing → Observing → Planning
                                           ↓                                           ↓
                                       Denied                                    Completed/Failed
```

### Implementation

Located in `pkg/runtime/state.go` and `pkg/runtime/engine.go`.

## Consequences

### Positive

- **Deterministic Replay**: Each state transition is recorded, enabling exact replay
- **Observable**: Clear visibility into where agent is in execution
- **Interruptible**: Can pause/cancel at state boundaries
- **Testable**: Easy to test state transitions in isolation
- **Policy Enforcement**: Natural checkpoints for validation

### Negative

- More verbose than implicit control flow
- Requires careful state transition management
- Additional complexity compared to simple loop

## Alternatives Considered

### 1. Event-Driven Architecture
- Pros: Flexible, decoupled
- Cons: Harder to replay, less predictable ordering

### 2. Simple Loop
- Pros: Simpler code
- Cons: No clear pause points, harder to debug, no replay support

### 3. Actor Model
- Pros: Concurrent, message-passing
- Cons: Over-engineered for single-agent execution, harder to implement deterministic replay

## Related

- ADR 0005: Deterministic Replay
- Implementation: `pkg/runtime/engine.go`
