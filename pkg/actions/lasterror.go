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
	"github.com/blackducksoftware/armada/pkg/api"
)

// GetLastErrorResponse defines the response for a GetLastError request
type GetLastErrorResponse struct {
	response *api.LastError
}

// ReplaceSource has nothing to do
func (resp *GetLastErrorResponse) ReplaceSource(ip string) {
	// Nothing to do
}

// GetResult returns the last error
func (resp *GetLastErrorResponse) GetResult() interface{} {
	return resp.response
}

// GetLastError handles creating a policy rule
// in all the hubs known to the federator
type GetLastError struct {
	endpoint   api.EndpointType
	method     string
	responseCh chan *GetLastErrorResponse
}

// NewGetLastError creates a new GetLastError object
func NewGetLastError(meth string, ep api.EndpointType) *GetLastError {
	return &GetLastError{endpoint: ep, method: meth, responseCh: make(chan *GetLastErrorResponse)}
}

// Execute will retrieve the last error from the provided federator
func (gle *GetLastError) Execute(fed FederatorInterface) error {
	resp := GetLastErrorResponse{
		response: fed.GetLastError(gle.method, gle.endpoint),
	}

	gle.responseCh <- &resp
	return nil
}

// GetResponse returns the response to the create policy rules request
func (gle *GetLastError) GetResponse() ActionResponseInterface {
	return <-gle.responseCh
}
