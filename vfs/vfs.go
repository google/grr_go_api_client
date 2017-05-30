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

// Package vfs provides a GRR client object which handles associated API calls via an arbitrary transfer protocol.
package vfs

import (
	"io"

	"context"
	"github.com/golang/protobuf/proto"
	avpb "github.com/google/grr_go_api_client/grrproto/grr/proto/api/vfs_proto"
)

// Connector abstracts GRR communication protocols (i.e. HTTP, Stubby).
type Connector interface {
	GetFileDetails(ctx context.Context, args *avpb.ApiGetFileDetailsArgs) (*avpb.ApiGetFileDetailsResult, error)
	GetVFSFilesArchive(ctx context.Context, args *avpb.ApiGetVfsFilesArchiveArgs) (io.ReadCloser, error)
	GetFileVersionTimes(ctx context.Context, args *avpb.ApiGetFileVersionTimesArgs) (*avpb.ApiGetFileVersionTimesResult, error)
	GetVFSRefreshOperationState(ctx context.Context, args *avpb.ApiGetVfsRefreshOperationStateArgs) (*avpb.ApiGetVfsRefreshOperationStateResult, error)
	CreateVFSRefreshOperation(ctx context.Context, args *avpb.ApiCreateVfsRefreshOperationArgs) (*avpb.ApiCreateVfsRefreshOperationResult, error)
	UpdateVFSFileContent(ctx context.Context, args *avpb.ApiUpdateVfsFileContentArgs) (*avpb.ApiUpdateVfsFileContentResult, error)
	GetVFSTimelineAsCSV(ctx context.Context, args *avpb.ApiGetVfsTimelineAsCsvArgs) (io.ReadCloser, error)
	GetVFSFileContentUpdateState(ctx context.Context, args *avpb.ApiGetVfsFileContentUpdateStateArgs) (*avpb.ApiGetVfsFileContentUpdateStateResult, error)
	GetFileBlob(ctx context.Context, args *avpb.ApiGetFileBlobArgs) (io.ReadCloser, error)
	GetVFSTimeline(ctx context.Context, args *avpb.ApiGetVfsTimelineArgs) (*avpb.ApiGetVfsTimelineResult, error)
	ListFiles(ctx context.Context, args *avpb.ApiListFilesArgs, fn func(m *avpb.ApiFile) error) error
}

// File object is a helper object to send requests to GRR server via an arbitrary transfer protocol.
type File struct {
	conn     Connector
	ClientID string
	Path     string
	Data     *avpb.ApiFile
}

// NewFile creates a new File object with given data.
func NewFile(conn Connector, clientID string, path string) *File {
	return &File{
		conn:     conn,
		ClientID: clientID,
		Path:     path,
	}
}

// Get sends request to GRR server to return file details for the file. If an error occurs, Data field will stay unchanged.
func (f *File) Get(ctx context.Context) error {
	data, err := f.conn.GetFileDetails(ctx, &avpb.ApiGetFileDetailsArgs{
		ClientId: proto.String(f.ClientID),
		FilePath: proto.String(f.Path),
	})

	if err != nil {
		return err
	}

	f.Data = data.GetFile()
	return nil
}

// GetFilesArchive sends request to GRR server to stream an archive of all the files in the VFS that are stored under the current file's path. This function starts a goroutine that runs until all data is read, or the readcloser is closed, or an error occurs in the go routine.
func (f *File) GetFilesArchive(ctx context.Context) (io.ReadCloser, error) {
	return f.conn.GetVFSFilesArchive(ctx, &avpb.ApiGetVfsFilesArchiveArgs{
		ClientId: proto.String(f.ClientID),
		FilePath: proto.String(f.Path),
	})
}

// BlobTimestamp is an optional argument that can be passed into GetBlob to restrict timestamp (in microseconds) of the request. Example usage could look like this: f.GetBlob(ctx, file.RestrictTimestamp(1491235296)).
func BlobTimestamp(timestamp uint64) func(args *avpb.ApiGetFileBlobArgs) {
	return func(args *avpb.ApiGetFileBlobArgs) {
		args.Timestamp = proto.Uint64(timestamp)
	}
}

// GetBlob sends request to GRR server to get byte contents of the VFS file. This function starts a goroutine that runs until all data is read, or the readcloser is closed, or an error occurs in the go routine.
func (f *File) GetBlob(ctx context.Context, options ...func(args *avpb.ApiGetFileBlobArgs)) (io.ReadCloser, error) {
	args := &avpb.ApiGetFileBlobArgs{
		ClientId: proto.String(f.ClientID),
		FilePath: proto.String(f.Path),
	}

	for _, option := range options {
		option(args)
	}

	return f.conn.GetFileBlob(ctx, args)
}

// GetVersionTimes sends request to GRR server to return a list of timestamps (in microseconds) corresponding to different historic versions of the file.
func (f *File) GetVersionTimes(ctx context.Context) ([]uint64, error) {
	result, err := f.conn.GetFileVersionTimes(ctx, &avpb.ApiGetFileVersionTimesArgs{
		ClientId: proto.String(f.ClientID),
		FilePath: proto.String(f.Path),
	})

	if err != nil {
		return nil, err
	}

	return result.GetTimes(), nil
}

// RefreshOperation wrappers refresh operations for VFS.
type RefreshOperation struct {
	conn        Connector
	ClientID    string
	OperationID string
}

// GetState sends request to GRR server to return a status of a "VFS refresh" of the operation.
func (o *RefreshOperation) GetState(ctx context.Context) (avpb.ApiGetVfsRefreshOperationStateResult_State, error) {
	result, err := o.conn.GetVFSRefreshOperationState(ctx, &avpb.ApiGetVfsRefreshOperationStateArgs{
		ClientId:    proto.String(o.ClientID),
		OperationId: proto.String(o.OperationID),
	})

	if err != nil {
		return avpb.ApiGetVfsRefreshOperationStateResult_RUNNING, err
	}

	return result.GetState(), nil
}

// Refresh sends request to GRR server to start a VFS "refresh" operation, that will refresh the stats data about files and directories at current VFS path on current client.
func (f *File) Refresh(ctx context.Context) (*RefreshOperation, error) {
	result, err := f.conn.CreateVFSRefreshOperation(ctx, &avpb.ApiCreateVfsRefreshOperationArgs{
		ClientId: proto.String(f.ClientID),
		FilePath: proto.String(f.Path),
	})

	if err != nil {
		return nil, err
	}

	return &RefreshOperation{
		conn:        f.conn,
		ClientID:    f.ClientID,
		OperationID: result.GetOperationId(),
	}, nil
}

// CollectOperation wrappers collect operations for VFS.
type CollectOperation struct {
	conn        Connector
	ClientID    string
	OperationID string
}

// GetState sends request to GRR server to get the state of the current operation.
func (o *CollectOperation) GetState(ctx context.Context) (avpb.ApiGetVfsFileContentUpdateStateResult_State, error) {
	result, err := o.conn.GetVFSFileContentUpdateState(ctx, &avpb.ApiGetVfsFileContentUpdateStateArgs{
		ClientId:    proto.String(o.ClientID),
		OperationId: proto.String(o.OperationID),
	})

	if err != nil {
		return avpb.ApiGetVfsFileContentUpdateStateResult_RUNNING, err
	}

	return result.GetState(), nil
}

// Collect sends request to GRR server to start a long-running operation that re-downloads the file from the GRR client.
func (f *File) Collect(ctx context.Context) (*CollectOperation, error) {
	result, err := f.conn.UpdateVFSFileContent(ctx, &avpb.ApiUpdateVfsFileContentArgs{
		ClientId: proto.String(f.ClientID),
		FilePath: proto.String(f.Path),
	})

	if err != nil {
		return nil, err
	}

	return &CollectOperation{
		conn:        f.conn,
		ClientID:    f.ClientID,
		OperationID: result.GetOperationId(),
	}, nil
}

// GetTimelineAsCSV sends request to GRR server to stream timeline information for the file in a CSV format. This function starts a goroutine that runs until all data is read, or the readcloser is closed, or an error occurs in the go routine.
func (f *File) GetTimelineAsCSV(ctx context.Context) (io.ReadCloser, error) {
	return f.conn.GetVFSTimelineAsCSV(ctx, &avpb.ApiGetVfsTimelineAsCsvArgs{
		ClientId: proto.String(f.ClientID),
		FilePath: proto.String(f.Path),
	})
}

// GetTimeline send a request to GRR server to get event timeline of VFS events
// for a given VFS path.
func (f *File) GetTimeline(ctx context.Context) ([]*avpb.ApiVfsTimelineItem, error) {
	result, err := f.conn.GetVFSTimeline(ctx, &avpb.ApiGetVfsTimelineArgs{
		ClientId: proto.String(f.ClientID),
		FilePath: proto.String(f.Path),
	})
	if err != nil {
		return nil, err
	}

	return result.GetItems(), nil
}

// ListFiles send a request to GRR server to list files in a given VFS directory
// of a given client and calls fn (a callback function) on each File object.
func (f *File) ListFiles(ctx context.Context, fn func(f *File) error) error {
	return f.conn.ListFiles(ctx, &avpb.ApiListFilesArgs{
		ClientId: proto.String(f.ClientID),
		FilePath: proto.String(f.Path),
	},
		func(item *avpb.ApiFile) error {
			return fn(&File{
				conn:     f.conn,
				ClientID: f.ClientID,
				Path:     item.GetPath(),
				Data:     item,
			})
		})
}
