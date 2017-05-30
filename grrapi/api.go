// Copyright 2017 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package grrapi provides a root object for interaction with GRR API library.
package grrapi

import (
	"io"
	"net/http"

	"context"
	"github.com/golang/protobuf/proto"
	"github.com/google/grr_go_api_client/client"
	"github.com/google/grr_go_api_client/connector"
	"github.com/google/grr_go_api_client/hunt"
)

// grrConnector abstracts GRR communication protocols (i.e. HTTP, Stubby)
// and handles API calls.
type grrConnector interface {
	client.Connector
	hunt.ConnectorSubset
}

// APIClient is a root object for accessing GRR API library.
type APIClient struct {
	conn grrConnector
}

// NewAPIClient creates a new root object with a given connector.
func NewAPIClient(conn grrConnector) *APIClient {
	return &APIClient{
		conn: conn,
	}
}

// NewHTTPClient creates a new root object using HTTP as protocol.
func NewHTTPClient(apiEndpoint string, pageSize int64, setAuth func(r *http.Request)) *APIClient {
	return &APIClient{
		conn: connector.NewHTTPConnector(apiEndpoint, pageSize, setAuth),
	}
}

// Close closes the underlying GRR connection.
func (c *APIClient) Close() error {
	if conn, ok := c.conn.(io.Closer); ok {
		return conn.Close()
	}
	return nil
}

// Client returns a new client ref with given id.
func (c *APIClient) Client(id string) *client.Client {
	return client.NewClient(id, c.conn)
}

// SearchClients sends a request to GRR server to search for clients using
// given query and calls fn (a callback function) on each Client object.
// An optional argument can be passed (see client.WithMaxCount).
func (c *APIClient) SearchClients(ctx context.Context, query string, fn func(c *client.Client) error, options ...client.SearchClientsOption) error {
	return client.SearchClients(ctx, c.conn, query, fn, options...)
}

// Hunt returns a new hunt ref with given id.
func (c *APIClient) Hunt(id string) *hunt.Hunt {
	return hunt.NewHunt(id, c.conn, nil)
}

// CreateHunt sends a request to GRR server to create a new hunt with
// given flow name and flow args. An optional argument can be passed
// (see hunt.CreateHuntRunnerArgs).
func (c *APIClient) CreateHunt(ctx context.Context, flowName string, flowArgs proto.Message, options ...hunt.CreateHuntOption) (*hunt.Hunt, error) {
	return hunt.CreateHunt(ctx, c.conn, flowName, flowArgs, options...)
}

// ListHunts sends a request to GRR server to list all GRR hunts and calls fn
// (a callback function) on each Hunt object. Total count of first request
// is returned.
func (c *APIClient) ListHunts(ctx context.Context, fn func(h *hunt.Hunt) error) (int64, error) {
	return hunt.ListHunts(ctx, c.conn, fn)
}

// ListHuntApprovals sends a request to GRR server to list all the hunt
// approvals belonging to requesting user and calls fn (a callback function)
// on each Approval object.
func (c *APIClient) ListHuntApprovals(ctx context.Context, fn func(a *hunt.Approval) error) error {
	return hunt.ListApprovals(ctx, c.conn, fn)
}
