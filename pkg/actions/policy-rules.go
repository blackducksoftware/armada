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

	"github.com/blackducksoftware/armada/pkg/hub"

	"github.com/blackducksoftware/hub-client-go/hubapi"

	log "github.com/sirupsen/logrus"
)

// GetPolicyRulesRequestType defines the type of policy rule request
type GetPolicyRulesRequestType int

const (
	PolicyRulesGetAll GetPolicyRulesRequestType = iota
	PolicyRulesGetOne
	PolicyRulesGetMany
)

// GetPolicyRulesResponse defines the response for a GetPolicyRules request
type GetPolicyRulesResponse struct {
	requestType    GetPolicyRulesRequestType
	policyRuleID   string
	allPolicyRules *hubapi.PolicyRuleList
}

// ReplaceSource will replace the source URL in the policy rule list metadata
// with the federator information
func (resp *GetPolicyRulesResponse) ReplaceSource(ip string) {
	if resp.requestType == PolicyRulesGetOne {
		resp.allPolicyRules.Items[0].Meta.Href = fmt.Sprintf("https://%s/api/policy-rules/%s", ip, resp.policyRuleID)
	} else {
		if resp.requestType == PolicyRulesGetMany {
			resp.allPolicyRules.Meta.Href = fmt.Sprintf("https://%s/api/all-policy-rules", ip)
		} else {
			resp.allPolicyRules.Meta.Href = fmt.Sprintf("https://%s/api/policy-rules", ip)
		}
		if len(resp.policyRuleID) > 0 {
			resp.allPolicyRules.Meta.Href += fmt.Sprintf("/%s", resp.policyRuleID)
		}
	}
}

// GetResult returns the policy rule list
func (resp *GetPolicyRulesResponse) GetResult() interface{} {
	if resp.requestType == PolicyRulesGetOne {
		return resp.allPolicyRules.Items[0]
	}
	return resp.allPolicyRules
}

// GetPolicyRules handles retrieving policy rules
// from all the hubs known to the federator
type GetPolicyRules struct {
	requestType  GetPolicyRulesRequestType
	policyRuleID string
	responseCh   chan *GetPolicyRulesResponse
}

// NewGetPolicyRules creates a new GetPolicyRules object
func NewGetPolicyRules(rt GetPolicyRulesRequestType, id string) *GetPolicyRules {
	return &GetPolicyRules{requestType: rt, policyRuleID: id, responseCh: make(chan *GetPolicyRulesResponse)}
}

// Execute will tell the provided federator to retrieve policy rules
func (gpr *GetPolicyRules) Execute(fed FederatorInterface) error {
	var wg sync.WaitGroup
	var policyRules hubapi.PolicyRuleList

	hubs := fed.GetHubs()
	log.Debugf("GetPolicyRules federator hubs: %+v", hubs)
	hubCount := len(hubs)
	policyRulesListCh := make(chan *hubapi.PolicyRuleList, hubCount)

	wg.Add(hubCount)
	for hubURL, client := range hubs {
		go func(client *hub.Client, url string, id string, rt GetPolicyRulesRequestType) {
			defer wg.Done()
			if rt == PolicyRulesGetAll {
				log.Debugf("querying all policy rules")
				list, err := client.ListAllPolicyRules()
				if err != nil {
					log.Warningf("failed to get policy rules from %s: %v", url, err)
					policyRulesListCh <- nil
				} else {
					policyRulesListCh <- list
				}
			} else {
				link := hubapi.ResourceLink{Href: fmt.Sprintf("https://%s/api/policy-rules/%s", url, id)}
				log.Debugf("querying policy rule %s", link.Href)
				p, err := client.GetPolicyRule(link)
				log.Debugf("response to policy rule query from %s: %+v", link.Href, p)
				if err != nil {
					policyRulesListCh <- nil
				} else {
					list := &hubapi.PolicyRuleList{
						TotalCount: 1,
						Items:      []hubapi.PolicyRule{*p},
					}
					policyRulesListCh <- list
				}
			}
		}(client, hubURL, gpr.policyRuleID, gpr.requestType)
	}

	wg.Wait()
	for i := 0; i < hubCount; i++ {
		response := <-policyRulesListCh
		if response != nil {
			log.Debugf("a hub responded with policy rule list: %+v", response)
			gpr.mergePolicyRuleList(&policyRules, response)
		}
	}

	getResponse := GetPolicyRulesResponse{
		requestType:    gpr.requestType,
		policyRuleID:   gpr.policyRuleID,
		allPolicyRules: &policyRules,
	}

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
	responseCh chan *CreateResponse
}

// NewCreatePolicyRule creates a new CreatePolicyRule object
func NewCreatePolicyRule(r *hubapi.PolicyRuleRequest) *CreatePolicyRule {
	return &CreatePolicyRule{request: r, responseCh: make(chan *CreateResponse)}
}

// Execute will tell the provided federator to create the policy rule in all hubs
func (cpr *CreatePolicyRule) Execute(fed FederatorInterface) error {
	var wg sync.WaitGroup
	var policyRules []string

	hubs := fed.GetHubs()
	log.Debugf("CreatePolicyRule federator hubs: %+v", hubs)
	hubCount := len(hubs)
	policyRulesCh := make(chan string, hubCount)

	wg.Add(hubCount)
	for hubURL, client := range hubs {
		go func(client *hub.Client, url string, req *hubapi.PolicyRuleRequest) {
			defer wg.Done()
			log.Debugf("creating policy rule %s", req.Name)
			pr, err := client.CreatePolicyRule(req)
			if err != nil {
				log.Warningf("failed to create policy rule %s in %s: %v", req.Name, url, err)
				policyRulesCh <- ""
			} else {
				policyRulesCh <- pr
			}
		}(client, hubURL, cpr.request)
	}

	wg.Wait()
	for i := 0; i < hubCount; i++ {
		response := <-policyRulesCh
		if len(response) > 0 {
			log.Debugf("a hub responded with policy rule: %+v", response)
			policyRules = append(policyRules, response)
		}
	}

	cpr.responseCh <- &CreateResponse{}
	return nil
}

// GetResponse returns the response to the create policyRules query
func (cpr *CreatePolicyRule) GetResponse() ActionResponseInterface {
	return <-cpr.responseCh
}
