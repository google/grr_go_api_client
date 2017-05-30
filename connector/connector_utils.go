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

package connector

import (
	"fmt"
	"reflect"

	"github.com/golang/protobuf/proto"
	"github.com/google/grr_go_api_client/errors"
)

func getItems(args proto.Message) ([]proto.Message, error) {
	if args == nil {
		return nil, fmt.Errorf("input to getItems is nil")
	}

	v := reflect.ValueOf(args).Elem()
	f := v.FieldByName("Items")

	if f.Kind() != reflect.Slice {
		return nil, fmt.Errorf("Items is not a slice")
	}

	items := make([]proto.Message, f.Len())
	for i := 0; i < f.Len(); i++ {
		items[i] = f.Index(i).Interface().(proto.Message)
	}
	return items, nil
}

// Callback is a func type that can be passed as an argument for SendIteratorRequest. sendIterator will call Callback on each item from results.
type Callback func(item proto.Message) error

type offsetGetter interface {
	GetOffset() int64
	proto.Message
}

// sendRequest is passed to SendIteratorRequest and should send a single request for count records starting at offset. It should return a proto with a field named 'items'. If the proto also has a field named 'total_count', totalCount (the second return value with type int64) will be set to that value; otherwise, it will be set to -1 to indicate the field does not exist.
type sendRequest func(offset, count int64) (proto.Message, int64, error)

func generatePages(sr sendRequest, fn Callback, offset, pageSize int64) error {
	for {
		// Ignore the total count returned here, because we get the value from the first call.
		result, _, err := sr(offset, pageSize)
		if err != nil {
			return err
		}

		items, err := getItems(result)
		if err != nil {
			return err
		}

		if len(items) == 0 {
			return nil
		}

		for _, item := range items {
			if err := fn(item); err != nil {
				return err
			}
		}

		offset += pageSize
	}
}

// SendIteratorRequest takes an argument, a function sr that sends a single request to GRR server, and a callback fn that applies to each result item returned by sr, and pageSize of the connector.
func SendIteratorRequest(args offsetGetter, sr sendRequest, fn Callback, pageSize int64) (int64, error) {
	offset := args.GetOffset()
	result, totalCount, err := sr(offset, pageSize)
	if err != nil {
		return totalCount, err
	}

	items, err := getItems(result)
	if err != nil {
		return totalCount, err
	}

	for _, item := range items {
		if err := fn(item); err != nil {
			return totalCount, err
		}
	}

	return totalCount, generatePages(sr, fn, offset+pageSize, pageSize)
}

// ConvertUserError converts an user error to ErrorWithKind.
func ConvertUserError(err error) ErrorWithKind {
	if err == nil {
		return nil
	}

	return errors.NewGRRAPIError(errors.UserCallbackError, err.Error())
}

// ErrorWithKind is an interface that wraps an error and its ErrorKind.
type ErrorWithKind interface {
	error
	Kind() errors.ErrorKind
}
