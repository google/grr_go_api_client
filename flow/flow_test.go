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

package flow

import (
	"fmt"
	"io"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"
	"time"

	"context"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	afpb "github.com/google/grr_go_api_client/grrproto/grr/proto/api/flow_proto"
	flowpb "github.com/google/grr_go_api_client/grrproto/grr/proto/flows_proto"
)

type call struct {
	handlerName string
	args        interface{}
}

const (
	testClientID     = "clientTest1"
	testFlowID       = "FlowTest1"
	testStatusString = "testStatusString"
)

var (
	testReadCloser = ioutil.NopCloser(strings.NewReader("reader"))
)

// mockConnector simply records information of all calls made via the connector.
type mockConnector struct {
	calls []call
	Connector
}

func createTestFlow() (*Flow, *mockConnector) {
	conn := &mockConnector{}

	return NewFlow(testClientID, testFlowID, conn), conn
}

func (c *mockConnector) GetFlowFilesArchive(ctx context.Context, args *afpb.ApiGetFlowFilesArchiveArgs) (io.ReadCloser, error) {
	c.calls = append(c.calls, call{
		handlerName: "GetFlowFilesArchive",
		args:        args,
	})
	return testReadCloser, nil
}

func (c *mockConnector) ListFlowResults(ctx context.Context, args *afpb.ApiListFlowResultsArgs, fn func(m *afpb.ApiFlowResult) error) (int64, error) {
	argsAny, err := ptypes.MarshalAny(args)
	if err != nil {
		return -1, err
	}

	fn(&afpb.ApiFlowResult{
		Payload: argsAny,
	})

	return 1, nil
}

func (c *mockConnector) CancelFlow(ctx context.Context, args *afpb.ApiCancelFlowArgs) error {
	c.calls = append(c.calls, call{
		handlerName: "CancelFlow",
		args:        args,
	})
	return nil
}

func (c *mockConnector) GetFlow(ctx context.Context, args *afpb.ApiGetFlowArgs) (*afpb.ApiFlow, error) {
	c.calls = append(c.calls, call{
		handlerName: "GetFlow",
		args:        args,
	})
	return &afpb.ApiFlow{
		State: afpb.ApiFlow_RUNNING.Enum(),
		Context: &flowpb.FlowContext{
			Status: proto.String(testStatusString),
		},
	}, nil
}

func TestGetFlowFilesArchive(t *testing.T) {
	flow, conn := createTestFlow()

	format := afpb.ApiGetFlowFilesArchiveArgs_ZIP
	rc, err := flow.GetFilesArchive(context.Background(), &format)
	if err != nil {
		t.Fatalf("GetFilesArchive for flows failed: %v", err)
	}

	if rc != testReadCloser {
		t.Errorf("The readCloser returned from GetFlowFilesArchive is %v; want: %v", rc, testReadCloser)
	}

	want := []call{
		call{
			handlerName: "GetFlowFilesArchive",
			args: &afpb.ApiGetFlowFilesArchiveArgs{
				ClientId:      proto.String(testClientID),
				FlowId:        proto.String(testFlowID),
				ArchiveFormat: &format,
			},
		},
	}

	if len(conn.calls) != len(want) ||
		conn.calls[0].handlerName != want[0].handlerName ||
		!proto.Equal(conn.calls[0].args.(proto.Message), want[0].args.(proto.Message)) {
		t.Errorf("The arguments of GetFlowFilesArchive diff (want: [%q] got: [%q])", want, conn.calls)
	}
}

func TestListFlowResults(t *testing.T) {
	flow, _ := createTestFlow()

	var items []*Result
	totalCount, err := flow.ListResults(context.Background(),
		func(r *Result) error {
			items = append(items, r)
			return nil
		})
	if err != nil {
		t.Fatal(err)
	}
	if totalCount != 1 {
		t.Errorf("the total count of results of ListResults is %v; want: 1", totalCount)
	}

	if len(items) == 0 {
		t.Fatalf("items should not be empty; should have an item")
	} else if len(items) != 1 {
		t.Errorf("items has a length %v; want 1", len(items))
	}

	got, err := items[0].Payload()
	if err != nil {
		t.Fatalf("payload fails: %v", err)
	}

	want := &afpb.ApiListFlowResultsArgs{
		ClientId: proto.String(flow.ClientID),
		FlowId:   proto.String(flow.ID),
	}

	if w, g := want, got; !proto.Equal(w, g) {
		t.Errorf("unexpected payload received; diff (want: [%q] got: [%q])", w, g)
	}
}

func TestCancelFlow(t *testing.T) {
	flow, conn := createTestFlow()

	if err := flow.Cancel(context.Background()); err != nil {
		t.Fatalf("cancel for a flow failed: %v", err)
	}

	want := []call{
		call{
			handlerName: "CancelFlow",
			args: &afpb.ApiCancelFlowArgs{
				ClientId: proto.String(flow.ClientID),
				FlowId:   proto.String(flow.ID),
			},
		},
	}

	if len(conn.calls) != len(want) ||
		conn.calls[0].handlerName != want[0].handlerName ||
		!proto.Equal(conn.calls[0].args.(proto.Message), want[0].args.(proto.Message)) {
		t.Errorf("the arguments of CancelFlow diff (want: [%q] got: [%q])", want, conn.calls)
	}
}

func TestGetFlow(t *testing.T) {
	flow, conn := createTestFlow()

	if err := flow.Get(context.Background()); err != nil {
		t.Fatalf("get method for a flow failed: %v", err)
	}

	// Make sure other fields of the flow object stay unchanged.
	want := &Flow{
		conn:     conn,
		ID:       testFlowID,
		ClientID: testClientID,
		Data: &afpb.ApiFlow{
			State: afpb.ApiFlow_RUNNING.Enum(),
			Context: &flowpb.FlowContext{
				Status: proto.String(testStatusString),
			},
		},
	}

	if w, g := want, flow; g.ID != w.ID ||
		g.ClientID != w.ClientID ||
		!proto.Equal(g.Data, w.Data) {
		t.Errorf("unexpected flow object got; diff (want: [%q] got: [%q])", w, g)
	}

	wantCalls := []call{
		call{
			handlerName: "GetFlow",
			args: &afpb.ApiGetFlowArgs{
				ClientId: proto.String(testClientID),
				FlowId:   proto.String(testFlowID),
			},
		},
	}

	if len(conn.calls) != len(wantCalls) ||
		conn.calls[0].handlerName != wantCalls[0].handlerName ||
		!proto.Equal(conn.calls[0].args.(proto.Message), wantCalls[0].args.(proto.Message)) {
		t.Errorf("the arguments of GetFlow diff (want: [%q] got: [%q])", wantCalls, conn.calls)
	}
}

func TestGetFlowStatus(t *testing.T) {
	flow, _ := createTestFlow()

	fs, err := flow.getState(context.Background())
	if err != nil {
		t.Fatalf("cannot get the status of a flow: %v", err)
	}

	want := &State{
		Name:       afpb.ApiFlow_RUNNING,
		StatusText: testStatusString,
	}

	if w, g := want, fs; !reflect.DeepEqual(w, g) {
		t.Errorf("unexpected flow object got; diff (want: [%q] got: [%q])", w, g)
	}

	if !fs.IsRunning() {
		t.Error("flow should be running, but it isn't")
	}

	if fs.IsTerminated() {
		t.Error("flow should not be terminated, but it is")
	}

	if fs.HasFailed() {
		t.Error("flow should not fail, but it has failed")
	}
}

type mockWaitFlowConnector struct {
	Connector
	state *afpb.ApiFlow_State
}

func createFlowAndMockWaitFlowConnector() (*Flow, *mockWaitFlowConnector) {
	conn := &mockWaitFlowConnector{
		state: afpb.ApiFlow_RUNNING.Enum(),
	}
	return NewFlow(testClientID, testFlowID, conn), conn
}

func (c *mockWaitFlowConnector) setState(state *afpb.ApiFlow_State) {
	c.state = state
}

func (c *mockWaitFlowConnector) GetFlow(_ context.Context, _ *afpb.ApiGetFlowArgs) (*afpb.ApiFlow, error) {
	return &afpb.ApiFlow{
		State: c.state,
		Context: &flowpb.FlowContext{
			Status: proto.String(testStatusString),
		},
	}, nil
}

func TestWaitFlow(t *testing.T) {
	flow, conn := createFlowAndMockWaitFlowConnector()

	tests := []struct {
		desc       string
		state      *afpb.ApiFlow_State
		err        error
		terminated bool
		failed     bool
	}{
		{"a flow is still running after maxWait (nanoseconds) have passed", afpb.ApiFlow_RUNNING.Enum(), fmt.Errorf("flow (id: %s) on client (id: %s) is still running", testFlowID, testClientID), false, false},
		{"a flow has terminated after maxWait (nanoseconds) have passed", afpb.ApiFlow_TERMINATED.Enum(), nil, true, false},
		{"a flow has raised an error after maxWait (nanoseconds) have passed", afpb.ApiFlow_ERROR.Enum(), nil, false, true},
		{"a flow has raised an error after maxWait (nanoseconds) have passed", afpb.ApiFlow_CLIENT_CRASHED.Enum(), nil, false, true},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("test %s:", test.desc), func(t *testing.T) {
			conn.setState(test.state)
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Microsecond)
			defer cancel()
			fs, err := flow.Wait(ctx, PollPeriod(time.Microsecond))

			if test.err == nil && err != nil {
				t.Fatalf("WaitFlow has an error %v, but it should not produce an error", err)
			} else if test.err != nil && err == nil {
				t.Fatalf("WaitFlow should have an error %v, but it doesn't", test.err)
			} else if test.err != nil && err != nil {
				if !strings.Contains(err.Error(), test.err.Error()) {
					t.Fatalf("WaitFlow has an error %v, but it should be %v", err, test.err)
				}
			}

			if err != nil {
				return
			}

			if terminated := fs.IsTerminated(); terminated != test.terminated {
				t.Errorf("result of IsTerminated of the FlowState is %v; want %v", terminated, test.terminated)
			}

			if failed := fs.HasFailed(); failed != test.failed {
				t.Errorf("result of HasFailed of the FlowState is %v; want %v", failed, test.failed)
			}
		})
	}
}
