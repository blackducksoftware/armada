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

package mock

import (
	"net/http"

	"github.com/blackducksoftware/armada/pkg/actions"
	"github.com/blackducksoftware/armada/pkg/api"
)

// Responder ...
type Responder struct {
	requestsCh chan actions.ActionInterface
}

// NewResponder ...
func NewResponder() *Responder {
	return &Responder{requestsCh: make(chan actions.ActionInterface)}
}

// SetHubs ...
func (mr *Responder) SetHubs(hubs *api.HubList) {}

/*
// UpdateConfig ...
func (mr *MockResponder) UpdateConfig(config *APIUpdateConfigRequest) {}

// FindProject ...
func (mr *MockResponder) FindProject(request APIProjectSearchRequest) *APIProjectSearchResponse {
	return &APIProjectSearchResponse{}
}
*/

// NotFound ...
func (mr *Responder) NotFound(w http.ResponseWriter, r *http.Request) {}

// Error ...
func (mr *Responder) Error(w http.ResponseWriter, r *http.Request, err error, statusCode int) {}

// GetRequestCh returns the next request for the Responder
func (mr *Responder) GetRequestCh() chan actions.ActionInterface {
	return mr.requestsCh
}

func (mr *Responder) ModelHandler(w http.ResponseWriter, r *http.Request)            {}
func (mr *Responder) CodeLocationsHandler(w http.ResponseWriter, r *http.Request)    {}
func (mr *Responder) AllCodeLocationsHandler(w http.ResponseWriter, r *http.Request) {}
func (mr *Responder) ProjectsHandler(w http.ResponseWriter, r *http.Request)         {}
func (mr *Responder) AllProjectsHandler(w http.ResponseWriter, r *http.Request)      {}
func (mr *Responder) PolicyRulesHandler(w http.ResponseWriter, r *http.Request)      {}
func (mr *Responder) AllPolicyRulesHandler(w http.ResponseWriter, r *http.Request)   {}
func (mr *Responder) UsersHandler(w http.ResponseWriter, r *http.Request)            {}
func (mr *Responder) AllUsersHandler(w http.ResponseWriter, r *http.Request)         {}
func (mr *Responder) ComponentsHandler(w http.ResponseWriter, r *http.Request)       {}
func (mr *Responder) AllComponentsHandler(w http.ResponseWriter, r *http.Request)    {}
