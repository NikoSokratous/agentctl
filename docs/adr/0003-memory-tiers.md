# ADR 0003: Memory Architecture (Three-Tier)

## Status

Accepted

## Context

AI agents need different types of memory for different purposes:
- Short-term state during execution
- Long-term facts and preferences
- Semantic search over past experiences
- Immutable audit trail

## Decision

We implement a **three-tier memory architecture** with a separate audit log:

### 1. Working Memory (In-Memory)
- **Purpose**: Session-scoped temporary state
- **Lifetime**: Duration of single run
- **Storage**: In-memory map
- **Use Cases**: Intermediate results, loop counters, flags

### 2. Persistent Memory (Key-Value Store)
- **Purpose**: Agent-scoped long-term storage
- **Lifetime**: Across multiple runs
- **Storage**: SQLite/PostgreSQL
- **Use Cases**: User preferences, learned facts, configuration

### 3. Semantic Memory (Vector Store)
- **Purpose**: Similarity search over experiences
- **Lifetime**: Permanent (until explicitly deleted)
- **Storage**: In-memory HNSW / Qdrant / Weaviate
- **Use Cases**: Finding similar past situations, RAG, deduplication

### 4. Execution Log (Append-Only)
- **Purpose**: Immutable audit trail
- **Lifetime**: Permanent (compliance requirement)
- **Storage**: SQLite/PostgreSQL events table
- **Use Cases**: Replay, debugging, compliance

## Architecture

```
┌─────────────────────────────────────────┐
│         Memory Manager                   │
│  (pkg/memory/manager.go)                │
└────┬─────────┬──────────┬───────────────┘
     │         │          │         
     v         v          v         
┌─────────┐ ┌──────────┐ ┌─────────────┐
│ Working │ │Persistent│ │  Semantic   │
│ Memory  │ │ Memory   │ │   Memory    │
│(in-mem) │ │(SQLite)  │ │  (Vector)   │
└─────────┘ └──────────┘ └─────────────┘
                │
                v
         ┌──────────────┐
         │ Event Log    │
         │ (immutable)  │
         └──────────────┘
```

## Consequences

### Positive

- **Clear Separation**: Each tier has specific purpose
- **Scalability**: Can swap implementations per tier
- **Performance**: Right storage for each access pattern
- **Compliance**: Immutable log meets audit requirements
- **GDPR**: Can delete agent data while preserving audit log

### Negative

- More complexity than single storage backend
- Need to manage consistency across tiers
- Multiple database connections

## Alternatives Considered

### 1. Single Unified Store
- Pros: Simpler, one connection
- Cons: Inefficient (vector search in SQL?), no clear boundaries

### 2. Two-Tier (Working + Persistent)
- Pros: Simpler than three-tier
- Cons: No semantic search, limited to key-value

### 3. External Memory Service
- Pros: Centralized, microservices-ready
- Cons: Network overhead, more moving parts

## Implementation

- Manager: `pkg/memory/manager.go`
- Working: `pkg/memory/working.go`
- Persistent: `pkg/memory/persistent.go`
- Semantic: `pkg/memory/semantic.go`
- Event Log: `internal/store/sqlite.go`

## Future Enhancements

- Memory compression (summarize old working memory)
- Cross-agent memory sharing
- Memory migration tools
- TTL for semantic embeddings

## Related

- ADR 0005: Deterministic Replay (uses execution log)
