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

// GetUsersRequestType defines the type of user request
type GetUsersRequestType int

const (
	UsersGetAll GetUsersRequestType = iota
	UsersGetOne
	UsersGetMany
)

// GetUsersResponse defines the response for a GetUsers request
type GetUsersResponse struct {
	requestType GetUsersRequestType
	userID      string
	allUsers    *hubapi.UserList
}

// ReplaceSource will replace the source URL in the user list metadata
// with the federator information
func (resp *GetUsersResponse) ReplaceSource(ip string) {
	if resp.requestType == UsersGetOne {
		resp.allUsers.Items[0].Meta.Href = fmt.Sprintf("https://%s/api/users/%s", ip, resp.userID)
	} else {
		/*
			if resp.requestType == UsersGetMany {
				resp.allUsers.Meta.Href = fmt.Sprintf("https://%s/api/all-users", ip)
			} else {
				resp.allUsers.Meta.Href = fmt.Sprintf("https://%s/api/users", ip)
			}
			if len(resp.userID) > 0 {
				resp.allUsers.Meta.Href += fmt.Sprintf("/%s", resp.userID)
			}
		*/
	}
}

// GetResult returns the user list
func (resp *GetUsersResponse) GetResult() interface{} {
	if resp.requestType == UsersGetOne {
		return resp.allUsers.Items[0]
	}
	return resp.allUsers
}

// GetUsers handles retrieving users
// from all the hubs known to the federator
type GetUsers struct {
	requestType GetUsersRequestType
	userID      string
	responseCh  chan *GetUsersResponse
}

// NewGetUsers creates a new GetUsers object
func NewGetUsers(rt GetUsersRequestType, id string) *GetUsers {
	return &GetUsers{requestType: rt, userID: id, responseCh: make(chan *GetUsersResponse)}
}

// Execute will tell the provided federator to retrieve users
func (gcl *GetUsers) Execute(fed FederatorInterface) error {
	var wg sync.WaitGroup
	var users hubapi.UserList

	hubs := fed.GetHubs()
	log.Debugf("GetUsers federator hubs: %+v", hubs)
	hubCount := len(hubs)
	usersListCh := make(chan *hubapi.UserList, hubCount)

	wg.Add(hubCount)
	for hubURL, client := range hubs {
		go func(client *hub.Client, url string, id string, rt GetUsersRequestType) {
			defer wg.Done()
			if rt == UsersGetAll {
				log.Debugf("querying all users")
				list, err := client.ListAllUsers()
				if err != nil {
					log.Warningf("failed to get users from %s: %v", url, err)
					usersListCh <- nil
				} else {
					usersListCh <- list
				}
			} else {
				link := hubapi.ResourceLink{Href: fmt.Sprintf("https://%s/api/users/%s", url, id)}
				log.Debugf("querying user %s", link.Href)
				cl, err := client.GetUser(link)
				log.Debugf("response to user query from %s: %+v", link.Href, cl)
				if err != nil {
					usersListCh <- nil
				} else {
					list := &hubapi.UserList{
						TotalCount: 1,
						Items:      []hubapi.User{*cl},
					}
					usersListCh <- list
				}
			}
		}(client, hubURL, gcl.userID, gcl.requestType)
	}

	wg.Wait()
	for i := 0; i < hubCount; i++ {
		response := <-usersListCh
		if response != nil {
			log.Debugf("a hub responded with codelocation list: %+v", response)
			gcl.mergeUserList(&users, response)
		}
	}

	getResponse := GetUsersResponse{
		requestType: gcl.requestType,
		userID:      gcl.userID,
		allUsers:    &users,
	}

	gcl.responseCh <- &getResponse
	return nil
}

func (gcl *GetUsers) mergeUserList(origList, newList *hubapi.UserList) {
	origList.TotalCount += newList.TotalCount
	origList.Items = append(origList.Items, newList.Items...)
	//	origList.Meta.Allow = append(origList.Meta.Allow, newList.Meta.Allow...)
	//	origList.Meta.Links = append(origList.Meta.Links, newList.Meta.Links...)
}

// GetResponse returns the response to the get users query
func (gcl *GetUsers) GetResponse() ActionResponseInterface {
	return <-gcl.responseCh
}
