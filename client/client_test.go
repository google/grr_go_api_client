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

package client

import (
	"reflect"
	"testing"

	"context"
	"github.com/golang/protobuf/proto"
	"github.com/google/grr_go_api_client/flow"
	acpb "github.com/google/grr_go_api_client/grrproto/grr/proto/api/client_proto"
	afpb "github.com/google/grr_go_api_client/grrproto/grr/proto/api/flow_proto"
	aupb "github.com/google/grr_go_api_client/grrproto/grr/proto/api/user_proto"
	flowpb "github.com/google/grr_go_api_client/grrproto/grr/proto/flows_proto"

	"github.com/golang/protobuf/ptypes"
)

type call struct {
	handlerName string
	args        interface{}
}

const (
	testClientID   = "clientTest1"
	testApprovalID = "approvalTest1"
	testUsername   = "usernameTest1"
	testFlowID     = "flowTest1"
	testFlowName   = "flow"
	testHuntName   = "hunt"
	testFilePath   = "filePath"
)

// fakeConnector simply records information of all calls made via the connector.
type fakeConnector struct {
	calls []call
	Connector
}

func (c *fakeConnector) AddClientsLabels(ctx context.Context, args *acpb.ApiAddClientsLabelsArgs) error {
	c.calls = append(c.calls, call{
		handlerName: "AddClientsLabels",
		args:        args,
	})
	return nil
}

func (c *fakeConnector) RemoveClientsLabels(ctx context.Context, args *acpb.ApiRemoveClientsLabelsArgs) error {
	c.calls = append(c.calls, call{
		handlerName: "RemoveClientsLabels",
		args:        args,
	})
	return nil
}

func (c *fakeConnector) GetClientApproval(ctx context.Context, args *aupb.ApiGetClientApprovalArgs) (*aupb.ApiClientApproval, error) {
	c.calls = append(c.calls, call{
		handlerName: "GetClientApproval",
		args:        args,
	})

	return &aupb.ApiClientApproval{
		Id: proto.String(testApprovalID),
	}, nil
}

func (c *fakeConnector) GetUsername(ctx context.Context) (string, error) {
	return testUsername, nil
}

func (c *fakeConnector) ListClientApprovals(ctx context.Context, args *aupb.ApiListClientApprovalsArgs, fn func(m *aupb.ApiClientApproval) error) error {
	c.calls = append(c.calls, call{
		handlerName: "ListClientApprovals",
		args:        args,
	})

	fn(&aupb.ApiClientApproval{
		Id: proto.String(testApprovalID),
	})

	return nil
}

func (c *fakeConnector) CreateFlow(ctx context.Context, args *afpb.ApiCreateFlowArgs) (*afpb.ApiFlow, error) {
	c.calls = append(c.calls, call{
		handlerName: "CreateFlow",
		args:        args,
	})

	return &afpb.ApiFlow{
		FlowId: proto.String(testFlowID),
		Name:   proto.String(testFlowName),
		RunnerArgs: &flowpb.FlowRunnerArgs{
			ClientId: proto.String(testClientID),
		},
	}, nil
}

func (c *fakeConnector) GetClient(ctx context.Context, args *acpb.ApiGetClientArgs) (*acpb.ApiClient, error) {
	c.calls = append(c.calls, call{
		handlerName: "GetClient",
		args:        args,
	})

	return &acpb.ApiClient{
		ClientId: proto.String(testClientID),
	}, nil
}

func (c *fakeConnector) SearchClients(ctx context.Context, args *acpb.ApiSearchClientsArgs, fn func(m *acpb.ApiClient) error) error {
	c.calls = append(c.calls, call{
		handlerName: "SearchClients",
		args:        args,
	})

	items := []*acpb.ApiClient{
		{ClientId: proto.String(testClientID)},
		{ClientId: proto.String(testClientID)},
	}

	if args.Count != nil && *args.Count <= 1 {
		items = items[:1]
	}

	for _, item := range items {
		fn(item)
	}

	return nil
}

func (c *fakeConnector) ListFlows(ctx context.Context, args *afpb.ApiListFlowsArgs, fn func(m *afpb.ApiFlow) error) error {
	c.calls = append(c.calls, call{
		handlerName: "ListFlows",
		args:        args,
	})

	fn(&afpb.ApiFlow{
		FlowId: proto.String(testFlowID),
	})

	return nil
}

func (c *fakeConnector) GrantClientApproval(ctx context.Context, args *aupb.ApiGrantClientApprovalArgs) (*aupb.ApiClientApproval, error) {
	c.calls = append(c.calls, call{
		handlerName: "GrantClientApproval",
		args:        args,
	})

	return &aupb.ApiClientApproval{
		Id: proto.String(testApprovalID),
	}, nil
}

func (c *fakeConnector) CreateClientApproval(ctx context.Context, args *aupb.ApiCreateClientApprovalArgs) (*aupb.ApiClientApproval, error) {
	c.calls = append(c.calls, call{
		handlerName: "CreateClientApproval",
		args:        args,
	})

	return &aupb.ApiClientApproval{
		Id: proto.String(testApprovalID),
	}, nil
}

func createTestClient() (*Client, *fakeConnector) {
	conn := &fakeConnector{}
	return NewClient(testClientID, conn), conn
}

func TestAddLabels(t *testing.T) {
	client, conn := createTestClient()
	testLabels := []string{"foo", "bar"}

	if err := client.AddLabels(context.Background(), testLabels); err != nil {
		t.Fatalf("AddLabels for clients failed: %v", err)
	}

	want := []call{
		call{
			handlerName: "AddClientsLabels",
			args: &acpb.ApiAddClientsLabelsArgs{
				ClientIds: []string{testClientID},
				Labels:    testLabels,
			},
		},
	}

	if len(conn.calls) != len(want) ||
		conn.calls[0].handlerName != want[0].handlerName ||
		!proto.Equal(conn.calls[0].args.(proto.Message), want[0].args.(proto.Message)) {
		t.Errorf("The arguments of AddClientsLabels are %v; want: %v", conn.calls, want)
	}
}

func TestRemoveLabels(t *testing.T) {
	client, conn := createTestClient()
	testLabels := []string{"foo", "bar"}

	if err := client.RemoveLabels(context.Background(), testLabels); err != nil {
		t.Fatalf("RemoveLabels for clients failed: %v", err)
	}

	want := []call{
		call{
			handlerName: "RemoveClientsLabels",
			args: &acpb.ApiRemoveClientsLabelsArgs{
				ClientIds: []string{testClientID},
				Labels:    testLabels,
			},
		},
	}

	if len(conn.calls) != len(want) ||
		conn.calls[0].handlerName != want[0].handlerName ||
		!proto.Equal(conn.calls[0].args.(proto.Message), want[0].args.(proto.Message)) {
		t.Errorf("The arguments of RemoveClientsLabels are %v; want: %v", conn.calls, want)
	}
}

func TestGetApproval(t *testing.T) {
	client, conn := createTestClient()

	approval := client.Approval(testApprovalID, testUsername)

	err := approval.Get(context.Background())
	if err != nil {
		t.Fatalf("get method for Approval fails: %v", err)
	}

	want := &aupb.ApiClientApproval{
		Id: proto.String(testApprovalID),
	}

	if w, g := want, approval.Data; !proto.Equal(w, g) {
		t.Errorf("The results of GetClientApproval diff (want: [%q] got: [%q])", w, g)
	}

	wantCalls := []call{{
		handlerName: "GetClientApproval",
		args: &aupb.ApiGetClientApprovalArgs{
			ClientId:   proto.String(client.ID),
			ApprovalId: proto.String(testApprovalID),
			Username:   proto.String(testUsername),
		},
	}}

	if len(conn.calls) != len(wantCalls) ||
		conn.calls[0].handlerName != wantCalls[0].handlerName ||
		!proto.Equal(conn.calls[0].args.(proto.Message), wantCalls[0].args.(proto.Message)) {
		t.Errorf("The arguments of GetClientApproval diff (want: [%q] got: [%q])", wantCalls, conn.calls)
	}
}

func TestListApprovals(t *testing.T) {
	client, conn := createTestClient()

	state := aupb.ApiListClientApprovalsArgs_ANY
	var gotItems []*Approval
	err := client.ListApprovals(context.Background(), &state,
		func(a *Approval) error {
			gotItems = append(gotItems, a)
			return nil
		})
	if err != nil {
		t.Fatalf("ListApprovals for clients failed: %v", err)
	}

	wantItems := []*Approval{{
		conn:     conn,
		ID:       testApprovalID,
		ClientID: testClientID,
		Username: testUsername,
		Data: &aupb.ApiClientApproval{
			Id: proto.String(testApprovalID),
		},
	}}

	if w, g := wantItems, gotItems; len(g) != len(w) ||
		g[0].conn != w[0].conn ||
		g[0].ID != w[0].ID ||
		g[0].ClientID != w[0].ClientID ||
		g[0].Username != w[0].Username ||
		!proto.Equal(g[0].Data, w[0].Data) {
		t.Errorf("The results of ListClientApprovals diff (want: [%q] got: [%q])", w, g)
	}

	wantCalls := []call{
		call{
			handlerName: "ListClientApprovals",
			args: &aupb.ApiListClientApprovalsArgs{
				ClientId: proto.String(testClientID),
				State:    &state,
			},
		},
	}

	if len(conn.calls) != len(wantCalls) ||
		conn.calls[0].handlerName != wantCalls[0].handlerName ||
		!proto.Equal(conn.calls[0].args.(proto.Message), wantCalls[0].args.(proto.Message)) {
		t.Errorf("The arguments of ListClientApprovals diff (want: [%q] got: [%q])", wantCalls, conn.calls)
	}
}

func TestListFlows(t *testing.T) {
	client, conn := createTestClient()

	var gotItems []*flow.Flow
	err := client.ListFlows(context.Background(), func(f *flow.Flow) error {
		gotItems = append(gotItems, f)
		return nil
	})
	if err != nil {
		t.Fatalf("ListFlows for clients failed: %v", err)
	}

	wantItems := []*flow.Flow{flow.NewFlow(testClientID, testFlowID, conn)}

	if w, g := wantItems, gotItems; len(g) != len(w) ||
		g[0].Data != nil || w[0].Data != nil ||
		!reflect.DeepEqual(w, g) {
		t.Errorf("The results of ListFlows diff (want: [%q] got: [%q])", w, g)
	}

	wantCalls := []call{
		call{
			handlerName: "ListFlows",
			args: &afpb.ApiListFlowsArgs{
				ClientId: proto.String(testClientID),
			},
		},
	}

	if len(conn.calls) != len(wantCalls) ||
		conn.calls[0].handlerName != wantCalls[0].handlerName ||
		!proto.Equal(conn.calls[0].args.(proto.Message), wantCalls[0].args.(proto.Message)) {
		t.Errorf("The arguments of ListFlows diff (want: [%q] got: [%q])", wantCalls, conn.calls)
	}
}

func TestGrantApproval(t *testing.T) {
	client, conn := createTestClient()
	approval := client.Approval(testApprovalID, testUsername)

	if err := approval.Grant(context.Background()); err != nil {
		t.Fatalf("cannot grant a client approval failed: %v", err)
	}

	want := &Approval{
		conn:     conn,
		ID:       testApprovalID,
		ClientID: testClientID,
		Username: testUsername,
		Data: &aupb.ApiClientApproval{
			Id: proto.String(testApprovalID),
		},
	}

	if w, g := want, approval; g.conn != w.conn ||
		g.ID != w.ID ||
		g.ClientID != w.ClientID ||
		g.Username != w.Username ||
		!proto.Equal(g.Data, w.Data) {
		t.Errorf("unexpected client approval got; (want: [%q] got: [%q])", w, g)
	}

	wantCalls := []call{{
		handlerName: "GrantClientApproval",
		args: &aupb.ApiGrantClientApprovalArgs{
			ClientId:   proto.String(testClientID),
			ApprovalId: proto.String(testApprovalID),
			Username:   proto.String(testUsername),
		},
	}}

	if len(conn.calls) != len(wantCalls) ||
		conn.calls[0].handlerName != wantCalls[0].handlerName ||
		!proto.Equal(conn.calls[0].args.(proto.Message), wantCalls[0].args.(proto.Message)) {
		t.Errorf("the arguments of GrantClientApproval diff (want: [%q] got: [%q])", wantCalls, conn.calls)
	}
}

func TestCreateApproval(t *testing.T) {
	client, conn := createTestClient()

	reason := "reason"
	notifiedUsers := []string{"user1", "user2"}
	emailAddresses := []string{"example@gmail.com", "test@gmail.com"}

	approval, err := client.CreateApproval(context.Background(), reason, notifiedUsers, EmailCCAddresses(emailAddresses), KeepClientAlive())
	if err != nil {
		t.Fatalf("CreateApproval for a Client object failed: %v", err)
	}

	want := &Approval{
		conn:     conn,
		ID:       testApprovalID,
		ClientID: testClientID,
		Username: testUsername,
		Data: &aupb.ApiClientApproval{
			Id: proto.String(testApprovalID),
		},
	}

	if w, g := want, approval; g.conn != w.conn ||
		g.ID != w.ID ||
		g.ClientID != w.ClientID ||
		g.Username != w.Username ||
		!proto.Equal(g.Data, w.Data) {
		t.Errorf("unexpected result of CreateApproval got (want: [%q] got: [%q])", w, g)
	}

	wantCalls := []call{{
		handlerName: "CreateClientApproval",
		args: &aupb.ApiCreateClientApprovalArgs{
			ClientId: proto.String(testClientID),
			Approval: &aupb.ApiClientApproval{
				Reason:           proto.String(reason),
				NotifiedUsers:    notifiedUsers,
				EmailCcAddresses: emailAddresses,
			},
			KeepClientAlive: proto.Bool(true),
		},
	}}

	if len(conn.calls) != len(wantCalls) ||
		conn.calls[0].handlerName != wantCalls[0].handlerName ||
		!proto.Equal(conn.calls[0].args.(proto.Message), wantCalls[0].args.(proto.Message)) {
		t.Errorf("the arguments of CreateClientApproval diff (want: [%q] got: [%q])", wantCalls, conn.calls)
	}
}

func TestCreateFlow(t *testing.T) {
	client, conn := createTestClient()

	wantFlowArgs := &flowpb.FileFinderArgs{
		Paths: []string{testFilePath},
	}

	flowRunnerArgs := &flowpb.FlowRunnerArgs{
		ClientId: proto.String(testClientID),
	}

	f, err := client.CreateFlow(context.Background(), testFlowName, wantFlowArgs, WithRunnerArgs(flowRunnerArgs))
	if err != nil {
		t.Fatalf("CreateFlow for Client fails: %v", err)
	}

	want := flow.NewFlow(testClientID, testFlowID, conn)
	want.Data = &afpb.ApiFlow{
		FlowId:     proto.String(testFlowID),
		Name:       proto.String(testFlowName),
		RunnerArgs: flowRunnerArgs,
	}

	if w, g := want, f; g.ID != w.ID ||
		g.ClientID != w.ClientID ||
		!proto.Equal(g.Data, w.Data) {
		t.Errorf("result of CreateFlow diff (want: [%q] got: [%q])", w, g)
	}

	// Check Args of the Flow is correct.
	anyArgs := conn.calls[0].args.(*afpb.ApiCreateFlowArgs).GetFlow().GetArgs()

	gotFlowArgs := &flowpb.FileFinderArgs{}
	if err := ptypes.UnmarshalAny(anyArgs, gotFlowArgs); err != nil {
		t.Fatalf("ptypes.UnmarshalAny failed: %v", err)
	}

	if w, g := wantFlowArgs, gotFlowArgs; !proto.Equal(w, g) {
		t.Fatalf("Args field of Flow in arguments of CreateFlow diff (want: [%q] got: [%q])", w, g)
	}

	wantCalls := []call{{
		handlerName: "CreateFlow",
		args: &afpb.ApiCreateFlowArgs{
			ClientId: proto.String(testClientID),
			Flow: &afpb.ApiFlow{
				Name:       proto.String(testFlowName),
				RunnerArgs: flowRunnerArgs,
				Args:       anyArgs,
			},
		},
	}}

	if len(conn.calls) != len(wantCalls) ||
		conn.calls[0].handlerName != wantCalls[0].handlerName ||
		!proto.Equal(conn.calls[0].args.(proto.Message), wantCalls[0].args.(proto.Message)) {
		t.Errorf("The arguments of CreateFlow diff (want: [%q] got: [%q])", wantCalls, conn.calls)
	}
}

func TestGetClient(t *testing.T) {
	client, conn := createTestClient()

	err := client.Get(context.Background())
	if err != nil {
		t.Fatalf("get method for Client fails: %v", err)
	}

	want := &Client{
		conn: conn,
		ID:   testClientID,
		Data: &acpb.ApiClient{
			ClientId: proto.String(testClientID),
		},
	}

	if w, g := want, client; g.ID != w.ID ||
		!proto.Equal(g.Data, w.Data) {
		t.Errorf("client objects diff (want: [%q] got: [%q])", w, g)
	}

	wantCalls := []call{{
		handlerName: "GetClient",
		args: &acpb.ApiGetClientArgs{
			ClientId: proto.String(client.ID),
		},
	}}

	if len(conn.calls) != len(wantCalls) ||
		conn.calls[0].handlerName != wantCalls[0].handlerName ||
		!proto.Equal(conn.calls[0].args.(proto.Message), wantCalls[0].args.(proto.Message)) {
		t.Errorf("The arguments of GetClient diff (want: [%q] got: [%q])", wantCalls, conn.calls)
	}

}

func TestSearchClients(t *testing.T) {
	const query = "Query"
	conn := &fakeConnector{}
	var gotItems []*Client

	err := SearchClients(context.Background(), conn, query,
		func(a *Client) error {
			gotItems = append(gotItems, a)
			return nil
		}, WithMaxCount(1))
	if err != nil {
		t.Fatalf("SearchClients for clients failed: %v", err)
	}

	wantItems := []*Client{{
		conn: conn,
		ID:   testClientID,
		Data: &acpb.ApiClient{
			ClientId: proto.String(testClientID),
		},
	}}

	if w, g := wantItems, gotItems; len(g) != len(w) ||
		g[0].ID != w[0].ID ||
		!proto.Equal(g[0].Data, w[0].Data) {
		t.Errorf("The results of SearchClients diff (want: [%q] got: [%q])", w, g)
	}

	wantCalls := []call{
		call{
			handlerName: "SearchClients",
			args: &acpb.ApiSearchClientsArgs{
				Query: proto.String(query),
				Count: proto.Int64(1),
			},
		},
	}

	if len(conn.calls) != len(wantCalls) ||
		conn.calls[0].handlerName != wantCalls[0].handlerName ||
		!proto.Equal(conn.calls[0].args.(proto.Message), wantCalls[0].args.(proto.Message)) {
		t.Errorf("The arguments of SearchClients diff (want: [%q] got: [%q])", wantCalls, conn.calls)
	}
}
