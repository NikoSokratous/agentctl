package graphql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/graphql-go/graphql"
)

// Schema defines the GraphQL schema
type Schema struct {
	db     *sql.DB
	schema graphql.Schema
}

// NewSchema creates a new GraphQL schema
func NewSchema(db *sql.DB) (*Schema, error) {
	s := &Schema{db: db}

	// Define types
	agentType := s.defineAgentType()
	workflowType := s.defineWorkflowType()
	runType := s.defineRunType()
	costType := s.defineCostType()

	// Define root query
	rootQuery := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"agent": &graphql.Field{
				Type: agentType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: s.resolveAgent,
			},
			"agents": &graphql.Field{
				Type: graphql.NewList(agentType),
				Args: graphql.FieldConfigArgument{
					"limit": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
				},
				Resolve: s.resolveAgents,
			},
			"workflow": &graphql.Field{
				Type: workflowType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: s.resolveWorkflow,
			},
			"workflows": &graphql.Field{
				Type:    graphql.NewList(workflowType),
				Resolve: s.resolveWorkflows,
			},
			"runs": &graphql.Field{
				Type: graphql.NewList(runType),
				Args: graphql.FieldConfigArgument{
					"agentId": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"status": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: s.resolveRuns,
			},
			"costs": &graphql.Field{
				Type: costType,
				Args: graphql.FieldConfigArgument{
					"tenantId": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"period": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: s.resolveCosts,
			},
		},
	})

	// Define mutations
	rootMutation := graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"createAgent": &graphql.Field{
				Type: agentType,
				Args: graphql.FieldConfigArgument{
					"role": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"goal": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"provider": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"model": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: s.createAgent,
			},
			"executeWorkflow": &graphql.Field{
				Type: runType,
				Args: graphql.FieldConfigArgument{
					"workflowId": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"inputs": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: s.executeWorkflow,
			},
		},
	})

	// Define subscriptions
	rootSubscription := graphql.NewObject(graphql.ObjectConfig{
		Name: "Subscription",
		Fields: graphql.Fields{
			"runUpdated": &graphql.Field{
				Type: runType,
				Args: graphql.FieldConfigArgument{
					"runId": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: s.subscribeRunUpdates,
			},
			"agentMetrics": &graphql.Field{
				Type: graphql.NewObject(graphql.ObjectConfig{
					Name: "AgentMetrics",
					Fields: graphql.Fields{
						"agentId":  &graphql.Field{Type: graphql.String},
						"cpu":      &graphql.Field{Type: graphql.Float},
						"memory":   &graphql.Field{Type: graphql.Float},
						"requests": &graphql.Field{Type: graphql.Int},
					},
				}),
				Args: graphql.FieldConfigArgument{
					"agentId": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: s.subscribeAgentMetrics,
			},
		},
	})

	// Create schema
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query:        rootQuery,
		Mutation:     rootMutation,
		Subscription: rootSubscription,
	})

	if err != nil {
		return nil, fmt.Errorf("create schema: %w", err)
	}

	s.schema = schema
	return s, nil
}

// defineAgentType defines the Agent GraphQL type
func (s *Schema) defineAgentType() *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: "Agent",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.String,
			},
			"role": &graphql.Field{
				Type: graphql.String,
			},
			"goal": &graphql.Field{
				Type: graphql.String,
			},
			"provider": &graphql.Field{
				Type: graphql.String,
			},
			"model": &graphql.Field{
				Type: graphql.String,
			},
			"status": &graphql.Field{
				Type: graphql.String,
			},
			"createdAt": &graphql.Field{
				Type: graphql.DateTime,
			},
		},
	})
}

// defineWorkflowType defines the Workflow GraphQL type
func (s *Schema) defineWorkflowType() *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: "Workflow",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.String,
			},
			"name": &graphql.Field{
				Type: graphql.String,
			},
			"description": &graphql.Field{
				Type: graphql.String,
			},
			"steps": &graphql.Field{
				Type: graphql.NewList(graphql.String),
			},
			"status": &graphql.Field{
				Type: graphql.String,
			},
		},
	})
}

// defineRunType defines the Run GraphQL type
func (s *Schema) defineRunType() *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: "Run",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.String,
			},
			"agentId": &graphql.Field{
				Type: graphql.String,
			},
			"status": &graphql.Field{
				Type: graphql.String,
			},
			"goal": &graphql.Field{
				Type: graphql.String,
			},
			"result": &graphql.Field{
				Type: graphql.String,
			},
			"createdAt": &graphql.Field{
				Type: graphql.DateTime,
			},
			"completedAt": &graphql.Field{
				Type: graphql.DateTime,
			},
		},
	})
}

// defineCostType defines the Cost GraphQL type
func (s *Schema) defineCostType() *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: "Cost",
		Fields: graphql.Fields{
			"total": &graphql.Field{
				Type: graphql.Float,
			},
			"byAgent": &graphql.Field{
				Type: graphql.NewList(graphql.NewObject(graphql.ObjectConfig{
					Name: "AgentCost",
					Fields: graphql.Fields{
						"agentId": &graphql.Field{Type: graphql.String},
						"cost":    &graphql.Field{Type: graphql.Float},
					},
				})),
			},
			"period": &graphql.Field{
				Type: graphql.String,
			},
		},
	})
}

// Resolvers

func (s *Schema) resolveAgent(p graphql.ResolveParams) (interface{}, error) {
	id := p.Args["id"].(string)

	// Query database
	query := `SELECT id, role, goal, llm_provider, llm_model, status, created_at FROM agents WHERE id = ?`

	var agentID, role, goal, provider, model, status, createdAt string
	err := s.db.QueryRow(query, id).Scan(
		&agentID, &role, &goal,
		&provider, &model,
		&status, &createdAt,
	)

	if err != nil {
		return nil, fmt.Errorf("query agent: %w", err)
	}

	agent := map[string]interface{}{
		"id":        agentID,
		"role":      role,
		"goal":      goal,
		"provider":  provider,
		"model":     model,
		"status":    status,
		"createdAt": createdAt,
	}

	return agent, nil
}

func (s *Schema) resolveAgents(p graphql.ResolveParams) (interface{}, error) {
	limit := 10
	if l, ok := p.Args["limit"].(int); ok {
		limit = l
	}

	query := `SELECT id, role, goal, llm_provider, llm_model, status, created_at FROM agents LIMIT ?`
	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("query agents: %w", err)
	}
	defer rows.Close()

	var agents []map[string]interface{}
	for rows.Next() {
		var agentID, role, goal, provider, model, status, createdAt string
		err := rows.Scan(
			&agentID, &role, &goal,
			&provider, &model,
			&status, &createdAt,
		)
		if err != nil {
			continue
		}
		agent := map[string]interface{}{
			"id":        agentID,
			"role":      role,
			"goal":      goal,
			"provider":  provider,
			"model":     model,
			"status":    status,
			"createdAt": createdAt,
		}
		agents = append(agents, agent)
	}

	return agents, nil
}

func (s *Schema) resolveWorkflow(p graphql.ResolveParams) (interface{}, error) {
	id := p.Args["id"].(string)
	// Mock implementation
	return map[string]interface{}{
		"id":          id,
		"name":        "Sample Workflow",
		"description": "A sample workflow",
		"status":      "active",
	}, nil
}

func (s *Schema) resolveWorkflows(p graphql.ResolveParams) (interface{}, error) {
	// Mock implementation
	return []map[string]interface{}{
		{
			"id":          "wf-1",
			"name":        "Code Review",
			"description": "Automated code review workflow",
			"status":      "active",
		},
	}, nil
}

func (s *Schema) resolveRuns(p graphql.ResolveParams) (interface{}, error) {
	// Mock implementation
	return []map[string]interface{}{
		{
			"id":      "run-1",
			"agentId": "agent-1",
			"status":  "completed",
			"goal":    "Write code",
		},
	}, nil
}

func (s *Schema) resolveCosts(p graphql.ResolveParams) (interface{}, error) {
	tenantID := p.Args["tenantId"].(string)
	_ = tenantID // Use tenantID in actual implementation

	// Mock implementation
	return map[string]interface{}{
		"total":  125.50,
		"period": "2026-02",
		"byAgent": []map[string]interface{}{
			{"agentId": "agent-1", "cost": 75.25},
			{"agentId": "agent-2", "cost": 50.25},
		},
	}, nil
}

func (s *Schema) createAgent(p graphql.ResolveParams) (interface{}, error) {
	// Mock implementation
	return map[string]interface{}{
		"id":       "agent-new",
		"role":     p.Args["role"],
		"goal":     p.Args["goal"],
		"provider": p.Args["provider"],
		"model":    p.Args["model"],
		"status":   "active",
	}, nil
}

func (s *Schema) executeWorkflow(p graphql.ResolveParams) (interface{}, error) {
	// Mock implementation
	return map[string]interface{}{
		"id":      "run-new",
		"agentId": "agent-1",
		"status":  "running",
		"goal":    "Execute workflow",
	}, nil
}

func (s *Schema) subscribeRunUpdates(p graphql.ResolveParams) (interface{}, error) {
	// In production, this would use channels/websockets for real-time updates
	return nil, fmt.Errorf("subscriptions require WebSocket connection")
}

func (s *Schema) subscribeAgentMetrics(p graphql.ResolveParams) (interface{}, error) {
	// In production, this would stream metrics via WebSocket
	return nil, fmt.Errorf("subscriptions require WebSocket connection")
}

// Execute executes a GraphQL query
func (s *Schema) Execute(query string, variables map[string]interface{}) *graphql.Result {
	params := graphql.Params{
		Schema:         s.schema,
		RequestString:  query,
		VariableValues: variables,
	}

	return graphql.Do(params)
}

// ExecuteWithContext executes a GraphQL query with context
func (s *Schema) ExecuteWithContext(ctx context.Context, query string, variables map[string]interface{}) *graphql.Result {
	params := graphql.Params{
		Schema:         s.schema,
		RequestString:  query,
		VariableValues: variables,
		Context:        ctx,
	}

	return graphql.Do(params)
}
