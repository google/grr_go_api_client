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

package vfs

import (
	"io"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"

	"context"
	"github.com/golang/protobuf/proto"
	avpb "github.com/google/grr_go_api_client/grrproto/grr/proto/api/vfs_proto"
)

const (
	testClientID    = "clientTest"
	testOperationID = "operationTest"
	testFilePath    = "someRandomPath"
	testFileName    = "testFile"
)

var (
	testReadCloser = ioutil.NopCloser(strings.NewReader("reader"))
)

type call struct {
	handlerName string
	args        interface{}
}

// mockConnector simply records information of all calls made via the connector.
type mockConnector struct {
	calls []call
	Connector
}

func (c *mockConnector) GetFileDetails(ctx context.Context, args *avpb.ApiGetFileDetailsArgs) (*avpb.ApiGetFileDetailsResult, error) {
	c.calls = append(c.calls, call{
		handlerName: "GetFileDetails",
		args:        args,
	})
	return &avpb.ApiGetFileDetailsResult{
		File: &avpb.ApiFile{
			Name: proto.String(testFileName),
			Path: proto.String(testFilePath),
		},
	}, nil
}

func (c *mockConnector) GetFileBlob(ctx context.Context, args *avpb.ApiGetFileBlobArgs) (io.ReadCloser, error) {
	c.calls = append(c.calls, call{
		handlerName: "GetFileBlob",
		args:        args,
	})
	return testReadCloser, nil
}

func (c *mockConnector) GetVFSFilesArchive(ctx context.Context, args *avpb.ApiGetVfsFilesArchiveArgs) (io.ReadCloser, error) {
	c.calls = append(c.calls, call{
		handlerName: "GetVFSFilesArchive",
		args:        args,
	})
	return testReadCloser, nil
}

func (c *mockConnector) GetFileVersionTimes(ctx context.Context, args *avpb.ApiGetFileVersionTimesArgs) (*avpb.ApiGetFileVersionTimesResult, error) {
	c.calls = append(c.calls, call{
		handlerName: "GetFileVersionTimes",
		args:        args,
	})
	return &avpb.ApiGetFileVersionTimesResult{
		Times: []uint64{0, 1, 100},
	}, nil
}

func (c *mockConnector) CreateVFSRefreshOperation(ctx context.Context, args *avpb.ApiCreateVfsRefreshOperationArgs) (*avpb.ApiCreateVfsRefreshOperationResult, error) {
	c.calls = append(c.calls, call{
		handlerName: "CreateVFSRefreshOperation",
		args:        args,
	})
	return &avpb.ApiCreateVfsRefreshOperationResult{
		OperationId: proto.String(testOperationID),
	}, nil
}

func (c *mockConnector) GetVFSRefreshOperationState(ctx context.Context, args *avpb.ApiGetVfsRefreshOperationStateArgs) (*avpb.ApiGetVfsRefreshOperationStateResult, error) {
	c.calls = append(c.calls, call{
		handlerName: "GetVFSRefreshOperationState",
		args:        args,
	})
	return &avpb.ApiGetVfsRefreshOperationStateResult{
		State: avpb.ApiGetVfsRefreshOperationStateResult_FINISHED.Enum(),
	}, nil
}

func (c *mockConnector) UpdateVFSFileContent(ctx context.Context, args *avpb.ApiUpdateVfsFileContentArgs) (*avpb.ApiUpdateVfsFileContentResult, error) {
	c.calls = append(c.calls, call{
		handlerName: "UpdateVFSFileContent",
		args:        args,
	})
	return &avpb.ApiUpdateVfsFileContentResult{
		OperationId: proto.String(testOperationID),
	}, nil
}

func (c *mockConnector) GetVFSFileContentUpdateState(ctx context.Context, args *avpb.ApiGetVfsFileContentUpdateStateArgs) (*avpb.ApiGetVfsFileContentUpdateStateResult, error) {
	c.calls = append(c.calls, call{
		handlerName: "GetVFSFileContentUpdateState",
		args:        args,
	})
	return &avpb.ApiGetVfsFileContentUpdateStateResult{
		State: avpb.ApiGetVfsFileContentUpdateStateResult_FINISHED.Enum(),
	}, nil
}

func (c *mockConnector) GetVFSTimelineAsCSV(ctx context.Context, args *avpb.ApiGetVfsTimelineAsCsvArgs) (io.ReadCloser, error) {
	c.calls = append(c.calls, call{
		handlerName: "GetVFSTimelineAsCSV",
		args:        args,
	})
	return testReadCloser, nil
}

func (c *mockConnector) GetVFSTimeline(ctx context.Context, args *avpb.ApiGetVfsTimelineArgs) (*avpb.ApiGetVfsTimelineResult, error) {
	c.calls = append(c.calls, call{
		handlerName: "GetVFSTimeline",
		args:        args,
	})

	return &avpb.ApiGetVfsTimelineResult{
		Items: []*avpb.ApiVfsTimelineItem{
			{FilePath: proto.String(testFilePath)},
			{FilePath: proto.String(testFilePath)},
			{FilePath: proto.String(testFilePath)},
		},
	}, nil
}

func (c *mockConnector) ListFiles(ctx context.Context, args *avpb.ApiListFilesArgs, fn func(m *avpb.ApiFile) error) error {
	c.calls = append(c.calls, call{
		handlerName: "ListFiles",
		args:        args,
	})

	fn(&avpb.ApiFile{Path: proto.String(testFilePath)})
	fn(&avpb.ApiFile{Path: proto.String(testFilePath)})
	fn(&avpb.ApiFile{Path: proto.String(testFilePath)})

	return nil
}

func createTestFile() (*File, *mockConnector) {
	conn := &mockConnector{}

	return &File{
		conn:     conn,
		ClientID: testClientID,
		Path:     testFilePath,
	}, conn
}

func TestGetFile(t *testing.T) {
	file, conn := createTestFile()

	if err := file.Get(context.Background()); err != nil {
		t.Fatalf("get method for a VFS file failed: %v", err)
	}

	want := &File{
		conn:     conn,
		ClientID: testClientID,
		Path:     testFilePath,
		Data: &avpb.ApiFile{
			Name: proto.String(testFileName),
			Path: proto.String(testFilePath),
		},
	}

	if w, g := want, file; w.ClientID != g.ClientID ||
		w.Path != g.Path ||
		!proto.Equal(w.Data, g.Data) {
		t.Errorf("unexpected file got; diff (want: [%q] got: [%q])", w, g)
	}

	wantCalls := []call{{
		handlerName: "GetFileDetails",
		args: &avpb.ApiGetFileDetailsArgs{
			ClientId: proto.String(testClientID),
			FilePath: proto.String(testFilePath),
		},
	}}

	if w, g := wantCalls, conn.calls; len(g) != len(w) ||
		g[0].handlerName != w[0].handlerName ||
		!proto.Equal(g[0].args.(proto.Message), w[0].args.(proto.Message)) {
		t.Errorf("the arguments of GetFileDetails diff (want: [%q] got: [%q])", w, g)
	}
}

func TestGetBlob(t *testing.T) {
	file, conn := createTestFile()

	// No optional argument passed.
	rc, err := file.GetBlob(context.Background())
	if err != nil {
		t.Fatalf("GetBlob for files (without optional argument) failed: %v", err)
	}

	if rc != testReadCloser {
		t.Errorf("The readCloser returned from GetFileBlob (without optional argument) is %v; want: %v", rc, testReadCloser)
	}

	want := []call{
		call{
			handlerName: "GetFileBlob",
			args: &avpb.ApiGetFileBlobArgs{
				ClientId: proto.String(testClientID),
				FilePath: proto.String(testFilePath),
			},
		},
	}

	if w, g := want, conn.calls; len(g) != len(w) ||
		g[0].handlerName != w[0].handlerName ||
		!proto.Equal(g[0].args.(proto.Message), w[0].args.(proto.Message)) {
		t.Errorf("The arguments of GetFileBlob (without optional argument) diff (want: [%q] got: [%q])", w, g)
	}

	// Optional arguments are passed.
	rc, err = file.GetBlob(context.Background(), BlobTimestamp(1491235296))
	if err != nil {
		t.Fatalf("GetBlob for files (with optional argument) failed: %v", err)
	}

	if rc != testReadCloser {
		t.Errorf("The readCloser returned from GetFileBlob (with optional argument) is %v; want: %v", rc, testReadCloser)
	}

	want = append(want, call{
		handlerName: "GetFileBlob",
		args: &avpb.ApiGetFileBlobArgs{
			ClientId:  proto.String(testClientID),
			FilePath:  proto.String(testFilePath),
			Timestamp: proto.Uint64(1491235296),
		},
	})

	if w, g := want, conn.calls; len(g) != len(w) ||
		g[0].handlerName != w[0].handlerName ||
		!proto.Equal(g[0].args.(proto.Message), w[0].args.(proto.Message)) {
		t.Errorf("The arguments of GetFileBlob (with optional argument) diff (want: [%q] got: [%q])", w, g)
	}
}

func TestFilesArchive(t *testing.T) {
	file, conn := createTestFile()

	rc, err := file.GetFilesArchive(context.Background())
	if err != nil {
		t.Fatalf("GetFilesArchive for files failed: %v", err)
	}

	if rc != testReadCloser {
		t.Errorf("The readCloser returned from GetVFSFilesArchive is %v; want: %v", rc, testReadCloser)
	}

	want := []call{
		call{
			handlerName: "GetVFSFilesArchive",
			args: &avpb.ApiGetVfsFilesArchiveArgs{
				ClientId: proto.String(testClientID),
				FilePath: proto.String(testFilePath),
			},
		},
	}

	if w, g := want, conn.calls; len(g) != len(w) ||
		g[0].handlerName != w[0].handlerName ||
		!proto.Equal(g[0].args.(proto.Message), w[0].args.(proto.Message)) {
		t.Errorf("The arguments of GetVFSFilesArchive diff (want: [%q] got: [%q])", w, g)
	}
}

func TestGetVersionTimes(t *testing.T) {
	file, conn := createTestFile()

	got, err := file.GetVersionTimes(context.Background())
	if err != nil {
		t.Fatalf("GetVersionTimes for files failed: %v", err)
	}

	want := []uint64{0, 1, 100}

	if w, g := want, got; !reflect.DeepEqual(w, g) {
		t.Errorf("unexpected timestamps got; diff (want: [%q] got: [%q])", w, g)
	}

	wantCalls := []call{{
		handlerName: "GetFileVersionTimes",
		args: &avpb.ApiGetFileVersionTimesArgs{
			ClientId: proto.String(testClientID),
			FilePath: proto.String(testFilePath),
		},
	}}

	if w, g := wantCalls, conn.calls; len(g) != len(w) ||
		g[0].handlerName != w[0].handlerName ||
		!proto.Equal(g[0].args.(proto.Message), w[0].args.(proto.Message)) {
		t.Errorf("the arguments of GetFileVersionTimes diff (want: [%q] got: [%q])", w, g)
	}
}

func TestRefresh(t *testing.T) {
	file, conn := createTestFile()

	ro, err := file.Refresh(context.Background())
	if err != nil {
		t.Fatalf("Refresh for files failed: %v", err)
	}

	wantRO := &RefreshOperation{
		conn:        conn,
		ClientID:    testClientID,
		OperationID: testOperationID,
	}

	if w, g := wantRO, ro; !reflect.DeepEqual(w, g) {
		t.Errorf("unexpected refresh operation got; diff (want: [%q] got: [%q])", w, g)
	}

	wantCalls := []call{{
		handlerName: "CreateVFSRefreshOperation",
		args: &avpb.ApiCreateVfsRefreshOperationArgs{
			ClientId: proto.String(testClientID),
			FilePath: proto.String(testFilePath),
		},
	}}

	if w, g := wantCalls, conn.calls; len(g) != len(w) ||
		g[0].handlerName != w[0].handlerName ||
		!proto.Equal(g[0].args.(proto.Message), w[0].args.(proto.Message)) {
		t.Errorf("the arguments of CreateVFSRefreshOperation diff (want: [%q] got: [%q])", w, g)
	}
}

func TestRefreshOperationGetState(t *testing.T) {
	conn := &mockConnector{}
	ro := &RefreshOperation{
		conn:        conn,
		ClientID:    testClientID,
		OperationID: testOperationID,
	}

	state, err := ro.GetState(context.Background())
	if err != nil {
		t.Fatalf("GetState for refresh operation failed: %v", err)
	}

	wantState := avpb.ApiGetVfsRefreshOperationStateResult_FINISHED

	if w, g := wantState, state; w != g {
		t.Errorf("unexpected state of refresh operation got; diff (want: [%q] got: [%q])", w, g)
	}

	wantCalls := []call{{
		handlerName: "GetVFSRefreshOperationState",
		args: &avpb.ApiGetVfsRefreshOperationStateArgs{
			ClientId:    proto.String(testClientID),
			OperationId: proto.String(testOperationID),
		},
	}}

	if w, g := wantCalls, conn.calls; len(g) != len(w) ||
		g[0].handlerName != w[0].handlerName ||
		!proto.Equal(g[0].args.(proto.Message), w[0].args.(proto.Message)) {
		t.Errorf("the arguments of GetVFSRefreshOperationState diff (want: [%q] got: [%q])", w, g)
	}
}

func TestCollect(t *testing.T) {
	file, conn := createTestFile()

	co, err := file.Collect(context.Background())
	if err != nil {
		t.Fatalf("Collect for files failed: %v", err)
	}

	wantCO := &CollectOperation{
		conn:        conn,
		ClientID:    testClientID,
		OperationID: testOperationID,
	}

	if w, g := wantCO, co; !reflect.DeepEqual(w, g) {
		t.Errorf("unexpected collect operation got; diff (want: [%q] got: [%q])", w, g)
	}

	wantCalls := []call{{
		handlerName: "UpdateVFSFileContent",
		args: &avpb.ApiUpdateVfsFileContentArgs{
			ClientId: proto.String(testClientID),
			FilePath: proto.String(testFilePath),
		},
	}}

	if w, g := wantCalls, conn.calls; len(g) != len(w) ||
		g[0].handlerName != w[0].handlerName ||
		!proto.Equal(g[0].args.(proto.Message), w[0].args.(proto.Message)) {
		t.Errorf("the arguments of UpdateVFSFileContent diff (want: [%q] got: [%q])", w, g)
	}
}

func TestCollectOperationGetState(t *testing.T) {
	conn := &mockConnector{}
	co := &CollectOperation{
		conn:        conn,
		ClientID:    testClientID,
		OperationID: testOperationID,
	}

	state, err := co.GetState(context.Background())
	if err != nil {
		t.Fatalf("GetState for collect operation failed: %v", err)
	}

	wantState := avpb.ApiGetVfsFileContentUpdateStateResult_FINISHED

	if w, g := wantState, state; w != g {
		t.Errorf("unexpected state of collect operation got; diff (want: [%q] got: [%q])", w, g)
	}

	wantCalls := []call{{
		handlerName: "GetVFSFileContentUpdateState",
		args: &avpb.ApiGetVfsFileContentUpdateStateArgs{
			ClientId:    proto.String(testClientID),
			OperationId: proto.String(testOperationID),
		},
	}}

	if w, g := wantCalls, conn.calls; len(g) != len(w) ||
		g[0].handlerName != w[0].handlerName ||
		!proto.Equal(g[0].args.(proto.Message), w[0].args.(proto.Message)) {
		t.Errorf("the arguments of GetVFSFileContentUpdateState diff (want: [%q] got: [%q])", w, g)
	}
}

func TestGetTimelineAsCSV(t *testing.T) {
	file, conn := createTestFile()

	rc, err := file.GetTimelineAsCSV(context.Background())
	if err != nil {
		t.Fatalf("GetTimelineAsCSV for files failed: %v", err)
	}

	if rc != testReadCloser {
		t.Errorf("The readCloser returned from GetTimelineAsCSV is %v; want: %v", rc, testReadCloser)
	}

	want := []call{
		call{
			handlerName: "GetVFSTimelineAsCSV",
			args: &avpb.ApiGetVfsTimelineAsCsvArgs{
				ClientId: proto.String(testClientID),
				FilePath: proto.String(testFilePath),
			},
		},
	}

	if w, g := want, conn.calls; len(g) != len(w) ||
		g[0].handlerName != w[0].handlerName ||
		!proto.Equal(g[0].args.(proto.Message), w[0].args.(proto.Message)) {
		t.Errorf("The arguments of GetVFSTimelineAsCSV diff (want: [%q] got: [%q])", w, g)
	}
}

func TestGetTimeline(t *testing.T) {
	file, conn := createTestFile()

	ti, err := file.GetTimeline(context.Background())
	if err != nil {
		t.Fatalf("GetTimeline for a VFS file failed: %v", err)
	}

	want := []*avpb.ApiVfsTimelineItem{
		{FilePath: proto.String(testFilePath)},
		{FilePath: proto.String(testFilePath)},
		{FilePath: proto.String(testFilePath)},
	}

	if w, g := want, ti; len(w) != len(g) {
		t.Errorf("unexpected file got; diff (want: [%q] got: [%q])", w, g)
	} else {
		for i := 0; i < len(w); i++ {
			if !proto.Equal(w[i], g[i]) {
				t.Errorf("unexpected file got; diff (want: [%q] got: [%q])", w, g)
			}
		}
	}

	wantCalls := []call{{
		handlerName: "GetVFSTimeline",
		args: &avpb.ApiGetVfsTimelineArgs{
			ClientId: proto.String(testClientID),
			FilePath: proto.String(testFilePath),
		},
	}}

	if w, g := wantCalls, conn.calls; len(g) != len(w) ||
		g[0].handlerName != w[0].handlerName ||
		!proto.Equal(g[0].args.(proto.Message), w[0].args.(proto.Message)) {
		t.Errorf("the arguments of GetVFSTimeline diff (want: [%q] got: [%q])", w, g)
	}
}

func TestListFiles(t *testing.T) {
	file, conn := createTestFile()

	var got []*File
	if err := file.ListFiles(context.Background(), func(f *File) error {
		got = append(got, f)
		return nil
	}); err != nil {
		t.Fatalf("ListFiles for a VFS file failed: %v", err)
	}

	item := &File{
		conn:     conn,
		ClientID: testClientID,
		Path:     testFilePath,
		Data: &avpb.ApiFile{
			Path: proto.String(testFilePath),
		},
	}

	want := []*File{item, item, item}

	if w, g := want, got; len(w) != len(g) {
		t.Errorf("results of ListFiles diff (want: [%q] got: [%q])", w, g)
	} else {
		for i := 0; i < len(w); i++ {
			if w[i].ClientID != g[i].ClientID ||
				w[i].Path != g[i].Path ||
				!proto.Equal(w[i].Data, g[i].Data) {
				t.Errorf("results of ListFiles diff (want: [%q] got: [%q])", w, g)
			}
		}
	}

	wantCalls := []call{{
		handlerName: "ListFiles",
		args: &avpb.ApiListFilesArgs{
			ClientId: proto.String(testClientID),
			FilePath: proto.String(testFilePath),
		},
	}}

	if w, g := wantCalls, conn.calls; len(g) != len(w) ||
		g[0].handlerName != w[0].handlerName ||
		!proto.Equal(g[0].args.(proto.Message), w[0].args.(proto.Message)) {
		t.Errorf("the arguments of ListFiles diff (want: [%q] got: [%q])", w, g)
	}
}
