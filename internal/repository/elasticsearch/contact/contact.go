package contact

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
)

type ContactSearchRepository struct {
	client *opensearch.Client
	index  string
}

func NewContactSearchRepository(client *opensearch.Client, index string) *ContactSearchRepository {
	return &ContactSearchRepository{
		client: client,
		index:  index,
	}
}

func (r *ContactSearchRepository) CreateIndex(ctx context.Context) error {
	const op = "ContactSearchRepository.CreateIndex"
	logger := domains.GetLogger(ctx).WithField("operation", op)

	logger.Info("checking if index exists")
	existsReq := opensearchapi.IndicesExistsRequest{
		Index: []string{r.index},
	}

	existsRes, err := existsReq.Do(ctx, r.client)
	if err != nil {
		logger.WithError(err).Error("failed to check if index exists")
		return fmt.Errorf("%s: failed to check if index exists: %w", op, err)
	}
	defer existsRes.Body.Close()

	if existsRes.StatusCode == 200 {
		logger.Warning("deleting existing index for clean state")
		deleteReq := opensearchapi.IndicesDeleteRequest{
			Index: []string{r.index},
		}
		deleteRes, err := deleteReq.Do(ctx, r.client)
		if err != nil {
			logger.WithError(err).Error("failed to delete existing index")
			return fmt.Errorf("%s: failed to delete existing index: %w", op, err)
		}
		defer deleteRes.Body.Close()
	}

	mapping := `{
		"settings": {
			"number_of_shards": 1,
			"number_of_replicas": 0,
			"analysis": {
				"tokenizer": {
					"edge_ngram_tokenizer": {
						"type": "edge_ngram",
						"min_gram": 2,
						"max_gram": 20,
						"token_chars": ["letter", "digit"]
					}
				},
				"analyzer": {
					"edge_ngram_analyzer": {
						"type": "custom",
						"tokenizer": "edge_ngram_tokenizer",
						"filter": ["lowercase"]
					}
				}
			}
		},
		"mappings": {
			"properties": {
				"user_id": {
					"type": "keyword"
				},
				"contact_user_id": {
					"type": "keyword"
				},
				"username": {
					"type": "text",
					"analyzer": "edge_ngram_analyzer",
					"search_analyzer": "standard"
				},
				"name": {
					"type": "text",
					"analyzer": "edge_ngram_analyzer",
					"search_analyzer": "standard"
				},
				"phone_number": {
					"type": "keyword"
				}
			}
		}
	}`

	createReq := opensearchapi.IndicesCreateRequest{
		Index: r.index,
		Body:  strings.NewReader(mapping),
	}

	logger.WithField("index", r.index).Info("creating elasticsearch index")
	createRes, err := createReq.Do(ctx, r.client)
	if err != nil {
		logger.WithError(err).Error("failed to create index")
		return fmt.Errorf("%s: failed to create index: %w", op, err)
	}
	defer createRes.Body.Close()

	if createRes.IsError() {
		var errBody map[string]interface{}
		if err := json.NewDecoder(createRes.Body).Decode(&errBody); err == nil {
			errBytes, _ := json.Marshal(errBody)
			logger.WithField("error_body", string(errBytes)).Error("failed to create index")
			return fmt.Errorf("%s: failed to create index: %s", op, string(errBytes))
		}
		logger.WithField("status", createRes.Status()).Error("failed to create index")
		return fmt.Errorf("%s: failed to create index: %s", op, createRes.Status())
	}

	logger.Info("elasticsearch index created successfully")
	return nil
}

func (r *ContactSearchRepository) IndexContact(ctx context.Context, userID, contactUserID, username, name, phoneNumber string) error {
	const op = "ContactSearchRepository.IndexContact"
	logger := domains.GetLogger(ctx).WithField("operation", op)

	doc := map[string]interface{}{
		"user_id":         userID,
		"contact_user_id": contactUserID,
		"username":        username,
		"name":            name,
		"phone_number":    phoneNumber,
	}

	data, err := json.Marshal(doc)
	if err != nil {
		logger.WithError(err).Error("failed to marshal document")
		return fmt.Errorf("%s: %w", op, err)
	}

	docID := fmt.Sprintf("%s_%s", userID, contactUserID)
	logger.WithField("doc_id", docID).Info("indexing contact")

	req := opensearchapi.IndexRequest{
		Index:      r.index,
		DocumentID: docID,
		Body:       bytes.NewReader(data),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, r.client)
	if err != nil {
		logger.WithError(err).Error("failed to index document")
		return fmt.Errorf("%s: %w", op, err)
	}
	defer res.Body.Close()

	if res.IsError() {
		logger.WithField("response", res.String()).Error("failed to index document")
		return fmt.Errorf("%s: failed to index document: %s", op, res.String())
	}

	logger.Info("contact indexed successfully")
	return nil
}

func (r *ContactSearchRepository) SearchContacts(ctx context.Context, userID, query string) ([]map[string]interface{}, error) {
	const op = "ContactSearchRepository.SearchContacts"
	logger := domains.GetLogger(ctx).WithField("operation", op)

	logger.WithField("query", query).WithField("user_id", userID).Info("searching contacts")

	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []interface{}{
					map[string]interface{}{
						"term": map[string]interface{}{
							"user_id": userID,
						},
					},
					map[string]interface{}{
						"bool": map[string]interface{}{
							"should": []interface{}{
								map[string]interface{}{
									"match": map[string]interface{}{
										"username": map[string]interface{}{
											"query": query,
											"boost": 2.0,
										},
									},
								},
								map[string]interface{}{
									"match": map[string]interface{}{
										"name": query,
									},
								},
							},
							"minimum_should_match": 1,
						},
					},
				},
			},
		},
		"size": 50,
	}

	queryBody, err := json.Marshal(searchQuery)
	if err != nil {
		logger.WithError(err).Error("failed to marshal search query")
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	req := opensearchapi.SearchRequest{
		Index: []string{r.index},
		Body:  bytes.NewReader(queryBody),
	}

	res, err := req.Do(ctx, r.client)
	if err != nil {
		logger.WithError(err).Error("failed to execute search")
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer res.Body.Close()

	if res.IsError() {
		logger.WithField("response", res.String()).Error("search returned error")
		return nil, fmt.Errorf("%s: search error: %s", op, res.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		logger.WithError(err).Error("failed to decode search response")
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	hits, ok := result["hits"].(map[string]interface{})
	if !ok {
		logger.Warning("no hits in search response")
		return []map[string]interface{}{}, nil
	}

	hitsArray, ok := hits["hits"].([]interface{})
	if !ok {
		logger.Warning("no hits array in search response")
		return []map[string]interface{}{}, nil
	}

	contacts := make([]map[string]interface{}, 0, len(hitsArray))
	for _, hit := range hitsArray {
		hitMap, ok := hit.(map[string]interface{})
		if !ok {
			continue
		}
		source, ok := hitMap["_source"].(map[string]interface{})
		if !ok {
			continue
		}
		contacts = append(contacts, source)
	}

	logger.WithField("results_count", len(contacts)).Info("search completed")
	return contacts, nil
}

func (r *ContactSearchRepository) DeleteContact(ctx context.Context, userID, contactUserID string) error {
	const op = "ContactSearchRepository.DeleteContact"
	logger := domains.GetLogger(ctx).WithField("operation", op)

	docID := fmt.Sprintf("%s_%s", userID, contactUserID)
	logger.WithField("doc_id", docID).Info("deleting contact from index")

	req := opensearchapi.DeleteRequest{
		Index:      r.index,
		DocumentID: docID,
		Refresh:    "true",
	}

	res, err := req.Do(ctx, r.client)
	if err != nil {
		logger.WithError(err).Error("failed to delete document")
		return fmt.Errorf("%s: %w", op, err)
	}
	defer res.Body.Close()

	if res.IsError() && res.StatusCode != 404 {
		logger.WithField("response", res.String()).Error("failed to delete document")
		return fmt.Errorf("%s: failed to delete document: %s", op, res.String())
	}

	if res.StatusCode == 404 {
		logger.Warning("document not found, nothing to delete")
	} else {
		logger.Info("contact deleted successfully")
	}
	return nil
}
