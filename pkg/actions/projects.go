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

// GetProjects handles retrieving projects
// from all the hubs known to the federator
type GetProjects struct {
	BasicGetRequest
}

// NewGetProjects creates a new GetProjects object
func NewGetProjects(id string, ep api.EndpointType) *GetProjects {
	return &GetProjects{BasicGetRequest{endPoint: ep, id: id, responseCh: make(chan *GetResponse)}}
}

// Execute will tell the provided federator to retrieve projects
func (gp *GetProjects) Execute(fed FederatorInterface) error {
	var projects hubapi.ProjectList

	funcs := api.GetFuncsType{
		Get:    "GetProject",
		GetAll: "ListAllProjects",
		SingleToList: func(single interface{}) interface{} {
			item := reflect.ValueOf(single).Interface()
			list := hubapi.ProjectList{
				TotalCount: 1,
				Items:      []hubapi.Project{*item.(*hubapi.Project)},
			}
			return &list
		},
	}
	fed.SendGetRequest(gp.endPoint, funcs, gp.id, &projects)

	gp.responseCh <- &GetResponse{
		endPoint: gp.endPoint,
		id:       gp.id,
		list:     &projects,
	}

	return nil
}
