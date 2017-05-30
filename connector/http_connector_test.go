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
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/google/grr_go_api_client/connectortestutils"
	"github.com/google/grr_go_api_client/errors"

	anypb "github.com/golang/protobuf/ptypes/any"
	acpb "github.com/google/grr_go_api_client/grrproto/grr/proto/api/client_proto"
	afpb "github.com/google/grr_go_api_client/grrproto/grr/proto/api/flow_proto"
	ahpb "github.com/google/grr_go_api_client/grrproto/grr/proto/api/hunt_proto"
	arpb "github.com/google/grr_go_api_client/grrproto/grr/proto/api/reflection_proto"
	aupb "github.com/google/grr_go_api_client/grrproto/grr/proto/api/user_proto"
	avpb "github.com/google/grr_go_api_client/grrproto/grr/proto/api/vfs_proto"
)

const (
	testCSRFToken = "testCSRFToken"
	testUsername  = "testUsername"
)

func TestCoerceValueToQueryStringType(t *testing.T) {
	field0 := "teststring"
	field1 := uint64(123456)
	field2 := int64(-123456)
	field3 := afpb.ApiGetFlowFilesArchiveArgs_TAR_GZ.Enum()
	field4 := []string{"teststring"}

	tests := []struct {
		desc   string
		field  interface{}
		output string
		err    error
	}{
		{"convert *string to string", &field0, "teststring", nil},
		{"convert *uint64 to string", &field1, "123456", nil},
		{"convert *int64  to string", &field2, "-123456", nil},
		{"convert enum value to string", field3, "TAR_GZ", nil},
		{"convert invalid type to string", field4, "", fmt.Errorf("unknown type")},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s: field needed to be converted is %v, and the expected result is %v", test.desc, test.field, test.output), func(t *testing.T) {
			output, err := coerceValueToQueryStringType(test.field)
			if output != test.output {
				t.Errorf("coerceValueToQueryStringType has an output %v; want %v", output, test.output)
			}

			if test.err == nil && err != nil {
				t.Errorf("coerceValueToQueryStringType has an error %v, but it should not produce an error", err)
			} else if test.err != nil && err == nil {
				t.Errorf("coerceValueToQueryStringType should have an error %v, but it doesn't", test.err)
			} else if test.err != nil && err != nil {
				if !strings.Contains(err.Error(), test.err.Error()) {
					t.Errorf("coerceValueToQueryStringType has an error %v, but it should be %v", err, test.err)
				}
			}
		})
	}
}

func TestBuildRoutingMap(t *testing.T) {
	data := &arpb.ApiListApiMethodsResult{
		Items: []*arpb.ApiMethod{
			&arpb.ApiMethod{
				Name:        proto.String("testAPIMethod1"),
				HttpRoute:   proto.String("/api/test/HTTPRoute1/<param1>/<path:param2>"),
				HttpMethods: []string{"GET"},
			},
			&arpb.ApiMethod{
				Name:        proto.String("testAPIMethod2"),
				HttpRoute:   proto.String("/api/v2/test/HTTPRoute2/noparam"),
				HttpMethods: []string{"POST"},
			},
		},
	}

	gotAPIMethods := make(map[string]*arpb.ApiMethod)
	gotURLVars := make(map[string][]string)

	err := buildRoutingMap(data, gotAPIMethods, gotURLVars)
	if err != nil {
		t.Fatalf("buildRoutingMap fails:\n%s", err)
	}

	wantAPIMethods := make(map[string]*arpb.ApiMethod)
	wantAPIMethods["testAPIMethod1"] = &arpb.ApiMethod{
		Name:        proto.String("testAPIMethod1"),
		HttpRoute:   proto.String("/api/v2/test/HTTPRoute1/<param1>/<path:param2>"),
		HttpMethods: []string{"GET"},
	}

	wantAPIMethods["testAPIMethod2"] = &arpb.ApiMethod{
		Name:        proto.String("testAPIMethod2"),
		HttpRoute:   proto.String("/api/v2/test/HTTPRoute2/noparam"),
		HttpMethods: []string{"POST"},
	}

	for k, v := range wantAPIMethods {
		if w, g := v, gotAPIMethods[k]; !proto.Equal(w, g) {
			t.Errorf("apiMethods of buildRoutingMap diff (want: [%q] got: [%q])", w, g)
		}
	}

	wantURLVars := make(map[string][]string)
	wantURLVars["testAPIMethod1"] = []string{"param1", "path:param2"}
	wantURLVars["testAPIMethod2"] = []string{}

	for k, v := range wantURLVars {
		if w, g := v, gotURLVars[k]; !reflect.DeepEqual(w, g) {
			t.Errorf("urlVars of buildRoutingMap diff (want: [%q] got: [%q])", w, g)
		}
	}
}

func TestGetMethodURLAndPathParamsNames(t *testing.T) {
	// Test when there is not exactly one HttpMethods: the function should catch the err.
	apiMethod := &arpb.ApiMethod{
		Name:        proto.String("GetFlowFilesArchiveArgs"),
		HttpRoute:   proto.String("/api/v2/clients/<client_id>/flows/<path:flow_id>/results/files-archive"),
		HttpMethods: []string{"GET", "POST"},
	}

	urlVars := []string{"client_id", "path:flow_id"}
	args := &afpb.ApiGetFlowFilesArchiveArgs{
		ClientId:      proto.String("testClientID"),
		FlowId:        proto.String("testFlowID"),
		ArchiveFormat: afpb.ApiGetFlowFilesArchiveArgs_ZIP.Enum(),
	}

	_, err := getMethodURLAndPathParamsNames(apiMethod, urlVars, args)

	want := fmt.Errorf("method %s has HTTP methods(s): %v ; want exactly one method", apiMethod.Name, apiMethod.HttpMethods)

	if err == nil {
		t.Errorf("error returend from GetMethodURLAndPathParamsNames is nil, but it should be %v", want)
	} else if !strings.Contains(err.Error(), want.Error()) {
		t.Errorf("error returend from GetMethodURLAndPathParamsNames is %v, but it should be %v", err, want)
	}

	// Function should not produce an error here.
	apiMethod.HttpMethods = []string{"GET"}
	gotInfo, err := getMethodURLAndPathParamsNames(apiMethod, urlVars, args)

	if err != nil {
		t.Errorf("error returend from GetMethodURLAndPathParamsNames is %v, but it should be nil", err)
	}

	wantInfo := &requestInfo{
		method:          "GET",
		url:             "/api/v2/clients/testClientID/flows/testFlowID/results/files-archive",
		pathParamsNames: []string{"ClientId", "FlowId"},
	}

	if w, g := wantInfo, gotInfo; !reflect.DeepEqual(w, g) {
		t.Errorf("requestInfo returned from GetMethodURLAndPathParamsNames diff (want: [%q] got: [%q])", w, g)
	}
}

func TestAddArgsToQueryParams(t *testing.T) {
	req, err := http.NewRequest("GET", "", nil)
	if err != nil {
		t.Fatalf("Cannnot create a http request %v", err)
	}

	args := &afpb.ApiGetFlowFilesArchiveArgs{
		ClientId:      proto.String("testClientID"),
		FlowId:        proto.String("testFlowID"),
		ArchiveFormat: afpb.ApiGetFlowFilesArchiveArgs_TAR_GZ.Enum(),
	}
	excludeNames := []string{"ClientId", "FlowId"}

	if err = addArgsToQueryParams(req, args, excludeNames); err != nil {
		t.Fatalf("addArgsToQueryParams should not return an error, but it returns %v", err)
	}

	want := make(map[string][]string)
	want["archive_format"] = []string{"TAR_GZ"}

	for k, v := range want {
		if w, g := v, req.URL.Query()[k]; !reflect.DeepEqual(w, g) {
			t.Errorf("URL query produced from addArgsToQueryParams diff (want: [%q] got: [%q])", w, g)
		}
	}
}

func TestArgsToBody(t *testing.T) {
	args := &afpb.ApiGetFlowFilesArchiveArgs{
		ClientId:      proto.String("testClientID"),
		FlowId:        proto.String("testFlowID"),
		ArchiveFormat: afpb.ApiGetFlowFilesArchiveArgs_TAR_GZ.Enum(),
	}
	excludeNames := []string{"ClientId"}

	reader, err := argsToBody(args, excludeNames)
	if err != nil {
		t.Fatalf("argsToBody should not return an error, but it returns %v", err)
	}

	want := &afpb.ApiGetFlowFilesArchiveArgs{
		FlowId:        proto.String("testFlowID"),
		ArchiveFormat: afpb.ApiGetFlowFilesArchiveArgs_TAR_GZ.Enum(),
	}

	got := &afpb.ApiGetFlowFilesArchiveArgs{}
	if err := jsonpb.Unmarshal(reader, got); err != nil {
		t.Fatalf("jsonpb.Unmarshal in argsToBody returns an error %v, but it shouldn't", err)
	}

	if w, g := want, got; !proto.Equal(w, g) {
		t.Errorf("json string in io.Reader returned from argsToQueryParams diff (want: [%q] got: [%q])", w, g)
	}
}

type testHTTPServer struct {
	mux       *http.ServeMux
	csrfToken string
}

func (s *testHTTPServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	// "/" pattern matches everything, so we need to check that we're at the root here.
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	cookie := http.Cookie{
		Name:  "csrftoken",
		Value: s.csrfToken,
	}

	http.SetCookie(w, &cookie)
}

func (s *testHTTPServer) checkHeaderOK(w http.ResponseWriter, r *http.Request) bool {
	csrfToken := r.Header.Get("x-csrftoken")
	if csrfToken != s.csrfToken {
		http.Error(w, fmt.Sprintf("HTTP request header doesn't have correct csrf token! have: %v; want: %v", csrfToken, s.csrfToken), http.StatusForbidden)
		return false
	}

	requestWith := r.Header.Get("x-requested-with")
	if requestWith != "XMLHttpRequest" {
		http.Error(w, fmt.Sprintf("HTTP request header doesn't have correct x-requested-with! have: %v; want: %v", requestWith, "XMLHttpRequest"), http.StatusForbidden)
		return false
	}

	cookie, err := r.Cookie("X-CSRF-Token")
	if err != nil {
		http.Error(w, fmt.Sprintf("HTTP request doesn't have cookie named X-CSRF-Token: %v", err), http.StatusForbidden)
		return false
	}

	if cookie.Value != s.csrfToken {
		http.Error(w, fmt.Sprintf("X-CSRF-Token cookie of HTTP request header has value %v; want: %v", cookie.Value, s.csrfToken), http.StatusForbidden)
		return false
	}

	return true
}

func (s *testHTTPServer) checkRequestBodyEmpty(w http.ResponseWriter, r *http.Request, methodName string) bool {
	if r.Body == nil {
		return true
	}

	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("request body of %s has an error %v", methodName, err), http.StatusInternalServerError)
		return false
	}
	if len(bytes) > 0 {
		http.Error(w, fmt.Sprintf("request for %s has an non-nil Body %v; want a nil body", methodName, r.Body), http.StatusInternalServerError)
		return false
	}

	return true
}

func (s *testHTTPServer) writeProtoResult(w http.ResponseWriter, result proto.Message) {
	jsonStr, err := (&jsonpb.Marshaler{}).MarshalToString(result)
	if err != nil {
		http.Error(w, "Error marshalling JSON with jsonpb", http.StatusInternalServerError)
		return
	}

	jsonBytes := []byte(jsonPrefix + jsonStr)
	w.Write(jsonBytes)
}

func (s *testHTTPServer) getQueryInt(w http.ResponseWriter, query url.Values, key string) (int64, bool) {
	value, ok := query[key]
	if !ok {
		http.Error(w, fmt.Sprintf("%s is not found in the query params", key), http.StatusInternalServerError)
		return -1, false
	}

	if len(value) != 1 {
		http.Error(w, fmt.Sprintf("value of %s is a string array with length %v; want: 1", key, len(value)), http.StatusInternalServerError)
		return -1, false
	}

	i, err := strconv.ParseInt(value[0], 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("an error %v occurs when converting query param %s into an int64", err, key), http.StatusInternalServerError)
		return -1, false
	}

	return i, true
}

func (s *testHTTPServer) getSendIteratorRequestOffsetAndCount(w http.ResponseWriter, r *http.Request, handlerName string) (bool, int64, int64) {
	// Check Header.
	if ok := s.checkHeaderOK(w, r); !ok {
		return false, -1, -1
	}

	// Check HTTP method.
	if r.Method != "GET" {
		http.Error(w, fmt.Sprintf("HTTP method for %s is %v; want GET", handlerName, r.Method), http.StatusInternalServerError)
		return false, -1, -1
	}

	// Check query params.
	query := r.URL.Query()
	offset, ok := s.getQueryInt(w, query, "offset")
	if !ok {
		return false, -1, -1
	}
	count, ok := s.getQueryInt(w, query, "count")
	if !ok {
		return false, -1, -1
	}

	if len(query) != 2 {
		http.Error(w, fmt.Sprintf("HTTP query params for %s are %v; want a map has keys 'count' and 'offset'", handlerName, query), http.StatusInternalServerError)
		return false, -1, -1
	}

	return true, offset, count
}

func makeDefaultAny(m proto.Message) *anypb.Any {
	ret, _ := ptypes.MarshalAny(m)
	return ret
}

func (s *testHTTPServer) handleListAPIMethods(w http.ResponseWriter, r *http.Request) {
	if ok := s.checkHeaderOK(w, r); !ok {
		return
	}

	apiMethods := []*arpb.ApiMethod{
		{
			Name:        proto.String("AddClientsLabels"),
			HttpRoute:   proto.String("/api/clients/labels/add"),
			HttpMethods: []string{"POST"},
		},
		{
			Name:        proto.String("GetFlowFilesArchive"),
			HttpRoute:   proto.String("/api/clients/<client_id>/flows/<path:flow_id>/results/files-archive"),
			HttpMethods: []string{"GET"},
		},
		{
			Name:        proto.String("ListHunts"),
			HttpRoute:   proto.String("/api/hunts"),
			HttpMethods: []string{"GET"},
			ResultTypeDescriptor: &arpb.ApiRDFValueDescriptor{
				Default: makeDefaultAny(&ahpb.ApiListHuntsResult{}),
			},
		},
		{
			Name:        proto.String("GetGrrUser"),
			HttpRoute:   proto.String("/api/users/me"),
			HttpMethods: []string{"GET"},
			ResultTypeDescriptor: &arpb.ApiRDFValueDescriptor{
				Default: makeDefaultAny(&aupb.ApiGrrUser{}),
			},
		},
		{
			Name:        proto.String("GetHunt"),
			HttpRoute:   proto.String("/api/hunts/<hunt_id>"),
			HttpMethods: []string{"GET"},
			ResultTypeDescriptor: &arpb.ApiRDFValueDescriptor{
				Default: makeDefaultAny(&ahpb.ApiHunt{}),
			},
		},
		{
			Name:        proto.String("ListHuntResults"),
			HttpRoute:   proto.String("/api/hunts/<hunt_id>/results"),
			HttpMethods: []string{"GET"},
			ResultTypeDescriptor: &arpb.ApiRDFValueDescriptor{
				Default: makeDefaultAny(&ahpb.ApiListHuntResultsResult{}),
			},
		},
		{
			Name:        proto.String("ListFlowResults"),
			HttpRoute:   proto.String("/api/clients/<client_id>/flows/<path:flow_id>/results"),
			HttpMethods: []string{"GET"},
			ResultTypeDescriptor: &arpb.ApiRDFValueDescriptor{
				Default: makeDefaultAny(&afpb.ApiListFlowResultsResult{}),
			},
		},
		{
			Name:        proto.String("ListClientApprovals"),
			HttpRoute:   proto.String("/api/users/me/approvals/client"),
			HttpMethods: []string{"GET"},
			ResultTypeDescriptor: &arpb.ApiRDFValueDescriptor{
				Default: makeDefaultAny(&aupb.ApiListClientApprovalsResult{}),
			},
		},
		{
			Name:        proto.String("ListHuntApprovals"),
			HttpRoute:   proto.String("/api/users/me/approvals/hunt"),
			HttpMethods: []string{"GET"},
			ResultTypeDescriptor: &arpb.ApiRDFValueDescriptor{
				Default: makeDefaultAny(&aupb.ApiListHuntApprovalsResult{}),
			},
		},
		{
			Name:        proto.String("ListFiles"),
			HttpRoute:   proto.String("/api/clients/<client_id>/vfs-index"),
			HttpMethods: []string{"GET"},
			ResultTypeDescriptor: &arpb.ApiRDFValueDescriptor{
				Default: makeDefaultAny(&avpb.ApiListFilesResult{}),
			},
		},
		{
			Name:        proto.String("ListFlows"),
			HttpRoute:   proto.String("/api/clients/<client_id>/flows"),
			HttpMethods: []string{"GET"},
			ResultTypeDescriptor: &arpb.ApiRDFValueDescriptor{
				Default: makeDefaultAny(&afpb.ApiListFlowsResult{}),
			},
		},
		{
			Name:        proto.String("SearchClients"),
			HttpRoute:   proto.String("/api/clients"),
			HttpMethods: []string{"GET"},
			ResultTypeDescriptor: &arpb.ApiRDFValueDescriptor{
				Default: makeDefaultAny(&acpb.ApiSearchClientsResult{}),
			},
		},
	}

	result := &arpb.ApiListApiMethodsResult{
		Items: apiMethods,
	}

	s.writeProtoResult(w, result)
}

func (s *testHTTPServer) handleAddClientsLabels(w http.ResponseWriter, r *http.Request) {
	url := "/api/v2/clients/labels/add"

	// Check Header.
	if ok := s.checkHeaderOK(w, r); !ok {
		return
	}

	// Check HTTP method.
	if r.Method != "POST" {
		http.Error(w, fmt.Sprintf("HTTP method for AddClientsLabels is %v; want POST", r.Method), http.StatusInternalServerError)
		return
	}

	// Check query params.
	query := r.URL.Query()
	if len(query) != 0 {
		http.Error(w, fmt.Sprintf("HTTP query params for AddClientsLabels are %v; want an empty map", query), http.StatusInternalServerError)
		return
	}

	// Check request body.
	got := &acpb.ApiAddClientsLabelsArgs{}
	if err := jsonpb.Unmarshal(r.Body, got); err != nil {
		http.Error(w, fmt.Sprintf("jsonpb unmarshal error for body arguments of %s: %v", url, err), http.StatusInternalServerError)
		return
	}

	want := &acpb.ApiAddClientsLabelsArgs{
		ClientIds: []string{"clientID1", "clientID2"},
		Labels:    []string{"foo", "bar"},
	}

	if !proto.Equal(want, got) {
		http.Error(w, fmt.Sprintf("HTTP request body for %s diff (want: [%q] got: [%q])", url, want, got), http.StatusInternalServerError)
	}

	// No need to write response.
}

func (s *testHTTPServer) handleGetFlowFilesArchive(w http.ResponseWriter, r *http.Request) {
	url := "/api/v2/clients/testClientID/flows/testFlowID/results/files-archive"

	// Check Header.
	if ok := s.checkHeaderOK(w, r); !ok {
		return
	}

	// Check HTTP method.
	if r.Method != "GET" {
		http.Error(w, fmt.Sprintf("HTTP method for GetFlowFilesArchive is %v; want GET", r.Method), http.StatusInternalServerError)
		return
	}

	// Check query params.
	gotQuery := r.URL.Query()

	wantQuery := make(map[string][]string)
	wantQuery["archive_format"] = []string{"ZIP"}
	for k, v := range wantQuery {
		if !reflect.DeepEqual(v, gotQuery[k]) {
			http.Error(w, fmt.Sprintf("HTTP request query params for %s diff (want: [%q] got: [%q])", url, wantQuery, gotQuery), http.StatusInternalServerError)
		}
	}

	// Check request body.
	if ok := s.checkRequestBodyEmpty(w, r, "GetFlowFilesArchive"); !ok {
		return
	}

	// Write response.
	testFlowFilesArchiveChunk := [][]byte{
		{1},
		{2, 3},
		{4, 5, 6},
		{7, 8, 9, 10},
	}

	for _, chunk := range testFlowFilesArchiveChunk {
		w.Write(chunk)
	}
}

func (s *testHTTPServer) handleListHunts(w http.ResponseWriter, r *http.Request) {
	ok, offset, count := s.getSendIteratorRequestOffsetAndCount(w, r, "ListHunts")
	if !ok {
		return
	}

	// Write response.
	items := []*ahpb.ApiHunt{
		{
			Name: proto.String("1"),
		},
		{
			Name: proto.String("2"),
		},
		{
			Name: proto.String("3"),
		},
		{
			Name: proto.String("4"),
		},
	}

	if offset > int64(len(items)) {
		items = items[:0]
	} else if count == -1 || offset+count > int64(len(items)) {
		items = items[offset:]
	} else {
		items = items[offset : offset+count]
	}

	result := &ahpb.ApiListHuntsResult{
		Items:      items,
		TotalCount: proto.Int64(4),
	}

	s.writeProtoResult(w, result)
}

func (s *testHTTPServer) handleListHuntResults(w http.ResponseWriter, r *http.Request) {
	ok, offset, count := s.getSendIteratorRequestOffsetAndCount(w, r, "ListHuntResults")
	if !ok {
		return
	}

	// Write response.
	items := []*ahpb.ApiHuntResult{
		{
			ClientId: proto.String("1"),
		},
		{
			ClientId: proto.String("2"),
		},
		{
			ClientId: proto.String("3"),
		},
		{
			ClientId: proto.String("4"),
		},
	}

	if offset > int64(len(items)) {
		items = items[:0]
	} else if count == -1 || offset+count > int64(len(items)) {
		items = items[offset:]
	} else {
		items = items[offset : offset+count]
	}

	result := &ahpb.ApiListHuntResultsResult{
		Items:      items,
		TotalCount: proto.Int64(4),
	}

	s.writeProtoResult(w, result)
}

func (s *testHTTPServer) handleListFlowResults(w http.ResponseWriter, r *http.Request) {
	ok, offset, count := s.getSendIteratorRequestOffsetAndCount(w, r, "ListFlowResults")
	if !ok {
		return
	}

	// Write response.
	items := []*afpb.ApiFlowResult{
		{
			Timestamp: proto.Uint64(1),
		},
		{
			Timestamp: proto.Uint64(2),
		},
		{
			Timestamp: proto.Uint64(3),
		},
		{
			Timestamp: proto.Uint64(4),
		},
	}

	if offset > int64(len(items)) {
		items = items[:0]
	} else if count == -1 || offset+count > int64(len(items)) {
		items = items[offset:]
	} else {
		items = items[offset : offset+count]
	}

	result := &afpb.ApiListFlowResultsResult{
		Items:      items,
		TotalCount: proto.Int64(4),
	}

	s.writeProtoResult(w, result)
}

func (s *testHTTPServer) handleListClientApprovals(w http.ResponseWriter, r *http.Request) {
	ok, offset, count := s.getSendIteratorRequestOffsetAndCount(w, r, "ListClientApprovals")
	if !ok {
		return
	}

	// Write response.
	items := []*aupb.ApiClientApproval{
		{
			Id: proto.String("1"),
		},
		{
			Id: proto.String("2"),
		},
		{
			Id: proto.String("3"),
		},
		{
			Id: proto.String("4"),
		},
	}

	if offset > int64(len(items)) {
		items = items[:0]
	} else if count == -1 || offset+count > int64(len(items)) {
		items = items[offset:]
	} else {
		items = items[offset : offset+count]
	}

	result := &aupb.ApiListClientApprovalsResult{
		Items: items,
	}

	s.writeProtoResult(w, result)
}

func (s *testHTTPServer) handleListHuntApprovals(w http.ResponseWriter, r *http.Request) {
	ok, offset, count := s.getSendIteratorRequestOffsetAndCount(w, r, "ListHuntApprovals")
	if !ok {
		return
	}

	// Write response.
	items := []*aupb.ApiHuntApproval{
		{
			Id: proto.String("1"),
		},
		{
			Id: proto.String("2"),
		},
		{
			Id: proto.String("3"),
		},
		{
			Id: proto.String("4"),
		},
	}

	if offset > int64(len(items)) {
		items = items[:0]
	} else if count == -1 || offset+count > int64(len(items)) {
		items = items[offset:]
	} else {
		items = items[offset : offset+count]
	}

	result := &aupb.ApiListHuntApprovalsResult{
		Items: items,
	}

	s.writeProtoResult(w, result)

}

func (s *testHTTPServer) handleGetGRRUser(w http.ResponseWriter, r *http.Request) {
	url := "/api/v2/users/me"

	// Check Header.
	if ok := s.checkHeaderOK(w, r); !ok {
		return
	}

	// Check HTTP method.
	if r.Method != "GET" {
		http.Error(w, fmt.Sprintf("HTTP method for GetGRRUser is %v; want GET", r.Method), http.StatusInternalServerError)
		return
	}

	// Check query params.
	query := r.URL.Query()
	if len(query) != 0 {
		http.Error(w, fmt.Sprintf("HTTP query params for %s are %v; want an empty map", url, query), http.StatusInternalServerError)
		return
	}

	// Check request body.
	if ok := s.checkRequestBodyEmpty(w, r, "GetHunt"); !ok {
		return
	}

	// Write response.
	result := &aupb.ApiGrrUser{
		Username: proto.String(testUsername),
	}

	s.writeProtoResult(w, result)
}

func (s *testHTTPServer) handleGetHunt(w http.ResponseWriter, r *http.Request) {
	url := "/api/v2/hunts/testHuntID"
	// Check Header.
	if ok := s.checkHeaderOK(w, r); !ok {
		return
	}

	// Check HTTP method.
	if r.Method != "GET" {
		http.Error(w, fmt.Sprintf("HTTP method for GetHunt is %v; want GET", r.Method), http.StatusInternalServerError)
		return
	}

	// Check query params.
	wantQuery := make(map[string][]string)
	for k, v := range wantQuery {
		if !reflect.DeepEqual(v, r.URL.Query()[k]) {
			http.Error(w, fmt.Sprintf("HTTP request query params for %s diff (want: [%q] got: [%q])", url, wantQuery, r.URL.Query()), http.StatusInternalServerError)
		}
	}

	// Check request body.
	if ok := s.checkRequestBodyEmpty(w, r, "GetHunt"); !ok {
		return
	}

	// Write response.
	s.writeProtoResult(w, &ahpb.ApiHunt{
		Name: proto.String("APIHunt"),
	})
}

func (s *testHTTPServer) handleListFiles(w http.ResponseWriter, r *http.Request) {
	ok, offset, count := s.getSendIteratorRequestOffsetAndCount(w, r, "ListFiles")
	if !ok {
		return
	}

	// Write response.
	items := []*avpb.ApiFile{
		{Path: proto.String("1")},
		{Path: proto.String("2")},
		{Path: proto.String("3")},
		{Path: proto.String("4")},
	}

	if offset > int64(len(items)) {
		items = items[:0]
	} else if count == -1 || offset+count > int64(len(items)) {
		items = items[offset:]
	} else {
		items = items[offset : offset+count]
	}

	result := &avpb.ApiListFilesResult{
		Items: items,
	}

	s.writeProtoResult(w, result)
}

func (s *testHTTPServer) handleListFlows(w http.ResponseWriter, r *http.Request) {
	ok, offset, count := s.getSendIteratorRequestOffsetAndCount(w, r, "ListFlows")
	if !ok {
		return
	}

	// Write response.
	items := []*afpb.ApiFlow{
		{FlowId: proto.String("1")},
		{FlowId: proto.String("2")},
		{FlowId: proto.String("3")},
		{FlowId: proto.String("4")},
	}

	if offset > int64(len(items)) {
		items = items[:0]
	} else if count == -1 || offset+count > int64(len(items)) {
		items = items[offset:]
	} else {
		items = items[offset : offset+count]
	}

	result := &afpb.ApiListFlowsResult{
		Items: items,
	}

	s.writeProtoResult(w, result)
}

func (s *testHTTPServer) handleSearchClients(w http.ResponseWriter, r *http.Request) {
	ok, offset, count := s.getSendIteratorRequestOffsetAndCount(w, r, "SearchClients")
	if !ok {
		return
	}

	// Write response.
	items := []*acpb.ApiClient{
		{ClientId: proto.String("1")},
		{ClientId: proto.String("2")},
		{ClientId: proto.String("3")},
		{ClientId: proto.String("4")},
	}

	if offset > int64(len(items)) {
		items = items[:0]
	} else if count == -1 || offset+count > int64(len(items)) {
		items = items[offset:]
	} else {
		items = items[offset : offset+count]
	}

	result := &acpb.ApiSearchClientsResult{
		Items: items,
	}

	s.writeProtoResult(w, result)
}

func newTestHTTPServer() *testHTTPServer {
	s := &testHTTPServer{
		csrfToken: testCSRFToken,
		mux:       http.NewServeMux(),
	}

	s.mux.HandleFunc("/", s.handleIndex)
	s.mux.HandleFunc("/api/v2/reflection/api-methods", s.handleListAPIMethods)
	s.mux.HandleFunc("/api/v2/clients/labels/add", s.handleAddClientsLabels)
	s.mux.HandleFunc("/api/v2/clients/testClientID/flows/testFlowID/results/files-archive", s.handleGetFlowFilesArchive)
	s.mux.HandleFunc("/api/v2/hunts", s.handleListHunts)
	s.mux.HandleFunc("/api/v2/users/me", s.handleGetGRRUser)
	s.mux.HandleFunc("/api/v2/hunts/testHuntID", s.handleGetHunt)
	s.mux.HandleFunc("/api/v2/hunts/testHuntID/results", s.handleListHuntResults)
	s.mux.HandleFunc("/api/v2/clients/testClientID/flows/testFlowID/results", s.handleListFlowResults)
	s.mux.HandleFunc("/api/v2/users/me/approvals/client", s.handleListClientApprovals)
	s.mux.HandleFunc("/api/v2/users/me/approvals/hunt", s.handleListHuntApprovals)
	s.mux.HandleFunc("/api/v2/clients/testClientID/vfs-index", s.handleListFiles)
	s.mux.HandleFunc("/api/v2/clients/testClientID/flows", s.handleListFlows)
	s.mux.HandleFunc("/api/v2/clients", s.handleSearchClients)

	return s
}

func createHTTPServerAndConnector(pageSize int64) (*httptest.Server, *HTTPConnector) {
	s := newTestHTTPServer()
	srv := httptest.NewServer(s.mux)
	return srv, NewHTTPConnector(srv.URL, pageSize, nil)
}

func TestGetCSRFToken(t *testing.T) {
	srv, conn := createHTTPServerAndConnector(1000)
	defer srv.Close()

	token, err := conn.getCSRFToken()
	if err != nil {
		t.Fatalf("getcsrf token fails: \n%v", err)
	}

	if token != testCSRFToken {
		t.Errorf("result of getCSRFToken is %s; want: %s", token, testCSRFToken)
	}
}

func TestAddClientsLabels(t *testing.T) {
	srv, conn := createHTTPServerAndConnector(1000)
	defer srv.Close()

	if err := conn.AddClientsLabels(context.Background(), &acpb.ApiAddClientsLabelsArgs{
		ClientIds: []string{"clientID1", "clientID2"},
		Labels:    []string{"foo", "bar"},
	}); err != nil {
		t.Fatalf("AddClientsLabels fails: \n%v", err)
	}
}

func TestGetFlowFilesArchive(t *testing.T) {
	srv, conn := createHTTPServerAndConnector(1000)
	defer srv.Close()

	tests := []struct {
		desc      string
		arraySize int
		results   [][]byte
	}{
		{"read one chunk at a time", 1, [][]byte{{1}, {2}, {3}, {4}, {5}, {6}, {7}, {8}, {9}, {10}}},
		{"number of total chunks is divisible by array size", 2, [][]byte{{1, 2}, {3, 4}, {5, 6}, {7, 8}, {9, 10}}},
		{"number of total chunks is not divisible by array size", 3, [][]byte{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}, {10}}},
		{"number of total chunks is less than a single array size", 11, [][]byte{{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}}},
	}

	args := &afpb.ApiGetFlowFilesArchiveArgs{
		ClientId:      proto.String("testClientID"),
		FlowId:        proto.String("testFlowID"),
		ArchiveFormat: afpb.ApiGetFlowFilesArchiveArgs_ZIP.Enum(),
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s : arraySize %v expected results: %v", test.desc, test.arraySize, test.results), func(t *testing.T) {
			reader, err := conn.GetFlowFilesArchive(context.Background(), args)

			if err != nil {
				t.Fatalf("GetFlowFilesArchive fails: \n%v", err)
			}

			var got [][]byte
			for {
				p := make([]byte, test.arraySize)
				n, err := reader.Read(p)
				got = append(got, p[:n])
				if err == io.EOF {
					break
				} else if err != nil {
					t.Errorf("binaryChunkReadCloser.Read failed: %v", err)
					break
				}
			}

			if w, g := test.results, got; !reflect.DeepEqual(w, g) {
				t.Errorf("The output of GetFlowFilesArchive diff (want: [%q] got: [%q])", w, g)
			}

			// Close the reader.
			if err := reader.Close(); err != nil {
				t.Errorf("Closing reader failed: %v", err)
			}
		})
	}

}

func TestSendIteratorRequestWithTotalCount(t *testing.T) {
	srv, conn := createHTTPServerAndConnector(1000)
	defer srv.Close()

	connectortestutils.PartialTestSendIteratorRequestWithTotalCount(t, conn)
}

func TestSendIteratorRequestWithoutTotalCount(t *testing.T) {
	srv, conn := createHTTPServerAndConnector(1000)
	defer srv.Close()

	connectortestutils.PartialTestSendIteratorRequestWithoutTotalCount(t, conn)
}

func TestGetUsername(t *testing.T) {
	srv, conn := createHTTPServerAndConnector(1000)
	defer srv.Close()

	username, err := conn.GetUsername(context.Background())
	if err != nil {
		t.Fatalf("GetUsername fails: \n%v", err)
	}

	if username != testUsername {
		t.Errorf("unexpected username got: %v; want %v", username, testUsername)
	}

	// Test that the username field of connector is also set.
	if conn.username != testUsername {
		t.Errorf("unexpected username of the connector: %v; want %v", conn.username, testUsername)
	}
}

func TestGetHunt(t *testing.T) {
	srv, conn := createHTTPServerAndConnector(1000)
	defer srv.Close()

	hunt, err := conn.GetHunt(context.Background(), &ahpb.ApiGetHuntArgs{
		HuntId: proto.String(connectortestutils.TestHuntID),
	})

	if err != nil {
		t.Fatalf("GetHunt fails: \n%v", err)
	}

	want := &ahpb.ApiHunt{
		Name: proto.String("APIHunt"),
	}

	if w, g := want, hunt; !proto.Equal(w, g) {
		t.Errorf("the result of GetHunt diff (want: [%q] got: [%q])", w, g)
	}
}

func convertHTTPStatusCode(errorKind errors.ErrorKind) int {
	switch errorKind {
	case errors.ResourceNotFoundError:
		return http.StatusNotFound
	case errors.AccessForbiddenError:
		return http.StatusForbidden
	case errors.APINotImplementedError:
		return http.StatusNotImplemented
	case errors.UnknownServerError:
		return http.StatusInternalServerError
	default:
		return http.StatusOK
	}
}

type testErrorHTTPServer struct {
	testHTTPServer
	statusCode             int
	csrfTokenHasError      bool
	listAPIMethodsHasError bool
}

func newTestErrorHTTPServer(errorKind errors.ErrorKind) *testErrorHTTPServer {
	s := &testErrorHTTPServer{
		testHTTPServer: testHTTPServer{
			csrfToken: testCSRFToken,
			mux:       http.NewServeMux(),
		},
		statusCode: convertHTTPStatusCode(errorKind),
	}

	s.mux.HandleFunc("/", s.handleIndex)
	s.mux.HandleFunc("/api/v2/reflection/api-methods", s.handleListAPIMethods)
	s.mux.HandleFunc("/api/v2/clients/labels/add", s.handleAddClientsLabels)
	s.mux.HandleFunc("/api/v2/clients/testClientID/flows/testFlowID/results/files-archive", s.handleGetFlowFilesArchive)
	s.mux.HandleFunc("/api/v2/clients/testClientID/flows/testFlowID/results", s.handleListFlowResults)
	s.mux.HandleFunc("/api/v2/hunts/testHuntID/results", s.handleListHuntResults)
	s.mux.HandleFunc("/api/v2/hunts", s.testHTTPServer.handleListHunts)

	return s
}

func createErrorHTTPServerAndConnector(errorKind errors.ErrorKind) (*testErrorHTTPServer, *httptest.Server, *HTTPConnector) {
	s := newTestErrorHTTPServer(errorKind)
	srv := httptest.NewServer(s.mux)
	return s, srv, NewHTTPConnector(srv.URL, 1000, nil)
}

func (s *testErrorHTTPServer) setCSRFTokenHasError(hasError bool) {
	s.csrfTokenHasError = hasError
}

func (s *testErrorHTTPServer) setListAPIMethodsHasError(hasError bool) {
	s.listAPIMethodsHasError = hasError
}

func (s *testErrorHTTPServer) setStatusCode(errorKind errors.ErrorKind) {
	s.statusCode = convertHTTPStatusCode(errorKind)
}

func (s *testErrorHTTPServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	if s.csrfTokenHasError && s.statusCode != http.StatusOK {
		w.WriteHeader(s.statusCode)
		return
	}

	s.testHTTPServer.handleIndex(w, r)
}

func (s *testErrorHTTPServer) handleListAPIMethods(w http.ResponseWriter, r *http.Request) {
	if s.listAPIMethodsHasError && s.statusCode != http.StatusOK {
		w.WriteHeader(s.statusCode)
		return
	}

	s.testHTTPServer.handleListAPIMethods(w, r)
}

func (s *testErrorHTTPServer) handleAddClientsLabels(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(s.statusCode)
}

func (s *testErrorHTTPServer) handleGetFlowFilesArchive(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(s.statusCode)
}

func (s *testErrorHTTPServer) handleListFlowResults(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(s.statusCode)
}

func (s *testErrorHTTPServer) handleListHuntResults(w http.ResponseWriter, r *http.Request) {
	ok, offset, count := s.getSendIteratorRequestOffsetAndCount(w, r, "ListHuntResults")
	if !ok {
		return
	}

	// Write response.
	items := []*ahpb.ApiHuntResult{
		{
			ClientId: proto.String("1"),
		},
		{
			ClientId: proto.String("2"),
		},
		{
			ClientId: proto.String("3"),
		},
		{
			ClientId: proto.String("4"),
		},
	}

	if offset > int64(len(items)) {
		items = items[:0]
	} else if count == -1 || offset+count > int64(len(items)) {
		items = items[offset:]
		w.WriteHeader(s.statusCode)
	} else {
		items = items[offset : offset+count]
	}

	result := &ahpb.ApiListHuntResultsResult{
		Items:      items,
		TotalCount: proto.Int64(4),
	}

	s.writeProtoResult(w, result)
}

func TestHTTPNonAPIErrors(t *testing.T) {
	s, srv, conn := createErrorHTTPServerAndConnector(errors.APINotImplementedError)
	defer srv.Close()
	var err error

	// Clear cached csrfToken and routing map.
	conn.csrfToken = ""

	// Test HTTP failure status code when fetching CSRF token.
	s.setCSRFTokenHasError(true)
	s.setListAPIMethodsHasError(false)
	err = conn.AddClientsLabels(context.Background(), &acpb.ApiAddClientsLabelsArgs{})
	if kind := errors.Kind(err); kind != errors.GenericError {
		t.Errorf("getCSRFToken has an error with kind %v; want: %v", kind, errors.GenericError)
	}

	// Clear cached csrfToken and routing map.
	conn.csrfToken = ""

	// Test HTTP failure status code when fetching routing map.
	s.setCSRFTokenHasError(false)
	s.setListAPIMethodsHasError(true)
	err = conn.AddClientsLabels(context.Background(), &acpb.ApiAddClientsLabelsArgs{})
	if kind := errors.Kind(err); kind != errors.GenericError {
		t.Errorf("fetchRoutingMap has an error with kind %v; want: %v", kind, errors.GenericError)
	}
}

func TestHTTPAPIErrors(t *testing.T) {
	s, srv, conn := createErrorHTTPServerAndConnector(errors.Invalid)
	defer srv.Close()

	calls := []struct {
		desc      string
		errorKind errors.ErrorKind
	}{
		{"HTTP 403 StatusForbidden error", errors.AccessForbiddenError},
		{"HTTP 404 StatusNotFound error", errors.ResourceNotFoundError},
		{"HTTP 501 StatusNotImplemented error", errors.APINotImplementedError},
		{"unknown HTTP error", errors.UnknownServerError},
	}

	tests := []struct {
		desc       string
		methodName string
		method     func() error
	}{
		{"test HTTP failure status code when calling a single request(getCSRFToken and fetchRoutingMap return no error)", "AddClientsLabels", func() error {
			// Clear cached csrfToken and routing map.
			conn.csrfToken = ""
			return conn.AddClientsLabels(context.Background(), &acpb.ApiAddClientsLabelsArgs{})
		}},
		{"test HTTP failure status code when calling a streaming request (getCSRFToken and fetchRoutingMap return no error)", "GetFlowFilesArchive", func() error {
			// Clear cached csrfToken and routing map.
			conn.csrfToken = ""
			_, err := conn.GetFlowFilesArchive(context.Background(), &afpb.ApiGetFlowFilesArchiveArgs{
				ClientId:      proto.String("testClientID"),
				FlowId:        proto.String("testFlowID"),
				ArchiveFormat: afpb.ApiGetFlowFilesArchiveArgs_ZIP.Enum(),
			})
			return err
		}},
		{"test HTTP failure status code when calling an iterator request in the first page (getCSRFToken and fetchRoutingMap return no error)", "ListFlowResults", func() error {
			// Clear cached csrfToken and routing map.
			conn.csrfToken = ""
			_, err := conn.ListFlowResults(context.Background(), &afpb.ApiListFlowResultsArgs{
				ClientId: proto.String("testClientID"),
				FlowId:   proto.String("testFlowID"),
			}, func(m *afpb.ApiFlowResult) error {
				return nil
			})
			return err
		}},
		{"test HTTP failure status code when calling an iterator request after fetching the first page (getCSRFToken and fetchRoutingMap return no error)", "ListHuntResults", func() error {
			// Clear cached csrfToken and routing map.
			conn.csrfToken = ""
			_, err := conn.ListHuntResults(context.Background(), &ahpb.ApiListHuntResultsArgs{
				HuntId: proto.String(connectortestutils.TestHuntID),
			}, func(m *ahpb.ApiHuntResult) error {
				return nil
			})
			return err
		}},
	}

	for _, test := range tests {
		for _, call := range calls {
			t.Run(fmt.Sprintf("%s: expected HTTP error: %s", test.desc, call.desc), func(t *testing.T) {
				s.setStatusCode(call.errorKind)
				err := test.method()
				if kind := errors.Kind(err); kind != call.errorKind {
					t.Fatalf("%s has an error with kind %v; want: %v", test.methodName, kind, call.errorKind)
				}
			})
		}
	}

	// Test UserCallback error.
	_, err := conn.ListHunts(context.Background(), &ahpb.ApiListHuntsArgs{}, func(m *ahpb.ApiHunt) error {
		return fmt.Errorf("testUserCallbackError")
	})

	if kind := errors.Kind(err); kind != errors.UserCallbackError {
		t.Fatalf("ListHunts has an error with kind %v; want: %v", kind, errors.UserCallbackError)
	}
}

const (
	basicAuthUsername = "username"
	basicAuthPassword = "password"
)

type testBasicAuthHTTPServer struct {
	testHTTPServer
	setAuth func(r *http.Request)
}

func newTestBasicAuthHTTPServer(setAuth func(r *http.Request)) *testBasicAuthHTTPServer {
	s := &testBasicAuthHTTPServer{
		testHTTPServer: testHTTPServer{
			csrfToken: testCSRFToken,
			mux:       http.NewServeMux(),
		},
		setAuth: setAuth,
	}

	s.mux.HandleFunc("/", s.handleIndex)
	s.mux.HandleFunc("/api/v2/reflection/api-methods", s.handleListAPIMethods)
	s.mux.HandleFunc("/api/v2/clients/labels/add", s.handleAddClientsLabels)
	s.mux.HandleFunc("/api/v2/clients/testClientID/flows/testFlowID/results/files-archive", s.handleGetFlowFilesArchive)
	s.mux.HandleFunc("/api/v2/hunts", s.handleListHunts)
	return s
}

func (s *testBasicAuthHTTPServer) checkBasicAuth(w http.ResponseWriter, r *http.Request) bool {
	username, password, ok := r.BasicAuth()
	if !ok || username != basicAuthUsername || password != basicAuthPassword {
		w.WriteHeader(http.StatusForbidden)
		return false
	}

	return true
}

func (s *testBasicAuthHTTPServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	if !s.checkBasicAuth(w, r) {
		return
	}
	s.testHTTPServer.handleIndex(w, r)
}

func (s *testBasicAuthHTTPServer) handleListAPIMethods(w http.ResponseWriter, r *http.Request) {
	if !s.checkBasicAuth(w, r) {
		return
	}
	s.testHTTPServer.handleListAPIMethods(w, r)
}

func (s *testBasicAuthHTTPServer) handleAddClientsLabels(w http.ResponseWriter, r *http.Request) {
	if !s.checkBasicAuth(w, r) {
		return
	}
	s.testHTTPServer.handleAddClientsLabels(w, r)
}

func (s *testBasicAuthHTTPServer) handleGetFlowFilesArchive(w http.ResponseWriter, r *http.Request) {
	if !s.checkBasicAuth(w, r) {
		return
	}
	s.testHTTPServer.handleGetFlowFilesArchive(w, r)
}

func (s *testBasicAuthHTTPServer) handleListHunts(w http.ResponseWriter, r *http.Request) {
	if !s.checkBasicAuth(w, r) {
		return
	}
	s.testHTTPServer.handleListHunts(w, r)
}

func createBasicAuthHTTPServerAndConnector(pageSize int64, setAuth func(r *http.Request)) (*httptest.Server, *HTTPConnector) {
	s := newTestBasicAuthHTTPServer(setAuth)
	srv := httptest.NewServer(s.mux)
	return srv, NewHTTPConnector(srv.URL, pageSize, setAuth)
}

func TestBasicAuthNotSet(t *testing.T) {
	srv, conn := createBasicAuthHTTPServerAndConnector(1000, nil)
	defer srv.Close()

	// Test Basic Auth for a single sendRequest.
	if err := conn.AddClientsLabels(context.Background(), &acpb.ApiAddClientsLabelsArgs{
		ClientIds: []string{"clientID1", "clientID2"},
		Labels:    []string{"foo", "bar"},
	}); err == nil {
		t.Fatal("AddClientsLabels succeeds, but it should fail as basic authentication is not set")
	}

	// Test Basic Auth for sendStreamingRequest.
	if _, err := conn.GetFlowFilesArchive(context.Background(), &afpb.ApiGetFlowFilesArchiveArgs{
		ClientId:      proto.String("testClientID"),
		FlowId:        proto.String("testFlowID"),
		ArchiveFormat: afpb.ApiGetFlowFilesArchiveArgs_ZIP.Enum(),
	}); err == nil {
		t.Fatal("GetFlowFilesArchive succeeds, but it should fail as basic authentication is not set")
	}

	// Test Basic Auth for SendIteratorRequest.
	if _, err := conn.ListHunts(context.Background(), &ahpb.ApiListHuntsArgs{}, func(m *ahpb.ApiHunt) error {
		return nil
	}); err == nil {
		t.Fatal("ListHunts succeeds, but it should fail as basic authentication is not set")
	}
}

func TestBasicAuth(t *testing.T) {
	srv, conn := createBasicAuthHTTPServerAndConnector(1000, func(r *http.Request) { r.SetBasicAuth(basicAuthUsername, basicAuthPassword) })
	defer srv.Close()

	// Test Basic Auth for a single sendRequest.
	if err := conn.AddClientsLabels(context.Background(), &acpb.ApiAddClientsLabelsArgs{
		ClientIds: []string{"clientID1", "clientID2"},
		Labels:    []string{"foo", "bar"},
	}); err != nil {
		t.Fatalf("BasicAuth for AddClientsLabels fails: \n%v", err)
	}

	// Test Basic Auth for sendStreamingRequest.
	if _, err := conn.GetFlowFilesArchive(context.Background(), &afpb.ApiGetFlowFilesArchiveArgs{
		ClientId:      proto.String("testClientID"),
		FlowId:        proto.String("testFlowID"),
		ArchiveFormat: afpb.ApiGetFlowFilesArchiveArgs_ZIP.Enum(),
	}); err != nil {
		t.Fatalf("BasicAuth for GetFlowFilesArchive fails: \n%v", err)
	}

	// Test Basic Auth for SendIteratorRequest.
	if _, err := conn.ListHunts(context.Background(), &ahpb.ApiListHuntsArgs{}, func(m *ahpb.ApiHunt) error {
		return nil
	}); err != nil {
		t.Fatalf("BasicAuth for ListHunts fails: \n%v", err)
	}
}
