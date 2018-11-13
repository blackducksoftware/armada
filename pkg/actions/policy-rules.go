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
	"sync"

	"github.com/blackducksoftware/armada/pkg/api"
	"github.com/blackducksoftware/armada/pkg/hub"

	"github.com/blackducksoftware/hub-client-go/hubapi"
	"github.com/blackducksoftware/hub-client-go/hubclient"

	log "github.com/sirupsen/logrus"
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

	funcs := api.GetFuncsType{
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
	fed.SendCreateRequest(api.PolicyRulesEndpoint, "CreatePolicyRule", cpr.request)
	cpr.responseCh <- &EmptyResponse{}
	return nil
}

// DeletePolicyRule handles deleting a policy rule
// in all the hubs known to the federator
type DeletePolicyRule struct {
	policyRuleID string
	responseCh   chan *EmptyResponse
}

// NewDeletePolicyRule creates a new DeletePolicyRule object
func NewDeletePolicyRule(id string) *DeletePolicyRule {
	return &DeletePolicyRule{policyRuleID: id, responseCh: make(chan *EmptyResponse)}
}

// Execute will tell the provided federator to delete the policy rule in all hubs
func (dpr *DeletePolicyRule) Execute(fed FederatorInterface) error {
	var wg sync.WaitGroup
	var errs api.LastError

	hubs := fed.GetHubs()
	log.Debugf("DeletePolicyRule federator hubs: %+v", hubs)
	hubCount := len(hubs)
	errCh := make(chan HubError, hubCount)
	errs.Errors = make(map[string]*hubclient.HubClientError)

	wg.Add(hubCount)
	for hubURL, client := range hubs {
		go func(client *hub.Client, url string, id string) {
			defer wg.Done()
			log.Debugf("deleting policy rule %s from hub %s", id, url)
			loc := fmt.Sprintf("https://%s/api/policy-rules/%s", url, id)
			err := client.DeletePolicyRule(loc)
			if err != nil {
				log.Warningf("failed to delete policy rule %s in %s: %v", id, url, err)
			}
			hubErr := err.(*hubclient.HubClientError)
			errCh <- HubError{Host: url, Err: hubErr}
		}(client, hubURL, dpr.policyRuleID)
	}
	wg.Wait()

	for i := 0; i < hubCount; i++ {
		err := <-errCh
		if err.Err != nil {
			errs.Errors[err.Host] = err.Err
		}
	}

	fed.SetLastError(api.PolicyRulesEndpoint, &errs)

	dpr.responseCh <- &EmptyResponse{}
	return nil
}

// GetResponse returns the response to the delete policy rules request
func (dpr *DeletePolicyRule) GetResponse() ActionResponseInterface {
	return <-dpr.responseCh
}
