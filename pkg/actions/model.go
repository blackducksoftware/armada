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

// GetModelResponse defines the response for a GetModel request
type GetModelResponse struct {
	allHubs map[string]*api.ModelHub
}

// ReplaceSource has nothing to do
func (resp *GetModelResponse) ReplaceSource(ip string) {}

// GetResult returns the component list
func (resp *GetModelResponse) GetResult() interface{} {
	return resp.allHubs
}

// GetModel will get the internal configuration of the hubs in the federator
type GetModel struct {
	doneCh chan *GetModelResponse
}

// NewGetModel gets the internal configuration of the hubs in the federator
func NewGetModel() *GetModel {
	return &GetModel{doneCh: make(chan *GetModelResponse)}
}

// Execute will get the hub informaton for the provided federator
func (gm *GetModel) Execute(fed FederatorInterface) error {
	hubs := map[string]*api.ModelHub{}
	for hubURL, hub := range fed.GetHubs() {
		hubs[hubURL] = <-hub.Model()
	}
	model := &GetModelResponse{allHubs: hubs}
	gm.doneCh <- model
	return nil
}

// GetResponse returns the response to the get model request
func (gm *GetModel) GetResponse() ActionResponseInterface {
	return <-gm.doneCh
}
