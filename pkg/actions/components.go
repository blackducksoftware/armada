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

// GetComponents handles retrieving components
// from all the hubs known to the federator
type GetComponents struct {
	BasicGetRequest
}

// NewGetComponents creates a new GetComponents object
func NewGetComponents(id string, ep api.EndpointType) *GetComponents {
	return &GetComponents{BasicGetRequest{endPoint: ep, id: id, responseCh: make(chan *GetResponse)}}
}

// Execute will tell the provided federator to retrieve components
func (gc *GetComponents) Execute(fed FederatorInterface) error {
	var components hubapi.ComponentList

	funcs := api.GetFuncsType{
		Get:    "GetComponent",
		GetAll: "ListAllComponents",
		SingleToList: func(single interface{}) interface{} {
			item := reflect.ValueOf(single).Interface()
			list := hubapi.ComponentList{
				TotalCount: 1,
				Items:      []hubapi.Component{*item.(*hubapi.Component)},
			}
			return &list
		},
	}
	fed.SendHubsGetRequest(gc.endPoint, funcs, gc.id, &components)

	gc.responseCh <- &GetResponse{
		endPoint: gc.endPoint,
		id:       gc.id,
		list:     &components,
	}

	return nil
}
