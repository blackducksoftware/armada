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
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/blackducksoftware/armada/pkg/actions"
	"github.com/blackducksoftware/armada/pkg/api"

	log "github.com/sirupsen/logrus"
)

// Responder implements fthe Responder interface
type Responder struct {
	localIP    string
	requestsCh chan actions.ActionInterface
}

// NewResponder creates a new Responder object
func NewResponder(ip net.IP) *Responder {
	return &Responder{localIP: ip.String(), requestsCh: make(chan actions.ActionInterface)}
}

/*
// GetModel .....
func (resp *Responder) GetModel() *APIModel {
	get := NewGetModelAction()
	go func() {
		resp.requestsCh <- get
	}()
	return <-get.Done
}
*/

// SetHubs ...
//func (r *Responder) SetHubs(hubs *APISetHubsRequest) {
func (resp *Responder) SetHubs(hubs *api.HubList) {
	resp.requestsCh <- &actions.SetHubs{Hubs: hubs}
}

/*
// UpdateConfig ...
func (resp *Responder) UpdateConfig(config *APIUpdateConfigRequest) {
	resp.requestsCh <- &UpdateConfigAction{}
}

// FindProject ...
func (resp *Responder) FindProject(request APIProjectSearchRequest) *APIProjectSearchResponse {
	req := NewFindProjectAction(request)
	resp.requestsCh <- req
	return <-req.Response
}
*/

// errors

// NotFound .....
func (resp *Responder) NotFound(w http.ResponseWriter, r *http.Request) {
	log.Errorf("Responder not found from request %+v", r)
	recordHTTPNotFound(r)
	http.NotFound(w, r)
}

// Error .....
func (resp *Responder) Error(w http.ResponseWriter, r *http.Request, err error, statusCode int) {
	log.Errorf("Responder error %s with code %d from request %+v", err.Error(), statusCode, r)
	recordHTTPError(r, err, statusCode)
	http.Error(w, err.Error(), statusCode)
}

// GetRequestCh returns the next request for the Responder
func (resp *Responder) GetRequestCh() chan actions.ActionInterface {
	return resp.requestsCh
}

func (resp *Responder) sendHTTPResponse(req actions.ActionInterface, w http.ResponseWriter, r *http.Request) {
	resp.requestsCh <- req
	actionResponse := req.GetResponse()
	actionResponse.ReplaceSource(resp.localIP)
	jsonBytes, err := json.MarshalIndent(actionResponse.GetResult(), "", "  ")
	if err != nil {
		resp.Error(w, r, err, 500)
	} else {
		header := w.Header()
		header.Set(http.CanonicalHeaderKey("content-type"), "application/json")
		fmt.Fprint(w, string(jsonBytes))
	}
}
