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

package http

import (
	"fmt"
	"net/http"

	"github.com/blackducksoftware/armada/pkg/actions"

	"github.com/blackducksoftware/hub-client-go/hubapi"

	"github.com/gorilla/mux"

	log "github.com/sirupsen/logrus"
)

// PolicyRulesHandler will process requests for policy rules
func (resp *Responder) PolicyRulesHandler(w http.ResponseWriter, r *http.Request) {
	var req actions.ActionInterface

	vars := mux.Vars(r)
	policyRule, ok := vars["policyRuleId"]
	if r.Method == http.MethodGet {
		if !ok {
			// No specific policy rule was given, so this is a request for all policy rules
			log.Info("retrieving all policy rules")
			req = actions.NewGetPolicyRules(actions.PolicyRulesGetAll, policyRule)
		} else {
			// Look up the specific policy rule
			log.Infof("retrieving policy rule %s", policyRule)
			req = actions.NewGetPolicyRules(actions.PolicyRulesGetOne, policyRule)
		}
		resp.sendHTTPResponse(req, w, r)
	} else if r.Method == http.MethodPost {
		// Create the policy rule
		var request hubapi.PolicyRuleRequest
		ok := resp.unmarshalRequest(w, r, &request)
		if ok {
			req = actions.NewCreatePolicyRule(&request)
			resp.sendHTTPResponse(req, w, r)
		}
	} else {
		resp.Error(w, r, fmt.Errorf("unsupported method %s", r.Method), 405)
	}
}

// AllPolicyRulesHandler will process requests for a policy rule from all hubs
func (resp *Responder) AllPolicyRulesHandler(w http.ResponseWriter, r *http.Request) {
	var req actions.ActionInterface

	vars := mux.Vars(r)
	policyRule, ok := vars["policyRuleId"]
	if r.Method == http.MethodGet {
		if !ok {
			log.Errorf("no policy rule provided")
		} else {
			// Look up the specific policy rule
			log.Infof("retrieving policy rule %s", policyRule)
			req = actions.NewGetPolicyRules(actions.PolicyRulesGetMany, policyRule)
		}
		resp.sendHTTPResponse(req, w, r)
	} else {
		resp.Error(w, r, fmt.Errorf("unsupported method %s", r.Method), 405)
	}
}
