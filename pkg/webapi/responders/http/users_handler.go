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

	"github.com/blackducksoftware/hub-client-go/hubapi"

	"github.com/gorilla/mux"

	log "github.com/sirupsen/logrus"
)

// UsersHandler will process requests for users
func (resp *Responder) UsersHandler(w http.ResponseWriter, r *http.Request) {
	var req actions.ActionInterface

	vars := mux.Vars(r)
	user, ok := vars["userId"]
	if r.Method == http.MethodGet {
		if !ok {
			// No specific user was given, so this is a request for all users
			log.Info("retrieving all users")
		} else {
			// Look up the specific user
			log.Infof("retrieving user %s", user)
		}
		req = actions.NewGetUsers(user, api.UsersEndpoint)
		resp.sendHTTPResponse(req, w, r)
	} else if r.Method == http.MethodPost {
		// Create the user
		var request hubapi.UserRequest
		ok := resp.unmarshalRequest(w, r, &request)
		if ok {
			req = actions.NewCreateUser(&request)
			resp.sendHTTPResponse(req, w, r)
		}
	} else {
		resp.Error(w, r, fmt.Errorf("unsupported method %s", r.Method), 405)
	}
}

// AllUsersHandler will process requests for a user from all hubs
func (resp *Responder) AllUsersHandler(w http.ResponseWriter, r *http.Request) {
	var req actions.ActionInterface

	vars := mux.Vars(r)
	user, ok := vars["userId"]
	if r.Method == http.MethodGet {
		if !ok {
			log.Errorf("no user provided")
		} else {
			// Look up the specific user
			log.Infof("retrieving user %s", user)
			req = actions.NewGetUsers(user, api.AllUsersEndpoint)
		}
		resp.sendHTTPResponse(req, w, r)
	} else {
		resp.Error(w, r, fmt.Errorf("unsupported method %s", r.Method), 405)
	}
}
