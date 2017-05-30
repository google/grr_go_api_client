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

package errors

import (
	"fmt"
	"testing"
)

type testErrorWithKind struct {
	err  string
	kind ErrorKind
}

func (e *testErrorWithKind) Error() string {
	return e.err
}

func (e *testErrorWithKind) Kind() ErrorKind {
	return e.kind
}

func createTestErrorWithKind(err string, kind ErrorKind) *testErrorWithKind {
	return &testErrorWithKind{
		err:  err,
		kind: kind,
	}
}

func TestKind(t *testing.T) {
	var err error

	if kind := Kind(err); kind != Invalid {
		t.Errorf("error has Kind %s; want: Invalid", kind.String())
	}

	err = fmt.Errorf("error should has Kind GenericError")
	if kind := Kind(err); kind != GenericError {
		t.Errorf("error has Kind %s; want: GenericError", kind.String())
	}

	errWithKind := createTestErrorWithKind("test", UnknownServerError)
	if kind := Kind(errWithKind); kind != UnknownServerError {
		t.Errorf("error has Kind %s; want: UnknownServerError", kind.String())
	}
}
