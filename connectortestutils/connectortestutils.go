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

// Package connectortestutils provides utilities for connector tests.
package connectortestutils

import (
	"fmt"
	"testing"

	"context"
	"github.com/golang/protobuf/proto"
	acpb "github.com/google/grr_go_api_client/grrproto/grr/proto/api/client_proto"
	afpb "github.com/google/grr_go_api_client/grrproto/grr/proto/api/flow_proto"
	ahpb "github.com/google/grr_go_api_client/grrproto/grr/proto/api/hunt_proto"
	aupb "github.com/google/grr_go_api_client/grrproto/grr/proto/api/user_proto"
	avpb "github.com/google/grr_go_api_client/grrproto/grr/proto/api/vfs_proto"
)

type connector interface {
	SetPageSize(pageSize int64)
	ListHunts(ctx context.Context, args *ahpb.ApiListHuntsArgs, fn func(m *ahpb.ApiHunt) error) (int64, error)
	ListHuntResults(ctx context.Context, args *ahpb.ApiListHuntResultsArgs, fn func(m *ahpb.ApiHuntResult) error) (int64, error)
	ListFlowResults(ctx context.Context, args *afpb.ApiListFlowResultsArgs, fn func(m *afpb.ApiFlowResult) error) (int64, error)
	ListClientApprovals(ctx context.Context, args *aupb.ApiListClientApprovalsArgs, fn func(m *aupb.ApiClientApproval) error) error
	ListHuntApprovals(ctx context.Context, args *aupb.ApiListHuntApprovalsArgs, fn func(m *aupb.ApiHuntApproval) error) error
	ListFiles(ctx context.Context, args *avpb.ApiListFilesArgs, fn func(m *avpb.ApiFile) error) error
	ListFlows(ctx context.Context, args *afpb.ApiListFlowsArgs, fn func(m *afpb.ApiFlow) error) error
	SearchClients(ctx context.Context, args *acpb.ApiSearchClientsArgs, fn func(m *acpb.ApiClient) error) error
}

const (
	// TestHuntID is a placeholder hunt ID.
	TestHuntID = "testHuntID"

	testClientID = "testClientID"
	testFlowID   = "testFlowID"
	testFilePath = "testFilePath"
)

var (
	iteratorRequestCalls = []struct {
		desc       string
		pageSize   int64
		totalCount int64
	}{
		{"pageSize is 1", 1, 4},
		{"totalCount is divisible by pageSize", 2, 4},
		{"totalCount is not divisible by pageSize", 3, 4},
		{"totalCount is equal to pageSize", 4, 4},
		{"totalCount is smaller than pageSize", 100, 4},
	}
)

// PartialTestSendIteratorRequestWithTotalCount is a common test implementation
// to be used with specific connector test.
func PartialTestSendIteratorRequestWithTotalCount(t *testing.T, conn connector) {
	tests := []struct {
		methodName string
		method     func() (int64, error)
	}{
		{"ListHunts", func() (int64, error) {
			var items []*ahpb.ApiHunt
			totalCount, err := conn.ListHunts(context.Background(), &ahpb.ApiListHuntsArgs{}, func(m *ahpb.ApiHunt) error {
				items = append(items, m)
				return nil
			})
			if err != nil {
				return totalCount, fmt.Errorf("method ListHunts failed: %v", err)
			}
			want := []*ahpb.ApiHunt{{Name: proto.String("1")}, {Name: proto.String("2")}, {Name: proto.String("3")}, {Name: proto.String("4")}}
			if w, g := want, items; len(w) != len(g) {
				for i := 0; i < len(w); i++ {
					if !proto.Equal(w[i], g[i]) {
						return totalCount, fmt.Errorf("output of ListHunts diff (want: [%q] got: [%q])", w[i], g[i])
					}
				}
			}
			return totalCount, nil
		},
		},
		{"ListHuntResults", func() (int64, error) {
			var items []*ahpb.ApiHuntResult
			totalCount, err := conn.ListHuntResults(context.Background(), &ahpb.ApiListHuntResultsArgs{
				HuntId: proto.String(TestHuntID),
			}, func(m *ahpb.ApiHuntResult) error {
				items = append(items, m)
				return nil
			})
			if err != nil {
				return totalCount, fmt.Errorf("method ListHuntResults failed: %v", err)
			}
			want := []*ahpb.ApiHuntResult{
				{ClientId: proto.String("1")},
				{ClientId: proto.String("2")},
				{ClientId: proto.String("3")},
				{ClientId: proto.String("4")},
			}
			if w, g := want, items; len(w) != len(g) {
				for i := 0; i < len(w); i++ {
					if !proto.Equal(w[i], g[i]) {
						return totalCount, fmt.Errorf("output of ListHuntResults diff (want: [%q] got: [%q])", w[i], g[i])
					}
				}
			}
			return totalCount, nil
		},
		},
		{"ListFlowResults", func() (int64, error) {
			var items []*afpb.ApiFlowResult
			totalCount, err := conn.ListFlowResults(context.Background(), &afpb.ApiListFlowResultsArgs{
				ClientId: proto.String("testClientID"),
				FlowId:   proto.String("testFlowID"),
			}, func(m *afpb.ApiFlowResult) error {
				items = append(items, m)
				return nil
			})
			if err != nil {
				return totalCount, fmt.Errorf("method ListFlowResults failed: %v", err)
			}
			want := []*afpb.ApiFlowResult{
				{Timestamp: proto.Uint64(1)},
				{Timestamp: proto.Uint64(2)},
				{Timestamp: proto.Uint64(3)},
				{Timestamp: proto.Uint64(4)}}
			if w, g := want, items; len(w) != len(g) {
				for i := 0; i < len(w); i++ {
					if !proto.Equal(w[i], g[i]) {
						return totalCount, fmt.Errorf("output of ListFlowResults diff (want: [%q] got: [%q])", w[i], g[i])
					}
				}
			}
			return totalCount, nil
		},
		},
	}

	for _, test := range tests {
		for _, call := range iteratorRequestCalls {
			t.Run(fmt.Sprintf("test %s for %s:", call.desc, test.methodName), func(t *testing.T) {
				conn.SetPageSize(call.pageSize)
				totalCount, err := test.method()
				if err != nil {
					t.Fatal(err)
				}
				if totalCount != call.totalCount {
					t.Fatalf("total count of the results of %s is %v; want %v", test.methodName, totalCount, call.totalCount)
				}
			})
		}
	}
}

// PartialTestSendIteratorRequestWithoutTotalCount is a common test
// implementation to be used with specific connector test.
func PartialTestSendIteratorRequestWithoutTotalCount(t *testing.T, conn connector) {
	tests := []struct {
		methodName string
		method     func() error
	}{
		{"ListClientApprovals", func() error {
			var items []*aupb.ApiClientApproval
			err := conn.ListClientApprovals(context.Background(), &aupb.ApiListClientApprovalsArgs{}, func(m *aupb.ApiClientApproval) error {
				items = append(items, m)
				return nil
			})
			if err != nil {
				return fmt.Errorf("method ListClientApprovals failed: %v", err)
			}
			want := []*aupb.ApiClientApproval{{Id: proto.String("1")}, {Id: proto.String("2")}, {Id: proto.String("3")}, {Id: proto.String("4")}}
			if w, g := want, items; len(w) != len(g) {
				for i := 0; i < len(w); i++ {
					if !proto.Equal(w[i], g[i]) {
						return fmt.Errorf("output of ListClientApprovals diff (want: [%q] got: [%q])", w[i], g[i])
					}
				}
			}
			return nil
		},
		},
		{"ListHuntApprovals", func() error {
			var items []*aupb.ApiHuntApproval
			err := conn.ListHuntApprovals(context.Background(), &aupb.ApiListHuntApprovalsArgs{}, func(m *aupb.ApiHuntApproval) error {
				items = append(items, m)
				return nil
			})
			if err != nil {
				return fmt.Errorf("method ListHuntApprovals failed: %v", err)
			}
			want := []*aupb.ApiHuntApproval{{Id: proto.String("1")}, {Id: proto.String("2")}, {Id: proto.String("3")}, {Id: proto.String("4")}}
			if w, g := want, items; len(w) != len(g) {
				for i := 0; i < len(w); i++ {
					if !proto.Equal(w[i], g[i]) {
						return fmt.Errorf("output of ListHuntApprovals diff (want: [%q] got: [%q])", w[i], g[i])
					}
				}
			}
			return nil
		},
		},
		{"ListFiles", func() error {
			var items []*avpb.ApiFile
			err := conn.ListFiles(context.Background(), &avpb.ApiListFilesArgs{
				ClientId: proto.String(testClientID),
			}, func(m *avpb.ApiFile) error {
				items = append(items, m)
				return nil
			})
			if err != nil {
				return fmt.Errorf("method ListFiles failed: %v", err)
			}
			want := []*avpb.ApiFile{
				{Path: proto.String("1")},
				{Path: proto.String("2")},
				{Path: proto.String("3")},
				{Path: proto.String("4")},
			}
			if w, g := want, items; len(w) != len(g) {
				for i := 0; i < len(w); i++ {
					if !proto.Equal(w[i], g[i]) {
						return fmt.Errorf("output of ListFiles diff (want: [%q] got: [%q])", w[i], g[i])
					}
				}
			}
			return nil
		},
		},
		{"ListFlows", func() error {
			var items []*afpb.ApiFlow
			err := conn.ListFlows(context.Background(), &afpb.ApiListFlowsArgs{
				ClientId: proto.String(testClientID),
			}, func(m *afpb.ApiFlow) error {
				items = append(items, m)
				return nil
			})
			if err != nil {
				return fmt.Errorf("method ListFlows failed: %v", err)
			}
			want := []*afpb.ApiFlow{
				{FlowId: proto.String("1")},
				{FlowId: proto.String("2")},
				{FlowId: proto.String("3")},
				{FlowId: proto.String("4")},
			}
			if w, g := want, items; len(w) != len(g) {
				for i := 0; i < len(w); i++ {
					if !proto.Equal(w[i], g[i]) {
						return fmt.Errorf("output of ListFlows diff (want: [%q] got: [%q])", w[i], g[i])
					}
				}
			}
			return nil
		},
		},
		{"SearchClients", func() error {
			var items []*acpb.ApiClient
			err := conn.SearchClients(context.Background(), &acpb.ApiSearchClientsArgs{}, func(m *acpb.ApiClient) error {
				items = append(items, m)
				return nil
			})
			if err != nil {
				return fmt.Errorf("method SearchClients failed: %v", err)
			}
			want := []*acpb.ApiClient{
				{ClientId: proto.String("1")},
				{ClientId: proto.String("2")},
				{ClientId: proto.String("3")},
				{ClientId: proto.String("4")},
			}
			if w, g := want, items; len(w) != len(g) {
				for i := 0; i < len(w); i++ {
					if !proto.Equal(w[i], g[i]) {
						return fmt.Errorf("output of SearchClients diff (want: [%q] got: [%q])", w[i], g[i])
					}
				}
			}
			return nil
		},
		},
	}

	for _, test := range tests {
		for _, call := range iteratorRequestCalls {
			t.Run(fmt.Sprintf("test %s for %s:", call.desc, test.methodName), func(t *testing.T) {
				conn.SetPageSize(call.pageSize)
				if err := test.method(); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}
