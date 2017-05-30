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

package hunt

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"context"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	ahpb "github.com/google/grr_go_api_client/grrproto/grr/proto/api/hunt_proto"
	aupb "github.com/google/grr_go_api_client/grrproto/grr/proto/api/user_proto"
	flowpb "github.com/google/grr_go_api_client/grrproto/grr/proto/flows_proto"
)

type call struct {
	handlerName string
	args        interface{}
}

const (
	huntID     = "huntTest1"
	approvalID = "approvalTest1"
	username   = "usernameTest1"
	flowName   = "flow"
	huntName   = "hunt"
	filePath   = "filePath"
)

var (
	testReadCloser = ioutil.NopCloser(strings.NewReader("reader"))
)

// mockConnector simply records information of all calls made via the connector.
type mockConnector struct {
	calls []call
	Connector
}

func (c *mockConnector) ListHunts(ctx context.Context, args *ahpb.ApiListHuntsArgs, fn func(m *ahpb.ApiHunt) error) (int64, error) {
	c.calls = append(c.calls, call{
		handlerName: "ListHunts",
		args:        args,
	})

	fn(&ahpb.ApiHunt{
		Urn:  proto.String("aff4:/hunts/1234"),
		Name: proto.String("testHunt"),
	})

	return 1, nil
}

func (c *mockConnector) ListHuntResults(ctx context.Context, args *ahpb.ApiListHuntResultsArgs, fn func(m *ahpb.ApiHuntResult) error) (int64, error) {
	c.calls = append(c.calls, call{
		handlerName: "ListHuntResults",
		args:        args,
	})

	fn(&ahpb.ApiHuntResult{
		ClientId: proto.String("1"),
	})

	return 1, nil
}

func (c *mockConnector) GetHunt(ctx context.Context, args *ahpb.ApiGetHuntArgs) (*ahpb.ApiHunt, error) {
	c.calls = append(c.calls, call{
		handlerName: "GetHunt",
		args:        args,
	})

	return &ahpb.ApiHunt{
		Urn: proto.String(huntID),
	}, nil
}

func (c *mockConnector) ModifyHunt(ctx context.Context, args *ahpb.ApiModifyHuntArgs) (*ahpb.ApiHunt, error) {
	c.calls = append(c.calls, call{
		handlerName: "ModifyHunt",
		args:        args,
	})

	return &ahpb.ApiHunt{
		Urn: proto.String(huntID),
	}, nil
}

func (c *mockConnector) GetHuntFilesArchive(ctx context.Context, args *ahpb.ApiGetHuntFilesArchiveArgs) (io.ReadCloser, error) {
	c.calls = append(c.calls, call{
		handlerName: "GetHuntFilesArchive",
		args:        args,
	})
	return testReadCloser, nil
}

func (c *mockConnector) GetHuntApproval(ctx context.Context, args *aupb.ApiGetHuntApprovalArgs) (*aupb.ApiHuntApproval, error) {
	c.calls = append(c.calls, call{
		handlerName: "GetHuntApproval",
		args:        args,
	})

	return &aupb.ApiHuntApproval{
		Id: proto.String(approvalID),
	}, nil
}

func (c *mockConnector) GrantHuntApproval(ctx context.Context, args *aupb.ApiGrantHuntApprovalArgs) (*aupb.ApiHuntApproval, error) {
	c.calls = append(c.calls, call{
		handlerName: "GrantHuntApproval",
		args:        args,
	})

	return &aupb.ApiHuntApproval{
		Id: proto.String(approvalID),
	}, nil
}

func (c *mockConnector) GetUsername(ctx context.Context) (string, error) {
	return username, nil
}

func (c *mockConnector) ListHuntApprovals(ctx context.Context, args *aupb.ApiListHuntApprovalsArgs, fn func(m *aupb.ApiHuntApproval) error) error {
	c.calls = append(c.calls, call{
		handlerName: "ListHuntApprovals",
		args:        args,
	})

	fn(&aupb.ApiHuntApproval{
		Id: proto.String(approvalID),
		Subject: &ahpb.ApiHunt{
			Urn: proto.String(huntID),
		},
	})

	return nil
}

func (c *mockConnector) CreateHunt(ctx context.Context, args *ahpb.ApiCreateHuntArgs) (*ahpb.ApiHunt, error) {
	c.calls = append(c.calls, call{
		handlerName: "CreateHunt",
		args:        args,
	})

	return &ahpb.ApiHunt{
		Urn: proto.String(fmt.Sprintf("aff4:/hunts/%s", huntID)),
	}, nil
}

func (c *mockConnector) CreateHuntApproval(ctx context.Context, args *aupb.ApiCreateHuntApprovalArgs) (*aupb.ApiHuntApproval, error) {
	c.calls = append(c.calls, call{
		handlerName: "CreateHuntApproval",
		args:        args,
	})

	return &aupb.ApiHuntApproval{
		Id: proto.String(approvalID),
	}, nil
}

func createTestHunt() (*Hunt, *mockConnector) {
	conn := &mockConnector{}

	return &Hunt{
		conn: conn,
		ID:   huntID,
	}, conn
}

func TestListHunts(t *testing.T) {
	conn := &mockConnector{}

	var items []*Hunt
	totalCount, err := ListHunts(context.Background(), conn,
		func(h *Hunt) error {
			items = append(items, h)
			return nil
		})
	if err != nil {
		t.Fatal(err)
	}

	if totalCount != 1 {
		t.Errorf("The total count of results of ListHunts is %v; want: 1", totalCount)
	}

	want := []*Hunt{{
		conn: conn,
		ID:   "1234",
		Data: &ahpb.ApiHunt{
			Urn:  proto.String("aff4:/hunts/1234"),
			Name: proto.String("testHunt"),
		},
	}}

	if w, g := want, items; len(g) != len(w) ||
		g[0].ID != w[0].ID ||
		!proto.Equal(g[0].Data, w[0].Data) {
		t.Errorf("The results of ListHunts diff (want: [%q] got: [%q])", w, g)
	}

	wantCall := []call{{
		handlerName: "ListHunts",
		args:        &ahpb.ApiListHuntsArgs{},
	}}

	if w, g := wantCall, conn.calls; len(g) != len(w) ||
		g[0].handlerName != w[0].handlerName ||
		!proto.Equal(g[0].args.(proto.Message), w[0].args.(proto.Message)) {
		t.Errorf("The arguments of ListHunts diff (want: [%q] got: [%q])", w, g)
	}
}

func TestListResults(t *testing.T) {
	hunt, conn := createTestHunt()

	var items []*Result
	totalCount, err := hunt.ListResults(context.Background(),
		func(r *Result) error {
			items = append(items, r)
			return nil
		})
	if err != nil {
		t.Fatal(err)
	}
	if totalCount != 1 {
		t.Errorf("The total count of results of ListResults is %v; want: 1", totalCount)
	}

	want := []*Result{{
		Data: &ahpb.ApiHuntResult{
			ClientId: proto.String("1"),
		},
	}}

	if w, g := want, items; len(g) != len(w) ||
		!proto.Equal(g[0].Data, w[0].Data) {
		t.Errorf("The results of ListResults diff (want: [%q] got: [%q])", w, g)
	}

	wantCall := []call{{
		handlerName: "ListHuntResults",
		args: &ahpb.ApiListHuntResultsArgs{
			HuntId: proto.String(huntID),
		},
	}}

	if w, g := wantCall, conn.calls; len(g) != len(w) ||
		g[0].handlerName != w[0].handlerName ||
		!proto.Equal(g[0].args.(proto.Message), w[0].args.(proto.Message)) {
		t.Errorf("The arguments of ListResults diff (want: [%q] got: [%q])", w, g)
	}
}

func TestGetHunt(t *testing.T) {
	hunt, conn := createTestHunt()

	if err := hunt.Get(context.Background()); err != nil {
		t.Fatalf("get method for a hunt failed: %v", err)
	}

	want := &Hunt{
		conn: conn,
		ID:   huntID,
		Data: &ahpb.ApiHunt{
			Urn: proto.String(huntID),
		},
	}

	if w, g := want, hunt; g.ID != w.ID ||
		!proto.Equal(g.Data, w.Data) {
		t.Errorf("unexpected hunt got; diff (want: [%q] got: [%q])", w, g)
	}

	wantCalls := []call{{
		handlerName: "GetHunt",
		args: &ahpb.ApiGetHuntArgs{
			HuntId: proto.String(hunt.ID),
		},
	}}

	if w, g := wantCalls, conn.calls; len(g) != len(w) ||
		g[0].handlerName != w[0].handlerName ||
		!proto.Equal(g[0].args.(proto.Message), w[0].args.(proto.Message)) {
		t.Errorf("the arguments of GetHunt diff (want: [%q] got: [%q])", w, g)
	}
}

func TestStart(t *testing.T) {
	hunt, conn := createTestHunt()

	if err := hunt.Start(context.Background()); err != nil {
		t.Fatalf("starting a hunt failed: %v", err)
	}

	want := &Hunt{
		conn: conn,
		ID:   huntID,
		Data: &ahpb.ApiHunt{
			Urn: proto.String(huntID),
		},
	}

	if w, g := want, hunt; g.ID != w.ID ||
		!proto.Equal(g.Data, w.Data) {
		t.Errorf("unexpected hunt got for starting a hunt; diff (want: [%q] got: [%q])", w, g)
	}

	wantCalls := []call{{
		handlerName: "ModifyHunt",
		args: &ahpb.ApiModifyHuntArgs{
			HuntId: proto.String(hunt.ID),
			State:  ahpb.ApiHunt_STARTED.Enum(),
		},
	}}

	if w, g := wantCalls, conn.calls; len(g) != len(w) ||
		g[0].handlerName != w[0].handlerName ||
		!proto.Equal(g[0].args.(proto.Message), w[0].args.(proto.Message)) {
		t.Errorf("the arguments of ModifyHunt diff (want: [%q] got: [%q])", w, g)
	}
}

func TestStop(t *testing.T) {
	hunt, conn := createTestHunt()

	if err := hunt.Stop(context.Background()); err != nil {
		t.Fatalf("stopping a hunt failed: %v", err)
	}

	want := &Hunt{
		conn: conn,
		ID:   huntID,
		Data: &ahpb.ApiHunt{
			Urn: proto.String(huntID),
		},
	}

	if w, g := want, hunt; g.ID != w.ID ||
		!proto.Equal(g.Data, w.Data) {
		t.Errorf("unexpected hunt got for stopping a hunt; diff (want: [%q] got: [%q])", w, g)
	}

	wantCalls := []call{{
		handlerName: "ModifyHunt",
		args: &ahpb.ApiModifyHuntArgs{
			HuntId: proto.String(hunt.ID),
			State:  ahpb.ApiHunt_STOPPED.Enum(),
		},
	}}

	if w, g := wantCalls, conn.calls; len(g) != len(w) ||
		g[0].handlerName != w[0].handlerName ||
		!proto.Equal(g[0].args.(proto.Message), w[0].args.(proto.Message)) {
		t.Errorf("the arguments of ModifyHunt diff (want: [%q] got: [%q])", w, g)
	}
}

func TestModify(t *testing.T) {
	hunt, conn := createTestHunt()

	// No optional argument passed.
	if err := hunt.Modify(context.Background()); err != nil {
		t.Fatalf("Modifying a hunt (without optional argument) failed: %v", err)
	}

	want := &Hunt{
		conn: conn,
		ID:   huntID,
		Data: &ahpb.ApiHunt{
			Urn: proto.String(huntID),
		},
	}

	if w, g := want, hunt; g.ID != w.ID ||
		!proto.Equal(g.Data, w.Data) {
		t.Errorf("unexpected hunt got for modifying a hunt (without optional argument); diff (want: [%q] got: [%q])", w, g)
	}

	wantCalls := []call{{
		handlerName: "ModifyHunt",
		args: &ahpb.ApiModifyHuntArgs{
			HuntId: proto.String(hunt.ID),
		},
	}}

	if w, g := wantCalls, conn.calls; len(g) != len(w) ||
		g[0].handlerName != w[0].handlerName ||
		!proto.Equal(g[0].args.(proto.Message), w[0].args.(proto.Message)) {
		t.Errorf("the arguments of ModifyHunt (without optional argument) diff (want: [%q] got: [%q])", w, g)
	}

	// Optional arguments are passed.
	if err := hunt.Modify(context.Background(), ModifyClientLimit(100), ModifyClientRate(1000)); err != nil {
		t.Fatalf("Modifying a hunt (with optional arguments) failed: %v", err)
	}

	if w, g := want, hunt; g.ID != w.ID ||
		!proto.Equal(g.Data, w.Data) {
		t.Errorf("unexpected hunt got for modifying a hunt (with optional argument); diff (want: [%q] got: [%q])", w, g)
	}

	wantCalls = append(wantCalls, call{
		handlerName: "ModifyHunt",
		args: &ahpb.ApiModifyHuntArgs{
			HuntId:      proto.String(hunt.ID),
			ClientLimit: proto.Int64(100),
			ClientRate:  proto.Int64(1000),
		},
	})

	if w, g := wantCalls, conn.calls; len(g) != len(w) ||
		g[0].handlerName != w[0].handlerName ||
		!proto.Equal(g[0].args.(proto.Message), w[0].args.(proto.Message)) {
		t.Errorf("the arguments of ModifyHunt (with optional argument) diff (want: [%q] got: [%q])", w, g)
	}
}

func TestGetHuntFilesArchive(t *testing.T) {
	hunt, conn := createTestHunt()

	format := ahpb.ApiGetHuntFilesArchiveArgs_TAR_GZ
	rc, err := hunt.GetFilesArchive(context.Background(), &format)
	if err != nil {
		t.Fatalf("GetFilesArchive for hunts failed: %v", err)
	}

	if rc != testReadCloser {
		t.Errorf("The readCloser returned from GetHuntFilesArchive is %v; want: %v", rc, testReadCloser)
	}

	want := []call{
		call{
			handlerName: "GetHuntFilesArchive",
			args: &ahpb.ApiGetHuntFilesArchiveArgs{
				HuntId:        proto.String(hunt.ID),
				ArchiveFormat: &format,
			},
		},
	}

	if w, g := want, conn.calls; len(g) != len(w) ||
		g[0].handlerName != w[0].handlerName ||
		!proto.Equal(g[0].args.(proto.Message), w[0].args.(proto.Message)) {
		t.Errorf("The arguments of GetHuntFilesArchive diff (want: [%q] got: [%q])", w, g)
	}
}

func TestGetApproval(t *testing.T) {
	hunt, conn := createTestHunt()
	approval := hunt.Approval(approvalID, username)

	if err := approval.Get(context.Background()); err != nil {
		t.Fatalf("get method for a hunt approval failed: %v", err)
	}

	want := &Approval{
		conn:     conn,
		ID:       approvalID,
		HuntID:   huntID,
		Username: username,
		Data: &aupb.ApiHuntApproval{
			Id: proto.String(approvalID),
		},
	}

	if w, g := want, approval; g.ID != w.ID ||
		g.HuntID != w.HuntID ||
		g.Username != w.Username ||
		!proto.Equal(g.Data, w.Data) {
		t.Errorf("unexpected hunt approval got; diff (want: [%q] got: [%q])", w, g)
	}

	wantCalls := []call{{
		handlerName: "GetHuntApproval",
		args: &aupb.ApiGetHuntApprovalArgs{
			HuntId:     proto.String(huntID),
			ApprovalId: proto.String(approvalID),
			Username:   proto.String(username),
		},
	}}

	if w, g := wantCalls, conn.calls; len(g) != len(w) ||
		g[0].handlerName != w[0].handlerName ||
		!proto.Equal(g[0].args.(proto.Message), w[0].args.(proto.Message)) {
		t.Errorf("the arguments of GetHuntApproval diff (want: [%q] got: [%q])", w, g)
	}
}

func TestGrantApproval(t *testing.T) {
	hunt, conn := createTestHunt()
	approval := hunt.Approval(approvalID, username)

	if err := approval.Grant(context.Background()); err != nil {
		t.Fatalf("cannot grant a hunt approval failed: %v", err)
	}

	want := &Approval{
		conn:     conn,
		ID:       approvalID,
		HuntID:   huntID,
		Username: username,
		Data: &aupb.ApiHuntApproval{
			Id: proto.String(approvalID),
		},
	}

	if w, g := want, approval; g.ID != w.ID ||
		g.HuntID != w.HuntID ||
		g.Username != w.Username ||
		!proto.Equal(g.Data, w.Data) {
		t.Errorf("unexpected hunt approval got; diff (want: [%q] got: [%q])", w, g)
	}

	wantCalls := []call{{
		handlerName: "GrantHuntApproval",
		args: &aupb.ApiGrantHuntApprovalArgs{
			HuntId:     proto.String(huntID),
			ApprovalId: proto.String(approvalID),
			Username:   proto.String(username),
		},
	}}

	if w, g := wantCalls, conn.calls; len(g) != len(w) ||
		g[0].handlerName != w[0].handlerName ||
		!proto.Equal(g[0].args.(proto.Message), w[0].args.(proto.Message)) {
		t.Errorf("the arguments of GrantHuntApproval diff (want: [%q] got: [%q])", w, g)
	}
}

func TestListApprovals(t *testing.T) {
	conn := &mockConnector{}

	var gotItems []*Approval
	err := ListApprovals(context.Background(), conn,
		func(a *Approval) error {
			gotItems = append(gotItems, a)
			return nil
		})
	if err != nil {
		t.Fatalf("ListApprovals for hunts failed: %v", err)
	}

	wantItems := []*Approval{{
		conn:     conn,
		ID:       approvalID,
		HuntID:   huntID,
		Username: username,
		Data: &aupb.ApiHuntApproval{
			Id: proto.String(approvalID),
			Subject: &ahpb.ApiHunt{
				Urn: proto.String(huntID),
			},
		},
	}}

	if w, g := wantItems, gotItems; len(g) != len(w) ||
		g[0].ID != w[0].ID ||
		g[0].HuntID != w[0].HuntID ||
		g[0].Username != w[0].Username ||
		!proto.Equal(g[0].Data, w[0].Data) {
		t.Errorf("The results of ListHuntApprovals diff (want: [%q] got: [%q])", w, g)
	}

	wantCalls := []call{
		call{
			handlerName: "ListHuntApprovals",
			args:        &aupb.ApiListHuntApprovalsArgs{},
		},
	}

	if w, g := wantCalls, conn.calls; len(g) != len(w) ||
		g[0].handlerName != w[0].handlerName ||
		!proto.Equal(g[0].args.(proto.Message), w[0].args.(proto.Message)) {
		t.Errorf("The arguments of ListHuntApprovals diff (want: [%q] got: [%q])", w, g)
	}
}

func TestCreateHunt(t *testing.T) {
	conn := &mockConnector{}

	wantFlowArgs := &flowpb.FileFinderArgs{
		Paths: []string{filePath},
	}

	huntRunnerArgs := &flowpb.HuntRunnerArgs{
		HuntName: proto.String(huntName),
	}

	hunt, err := CreateHunt(context.Background(), conn, flowName, wantFlowArgs, CreateHuntRunnerArgs(huntRunnerArgs))
	if err != nil {
		t.Fatalf("CreateHunt failed: %v", err)
	}

	want := &Hunt{
		conn: conn,
		ID:   huntID,
		Data: &ahpb.ApiHunt{
			Urn: proto.String(fmt.Sprintf("aff4:/hunts/%s", huntID)),
		},
	}

	if w, g := want, hunt; g.ID != w.ID ||
		!proto.Equal(g.Data, w.Data) {
		t.Errorf("unexpected result of CreateHunt got; diff (want: [%q] got: [%q])", w, g)
	}

	if calls := len(conn.calls); calls != 1 {
		t.Fatalf("connector has been called %v times; want exactly one", calls)
	}

	ha, ok := conn.calls[0].args.(*ahpb.ApiCreateHuntArgs)
	if !ok {
		t.Fatal("type assertion to ApiCreateHuntArgs failed")
	}

	anyArgs := ha.GetFlowArgs()

	gotFlowArgs := &flowpb.FileFinderArgs{}
	if err := ptypes.UnmarshalAny(anyArgs, gotFlowArgs); err != nil {
		t.Fatalf("ptypes.UnmarshalAny failed: %v", err)
	}

	if w, g := wantFlowArgs, gotFlowArgs; !proto.Equal(w, g) {
		t.Fatalf("Args field of the argument of CreateHunt diff (want: [%q] got: [%q])", w, g)
	}

	wantCalls := []call{{
		handlerName: "CreateHunt",
		args: &ahpb.ApiCreateHuntArgs{
			FlowName:       proto.String(flowName),
			FlowArgs:       anyArgs,
			HuntRunnerArgs: huntRunnerArgs,
		},
	}}

	if w, g := wantCalls, conn.calls; len(g) != len(w) ||
		g[0].handlerName != w[0].handlerName ||
		!proto.Equal(g[0].args.(proto.Message), w[0].args.(proto.Message)) {
		t.Errorf("the arguments of CreateHunt diff (want: [%q] got: [%q])", w, g)
	}
}

func TestCreateApproval(t *testing.T) {
	hunt, conn := createTestHunt()

	const reason = "reason"
	notifiedUsers := []string{"user1", "user2"}
	emailAddresses := []string{"example@gmail.com", "test@gmail.com"}

	approval, err := hunt.CreateApproval(context.Background(), reason, notifiedUsers, EmailCCAddresses(emailAddresses))
	if err != nil {
		t.Fatalf("CreateApproval for a Hunt object failed: %v", err)
	}

	want := &Approval{
		conn:     conn,
		ID:       approvalID,
		HuntID:   huntID,
		Username: username,
		Data: &aupb.ApiHuntApproval{
			Id: proto.String(approvalID),
		},
	}

	if w, g := want, approval; g.ID != w.ID ||
		g.HuntID != w.HuntID ||
		g.Username != w.Username ||
		!proto.Equal(g.Data, w.Data) {
		t.Errorf("unexpected result of CreateApproval got; diff (want: [%q] got: [%q])", w, g)
	}

	wantCalls := []call{{
		handlerName: "CreateHuntApproval",
		args: &aupb.ApiCreateHuntApprovalArgs{
			HuntId: proto.String(huntID),
			Approval: &aupb.ApiHuntApproval{
				Reason:           proto.String(reason),
				NotifiedUsers:    notifiedUsers,
				EmailCcAddresses: emailAddresses,
			},
		},
	}}

	if w, g := wantCalls, conn.calls; len(g) != len(w) ||
		g[0].handlerName != w[0].handlerName ||
		!proto.Equal(g[0].args.(proto.Message), w[0].args.(proto.Message)) {
		t.Errorf("the arguments of CreateHuntApproval diff (want: [%q] got: [%q])", w, g)
	}
}
