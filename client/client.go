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

// Package client provides a GRR client object which handles associated API calls via an arbitrary transfer protocol.
package client

import (
	"fmt"

	"context"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/google/grr_go_api_client/flow"
	acpb "github.com/google/grr_go_api_client/grrproto/grr/proto/api/client_proto"
	afpb "github.com/google/grr_go_api_client/grrproto/grr/proto/api/flow_proto"
	aupb "github.com/google/grr_go_api_client/grrproto/grr/proto/api/user_proto"
	flowpb "github.com/google/grr_go_api_client/grrproto/grr/proto/flows_proto"
	"github.com/google/grr_go_api_client/vfs"
)

// Connector abstracts GRR communication protocols (i.e. HTTP, Stubby) and handles API calls related to clients.
type Connector interface {
	// Adds given list of labels to given clients.
	AddClientsLabels(ctx context.Context, args *acpb.ApiAddClientsLabelsArgs) error
	// Removes given labels from given clients.
	RemoveClientsLabels(ctx context.Context, args *acpb.ApiRemoveClientsLabelsArgs) error
	// Returns information for a client approval with a given reason.
	GetClientApproval(ctx context.Context, args *aupb.ApiGetClientApprovalArgs) (*aupb.ApiClientApproval, error)
	// GetUsername returns username of the current user.
	GetUsername(ctx context.Context) (string, error)
	// Lists all the client approvals.
	ListClientApprovals(ctx context.Context, args *aupb.ApiListClientApprovalsArgs, fn func(m *aupb.ApiClientApproval) error) error
	// List flows on a given client.
	ListFlows(ctx context.Context, args *afpb.ApiListFlowsArgs, fn func(m *afpb.ApiFlow) error) error
	// Grant client approval.
	GrantClientApproval(ctx context.Context, args *aupb.ApiGrantClientApprovalArgs) (*aupb.ApiClientApproval, error)
	// Create a new client approval.
	CreateClientApproval(ctx context.Context, args *aupb.ApiCreateClientApprovalArgs) (*aupb.ApiClientApproval, error)
	// Start a new flow on a given client.
	CreateFlow(ctx context.Context, args *afpb.ApiCreateFlowArgs) (*afpb.ApiFlow, error)
	// Get client with a given client id.
	GetClient(ctx context.Context, args *acpb.ApiGetClientArgs) (*acpb.ApiClient, error)
	// Search for clients using a search query
	SearchClients(ctx context.Context, args *acpb.ApiSearchClientsArgs, fn func(m *acpb.ApiClient) error) error

	flow.Connector
	vfs.Connector
}

// Client object is a helper object to send requests to GRR server via an arbitrary transfer protocol.
type Client struct {
	conn Connector
	ID   string
	Data *acpb.ApiClient
}

// NewClient returns a new client ref with given id and connector.
func NewClient(id string, conn Connector) *Client {
	return &Client{
		ID:   id,
		conn: conn,
	}
}

// File returns a file object with a given path on client's VFS.
func (c *Client) File(path string) *vfs.File {
	return vfs.NewFile(c.conn, c.ID, path)
}

// Get sends a request to GRR server to get information for a client and fills it in Data field.
// If an error occurs, Data field will stay unchanged.
func (c *Client) Get(ctx context.Context) error {
	data, err := c.conn.GetClient(ctx, &acpb.ApiGetClientArgs{
		ClientId: proto.String(c.ID),
	})

	if err != nil {
		return err
	}

	c.Data = data
	return nil
}

// Flow returns a new flow object with given flow ID and current client id and connector.
func (c *Client) Flow(id string) *flow.Flow {
	return flow.NewFlow(c.ID, id, c.conn)
}

// WithRunnerArgs is an optional argument that can be passed into
// CreateFlow to define the runner arguments of new flow. Example usage
// could look like this: CreateFlow(ctx, conn, WithRunnerArgs(&flowpb.FlowRunnerArgs{...})).
func WithRunnerArgs(flowRunnerArgs *flowpb.FlowRunnerArgs) func(args *afpb.ApiCreateFlowArgs) {
	return func(args *afpb.ApiCreateFlowArgs) {
		args.Flow.RunnerArgs = flowRunnerArgs
	}
}

// CreateFlow sends a request to GRR server to create a new flow with
// given name and arguments of the flow.
func (c *Client) CreateFlow(ctx context.Context, name string, flowArgs proto.Message, options ...func(args *afpb.ApiCreateFlowArgs)) (*flow.Flow, error) {
	args := &afpb.ApiCreateFlowArgs{
		ClientId: proto.String(c.ID),
		Flow: &afpb.ApiFlow{
			Name: proto.String(name),
		},
	}

	anyArgs, err := ptypes.MarshalAny(flowArgs)
	if err != nil {
		return nil, fmt.Errorf("cannot create a flow because ptypes.MarshalAny failed: %v", err)
	}
	args.Flow.Args = anyArgs

	for _, option := range options {
		option(args)
	}

	data, err := c.conn.CreateFlow(ctx, args)

	if err != nil {
		return nil, err
	}

	f := flow.NewFlow(c.ID, data.GetFlowId(), c.conn)
	f.Data = data
	return f, nil
}

// AddLabels sends a request to GRR server to add labels for the client.
func (c *Client) AddLabels(ctx context.Context, Labels []string) error {
	args := &acpb.ApiAddClientsLabelsArgs{
		ClientIds: []string{c.ID},
		Labels:    Labels,
	}

	return c.conn.AddClientsLabels(ctx, args)
}

// RemoveLabels sends a request to GRR server to remove labels from the client.
func (c *Client) RemoveLabels(ctx context.Context, Labels []string) error {
	args := &acpb.ApiRemoveClientsLabelsArgs{
		ClientIds: []string{c.ID},
		Labels:    Labels,
	}

	return c.conn.RemoveClientsLabels(ctx, args)
}

// Approval object is a helper object to send requests to GRR server via an arbitrary transfer protocol.
type Approval struct {
	conn     Connector
	ID       string
	ClientID string
	Username string
	Data     *aupb.ApiClientApproval
}

// Approval returns a Approval ref with given approval id, username and current client id.
func (c *Client) Approval(id string, username string) *Approval {
	return &Approval{
		conn:     c.conn,
		ID:       id,
		ClientID: c.ID,
		Username: username,
	}
}

// KeepClientAlive is an optional argument that can be passed into
// CreateApproval to keep client alive. Example usage could look
// like this: CreateApproval(ctx, conn, KeepClientAlive())
func KeepClientAlive() func(args *aupb.ApiCreateClientApprovalArgs) {
	return func(args *aupb.ApiCreateClientApprovalArgs) {
		args.KeepClientAlive = proto.Bool(true)
	}
}

// EmailCCAddresses is an optional argument that can be passed into
// CreateApproval to add email cc addresses for the operation. Example usage
// could look like this:
// CreateApproval(ctx, conn, EmailCCAddresses([]string{...})).
func EmailCCAddresses(emailAddresses []string) func(args *aupb.ApiCreateClientApprovalArgs) {
	return func(args *aupb.ApiCreateClientApprovalArgs) {
		args.Approval.EmailCcAddresses = emailAddresses
	}
}

// CreateApproval sends a request to GRR server to create a new client approval
// with given reason and notify users. KeepClientAlive is set to false by default.
func (c *Client) CreateApproval(ctx context.Context, reason string, notifiedUsers []string, options ...func(args *aupb.ApiCreateClientApprovalArgs)) (*Approval, error) {
	args := &aupb.ApiCreateClientApprovalArgs{
		ClientId: proto.String(c.ID),
		Approval: &aupb.ApiClientApproval{
			Reason:        proto.String(reason),
			NotifiedUsers: notifiedUsers,
		},
		KeepClientAlive: proto.Bool(false),
	}

	for _, option := range options {
		option(args)
	}

	data, err := c.conn.CreateClientApproval(ctx, args)
	if err != nil {
		return nil, err
	}

	username, err := c.conn.GetUsername(ctx)
	if err != nil {
		return nil, err
	}

	return &Approval{
		conn:     c.conn,
		ID:       data.GetId(),
		ClientID: c.ID,
		Username: username,
		Data:     data,
	}, nil
}

// Grant sends a request to GRR server to grand a client approval.
// If an error occurs, Data field will stay unchanged.
func (a *Approval) Grant(ctx context.Context) error {
	data, err := a.conn.GrantClientApproval(ctx, &aupb.ApiGrantClientApprovalArgs{
		ClientId:   proto.String(a.ClientID),
		ApprovalId: proto.String(a.ID),
		Username:   proto.String(a.Username),
	})

	if err != nil {
		return err
	}

	a.Data = data
	return nil
}

// Get sends a request to GRR server to get information for a client approval and fills it in Data field.
// If an error occurs, Data field will stay unchanged.
func (a *Approval) Get(ctx context.Context) error {
	data, err := a.conn.GetClientApproval(ctx, &aupb.ApiGetClientApprovalArgs{
		ClientId:   proto.String(a.ClientID),
		ApprovalId: proto.String(a.ID),
		Username:   proto.String(a.Username),
	})

	if err != nil {
		return err
	}

	a.Data = data
	return nil
}

// ListApprovals sends a request to GRR server to list all approvals of the client and calls fn (a callback function) on each Approval object.
func (c *Client) ListApprovals(ctx context.Context, state *aupb.ApiListClientApprovalsArgs_State, fn func(a *Approval) error) error {
	return c.conn.ListClientApprovals(ctx, &aupb.ApiListClientApprovalsArgs{
		ClientId: proto.String(c.ID),
		State:    state,
	},
		func(item *aupb.ApiClientApproval) error {
			username, err := c.conn.GetUsername(ctx)
			if err != nil {
				return err
			}

			return fn(&Approval{
				conn:     c.conn,
				ID:       item.GetId(),
				ClientID: c.ID,
				Username: username,
				Data:     item,
			})
		})
}

// ListFlows sends a request to GRR server to list flows on a given client,
// and calls fn (a callback function) on each Flow object.
func (c *Client) ListFlows(ctx context.Context, fn func(f *flow.Flow) error) error {
	return c.conn.ListFlows(ctx, &afpb.ApiListFlowsArgs{
		ClientId: proto.String(c.ID),
	},
		func(item *afpb.ApiFlow) error {
			return fn(flow.NewFlow(c.ID, item.GetFlowId(), c.conn))
		})
}

// SearchClientsOption is a function signature for avoiding repetitive typing.
type SearchClientsOption func(args *acpb.ApiSearchClientsArgs)

// WithMaxCount is an optional argument that can be passed into
// SearchClients to restrict the maximum number of clients returned from the
// response. Example usage could look like this:
// SearchClients(ctx, conn, WithMaxCount(1)).
func WithMaxCount(count int64) SearchClientsOption {
	return func(args *acpb.ApiSearchClientsArgs) {
		args.Count = proto.Int64(count)
	}
}

// SearchClients sends a request to GRR server to search for clients using
// given query and calls fn (a callback function) on each Client object.
// An optional argument can be passed (see function WithMaxCount).
func SearchClients(ctx context.Context, conn Connector, query string, fn func(c *Client) error, options ...SearchClientsOption) error {
	args := &acpb.ApiSearchClientsArgs{
		Query: proto.String(query),
	}

	for _, option := range options {
		option(args)
	}

	return conn.SearchClients(ctx, args,
		func(item *acpb.ApiClient) error {
			return fn(&Client{
				conn: conn,
				ID:   item.GetClientId(),
				Data: item,
			})
		})
}
