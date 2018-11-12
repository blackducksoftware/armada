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
	"sync"

	"github.com/blackducksoftware/armada/pkg/api"
	"github.com/blackducksoftware/armada/pkg/hub"

	"github.com/blackducksoftware/hub-client-go/hubapi"
	"github.com/blackducksoftware/hub-client-go/hubclient"

	log "github.com/sirupsen/logrus"
)

// GetUsers handles retrieving users
// from all the hubs known to the federator
type GetUsers struct {
	BasicGetRequest
}

// NewGetUsers creates a new GetUsers object
func NewGetUsers(id string, ep api.EndpointType) *GetUsers {
	return &GetUsers{BasicGetRequest{endPoint: ep, id: id, responseCh: make(chan *GetResponse)}}
}

// Execute will tell the provided federator to retrieve users
func (gu *GetUsers) Execute(fed FederatorInterface) error {
	var users hubapi.UserList

	funcs := api.GetFuncsType{
		Get:    "GetUser",
		GetAll: "ListAllUsers",
		SingleToList: func(single interface{}) interface{} {
			item := reflect.ValueOf(single).Interface()
			list := hubapi.UserList{
				TotalCount: 1,
				Items:      []hubapi.User{*item.(*hubapi.User)},
			}
			return &list
		},
	}
	fed.SendHubsGetRequest(gu.endPoint, funcs, gu.id, &users)

	gu.responseCh <- &GetResponse{
		endPoint: gu.endPoint,
		id:       gu.id,
		list:     &users,
	}

	return nil
}

// CreateUser handles creating a user
// in all the hubs known to the federator
type CreateUser struct {
	request    *hubapi.UserRequest
	responseCh chan *EmptyResponse
}

// NewCreateUser creates a new CreateUser object
func NewCreateUser(r *hubapi.UserRequest) *CreateUser {
	return &CreateUser{request: r, responseCh: make(chan *EmptyResponse)}
}

// Execute will tell the provided federator to create the user in all hubs
func (cu *CreateUser) Execute(fed FederatorInterface) error {
	var wg sync.WaitGroup
	var errs api.LastError

	hubs := fed.GetHubs()
	log.Debugf("CreateUser federator hubs: %+v", hubs)
	hubCount := len(hubs)
	usersCh := make(chan *hubapi.User, hubCount)
	errCh := make(chan HubError, hubCount)
	errs.Errors = make(map[string]*hubclient.HubClientError)

	wg.Add(hubCount)
	for hubURL, client := range hubs {
		go func(client *hub.Client, url string, req *hubapi.UserRequest) {
			defer wg.Done()
			log.Debugf("creating user %s", req.UserName)
			user, err := client.CreateUser(req)
			if err != nil {
				log.Warningf("failed to create user %s in %s: %v", req.UserName, url, err)
				hubErr := err.(*hubclient.HubClientError)
				errCh <- HubError{Host: url, Err: hubErr}
			} else {
				usersCh <- user
			}
		}(client, hubURL, cu.request)
	}

	wg.Wait()
	for i := 0; i < hubCount; i++ {
		select {
		case response := <-usersCh:
			if response != nil {
				log.Debugf("a hub responded with user: %+v", response)
			}
		case err := <-errCh:
			errs.Errors[err.Host] = err.Err
		}
	}

	fed.SetLastError(api.UsersEndpoint, &errs)

	cu.responseCh <- &EmptyResponse{}
	return nil
}

// GetResponse returns the response to the create users query
func (cu *CreateUser) GetResponse() ActionResponseInterface {
	return <-cu.responseCh
}
