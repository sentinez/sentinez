// Copyright 2025 Duc-Hung Ho.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package es provides an implementation of the database using elasticsearch.
package es

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/sentinez/sentinez/pkg/common/utils"

	"github.com/elastic/go-elasticsearch/v8"
)

// ElasticSearch is a generic Elasticsearch repository.
type ElasticSearch[T any, ID comparable] struct {
	client    *elasticsearch.Client
	indexName string
}

// New initializes a new SearchLayer with Elasticsearch.
// SearchLayer is a generic Elasticsearch repository.
// Must be implemented Create, Get method by the user.
func New[T any, ID comparable](
	client *elasticsearch.Client, indexName string) *ElasticSearch[T, ID] {

	return &ElasticSearch[T, ID]{
		client:    client,
		indexName: indexName,
	}
}

// Create inserts a new document into Elasticsearch.
func (es *ElasticSearch[T, ID]) Create(
	ctx context.Context, entity T) (T, error) {

	_ = ctx
	_ = entity

	panic("not implemented")
}

// Get retrieves a document from Elasticsearch by ID.
func (es *ElasticSearch[T, ID]) Get(ctx context.Context, id ID) (T, error) {
	_ = ctx
	_ = id

	panic("not implemented")
}

// GetAll retrieves all documents from the index.
func (es *ElasticSearch[T, ID]) GetAll(ctx context.Context) ([]T, error) {
	query := `{"query": {"match_all": {}}}`

	res, err := es.client.Search(
		es.client.Search.WithContext(ctx),
		es.client.Search.WithIndex(es.indexName),
		es.client.Search.WithBody(strings.NewReader(query)),
		es.client.Search.WithSize(1000))
	if err != nil {
		return nil, err
	}
	defer utils.CallBack(res.Body.Close)

	// Decode the JSON response
	var result struct {
		Hits struct {
			Hits []struct {
				Source T `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	var entities []T
	for _, hit := range result.Hits.Hits {
		entities = append(entities, hit.Source)
	}

	return entities, nil
}

// Update modifies an existing document in Elasticsearch by ID.
func (es *ElasticSearch[T, ID]) Update(
	ctx context.Context, id ID, entity T) (T, error) {

	_ = ctx
	data, err := json.Marshal(map[string]any{
		"doc": entity,
	})
	if err != nil {
		return entity, err
	}

	// Send the update request
	res, err := es.client.Update(es.indexName,
		fmt.Sprintf("%v", id), bytes.NewReader(data))
	if err != nil {
		return entity, err
	}
	defer utils.CallBack(res.Body.Close)

	// Check for errors
	if res.IsError() {
		return entity, fmt.Errorf("error updating document: %s", res.String())
	}

	return entity, nil
}

// Delete removes a document from Elasticsearch by ID.
func (es *ElasticSearch[T, ID]) Delete(ctx context.Context, id ID) error {
	_ = ctx
	res, err := es.client.Delete(es.indexName, fmt.Sprintf("%v", id))
	if err != nil {
		return err
	}
	defer utils.CallBack(res.Body.Close)

	if res.StatusCode == 404 {
		return errors.New("document not found")
	}

	return nil
}

// Exists checks if a document exists in Elasticsearch.
func (es *ElasticSearch[T, ID]) Exists(
	ctx context.Context, id ID) (bool, error) {

	_ = ctx
	res, err := es.client.Exists(es.indexName, fmt.Sprintf("%v", id))
	if err != nil {
		return false, err
	}
	defer utils.CallBack(res.Body.Close)

	return res.StatusCode == 200, nil
}

// Count returns the number of documents in the Elasticsearch index.
func (es *ElasticSearch[T, ID]) Count(ctx context.Context) (int64, error) {
	_ = ctx
	res, err := es.client.Count(es.client.Count.WithIndex(es.indexName))
	if err != nil {
		return 0, err
	}
	defer utils.CallBack(res.Body.Close)

	var result struct {
		Count int64 `json:"count"`
	}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return 0, err
	}

	return result.Count, nil
}
