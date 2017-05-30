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
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"context"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/google/grr_go_api_client/errors"
	acpb "github.com/google/grr_go_api_client/grrproto/grr/proto/api/client_proto"
	afpb "github.com/google/grr_go_api_client/grrproto/grr/proto/api/flow_proto"
	ahpb "github.com/google/grr_go_api_client/grrproto/grr/proto/api/hunt_proto"
	arpb "github.com/google/grr_go_api_client/grrproto/grr/proto/api/reflection_proto"
	aupb "github.com/google/grr_go_api_client/grrproto/grr/proto/api/user_proto"
	avpb "github.com/google/grr_go_api_client/grrproto/grr/proto/api/vfs_proto"

	anypb "github.com/golang/protobuf/ptypes/any"
	durpb "github.com/golang/protobuf/ptypes/duration"
	stpb "github.com/golang/protobuf/ptypes/struct"
	tspb "github.com/golang/protobuf/ptypes/timestamp"
	wpb "github.com/golang/protobuf/ptypes/wrappers"
)

const (
	jsonPrefix = ")]}'\n"
)

var (
	notUsedAnypb = &anypb.Any{}
	notUsedDurpb = &durpb.Duration{}
	notUsedStpb  = &stpb.Value{}
	notUsedTspb  = &tspb.Timestamp{}
	notUsedWpb   = &wpb.BoolValue{}
)

func convertHTTPError(statusCode int, msg string) ErrorWithKind {
	switch statusCode {
	case 200:
		return nil
	case 404:
		return errors.NewGRRAPIError(errors.ResourceNotFoundError, msg)
	case 403:
		return errors.NewGRRAPIError(errors.AccessForbiddenError, msg)
	case 501:
		return errors.NewGRRAPIError(errors.APINotImplementedError, msg)
	default:
		return errors.NewGRRAPIError(errors.UnknownServerError, msg)
	}
}

func stringInSlice(slice []string, str string) bool {
	for _, s := range slice {
		if str == s {
			return true
		}
	}
	return false
}

// HTTPConnector allows GRR clients (defined in client package) to send request and receive response from GRR server via HTTP.
type HTTPConnector struct {
	apiEndpoint string
	auth        string
	csrfToken   string
	client      *http.Client
	apiMethods  map[string]*arpb.ApiMethod
	urlVars     map[string][]string
	setAuth     func(r *http.Request)
	pageSize    int64
	username    string
}

// NewHTTPConnector returns a newly created HTTP connector, in which setAuth adds authentication for each single HTTP request.
// It can be nil if no authentication is required.
// Example usage: NewHTTPConnector("http://localhost:8081", func(r *http.Request) { r.SetBasicAuth("username", "password") })
func NewHTTPConnector(apiEndpoint string, pageSize int64, setAuth func(r *http.Request)) *HTTPConnector {
	return &HTTPConnector{
		apiEndpoint: strings.Trim(apiEndpoint, "/ "),
		client:      &http.Client{},
		apiMethods:  make(map[string]*arpb.ApiMethod),
		urlVars:     make(map[string][]string),
		setAuth:     setAuth,
		pageSize:    pageSize,
	}
}

// SetPageSize is a setter for HTTPConnector.pageSize .
func (c *HTTPConnector) SetPageSize(pageSize int64) {
	c.pageSize = pageSize
}

func (c *HTTPConnector) getCSRFToken() (string, error) {
	req, err := http.NewRequest("GET", c.apiEndpoint, nil)
	if err != nil {
		return "", fmt.Errorf("cannot create request for a CSRF token: %v", err)
	}
	if c.setAuth != nil {
		c.setAuth(req)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot create request for a CSRF token: %v", err)
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf(
			"cannot get CSRF token: %s resulted in a HTTP error with status code %d; want 200",
			c.apiEndpoint,
			resp.StatusCode)
	}

	cookies := resp.Cookies()
	for _, cookie := range cookies {
		if cookie.Name == "csrftoken" {
			return cookie.Value, nil
		}
	}

	return "", fmt.Errorf("no CSRF token included in cookies of %s", c.apiEndpoint)
}

func (c *HTTPConnector) addHTTPHeader(req *http.Request) {
	req.Header.Add("x-csrftoken", c.csrfToken)
	req.Header.Add("x-requested-with", "XMLHttpRequest")
	cookie := http.Cookie{
		Name:  "X-CSRF-Token",
		Value: c.csrfToken,
	}
	req.AddCookie(&cookie)
}

func buildRoutingMap(data *arpb.ApiListApiMethodsResult, apiMethods map[string]*arpb.ApiMethod, urlVars map[string][]string) error {
	re := regexp.MustCompile("<([a-zA-Z0-9_:]+)>")

	for _, method := range data.Items {
		// Make sure api/v2 is used
		if !strings.HasPrefix(*method.HttpRoute, "/api/v2") {
			*method.HttpRoute = strings.Replace(*method.HttpRoute, "/api/", "/api/v2/", 1)
		}

		apiMethods[*method.Name] = method

		// Exact out all url variables.
		matches := re.FindAllStringSubmatch(*method.HttpRoute, -1)
		vars := make([]string, len(matches))
		for i, variable := range matches {
			if len(variable) != 2 {
				return fmt.Errorf("each element for %v should has at length of 2, but %v has a length of %v", matches, variable, len(variable))
			}

			vars[i] = variable[1]
		}
		urlVars[*method.Name] = vars
	}

	return nil
}

func (c *HTTPConnector) fetchRoutingMap() error {
	url := c.apiEndpoint + "/api/v2/reflection/api-methods"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	c.addHTTPHeader(req)
	if c.setAuth != nil {
		c.setAuth(req)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf(
			"cannot fetch routing map: %s resulted in a HTTP error with status code %d; want 200",
			url,
			resp.StatusCode)
	}

	all, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("unable to read response body of %s due to an error: %v", url, err)
	}

	data := &arpb.ApiListApiMethodsResult{}
	r := bytes.NewReader(all[len(jsonPrefix):])
	if err := jsonpb.Unmarshal(r, data); err != nil {
		return fmt.Errorf("jsonpb unmarshal error: %v", err)
	}

	if err = buildRoutingMap(data, c.apiMethods, c.urlVars); err != nil {
		return fmt.Errorf("cannot build routing map because of an error:\n%v", err)
	}

	return nil
}

func (c *HTTPConnector) initializeIfNeeded() error {
	if c.csrfToken != "" {
		return nil
	}

	token, err := c.getCSRFToken()
	if err != nil {
		return err
	}

	c.csrfToken = token

	return c.fetchRoutingMap()
}

func coerceValueToQueryStringType(field interface{}) (string, error) {
	var value string

	switch fieldType := field.(type) {
	case fmt.Stringer:
		value = field.(fmt.Stringer).String()
	case *string:
		value = *field.(*string)
	case *int64:
		value = strconv.FormatInt(*field.(*int64), 10)
	case *uint64:
		value = strconv.FormatUint(*field.(*uint64), 10)
	default:
		return "", fmt.Errorf("unknown type %v", fieldType)
	}

	return value, nil
}

type field struct {
	camelName string
	name      string
	value     interface{}
}

func listFields(args proto.Message) ([]field, error) {
	var res []field
	if args == nil {
		return res, nil
	}

	st := reflect.ValueOf(args).Elem()
	re := regexp.MustCompile("name=([a-zA-Z0-9_:]+)")
	for i := 0; i < st.NumField(); i++ {
		typeField := st.Type().Field(i)
		tagStr := typeField.Tag.Get("protobuf")

		// Skip if the field doesn't have a tag with key 'protobuf'.
		if tagStr == "" {
			continue
		}

		matches := re.FindStringSubmatch(tagStr)
		if len(matches) == 0 {
			continue
		} else if len(matches) != 2 {
			return nil, fmt.Errorf("there should be exactly one match for 'name=[...]' in %s", tagStr)
		}

		name := matches[1]
		res = append(res, field{
			camelName: typeField.Name,
			name:      name,
			value:     st.Field(i).Interface(),
		})
	}

	return res, nil
}

type requestInfo struct {
	method          string
	url             string
	pathParamsNames []string
}

func getMethodURLAndPathParamsNames(apiMethod *arpb.ApiMethod, urlVars []string, args proto.Message) (*requestInfo, error) {
	methods := apiMethod.HttpMethods
	if len(methods) != 1 {
		return nil, fmt.Errorf("method %s has HTTP methods(s): %v ; want exactly one method", apiMethod.Name, methods)
	}

	info := &requestInfo{
		method: methods[0],
		url:    *apiMethod.HttpRoute,
	}

	fieldNames, err := listFields(args)
	if err != nil {
		return nil, fmt.Errorf("cannot get field names of the argument in getMethodURLAndPathParamsNames: %v", err)
	}

	for _, f := range fieldNames {
		// Skip if the field is not required to build the url.
		variable := f.name
		if stringInSlice(urlVars, "path:"+variable) {
			variable = "path:" + variable
		} else if !stringInSlice(urlVars, variable) {
			continue
		}

		value, err := coerceValueToQueryStringType(f.value)
		if err != nil {
			return nil, err
		}

		info.url = strings.Replace(info.url, "<"+variable+">", value, 1)
		info.pathParamsNames = append(info.pathParamsNames, f.camelName)
	}

	return info, nil
}

func addArgsToQueryParams(req *http.Request, args proto.Message, excludeNames []string) error {
	q := req.URL.Query()

	fieldNames, err := listFields(args)
	if err != nil {
		return fmt.Errorf("cannot get field names of the argument in addArgsToQueryParams: %v", err)
	}

	for _, f := range fieldNames {
		// Skip excluded fields.
		if stringInSlice(excludeNames, f.camelName) {
			continue
		}

		if !reflect.ValueOf(f.value).Elem().IsValid() {
			continue
		}

		value, err := coerceValueToQueryStringType(f.value)
		if err != nil {
			return err
		}

		q.Add(f.name, value)
	}

	req.URL.RawQuery = q.Encode()
	return nil
}

func argsToBody(args proto.Message, excludeNames []string) (io.Reader, error) {
	if args == nil {
		return nil, nil
	}

	argsCopy := proto.Clone(args)

	// Clear excluded fields.
	st := reflect.ValueOf(argsCopy).Elem()
	for i := 0; i < st.NumField(); i++ {
		typeField := st.Type().Field(i)
		if stringInSlice(excludeNames, typeField.Name) {
			field := st.Field(i)
			field.Set(reflect.Zero(field.Type()))
		}
	}

	jsonStr, err := (&jsonpb.Marshaler{}).MarshalToString(argsCopy)
	if err != nil {
		return nil, err
	}

	return strings.NewReader(jsonStr), nil
}

func (c *HTTPConnector) buildRequest(handlerName string, args proto.Message) (*http.Request, error) {
	apiMethod, ok := c.apiMethods[handlerName]
	if !ok {
		return nil, fmt.Errorf("%s cannot be found in available API methods", handlerName)
	}

	info, err := getMethodURLAndPathParamsNames(apiMethod, c.urlVars[handlerName], args)
	if err != nil {
		return nil, err
	}

	var body io.Reader

	if info.method != "GET" {
		body, err = argsToBody(args, info.pathParamsNames)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(info.method, c.apiEndpoint+info.url, body)
	if err != nil {
		return nil, err
	}

	if info.method == "GET" {
		if err = addArgsToQueryParams(req, args, info.pathParamsNames); err != nil {
			return nil, err
		}
	}

	c.addHTTPHeader(req)
	if c.setAuth != nil {
		c.setAuth(req)
	}

	return req, nil
}

func (c *HTTPConnector) sendRequest(handlerName string, args proto.Message) (proto.Message, error) {
	if err := c.initializeIfNeeded(); err != nil {
		return nil, err
	}

	req, err := c.buildRequest(handlerName, args)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, convertHTTPError(resp.StatusCode, fmt.Sprintf(
			"api call %s resulted in a HTTP error with status code %d; want 200",
			handlerName,
			resp.StatusCode))
	}

	methodDescriptor := c.apiMethods[handlerName]

	if methodDescriptor == nil {
		return nil, fmt.Errorf("cannot find the descriptor of %s", handlerName)
	}

	if methodDescriptor.ResultTypeDescriptor != nil {
		defaultValue := methodDescriptor.ResultTypeDescriptor.Default
		var data ptypes.DynamicAny
		if err = ptypes.UnmarshalAny(defaultValue, &data); err != nil {
			return nil, fmt.Errorf("cannot unmarshal a proto message into type %v: %v", defaultValue, err)
		}

		all, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("unable to read response body of api call %s: %v", handlerName, err)
		}

		r := bytes.NewReader(all[len(jsonPrefix):])
		if err = jsonpb.Unmarshal(r, data.Message); err != nil {
			return nil, fmt.Errorf("jsonpb unmarshal error: %v", err)
		}

		return data.Message, nil
	}

	return nil, nil
}

func (c *HTTPConnector) sendStreamingRequest(handlerName string, args proto.Message) (io.ReadCloser, error) {
	if err := c.initializeIfNeeded(); err != nil {
		return nil, err
	}

	req, err := c.buildRequest(handlerName, args)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, convertHTTPError(resp.StatusCode, fmt.Sprintf(
			"api call %s resulted in a HTTP error with status code %d; want 200",
			handlerName,
			resp.StatusCode))
	}

	return resp.Body, nil
}

// AddClientsLabels sends a request to GRR server to add labels for the client via HTTP.
func (c *HTTPConnector) AddClientsLabels(_ context.Context, args *acpb.ApiAddClientsLabelsArgs) error {
	_, err := c.sendRequest("AddClientsLabels", args)
	return err
}

// RemoveClientsLabels sends a request to GRR server to remove labels for the client via HTTP.
func (c *HTTPConnector) RemoveClientsLabels(_ context.Context, args *acpb.ApiRemoveClientsLabelsArgs) error {
	_, err := c.sendRequest("RemoveClientsLabels", args)
	return err
}

// GetFlowFilesArchive sends a request to GRR server to get files archive of the flow via HTTP.
// This function starts a goroutine that runs until all data is read, or the readcloser is closed, or an error occurs in the go routine.
func (c *HTTPConnector) GetFlowFilesArchive(_ context.Context, args *afpb.ApiGetFlowFilesArchiveArgs) (io.ReadCloser, error) {
	return c.sendStreamingRequest("GetFlowFilesArchive", args)
}

// ListHunts sends a request to GRR server to list all GRR hunts via HTTP and calls fn on each hunt.
func (c *HTTPConnector) ListHunts(_ context.Context, args *ahpb.ApiListHuntsArgs, fn func(m *ahpb.ApiHunt) error) (int64, error) {
	sr := func(offset, count int64) (proto.Message, int64, error) {
		args = proto.Clone(args).(*ahpb.ApiListHuntsArgs)
		args.Offset = proto.Int64(offset)
		args.Count = proto.Int64(count)
		resp, err := c.sendRequest("ListHunts", args)
		if err != nil {
			return nil, -1, err
		}
		result := resp.(*ahpb.ApiListHuntsResult)
		return result, result.GetTotalCount(), nil
	}

	callback := func(m proto.Message) error {
		return ConvertUserError(fn(m.(*ahpb.ApiHunt)))
	}

	return SendIteratorRequest(args, sr, callback, c.pageSize)
}

// GetUsername returns the username of current user.
func (c *HTTPConnector) GetUsername(ctx context.Context) (string, error) {
	if c.username != "" {
		return c.username, nil
	}

	user, err := c.GetGRRUser(ctx)
	if err != nil {
		return "", err
	}

	c.username = user.GetUsername()
	return c.username, nil
}

// GetGRRUser sends a request to GRR server to return the current user via HTTP.
func (c *HTTPConnector) GetGRRUser(_ context.Context) (*aupb.ApiGrrUser, error) {
	user, err := c.sendRequest("GetGrrUser", nil)
	if err != nil {
		return nil, err
	}
	return user.(*aupb.ApiGrrUser), nil
}

// CancelFlow sends a request to GRR server to cancel the flow via HTTP.
func (c *HTTPConnector) CancelFlow(_ context.Context, args *afpb.ApiCancelFlowArgs) error {
	// Ignore void result.
	_, err := c.sendRequest("CancelFlow", args)
	return err
}

// GetClientApproval sends a request to GRR server to get information for a client approval.
func (c *HTTPConnector) GetClientApproval(_ context.Context, args *aupb.ApiGetClientApprovalArgs) (*aupb.ApiClientApproval, error) {
	approval, err := c.sendRequest("GetClientApproval", args)
	if err != nil {
		return nil, err
	}
	return approval.(*aupb.ApiClientApproval), nil
}

// GetHuntApproval sends a request to GRR server to get a hunt approval via HTTP.
func (c *HTTPConnector) GetHuntApproval(_ context.Context, args *aupb.ApiGetHuntApprovalArgs) (*aupb.ApiHuntApproval, error) {
	approval, err := c.sendRequest("GetHuntApproval", args)
	if err != nil {
		return nil, err
	}
	return approval.(*aupb.ApiHuntApproval), nil
}

// ModifyHunt sends a request to GRR server to start/stop hunt or modify its attributes via HTTP.
func (c *HTTPConnector) ModifyHunt(_ context.Context, args *ahpb.ApiModifyHuntArgs) (*ahpb.ApiHunt, error) {
	hunt, err := c.sendRequest("ModifyHunt", args)
	if err != nil {
		return nil, err
	}
	return hunt.(*ahpb.ApiHunt), nil
}

// GetHunt sends a request to GRR server to get a hunt via HTTP.
func (c *HTTPConnector) GetHunt(_ context.Context, args *ahpb.ApiGetHuntArgs) (*ahpb.ApiHunt, error) {
	hunt, err := c.sendRequest("GetHunt", args)
	if err != nil {
		return nil, err
	}
	return hunt.(*ahpb.ApiHunt), nil
}

// GrantHuntApproval sends a request to GRR server to grant a hunt approval via HTTP.
func (c *HTTPConnector) GrantHuntApproval(_ context.Context, args *aupb.ApiGrantHuntApprovalArgs) (*aupb.ApiHuntApproval, error) {
	approval, err := c.sendRequest("GrantHuntApproval", args)
	if err != nil {
		return nil, err
	}
	return approval.(*aupb.ApiHuntApproval), nil

}

// GetHuntFilesArchive sends a request to GRR server to get ZIP or TAR.GZ
// archive with all the files downloaded by the hunt via HTTP.
// This function starts a goroutine that runs until all data is read,
// or the readcloser is closed, or an error occurs in the go routine.
func (c *HTTPConnector) GetHuntFilesArchive(_ context.Context, args *ahpb.ApiGetHuntFilesArchiveArgs) (io.ReadCloser, error) {
	return c.sendStreamingRequest("GetHuntFilesArchive", args)
}

// GetFileBlob sends a request to GRR server to get byte contents of
// a VFS file on a given client via HTTP. This function starts a goroutine that
// runs until all data is read, or the readcloser is closed,
// or an error occurs in the go routine.
func (c *HTTPConnector) GetFileBlob(_ context.Context, args *avpb.ApiGetFileBlobArgs) (io.ReadCloser, error) {
	return c.sendStreamingRequest("GetFileBlob", args)
}

// GetVFSFilesArchive sends a request to GRR server to stream an archive
// with files collected and stored in the VFS of a client via HTTP.
// This function starts a goroutine that runs until all data is read,
// or the readcloser is closed, or an error occurs in the go routine.
func (c *HTTPConnector) GetVFSFilesArchive(_ context.Context, args *avpb.ApiGetVfsFilesArchiveArgs) (io.ReadCloser, error) {
	return c.sendStreamingRequest("GetVFSFilesArchive", args)
}

// GetVFSTimelineAsCSV sends a request to GRR server to stream timeline
// information for a given VFS path on a given client in a CSV format via HTTP.
// This function starts a goroutine that runs until all data is read,
// or the readcloser is closed, or an error occurs in the go routine.
func (c *HTTPConnector) GetVFSTimelineAsCSV(_ context.Context, args *avpb.ApiGetVfsTimelineAsCsvArgs) (io.ReadCloser, error) {
	return c.sendStreamingRequest("GetVFSTimelineAsCSV", args)
}

// ListHuntResults sends a request to GRR server to list all the results
// returned by a given hunt via HTTP and calls fn on each hunt result.
func (c *HTTPConnector) ListHuntResults(_ context.Context, args *ahpb.ApiListHuntResultsArgs, fn func(m *ahpb.ApiHuntResult) error) (int64, error) {
	sr := func(offset, count int64) (proto.Message, int64, error) {
		args := proto.Clone(args).(*ahpb.ApiListHuntResultsArgs)
		args.Offset = proto.Int64(offset)
		args.Count = proto.Int64(count)
		resp, err := c.sendRequest("ListHuntResults", args)
		if err != nil {
			return nil, -1, err
		}
		result := resp.(*ahpb.ApiListHuntResultsResult)
		return result, result.GetTotalCount(), nil
	}

	callback := func(m proto.Message) error {
		return ConvertUserError(fn(m.(*ahpb.ApiHuntResult)))
	}

	return SendIteratorRequest(args, sr, callback, c.pageSize)
}

// ListFlowResults sends a request to GRR server to list all the results
// returned by a given flow via HTTP and calls fn on each flow result.
func (c *HTTPConnector) ListFlowResults(_ context.Context, args *afpb.ApiListFlowResultsArgs, fn func(m *afpb.ApiFlowResult) error) (int64, error) {
	sr := func(offset, count int64) (proto.Message, int64, error) {
		args = proto.Clone(args).(*afpb.ApiListFlowResultsArgs)
		args.Offset = proto.Int64(offset)
		args.Count = proto.Int64(count)
		resp, err := c.sendRequest("ListFlowResults", args)
		if err != nil {
			return nil, -1, err
		}
		result := resp.(*afpb.ApiListFlowResultsResult)
		return result, result.GetTotalCount(), nil
	}

	callback := func(m proto.Message) error {
		return ConvertUserError(fn(m.(*afpb.ApiFlowResult)))
	}

	return SendIteratorRequest(args, sr, callback, c.pageSize)
}

// ListClientApprovals sends a request to GRR server to list all approvals of
// the client via HTTP and calls fn (a callback function) on each approval.
func (c *HTTPConnector) ListClientApprovals(_ context.Context, args *aupb.ApiListClientApprovalsArgs, fn func(m *aupb.ApiClientApproval) error) error {
	sr := func(offset, count int64) (proto.Message, int64, error) {
		args = proto.Clone(args).(*aupb.ApiListClientApprovalsArgs)
		args.Offset = proto.Int64(offset)
		args.Count = proto.Int64(count)
		resp, err := c.sendRequest("ListClientApprovals", args)
		if err != nil {
			return nil, -1, err
		}
		return resp.(*aupb.ApiListClientApprovalsResult), -1, nil
	}

	callback := func(m proto.Message) error {
		return ConvertUserError(fn(m.(*aupb.ApiClientApproval)))
	}

	// ApiListClientApprovalsResult doesn't have field total_count.
	_, err := SendIteratorRequest(args, sr, callback, c.pageSize)
	return err
}

// ListHuntApprovals sends a request to GRR server to list all the hunt
// approvals belonging to requesting user via HTTP and calls fn
// (a callback function) on each approval.
func (c *HTTPConnector) ListHuntApprovals(_ context.Context, args *aupb.ApiListHuntApprovalsArgs, fn func(m *aupb.ApiHuntApproval) error) error {
	sr := func(offset, count int64) (proto.Message, int64, error) {
		args = proto.Clone(args).(*aupb.ApiListHuntApprovalsArgs)
		args.Offset = proto.Int64(offset)
		args.Count = proto.Int64(count)
		resp, err := c.sendRequest("ListHuntApprovals", args)
		if err != nil {
			return nil, -1, err
		}
		return resp.(*aupb.ApiListHuntApprovalsResult), -1, nil
	}

	callback := func(m proto.Message) error {
		return ConvertUserError(fn(m.(*aupb.ApiHuntApproval)))
	}

	// ApiListHuntApprovalsResult doesn't have field total_count.
	_, err := SendIteratorRequest(args, sr, callback, c.pageSize)

	return err
}

// GetVFSTimeline sends a request to GRR server to get event timeline of VFS events
// for a given VFS path via HTTP.
func (c *HTTPConnector) GetVFSTimeline(_ context.Context, args *avpb.ApiGetVfsTimelineArgs) (*avpb.ApiGetVfsTimelineResult, error) {
	timeline, err := c.sendRequest("GetVFSTimeline", args)
	if err != nil {
		return nil, err
	}
	return timeline.(*avpb.ApiGetVfsTimelineResult), err
}

// ListFiles sends a request to GRR server to list files in a given VFS directory
// of a given client via HTTP and calls fn (a callback function) on each file.
func (c *HTTPConnector) ListFiles(_ context.Context, args *avpb.ApiListFilesArgs, fn func(m *avpb.ApiFile) error) error {
	sr := func(offset, count int64) (proto.Message, int64, error) {
		args = proto.Clone(args).(*avpb.ApiListFilesArgs)
		args.Offset = proto.Int64(offset)
		args.Count = proto.Int64(count)
		resp, err := c.sendRequest("ListFiles", args)
		if err != nil {
			return nil, -1, err
		}
		result := resp.(*avpb.ApiListFilesResult)
		return result, -1, nil
	}

	callback := func(m proto.Message) error {
		return ConvertUserError(fn(m.(*avpb.ApiFile)))
	}

	// ApiListFilesResult doesn't have field total_count
	_, err := SendIteratorRequest(args, sr, callback, c.pageSize)

	return err
}

// ListFlows sends a request to GRR server to list flows in a given VFS directory
// of a given client via HTTP and calls fn (a callback function) on each flow.
func (c *HTTPConnector) ListFlows(_ context.Context, args *afpb.ApiListFlowsArgs, fn func(m *afpb.ApiFlow) error) error {
	sr := func(offset, count int64) (proto.Message, int64, error) {
		args = proto.Clone(args).(*afpb.ApiListFlowsArgs)
		args.Offset = proto.Int64(offset)
		args.Count = proto.Int64(count)
		resp, err := c.sendRequest("ListFlows", args)
		if err != nil {
			return nil, -1, err
		}
		result := resp.(*afpb.ApiListFlowsResult)
		return result, -1, nil
	}

	callback := func(m proto.Message) error {
		return ConvertUserError(fn(m.(*afpb.ApiFlow)))
	}

	// ApiListFlowsResult doesn't have field total_count
	_, err := SendIteratorRequest(args, sr, callback, c.pageSize)

	return err
}

// CreateClientApproval sends a request to GRR server to create a new client approval via HTTP.
func (c *HTTPConnector) CreateClientApproval(_ context.Context, args *aupb.ApiCreateClientApprovalArgs) (*aupb.ApiClientApproval, error) {
	approval, err := c.sendRequest("CreateClientApproval", args)
	if err != nil {
		return nil, err
	}
	return approval.(*aupb.ApiClientApproval), err
}

// GrantClientApproval sends a request to GRR server to grant a client approval via HTTP.
func (c *HTTPConnector) GrantClientApproval(_ context.Context, args *aupb.ApiGrantClientApprovalArgs) (*aupb.ApiClientApproval, error) {
	approval, err := c.sendRequest("GrantClientApproval", args)
	if err != nil {
		return nil, err
	}
	return approval.(*aupb.ApiClientApproval), err
}

// CreateHunt sends a request to GRR server to create a new hunt via HTTP.
func (c *HTTPConnector) CreateHunt(_ context.Context, args *ahpb.ApiCreateHuntArgs) (*ahpb.ApiHunt, error) {
	hunt, err := c.sendRequest("CreateHunt", args)
	if err != nil {
		return nil, err
	}
	return hunt.(*ahpb.ApiHunt), err
}

// CreateHuntApproval sends a request to GRR server to create a new hunt approval via HTTP.
func (c *HTTPConnector) CreateHuntApproval(_ context.Context, args *aupb.ApiCreateHuntApprovalArgs) (*aupb.ApiHuntApproval, error) {
	approval, err := c.sendRequest("CreateHuntApproval", args)
	if err != nil {
		return nil, err
	}
	return approval.(*aupb.ApiHuntApproval), err
}

// GetFlow sends a request to GRR server to get information for a flow via HTTP.
func (c *HTTPConnector) GetFlow(_ context.Context, args *afpb.ApiGetFlowArgs) (*afpb.ApiFlow, error) {
	flow, err := c.sendRequest("GetFlow", args)
	if err != nil {
		return nil, err
	}
	return flow.(*afpb.ApiFlow), nil
}

// CreateFlow sends a request to GRR server to start a new flow on a
// given client via HTTP.
func (c *HTTPConnector) CreateFlow(_ context.Context, args *afpb.ApiCreateFlowArgs) (*afpb.ApiFlow, error) {
	flow, err := c.sendRequest("CreateFlow", args)
	if err != nil {
		return nil, err
	}
	return flow.(*afpb.ApiFlow), nil
}

// GetClient sends a request to GRR server to get client with a given
// client id via HTTP.
func (c *HTTPConnector) GetClient(_ context.Context, args *acpb.ApiGetClientArgs) (*acpb.ApiClient, error) {
	client, err := c.sendRequest("GetClient", args)
	if err != nil {
		return nil, err
	}
	return client.(*acpb.ApiClient), err
}

// SearchClients sends a request to GRR server to search for clients using
// a search query via HTTP.
func (c *HTTPConnector) SearchClients(_ context.Context, args *acpb.ApiSearchClientsArgs, fn func(m *acpb.ApiClient) error) error {
	sr := func(offset, count int64) (proto.Message, int64, error) {
		args = proto.Clone(args).(*acpb.ApiSearchClientsArgs)
		args.Offset = proto.Int64(offset)
		args.Count = proto.Int64(count)
		resp, err := c.sendRequest("SearchClients", args)
		if err != nil {
			return nil, -1, err
		}
		return resp.(*acpb.ApiSearchClientsResult), -1, nil
	}

	callback := func(m proto.Message) error {
		return ConvertUserError(fn(m.(*acpb.ApiClient)))
	}

	// ApiSearchClientsResult doesn't have field total_count.
	_, err := SendIteratorRequest(args, sr, callback, c.pageSize)

	return err
}

// GetVFSRefreshOperationState sends a request to GRR server to return a status
// of a "VFS refresh" operation started by CreateVFSRefreshOperation via HTTP.
func (c *HTTPConnector) GetVFSRefreshOperationState(_ context.Context, args *avpb.ApiGetVfsRefreshOperationStateArgs) (*avpb.ApiGetVfsRefreshOperationStateResult, error) {
	result, err := c.sendRequest("GetVfsRefreshOperationState", args)
	if err != nil {
		return nil, err
	}
	return result.(*avpb.ApiGetVfsRefreshOperationStateResult), err
}

// CreateVFSRefreshOperation sends a request to GRR server to start a VFS
// "refresh" operation, that will refresh the stats data about files and
// directories at a given VFS path on a given client, via HTTP.
func (c *HTTPConnector) CreateVFSRefreshOperation(_ context.Context, args *avpb.ApiCreateVfsRefreshOperationArgs) (*avpb.ApiCreateVfsRefreshOperationResult, error) {
	result, err := c.sendRequest("CreateVfsRefreshOperation", args)
	if err != nil {
		return nil, err
	}
	return result.(*avpb.ApiCreateVfsRefreshOperationResult), err
}

// GetFileDetails sends a request to GRR server to return file details
// for the file on a given client at a given VFS path via HTTP.
func (c *HTTPConnector) GetFileDetails(_ context.Context, args *avpb.ApiGetFileDetailsArgs) (*avpb.ApiGetFileDetailsResult, error) {
	result, err := c.sendRequest("GetFileDetails", args)
	if err != nil {
		return nil, err
	}
	return result.(*avpb.ApiGetFileDetailsResult), err
}

// GetFileVersionTimes sends a request to GRR server to return a list of
// timestamps (in microseconds) corresponding to different historic
// versions of a VFS file via HTTP.
func (c *HTTPConnector) GetFileVersionTimes(_ context.Context, args *avpb.ApiGetFileVersionTimesArgs) (*avpb.ApiGetFileVersionTimesResult, error) {
	result, err := c.sendRequest("GetFileVersionTimes", args)
	if err != nil {
		return nil, err
	}
	return result.(*avpb.ApiGetFileVersionTimesResult), err
}

// GetVFSFileContentUpdateState sends a request to GRR server to get the state
// of a long-running operation started by UpdateVFSFileContent via HTTP.
func (c *HTTPConnector) GetVFSFileContentUpdateState(_ context.Context, args *avpb.ApiGetVfsFileContentUpdateStateArgs) (*avpb.ApiGetVfsFileContentUpdateStateResult, error) {
	result, err := c.sendRequest("GetVFSFileContentUpdateState", args)
	if err != nil {
		return nil, err
	}
	return result.(*avpb.ApiGetVfsFileContentUpdateStateResult), err
}

// UpdateVFSFileContent sends a request to GRR server to start a long-running
// operation that re-downloads the file from the GRR client via HTTP.
func (c *HTTPConnector) UpdateVFSFileContent(_ context.Context, args *avpb.ApiUpdateVfsFileContentArgs) (*avpb.ApiUpdateVfsFileContentResult, error) {
	result, err := c.sendRequest("UpdateVFSFileContent", args)
	if err != nil {
		return nil, err
	}
	return result.(*avpb.ApiUpdateVfsFileContentResult), err
}
