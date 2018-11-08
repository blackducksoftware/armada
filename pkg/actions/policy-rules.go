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
	endPoint     api.EndpointType
	policyRuleID string
	responseCh   chan *GetResponse
}

// NewGetPolicyRules creates a new GetPolicyRules object
func NewGetPolicyRules(id string, ep api.EndpointType) *GetPolicyRules {
	return &GetPolicyRules{policyRuleID: id, responseCh: make(chan *GetResponse), endPoint: ep}
}

// Execute will tell the provided federator to retrieve policy rules
func (gpr *GetPolicyRules) Execute(fed FederatorInterface) error {
	var wg sync.WaitGroup
	var policyRules hubapi.PolicyRuleList
	var errs api.LastError

	hubs := fed.GetHubs()
	log.Debugf("GetPolicyRules federator hubs: %+v", hubs)
	hubCount := len(hubs)
	policyRulesListCh := make(chan *hubapi.PolicyRuleList, hubCount)
	errCh := make(chan HubError, hubCount)
	errs.Errors = make(map[string]*hubclient.HubClientError)

	wg.Add(hubCount)
	for hubURL, client := range hubs {
		go func(client *hub.Client, url string, id string) {
			defer wg.Done()
			if len(id) > 0 {
				link := hubapi.ResourceLink{Href: fmt.Sprintf("https://%s/api/policy-rules/%s", url, id)}
				log.Debugf("querying policy rule %s", link.Href)
				p, err := client.GetPolicyRule(link)
				log.Debugf("response to policy rule query from %s: %+v", link.Href, p)
				if err != nil {
					hubErr := err.(*hubclient.HubClientError)
					errCh <- HubError{Host: url, Err: hubErr}
				} else {
					list := &hubapi.PolicyRuleList{
						TotalCount: 1,
						Items:      []hubapi.PolicyRule{*p},
					}
					policyRulesListCh <- list
				}
			} else {
				log.Debugf("querying all policy rules")
				list, err := client.ListAllPolicyRules()
				if err != nil {
					log.Warningf("failed to get policy rules from %s: %v", url, err)
					hubErr := err.(*hubclient.HubClientError)
					errCh <- HubError{Host: url, Err: hubErr}
				} else {
					policyRulesListCh <- list
				}
			}
		}(client, hubURL, gpr.policyRuleID)
	}

	wg.Wait()
	for i := 0; i < hubCount; i++ {
		select {
		case response := <-policyRulesListCh:
			if response != nil {
				log.Debugf("a hub responded with policy rule list: %+v", response)
				gpr.mergePolicyRuleList(&policyRules, response)
			}
		case err := <-errCh:
			errs.Errors[err.Host] = err.Err
		}
	}

	getResponse := GetResponse{
		endPoint: gpr.endPoint,
		id:       gpr.policyRuleID,
		list:     &policyRules,
	}

	fed.SetLastError(gpr.endPoint, &errs)

	gpr.responseCh <- &getResponse
	return nil
}

func (gpr *GetPolicyRules) mergePolicyRuleList(origList, newList *hubapi.PolicyRuleList) {
	origList.TotalCount += newList.TotalCount
	origList.Items = append(origList.Items, newList.Items...)
	origList.Meta.Allow = append(origList.Meta.Allow, newList.Meta.Allow...)
	origList.Meta.Links = append(origList.Meta.Links, newList.Meta.Links...)
}

// GetResponse returns the response to the get policy rules query
func (gpr *GetPolicyRules) GetResponse() ActionResponseInterface {
	return <-gpr.responseCh
}

// CreatePolicyRule handles creating a policy rule
// in all the hubs known to the federator
type CreatePolicyRule struct {
	request    *hubapi.PolicyRuleRequest
	responseCh chan *EmptyResponse
}

// NewCreatePolicyRule creates a new CreatePolicyRule object
func NewCreatePolicyRule(r *hubapi.PolicyRuleRequest) *CreatePolicyRule {
	return &CreatePolicyRule{request: r, responseCh: make(chan *EmptyResponse)}
}

// Execute will tell the provided federator to create the policy rule in all hubs
func (cpr *CreatePolicyRule) Execute(fed FederatorInterface) error {
	var wg sync.WaitGroup
	var policyRules []string
	var errs api.LastError

	hubs := fed.GetHubs()
	log.Debugf("CreatePolicyRule federator hubs: %+v", hubs)
	hubCount := len(hubs)
	policyRulesCh := make(chan string, hubCount)
	errCh := make(chan HubError, hubCount)
	errs.Errors = make(map[string]*hubclient.HubClientError)

	wg.Add(hubCount)
	for hubURL, client := range hubs {
		go func(client *hub.Client, url string, req *hubapi.PolicyRuleRequest) {
			defer wg.Done()
			log.Debugf("creating policy rule %s", req.Name)
			pr, err := client.CreatePolicyRule(req)
			if err != nil {
				log.Warningf("failed to create policy rule %s in %s: %v", req.Name, url, err)
				hubErr := err.(*hubclient.HubClientError)
				errCh <- HubError{Host: url, Err: hubErr}
			} else {
				policyRulesCh <- pr
			}
		}(client, hubURL, cpr.request)
	}

	wg.Wait()
	for i := 0; i < hubCount; i++ {
		select {
		case response := <-policyRulesCh:
			if len(response) > 0 {
				log.Debugf("a hub responded with policy rule: %+v", response)
				policyRules = append(policyRules, response)
			}
		case err := <-errCh:
			errs.Errors[err.Host] = err.Err
		}
	}

	fed.SetLastError(api.PolicyRulesEndpoint, &errs)

	cpr.responseCh <- &EmptyResponse{}
	return nil
}

// GetResponse returns the response to the create policy rules request
func (cpr *CreatePolicyRule) GetResponse() ActionResponseInterface {
	return <-cpr.responseCh
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
