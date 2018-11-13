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
	"reflect"

	"github.com/blackducksoftware/armada/pkg/api"

	"github.com/blackducksoftware/hub-client-go/hubapi"
)

// GetCodeLocations handles retrieving code locations
// from all the hubs known to the federator
type GetCodeLocations struct {
	BasicGetRequest
}

// NewGetCodeLocations creates a new GetCodeLocations object
func NewGetCodeLocations(id string, ep api.EndpointType) *GetCodeLocations {
	return &GetCodeLocations{BasicGetRequest{endPoint: ep, id: id, responseCh: make(chan *GetResponse)}}
}

// Execute will tell the provided federator to retrieve code locations
func (gcl *GetCodeLocations) Execute(fed FederatorInterface) error {
	var codeLocations hubapi.CodeLocationList

	funcs := api.GetFuncsType{
		Get:    "GetCodeLocation",
		GetAll: "ListAllCodeLocations",
		SingleToList: func(single interface{}) interface{} {
			item := reflect.ValueOf(single).Interface()
			list := hubapi.CodeLocationList{
				TotalCount: 1,
				Items:      []hubapi.CodeLocation{*item.(*hubapi.CodeLocation)},
			}
			return &list
		},
	}
	fed.SendGetRequest(gcl.endPoint, funcs, gcl.id, &codeLocations)

	gcl.responseCh <- &GetResponse{
		endPoint: gcl.endPoint,
		id:       gcl.id,
		list:     &codeLocations,
	}

	return nil
}
