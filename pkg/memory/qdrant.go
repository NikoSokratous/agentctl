package memory

import (
	"context"
	"encoding/json"
	"fmt"

	pb "github.com/qdrant/go-client/qdrant"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// QdrantStore implements semantic memory using Qdrant vector database.
type QdrantStore struct {
	conn              *grpc.ClientConn
	collectionsClient pb.CollectionsClient
	pointsClient      pb.PointsClient
	collectionName    string
	vectorSize        uint64
}

// QdrantConfig holds configuration for Qdrant connection.
type QdrantConfig struct {
	Host           string `yaml:"host" json:"host"`
	Port           int    `yaml:"port" json:"port"`
	CollectionName string `yaml:"collection" json:"collection"`
	VectorSize     int    `yaml:"vector_size" json:"vector_size"`
	APIKey         string `yaml:"api_key" json:"api_key,omitempty"`
}

// NewQdrantStore creates a new Qdrant-backed semantic store.
func NewQdrantStore(ctx context.Context, config QdrantConfig) (*QdrantStore, error) {
	// Connect to Qdrant
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	conn, err := grpc.DialContext(
		ctx,
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("connect to Qdrant: %w", err)
	}

	store := &QdrantStore{
		conn:              conn,
		collectionsClient: pb.NewCollectionsClient(conn),
		pointsClient:      pb.NewPointsClient(conn),
		collectionName:    config.CollectionName,
		vectorSize:        uint64(config.VectorSize),
	}

	// Ensure collection exists
	if err := store.ensureCollection(ctx); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure collection: %w", err)
	}

	return store, nil
}

// ensureCollection creates the collection if it doesn't exist.
func (q *QdrantStore) ensureCollection(ctx context.Context) error {
	// Check if collection exists
	collections, err := q.collectionsClient.List(ctx, &pb.ListCollectionsRequest{})
	if err != nil {
		return fmt.Errorf("list collections: %w", err)
	}

	// Check if our collection exists
	for _, collection := range collections.GetCollections() {
		if collection.Name == q.collectionName {
			return nil // Collection exists
		}
	}

	// Create collection
	_, err = q.collectionsClient.Create(ctx, &pb.CreateCollection{
		CollectionName: q.collectionName,
		VectorsConfig: &pb.VectorsConfig{
			Config: &pb.VectorsConfig_Params{
				Params: &pb.VectorParams{
					Size:     q.vectorSize,
					Distance: pb.Distance_Cosine,
				},
			},
		},
	})

	if err != nil {
		return fmt.Errorf("create collection: %w", err)
	}

	return nil
}

// Store adds a memory with embedding to Qdrant (implements custom interface).
func (q *QdrantStore) Store(ctx context.Context, agentName, key string, value interface{}, embedding []float32) error {
	// Serialize value
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal value: %w", err)
	}

	// Create point
	point := &pb.PointStruct{
		Id: &pb.PointId{
			PointIdOptions: &pb.PointId_Uuid{
				Uuid: key,
			},
		},
		Vectors: &pb.Vectors{
			VectorsOptions: &pb.Vectors_Vector{
				Vector: &pb.Vector{
					Data: embedding,
				},
			},
		},
		Payload: map[string]*pb.Value{
			"agent": {
				Kind: &pb.Value_StringValue{StringValue: agentName},
			},
			"key": {
				Kind: &pb.Value_StringValue{StringValue: key},
			},
			"value": {
				Kind: &pb.Value_StringValue{StringValue: string(valueBytes)},
			},
		},
	}

	// Upsert point
	waitUpsert := true
	_, err = q.pointsClient.Upsert(ctx, &pb.UpsertPoints{
		CollectionName: q.collectionName,
		Wait:           &waitUpsert,
		Points:         []*pb.PointStruct{point},
	})

	return err
}

// Upsert implements SemanticStore interface.
func (q *QdrantStore) Upsert(ctx context.Context, agentID string, id string, embedding Embedding, metadata map[string]any) error {
	return q.Store(ctx, agentID, id, metadata, embedding)
}

// Search implements SemanticStore.Search interface (uses Embedding type).
func (q *QdrantStore) Search(ctx context.Context, agentID string, query Embedding, topK int) ([]SearchResult, error) {
	// Build filter for agent
	filter := &pb.Filter{
		Must: []*pb.Condition{
			{
				ConditionOneOf: &pb.Condition_Field{
					Field: &pb.FieldCondition{
						Key: "agent",
						Match: &pb.Match{
							MatchValue: &pb.Match_Keyword{
								Keyword: agentID,
							},
						},
					},
				},
			},
		},
	}

	// Search
	response, err := q.pointsClient.Search(ctx, &pb.SearchPoints{
		CollectionName: q.collectionName,
		Vector:         query,
		Limit:          uint64(topK),
		Filter:         filter,
		WithPayload:    &pb.WithPayloadSelector{SelectorOptions: &pb.WithPayloadSelector_Enable{Enable: true}},
	})

	if err != nil {
		return nil, err
	}

	// Parse results
	results := make([]SearchResult, 0, len(response.Result))
	for _, hit := range response.Result {
		valuePayload := hit.Payload["value"]
		if valuePayload == nil {
			continue
		}

		var value interface{}
		metadata := make(map[string]any)

		if strValue, ok := valuePayload.Kind.(*pb.Value_StringValue); ok {
			json.Unmarshal([]byte(strValue.StringValue), &value)
			metadata["value"] = value
		}

		if keyValue := hit.Payload["key"]; keyValue != nil {
			metadata["key"] = keyValue.GetStringValue()
		}

		results = append(results, SearchResult{
			ID:       hit.Id.GetUuid(),
			Score:    float64(hit.Score),
			Metadata: metadata,
		})
	}

	return results, nil
}

// SemanticResult is the result format for legacy custom Search method.
type SemanticResult struct {
	Key   string
	Value interface{}
	Score float64
}

// Delete removes a memory from Qdrant.
func (q *QdrantStore) Delete(ctx context.Context, agentName, key string) error {
	filter := &pb.Filter{
		Must: []*pb.Condition{
			{
				ConditionOneOf: &pb.Condition_Field{
					Field: &pb.FieldCondition{
						Key: "key",
						Match: &pb.Match{
							MatchValue: &pb.Match_Keyword{
								Keyword: key,
							},
						},
					},
				},
			},
		},
	}

	waitDelete := true
	_, err := q.pointsClient.Delete(ctx, &pb.DeletePoints{
		CollectionName: q.collectionName,
		Wait:           &waitDelete,
		Points: &pb.PointsSelector{
			PointsSelectorOneOf: &pb.PointsSelector_Filter{
				Filter: filter,
			},
		},
	})

	return err
}

// DeleteAgent implements SemanticStore interface.
func (q *QdrantStore) DeleteAgent(ctx context.Context, agentID string) error {
	filter := &pb.Filter{
		Must: []*pb.Condition{
			{
				ConditionOneOf: &pb.Condition_Field{
					Field: &pb.FieldCondition{
						Key: "agent",
						Match: &pb.Match{
							MatchValue: &pb.Match_Keyword{
								Keyword: agentID,
							},
						},
					},
				},
			},
		},
	}

	waitDelete := true
	_, err := q.pointsClient.Delete(ctx, &pb.DeletePoints{
		CollectionName: q.collectionName,
		Wait:           &waitDelete,
		Points: &pb.PointsSelector{
			PointsSelectorOneOf: &pb.PointsSelector_Filter{
				Filter: filter,
			},
		},
	})

	return err
}

// Close closes the Qdrant connection.
func (q *QdrantStore) Close() error {
	if q.conn != nil {
		return q.conn.Close()
	}
	return nil
}
