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

// GetPolicyRules handles retrieving policy rules
// from all the hubs known to the federator
type GetPolicyRules struct {
	BasicGetRequest
}

// NewGetPolicyRules creates a new GetPolicyRules object
func NewGetPolicyRules(id string, ep api.EndpointType) *GetPolicyRules {
	return &GetPolicyRules{BasicGetRequest{endPoint: ep, id: id, responseCh: make(chan *GetResponse)}}
}

// Execute will tell the provided federator to retrieve policy rules
func (gpr *GetPolicyRules) Execute(fed FederatorInterface) error {
	var policyRules hubapi.PolicyRuleList

	funcs := api.HubFuncsType{
		Get:    "GetPolicyRule",
		GetAll: "ListAllPolicyRules",
		SingleToList: func(single interface{}) interface{} {
			item := reflect.ValueOf(single).Interface()
			list := hubapi.PolicyRuleList{
				TotalCount: 1,
				Items:      []hubapi.PolicyRule{*item.(*hubapi.PolicyRule)},
			}
			return &list
		},
	}
	fed.SendGetRequest(gpr.endPoint, funcs, gpr.id, &policyRules)

	gpr.responseCh <- &GetResponse{
		endPoint: gpr.endPoint,
		id:       gpr.id,
		list:     &policyRules,
	}

	return nil
}

// CreatePolicyRule handles creating a policy rule
// in all the hubs known to the federator
type CreatePolicyRule struct {
	BasicCreateRequest
}

// NewCreatePolicyRule creates a new CreatePolicyRule object
func NewCreatePolicyRule(r *hubapi.PolicyRuleRequest) *CreatePolicyRule {
	return &CreatePolicyRule{BasicCreateRequest{request: r, responseCh: make(chan *EmptyResponse)}}
}

// Execute will tell the provided federator to create the policy rule in all hubs
func (cpr *CreatePolicyRule) Execute(fed FederatorInterface) error {
	funcs := api.HubFuncsType{
		Create: "CreatePolicyRule",
	}
	fed.SendCreateRequest(api.PolicyRulesEndpoint, funcs, cpr.request)
	cpr.responseCh <- &EmptyResponse{}
	return nil
}

// DeletePolicyRule handles deleting a policy rule
// in all the hubs known to the federator
type DeletePolicyRule struct {
	BasicDeleteRequest
}

// NewDeletePolicyRule creates a new DeletePolicyRule object
func NewDeletePolicyRule(id string) *DeletePolicyRule {
	return &DeletePolicyRule{BasicDeleteRequest{id: id, responseCh: make(chan *EmptyResponse)}}
}

// Execute will tell the provided federator to delete the policy rule in all hubs
func (dpr *DeletePolicyRule) Execute(fed FederatorInterface) error {
	funcs := api.HubFuncsType{
		Get:    "GetPolicyRule",
		GetAll: "ListAllPolicyRules",
		Delete: "DeletePolicyRule",
		SingleToList: func(single interface{}) interface{} {
			item := reflect.ValueOf(single).Interface()
			list := hubapi.PolicyRuleList{
				TotalCount: 1,
				Items:      []hubapi.PolicyRule{*item.(*hubapi.PolicyRule)},
			}
			return &list
		},
	}
	fed.SendDeleteRequest(api.PolicyRulesEndpoint, funcs, dpr.id)
	dpr.responseCh <- &EmptyResponse{}
	return nil
}
