package infra

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/elastic/go-elasticsearch/v9"
)

type ElastickInfra struct {
	Client *elasticsearch.Client
}

func NewElastickInfra() *ElastickInfra {
	cl, err := elasticsearch.NewClient(
		elasticsearch.Config{
			Addresses: []string{"http://localhost:9200"},
		})
	if err != nil {
		panic(err)
	}
	return &ElastickInfra{Client: cl}
}

func (e *ElastickInfra) Handle(input_data []byte) error {
	var data struct {
		Payload struct {
			Before map[string]interface{} `json:"before"`
			After  map[string]interface{} `json:"after"`
			Source struct {
				Table string `json:"table"`
			} `json:"source"`
			Op string `json:"op"`
		} `json:"payload"`
	}

	if err := json.Unmarshal(input_data, &data); err != nil {
		return fmt.Errorf("unmarshal error: %w", err)
	}

	p := data.Payload
	index := "log_shelter-" + p.Source.Table

	var docID string
	if p.After != nil && p.After["id"] != nil {
		docID = fmt.Sprintf("%v", p.After["id"])
	} else if p.Before != nil && p.Before["id"] != nil {
		docID = fmt.Sprintf("%v", p.Before["id"])
	}

	ctx := context.Background()

	if p.Op == "c" || p.Op == "u" || p.Op == "r" {
		body, _ := json.Marshal(p.After)
		res, err := e.Client.Index(
			index,
			bytes.NewReader(body),
			e.Client.Index.WithDocumentID(docID),
			e.Client.Index.WithContext(ctx),
		)
		if err != nil {
			return fmt.Errorf("index failed: %w", err)
		}
		defer res.Body.Close()
	}

	if p.Op == "d" {
		res, err := e.Client.Delete(index, docID, e.Client.Delete.WithContext(ctx))
		if err != nil {
			return fmt.Errorf("delete failed: %w", err)
		}
		defer res.Body.Close()
	}

	return nil
}

func (e *ElastickInfra) Search(q string) []string {
	ctx := context.Background()

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"query_string": map[string]interface{}{
				"query": q,
			},
		},
	}

	body, _ := json.Marshal(query)

	res, err := e.Client.Search(
		e.Client.Search.WithContext(ctx),
		e.Client.Search.WithIndex("log_shelter-*"),
		e.Client.Search.WithBody(bytes.NewReader(body)),
	)
	if err != nil {
		return []string{}
	}
	defer res.Body.Close()

	if res.IsError() {
		slog.Error("error", "err", res.Status())
		return []string{}
	}

	var result struct {
		Hits struct {
			Hits []struct {
				Source map[string]interface{} `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	json.NewDecoder(res.Body).Decode(&result)

	var results []string
	for _, hit := range result.Hits.Hits {
		data, _ := json.Marshal(hit.Source)
		results = append(results, string(data))
	}

	return results
}
