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

// Package flow provides GRR flow object which handles associated API calls via an arbitrary transfer protocol.
package flow

import (
	"fmt"
	"io"
	"time"

	"context"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	afpb "github.com/google/grr_go_api_client/grrproto/grr/proto/api/flow_proto"
)

// Connector abstracts GRR communication protocols (i.e. HTTP, Stubby).
type Connector interface {
	GetFlowFilesArchive(ctx context.Context, args *afpb.ApiGetFlowFilesArchiveArgs) (io.ReadCloser, error)
	ListFlowResults(ctx context.Context, args *afpb.ApiListFlowResultsArgs, fn func(m *afpb.ApiFlowResult) error) (int64, error)
	CancelFlow(ctx context.Context, args *afpb.ApiCancelFlowArgs) error
	GetFlow(ctx context.Context, args *afpb.ApiGetFlowArgs) (*afpb.ApiFlow, error)
}

// Flow object is a helper object to send requests to GRR server via an arbitrary transfer protocol.
type Flow struct {
	conn     Connector
	ID       string
	ClientID string
	Data     *afpb.ApiFlow
}

// NewFlow returns a new flow with given clientID, flow id and connector.
func NewFlow(clientID string, id string, conn Connector) *Flow {
	return &Flow{
		conn:     conn,
		ID:       id,
		ClientID: clientID,
	}
}

// GetFilesArchive sends a request to GRR server to get files archive of the flow.
func (f *Flow) GetFilesArchive(ctx context.Context, format *afpb.ApiGetFlowFilesArchiveArgs_ArchiveFormat) (io.ReadCloser, error) {
	args := &afpb.ApiGetFlowFilesArchiveArgs{
		ClientId:      proto.String(f.ClientID),
		FlowId:        proto.String(f.ID),
		ArchiveFormat: format,
	}

	return f.conn.GetFlowFilesArchive(ctx, args)
}

// Result wraps ApiFlowResult data.
type Result struct {
	Data *afpb.ApiFlowResult
}

// Payload returns the payload of a flow Result.
func (r *Result) Payload() (proto.Message, error) {
	var x ptypes.DynamicAny
	if err := ptypes.UnmarshalAny(r.Data.GetPayload(), &x); err != nil {
		return nil, err
	}
	return x.Message, nil
}

// ListResults sends a request to GRR server to list results of the flow.
func (f *Flow) ListResults(ctx context.Context, fn func(r *Result) error) (int64, error) {
	return f.conn.ListFlowResults(ctx, &afpb.ApiListFlowResultsArgs{
		ClientId: proto.String(f.ClientID),
		FlowId:   proto.String(f.ID),
	},
		func(item *afpb.ApiFlowResult) error {
			return fn(&Result{Data: item})
		})
}

// Cancel sends a request to GRR server to cancel the flow.
func (f *Flow) Cancel(ctx context.Context) error {
	args := &afpb.ApiCancelFlowArgs{
		ClientId: proto.String(f.ClientID),
		FlowId:   proto.String(f.ID),
	}

	return f.conn.CancelFlow(ctx, args)
}

// Get sends a request to GRR server to get information for a flow and fills it in Data field.
// If an error occurs, Data field will stay unchanged.
func (f *Flow) Get(ctx context.Context) error {
	data, err := f.conn.GetFlow(ctx, &afpb.ApiGetFlowArgs{
		ClientId: proto.String(f.ClientID),
		FlowId:   proto.String(f.ID),
	})

	if err != nil {
		return err
	}

	f.Data = data
	return nil
}

// State contains a flow state (e.g. "RUNNING") and the completion status string for humans.
type State struct {
	// an enum presenting the status of a flow.
	Name afpb.ApiFlow_State
	// a user-friendly status message.
	StatusText string
}

// IsRunning returns true if the flow is still running.
func (fs *State) IsRunning() bool {
	return fs.Name == afpb.ApiFlow_RUNNING
}

// IsTerminated returns true if the flow has been completed and the actually performed on the agent.
func (fs *State) IsTerminated() bool {
	return fs.Name == afpb.ApiFlow_TERMINATED
}

// HasFailed returns true if the flow somehow failed with a agent or server error.
func (fs *State) HasFailed() bool {
	return !fs.IsRunning() && !fs.IsTerminated()
}

// getState sends a request to GRR server to get information for a flow and fills it in Data field.
// If an error occurs, Data field will stay unchanged; otherwise, the function will return a State object.
func (f *Flow) getState(ctx context.Context) (*State, error) {
	if err := f.Get(ctx); err != nil {
		return nil, err
	}

	return &State{
		Name:       f.Data.GetState(),
		StatusText: f.Data.GetContext().GetStatus(),
	}, nil
}

// PollPeriod is an optional argument that can be passed into Wait to set pollPeriod (time.Duration).
// Example usage could look like this: f.Wait(ctx, flow.PollPeriod(time.Duration(40)*time.Second)).
func PollPeriod(pollPeriod time.Duration) func(pollPeriodPtr *time.Duration) {
	return func(pollPeriodPtr *time.Duration) {
		*pollPeriodPtr = pollPeriod
	}
}

const (
	// defaultPollPeriod is how often we ask GRR for the status of the flow.
	defaultPollPeriod = 30 * time.Second
)

// Wait periodically pulls flow status from the server and returns when it's
// completed or context expires. It's users' responsibility to specify the
// WithDeadline/WithTimeout func for context. Example usage could look like this:
// ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Minute)
// defer cancel()
// fs, err := flow.Wait(ctx)
func (f *Flow) Wait(ctx context.Context, options ...func(pollPeriodPtr *time.Duration)) (*State, error) {
	pollPeriod := defaultPollPeriod

	for _, option := range options {
		option(&pollPeriod)
	}

	ticker := time.NewTicker(pollPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fs, err := f.getState(ctx)
			if err != nil {
				return nil, err
			}
			if !fs.IsRunning() {
				return fs, nil
			}
		case <-ctx.Done():
			return nil, fmt.Errorf("flow (id: %s) on client (id: %s) is still running: %v", f.ID, f.ClientID, ctx.Err())
		}
	}
}
