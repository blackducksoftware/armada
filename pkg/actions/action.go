/*
Copyright (C) 2018 Synopsys, Inc.

Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements. See the NOTICE file
distributed with this work for additional information
regarding copyright ownership. The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied. See the License for the
specific language governing permissions and limitations
under the License.
*/

package actions

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/blackducksoftware/armada/pkg/api"
	"github.com/blackducksoftware/armada/pkg/hub"

	"github.com/blackducksoftware/hub-client-go/hubclient"
)

// ActionInterface defines an interface for actions of the federator
type ActionInterface interface {
	// Apply the action to the provided federator
	Execute(FederatorInterface) error
	GetResponse() ActionResponseInterface
}

// ActionResponseInterface defines the interface for action responses
type ActionResponseInterface interface {
	ReplaceSource(string)
	GetResult() interface{}
}

// FederatorInterface defines the interface for federator objects
type FederatorInterface interface {
	DeleteHub(string)
	CreateHubClients(*api.HubList)
	AddHub(string, *hub.Client)
	GetHubs() map[string]*hub.Client
	SetLastError(api.EndpointType, *api.LastError)
	GetLastError(api.EndpointType) *api.LastError
	SendGetRequest(api.EndpointType, api.GetFuncsType, string, interface{})
	SendCreateRequest(api.EndpointType, string, interface{})
}

// EmptyResponse is a generic response to a request which
// provides no response that satisfies the ActionResponseInterface
type EmptyResponse struct{}

// ReplaceSource does nothing for the generic response
func (cr *EmptyResponse) ReplaceSource(string) {}

// GetResult returns nothing for the generic response
func (cr *EmptyResponse) GetResult() interface{} {
	return nil
}

// HubError defines an error from a hub
type HubError struct {
	Host string
	Err  *hubclient.HubClientError
}

// BasicGetRequest defines common elements of a get request
type BasicGetRequest struct {
	endPoint   api.EndpointType
	id         string
	responseCh chan *GetResponse
}

// GetResponse returns the response to the get request
func (gpr *BasicGetRequest) GetResponse() ActionResponseInterface {
	return <-gpr.responseCh
}

// GetResponse defines a generic response for a get request
type GetResponse struct {
	endPoint api.EndpointType
	id       string
	list     interface{}
}

// ReplaceSource will replace the source URL in the metadata
// with the federator information
func (gr *GetResponse) ReplaceSource(ip string) {
	href := fmt.Sprintf("https://%s/api/%s", ip, gr.endPoint)
	if len(gr.id) > 0 {
		href += fmt.Sprintf("/%s", gr.id)
	}
	if gr.isSingleResponse() {
		items := reflect.ValueOf(gr.list).Elem().FieldByName("Items").Index(0)
		meta := items.FieldByName("Meta")
		meta.FieldByName("Href").SetString(href)
	} else {
		list := reflect.ValueOf(gr.list).Elem()
		meta := list.FieldByName("Meta")
		if meta.IsValid() {
			meta.FieldByName("Href").SetString(href)
		}
	}
}

func (gr *GetResponse) isSingleResponse() bool {
	return !strings.HasPrefix(string(gr.endPoint), "all-") && len(gr.id) > 0
}

// GetResult returns the result for a get request
func (gr *GetResponse) GetResult() interface{} {
	if gr.isSingleResponse() {
		return reflect.ValueOf(gr.list).Elem().FieldByName("Items").Index(0).Interface()
	}
	return gr.list
}

// ConvertAllEndpoint will take an endpoint starting with the prefix "all-" and return
// the proper hub endpoint
func ConvertAllEndpoint(endPoint api.EndpointType) string {
	return strings.TrimPrefix(string(endPoint), "all-")
}

// BasicCreateRequest defines common elements of a create request
type BasicCreateRequest struct {
	request    interface{}
	responseCh chan *EmptyResponse
}

// GetResponse returns the response to the create request
func (gpr *BasicCreateRequest) GetResponse() ActionResponseInterface {
	return <-gpr.responseCh
}
