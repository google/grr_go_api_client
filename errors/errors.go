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

// Package errors defines objects that provide detailed information about GRR API errors.
package errors

// ErrorKind categorizes errors into different kinds.
type ErrorKind int

const (
	// Invalid indicates no error.
	Invalid ErrorKind = iota

	// UserCallbackError indicates an error inside a callback function passed by users.
	UserCallbackError

	// ResourceNotFoundError indicates resource is not found on GRR server.
	ResourceNotFoundError

	// AccessForbiddenError indicates resource access is forbidden on GRR server.
	AccessForbiddenError

	// APINotImplementedError indicates the API method is not implemented on GRR server.
	APINotImplementedError

	// UnknownServerError indicates other errors occur on GRR server.
	UnknownServerError

	// GenericError indicates an unknown error.
	GenericError
)

// String returns the description of an ErrorKind.
func (e ErrorKind) String() string {
	switch e {
	case UserCallbackError:
		return "UserCallbackError"
	case ResourceNotFoundError:
		return "ResourceNotFoundError"
	case AccessForbiddenError:
		return "AccessForbiddenError"
	case APINotImplementedError:
		return "APINotImplementedError"
	case UnknownServerError:
		return "UnknownServerError"
	case GenericError:
		return "GenericError"
	default:
		return "Invalid"
	}
}

// GRRAPIError wraps API errors with ErrorKind.
type GRRAPIError struct {
	kind ErrorKind
	msg  string
}

// NewGRRAPIError returns a grrAPIError with given msg string and ErrorKind.
func NewGRRAPIError(kind ErrorKind, msg string) *GRRAPIError {
	return &GRRAPIError{
		kind: kind,
		msg:  msg,
	}
}

// Error returns the string that describes the error.
func (e *GRRAPIError) Error() string {
	return e.msg
}

// Kind returns the ErrorKind of the error.
func (e *GRRAPIError) Kind() ErrorKind {
	return e.kind
}

// Kind returns Invalid if the err is nil. For a non-nil error, the function returns the value of method Kind() if it has one, or GenericError otherwise.
func Kind(err error) ErrorKind {
	if err == nil {
		return Invalid
	}

	type kind interface {
		Kind() ErrorKind
	}

	e, ok := err.(kind)

	if ok {
		return e.Kind()
	}

	return GenericError
}
