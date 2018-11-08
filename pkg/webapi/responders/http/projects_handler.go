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
	"github.com/blackducksoftware/armada/pkg/api"

	"github.com/gorilla/mux"

	log "github.com/sirupsen/logrus"
)

// ProjectsHandler will process requests for projects
func (resp *Responder) ProjectsHandler(w http.ResponseWriter, r *http.Request) {
	var req actions.ActionInterface

	vars := mux.Vars(r)
	project, ok := vars["projectId"]
	if r.Method == http.MethodGet {
		if !ok {
			// No specific project was given, so this is a request for all projects
			log.Info("retrieving all projects")
		} else {
			// Look up the specific project
			log.Infof("retrieving project %s", project)
		}
		req = actions.NewGetProjects(project, api.ProjectsEndpoint)
		resp.sendHTTPResponse(req, w, r)
	} else {
		resp.Error(w, r, fmt.Errorf("unsupported method %s", r.Method), 405)
	}
}

// AllProjectsHandler will process requests for a project from all hubs
func (resp *Responder) AllProjectsHandler(w http.ResponseWriter, r *http.Request) {
	var req actions.ActionInterface

	vars := mux.Vars(r)
	project, ok := vars["projectId"]
	if r.Method == http.MethodGet {
		if !ok {
			log.Errorf("no project provided")
		} else {
			// Look up the specific project
			log.Infof("retrieving project %s", project)
			req = actions.NewGetProjects(project, api.AllProjectsEndpoint)
		}
		resp.sendHTTPResponse(req, w, r)
	} else {
		resp.Error(w, r, fmt.Errorf("unsupported method %s", r.Method), 405)
	}
}
