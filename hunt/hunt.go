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

// Package hunt provides GRR flow object which handles associated API calls via an arbitrary transfer protocol.
package hunt

import (
	"fmt"
	"io"
	"strings"

	"context"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	ahpb "github.com/google/grr_go_api_client/grrproto/grr/proto/api/hunt_proto"
	aupb "github.com/google/grr_go_api_client/grrproto/grr/proto/api/user_proto"
	flowpb "github.com/google/grr_go_api_client/grrproto/grr/proto/flows_proto"
)

// ConnectorSubset is introduced due to root object GRRAPI in api.go.
// GetUsername is an API method that will be used by multiple objects
// (e.g. Hunt, Client), so ConnectorSubset, rather than Connector defined below,
// is embedded to Connector in api.go, in order to avoid duplication of
// GetUsername of that interface.
type ConnectorSubset interface {
	// Lists GRR hunts.
	ListHunts(ctx context.Context, args *ahpb.ApiListHuntsArgs, fn func(m *ahpb.ApiHunt) error) (int64, error)
	// Lists all the results returned by a given hunt.
	ListHuntResults(ctx context.Context, args *ahpb.ApiListHuntResultsArgs, fn func(m *ahpb.ApiHuntResult) error) (int64, error)
	// Gets detailed information about a hunt with a given id.
	GetHunt(ctx context.Context, args *ahpb.ApiGetHuntArgs) (*ahpb.ApiHunt, error)
	// Modifies a hunt.
	ModifyHunt(ctx context.Context, args *ahpb.ApiModifyHuntArgs) (*ahpb.ApiHunt, error)
	// Streams ZIP/TAR.GZ archive with all the files downloaded during the hunt.
	GetHuntFilesArchive(ctx context.Context, args *ahpb.ApiGetHuntFilesArchiveArgs) (io.ReadCloser, error)
	// Returns information for a hunt approval with a given reason.
	GetHuntApproval(ctx context.Context, args *aupb.ApiGetHuntApprovalArgs) (*aupb.ApiHuntApproval, error)
	// Grants (adds caller to the approvers list) hunt approval.
	GrantHuntApproval(ctx context.Context, args *aupb.ApiGrantHuntApprovalArgs) (*aupb.ApiHuntApproval, error)
	// List hunt approvals of a current user.
	ListHuntApprovals(ctx context.Context, args *aupb.ApiListHuntApprovalsArgs, fn func(m *aupb.ApiHuntApproval) error) error
	// Create a new hunt.
	CreateHunt(ctx context.Context, args *ahpb.ApiCreateHuntArgs) (*ahpb.ApiHunt, error)
	// Create a new hunt approval.
	CreateHuntApproval(ctx context.Context, args *aupb.ApiCreateHuntApprovalArgs) (*aupb.ApiHuntApproval, error)
}

// Connector abstracts GRR communication protocols (i.e. HTTP, Stubby) and handles API calls related to hunts.
type Connector interface {
	ConnectorSubset
	// GetUsername returns username of the current user.
	GetUsername(ctx context.Context) (string, error)
}

// Hunt object is a helper object to send requests to GRR server via an arbitrary transfer protocol.
type Hunt struct {
	// conn is an implementation detail that should be hidden from users. Users do not need to know which protocol (HTTP or Stubby) used to talk to GRR server.
	conn Connector

	ID string

	// Proto with hunt information returned from GRR server.
	Data *ahpb.ApiHunt
}

// NewHunt returns a new hunt with given huntID and connector.
func NewHunt(huntID string, conn Connector, data *ahpb.ApiHunt) *Hunt {
	return &Hunt{
		conn: conn,
		ID:   huntID,
		Data: data,
	}
}

// CreateHuntOption is a function signature for avoiding repetitive typing.
type CreateHuntOption func(args *ahpb.ApiCreateHuntArgs)

// CreateHuntRunnerArgs is an optional argument that can be passed into
// CreateHunt to define the runner arguments of new hunt. Example usage
// could look like this: CreateHunt(ctx, conn, CreateHuntRunnerArgs(&flowpb.HuntRunnerArgs{...})).
func CreateHuntRunnerArgs(huntRunnerArgs *flowpb.HuntRunnerArgs) CreateHuntOption {
	return func(args *ahpb.ApiCreateHuntArgs) {
		args.HuntRunnerArgs = huntRunnerArgs
	}
}

// CreateHunt sends a request to GRR server to create a new hunt with
// given flow name and flow args.
func CreateHunt(ctx context.Context, conn Connector, flowName string, flowArgs proto.Message, options ...CreateHuntOption) (*Hunt, error) {
	args := &ahpb.ApiCreateHuntArgs{
		FlowName: proto.String(flowName),
	}

	anyArgs, err := ptypes.MarshalAny(flowArgs)
	if err != nil {
		return nil, fmt.Errorf("cannot create a hunt because ptypes.MarshalAny failed: %v", err)
	}
	args.FlowArgs = anyArgs

	for _, option := range options {
		option(args)
	}

	data, err := conn.CreateHunt(ctx, args)

	if err != nil {
		return nil, err
	}

	h := &Hunt{
		conn: conn,
		ID:   strings.TrimPrefix(data.GetUrn(), "aff4:/hunts/"),
		Data: data,
	}
	return h, nil
}

// Get sends a request to GRR server to get information for a hunt and fills it in Data field. If an error occurs, Data field will stay unchanged.
func (h *Hunt) Get(ctx context.Context) error {
	data, err := h.conn.GetHunt(ctx, &ahpb.ApiGetHuntArgs{
		HuntId: proto.String(h.ID),
	})

	if err != nil {
		return err
	}

	h.Data = data
	return nil
}

// Result wraps ApiHuntResult data.
type Result struct {
	Data *ahpb.ApiHuntResult
}

// Payload returns the payload of a hunt Result.
func (r *Result) Payload() (proto.Message, error) {
	var x ptypes.DynamicAny
	if err := ptypes.UnmarshalAny(r.Data.GetPayload(), &x); err != nil {
		return nil, err
	}
	return x.Message, nil
}

// ListHunts sends a request to GRR server to list all GRR hunts and calls fn (a callback function) on each Hunt object. Total count of first request is returned.
func ListHunts(ctx context.Context, conn Connector, fn func(h *Hunt) error) (int64, error) {
	return conn.ListHunts(ctx, &ahpb.ApiListHuntsArgs{},
		func(item *ahpb.ApiHunt) error {
			return fn(&Hunt{
				conn: conn,
				ID:   strings.TrimPrefix(item.GetUrn(), "aff4:/hunts/"),
				Data: item,
			})
		})
}

// ListResults sends a request to GRR server to list all the results returned by a given hunt and calls fn (a callback function) on each Result object. Total count of first request is returned.
func (h *Hunt) ListResults(ctx context.Context, fn func(r *Result) error) (int64, error) {
	return h.conn.ListHuntResults(ctx, &ahpb.ApiListHuntResultsArgs{
		HuntId: proto.String(h.ID),
	},
		func(item *ahpb.ApiHuntResult) error {
			return fn(&Result{
				Data: item,
			})
		})
}

// Start sends a request to GRR server to start a hunt. If an error occurs, Data field will stay unchanged.
func (h *Hunt) Start(ctx context.Context) error {
	data, err := h.conn.ModifyHunt(ctx, &ahpb.ApiModifyHuntArgs{
		HuntId: proto.String(h.ID),
		State:  ahpb.ApiHunt_STARTED.Enum(),
	})

	if err != nil {
		return err
	}

	h.Data = data
	return nil
}

// Stop sends a request to GRR server to stop a hunt. If an error occurs, Data field will stay unchanged.
func (h *Hunt) Stop(ctx context.Context) error {
	data, err := h.conn.ModifyHunt(ctx, &ahpb.ApiModifyHuntArgs{
		HuntId: proto.String(h.ID),
		State:  ahpb.ApiHunt_STOPPED.Enum(),
	})

	if err != nil {
		return err
	}

	h.Data = data
	return nil
}

// ModifyClientLimit is an optional argument that can be passed into ModifyHunt to change the client limit of the request. Example usage could look like this: h.ModifyHunt(ctx, hunt.ModifyClientLimit(10)).
func ModifyClientLimit(limit int64) func(args *ahpb.ApiModifyHuntArgs) {
	return func(args *ahpb.ApiModifyHuntArgs) {
		args.ClientLimit = proto.Int64(limit)
	}
}

// ModifyClientRate is an optional argument that can be passed into ModifyHunt to change the client rate of the request. Example usage could look like this: h.ModifyHunt(ctx, hunt.ModifyClientRate(10)).
func ModifyClientRate(rate int64) func(args *ahpb.ApiModifyHuntArgs) {
	return func(args *ahpb.ApiModifyHuntArgs) {
		args.ClientRate = proto.Int64(rate)
	}
}

// ModifyExpires is an optional argument that can be passed into ModifyHunt to change expires of the request. Example usage could look like this: h.ModifyHunt(ctx, hunt.ModifyExpires(10)).
func ModifyExpires(expires uint64) func(args *ahpb.ApiModifyHuntArgs) {
	return func(args *ahpb.ApiModifyHuntArgs) {
		args.Expires = proto.Uint64(expires)
	}
}

// Modify sends a request to GRR server to modify a hunt. If an error occurs, Data field will stay unchanged.
func (h *Hunt) Modify(ctx context.Context, options ...func(args *ahpb.ApiModifyHuntArgs)) error {
	args := &ahpb.ApiModifyHuntArgs{
		HuntId: proto.String(h.ID),
	}

	for _, option := range options {
		option(args)
	}

	data, err := h.conn.ModifyHunt(ctx, args)

	if err != nil {
		return err
	}

	h.Data = data
	return nil
}

// GetFilesArchive sends a request to GRR server to get ZIP or TAR.GZ archive with all the files downloaded by the hunt. This function starts a goroutine that runs until all data is read, or the readcloser is closed, or an error occurs in the go routine.
func (h *Hunt) GetFilesArchive(ctx context.Context, format *ahpb.ApiGetHuntFilesArchiveArgs_ArchiveFormat) (io.ReadCloser, error) {
	return h.conn.GetHuntFilesArchive(ctx, &ahpb.ApiGetHuntFilesArchiveArgs{
		HuntId:        proto.String(h.ID),
		ArchiveFormat: format,
	})
}

// Approval wraps ApiHuntApproval data.
type Approval struct {
	conn     Connector
	ID       string
	HuntID   string
	Username string
	Data     *aupb.ApiHuntApproval
}

// Approval returns a Approval ref with given approval id, username and current hunt id.
func (h *Hunt) Approval(id string, username string) *Approval {
	return &Approval{
		conn:     h.conn,
		ID:       id,
		HuntID:   h.ID,
		Username: username,
	}
}

// EmailCCAddresses is an optional argument that can be passed into
// CreateApproval to add email cc addresses for the operation. Example usage
// could look like this:
// CreateApproval(ctx, conn, EmailCCAddresses([]string{...})).
func EmailCCAddresses(emailAddresses []string) func(args *aupb.ApiCreateHuntApprovalArgs) {
	return func(args *aupb.ApiCreateHuntApprovalArgs) {
		args.Approval.EmailCcAddresses = emailAddresses
	}
}

// CreateApproval sends a request to GRR server to create a new hunt approval
// with given reason and notify users.
func (h *Hunt) CreateApproval(ctx context.Context, reason string, notifiedUsers []string, options ...func(args *aupb.ApiCreateHuntApprovalArgs)) (*Approval, error) {
	args := &aupb.ApiCreateHuntApprovalArgs{
		HuntId: proto.String(h.ID),
		Approval: &aupb.ApiHuntApproval{
			Reason:        proto.String(reason),
			NotifiedUsers: notifiedUsers,
		},
	}

	for _, option := range options {
		option(args)
	}

	data, err := h.conn.CreateHuntApproval(ctx, args)
	if err != nil {
		return nil, err
	}

	username, err := h.conn.GetUsername(ctx)
	if err != nil {
		return nil, err
	}

	return &Approval{
		conn:     h.conn,
		ID:       data.GetId(),
		HuntID:   h.ID,
		Username: username,
		Data:     data,
	}, nil
}

// Get sends a request to GRR server to get information for a hunt approval and fills it in Data field. If an error occurs, Data field will stay unchanged.
func (a *Approval) Get(ctx context.Context) error {
	data, err := a.conn.GetHuntApproval(ctx, &aupb.ApiGetHuntApprovalArgs{
		HuntId:     proto.String(a.HuntID),
		ApprovalId: proto.String(a.ID),
		Username:   proto.String(a.Username),
	})

	if err != nil {
		return err
	}

	a.Data = data
	return nil
}

// Grant sends a request to GRR server to grand a hunt approval. If an error occurs, Data field will stay unchanged.
func (a *Approval) Grant(ctx context.Context) error {
	data, err := a.conn.GrantHuntApproval(ctx, &aupb.ApiGrantHuntApprovalArgs{
		HuntId:     proto.String(a.HuntID),
		ApprovalId: proto.String(a.ID),
		Username:   proto.String(a.Username),
	})

	if err != nil {
		return err
	}

	a.Data = data
	return nil
}

// ListApprovals sends a request to GRR server to list all the hunt approvals belonging to requesting user and calls fn (a callback function) on each Approval object.
func ListApprovals(ctx context.Context, conn Connector, fn func(a *Approval) error) error {
	return conn.ListHuntApprovals(ctx, &aupb.ApiListHuntApprovalsArgs{},
		func(item *aupb.ApiHuntApproval) error {
			username, err := conn.GetUsername(ctx)
			if err != nil {
				return err
			}

			var huntID string
			hunt := item.GetSubject()
			if hunt != nil {
				huntID = strings.TrimPrefix(hunt.GetUrn(), "aff4:/hunts/")
			}

			return fn(&Approval{
				conn:     conn,
				ID:       item.GetId(),
				HuntID:   huntID,
				Username: username,
				Data:     item,
			})
		})
}
