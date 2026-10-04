package storage

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/qdrant/go-client/qdrant"
)

const vectorDimension = 768

// sourceIDKey holds the original (non-UUID) id in the point payload so it can
// be recovered on search. Qdrant point IDs must be UUID or uint.
const sourceIDKey = "source_id"

// pointID derives a deterministic UUIDv5 from an arbitrary string id (e.g. URL),
// so the same source id always maps to the same qdrant point.
func pointID(id string) string {
	return uuid.NewSHA1(uuid.NameSpaceURL, []byte(id)).String()
}

type QdrantClient struct {
	client         *qdrant.Client
	collectionName string
}

func NewQdrantClient(ctx context.Context, address, collectionName string) (*QdrantClient, error) {
	if address == "" {
		address = "localhost"
	}

	if collectionName == "" {
		collectionName = "articles"
	}

	conn, err := qdrant.NewClient(&qdrant.Config{
		Host: address,
		Port: 6334,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create qdrant client: %w", err)
	}

	client := &QdrantClient{
		client:         conn,
		collectionName: collectionName,
	}

	if err := client.ensureCollection(ctx); err != nil {
		return nil, err
	}
	return client, nil
}

func (qc *QdrantClient) ensureCollection(ctx context.Context) error {
	exists, err := qc.client.CollectionExists(ctx, qc.collectionName)
	if err != nil {
		return fmt.Errorf("unable to check if collection %q exists: %w",
			qc.collectionName, err)
	}
	if exists {
		return nil
	}

	err = qc.client.CreateCollection(ctx, &qdrant.CreateCollection{
		CollectionName: qc.collectionName,
		VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
			Size:     vectorDimension,
			Distance: qdrant.Distance_Cosine,
		}),
	})
	if err != nil {
		return fmt.Errorf("unable to create collection %q: %w", qc.collectionName, err)
	}
	return nil
}

func (qc *QdrantClient) Store(ctx context.Context, id string, vector []float32, metadata map[string]any) error {
	payload := make(map[string]*qdrant.Value)
	for k, v := range metadata {
		payload[k] = valueToQdrant(v)
	}
	payload[sourceIDKey] = qdrant.NewValueString(id)
	point := &qdrant.PointStruct{
		Id:      qdrant.NewID(pointID(id)),
		Vectors: qdrant.NewVectors(vector...),
		Payload: payload,
	}

	_, err := qc.client.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: qc.collectionName,
		Points:         []*qdrant.PointStruct{point},
	})

	return err
}

func (qc *QdrantClient) Search(ctx context.Context, vector []float32, limit int) ([]SearchMatch, error) {
	response, err := qc.client.Query(ctx, &qdrant.QueryPoints{
		CollectionName: qc.collectionName,
		Query:          qdrant.NewQuery(vector...),
		Limit:          qdrant.PtrOf(uint64(limit)),
		WithPayload:    qdrant.NewWithPayload(true),
	})
	if err != nil {
		return nil, err
	}
	matches := make([]SearchMatch, len(response))

	for i, hit := range response {
		matches[i] = SearchMatch{
			ID:    hit.Payload[sourceIDKey].GetStringValue(),
			Score: hit.Score,
		}
	}
	return matches, nil
}

type SearchMatch struct {
	ID    string
	Score float32
}

func valueToQdrant(v any) *qdrant.Value {
	switch val := v.(type) {
	case string:
		return qdrant.NewValueString(val)
	case int64:
		return qdrant.NewValueInt(val)
	case float64:
		return qdrant.NewValueDouble(val)
	case bool:
		return qdrant.NewValueBool(val)
	default:
		return qdrant.NewValueString(fmt.Sprintf("%v", v))
	}
}
