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

package responders

import (
	"net/http"

	"github.com/blackducksoftware/armada/pkg/actions"
	"github.com/blackducksoftware/armada/pkg/api"
)

// ResponderInterface defines the interface for REST handlers
type ResponderInterface interface {
	GetRequestCh() chan actions.ActionInterface

	//
	SetHubs(hubs *api.HubList)
	//	UpdateConfig(config *APIUpdateConfigRequest)
	//
	//	FindProject(request APIProjectSearchRequest) *APIProjectSearchResponse

	// Handlers
	ModelHandler(w http.ResponseWriter, r *http.Request)
	CodeLocationsHandler(w http.ResponseWriter, r *http.Request)
	AllCodeLocationsHandler(w http.ResponseWriter, r *http.Request)
	ProjectsHandler(w http.ResponseWriter, r *http.Request)
	AllProjectsHandler(w http.ResponseWriter, r *http.Request)
	PolicyRulesHandler(w http.ResponseWriter, r *http.Request)
	AllPolicyRulesHandler(w http.ResponseWriter, r *http.Request)
	UsersHandler(w http.ResponseWriter, r *http.Request)
	AllUsersHandler(w http.ResponseWriter, r *http.Request)
	ComponentsHandler(w http.ResponseWriter, r *http.Request)
	AllComponentsHandler(w http.ResponseWriter, r *http.Request)
	LastErrorHandler(w http.ResponseWriter, r *http.Request)

	// errors
	NotFound(w http.ResponseWriter, r *http.Request)
	Error(w http.ResponseWriter, r *http.Request, err error, statusCode int)
}
