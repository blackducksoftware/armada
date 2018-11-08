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

// GetProjects handles retrieving projects
// from all the hubs known to the federator
type GetProjects struct {
	endPoint   api.EndpointType
	projectID  string
	responseCh chan *GetResponse
}

// NewGetProjects creates a new GetProjects object
func NewGetProjects(id string, ep api.EndpointType) *GetProjects {
	return &GetProjects{projectID: id, endPoint: ep, responseCh: make(chan *GetResponse)}
}

// Execute will tell the provided federator to retrieve projects
func (gp *GetProjects) Execute(fed FederatorInterface) error {
	var wg sync.WaitGroup
	var projects hubapi.ProjectList
	var errs api.LastError

	hubs := fed.GetHubs()
	log.Debugf("GetProjects federator hubs: %+v", hubs)
	hubCount := len(hubs)
	projectsListCh := make(chan *hubapi.ProjectList, hubCount)
	errCh := make(chan HubError, hubCount)
	errs.Errors = make(map[string]*hubclient.HubClientError)

	wg.Add(hubCount)
	for hubURL, client := range hubs {
		go func(client *hub.Client, url string, id string) {
			defer wg.Done()
			if len(id) > 0 {
				link := hubapi.ResourceLink{Href: fmt.Sprintf("https://%s/api/projects/%s", url, id)}
				log.Debugf("querying project %s", link.Href)
				cl, err := client.GetProject(link)
				log.Debugf("response to project query from %s: %+v", link.Href, cl)
				if err != nil {
					hubErr := err.(*hubclient.HubClientError)
					errCh <- HubError{Host: url, Err: hubErr}
				} else {
					list := &hubapi.ProjectList{
						TotalCount: 1,
						Items:      []hubapi.Project{*cl},
					}
					projectsListCh <- list
				}
			} else {
				log.Debugf("querying all projects")
				list, err := client.ListAllProjects()
				if err != nil {
					log.Warningf("failed to get projects from %s: %v", url, err)
					hubErr := err.(*hubclient.HubClientError)
					errCh <- HubError{Host: url, Err: hubErr}
				} else {
					projectsListCh <- list
				}
			}
		}(client, hubURL, gp.projectID)
	}

	wg.Wait()
	for i := 0; i < hubCount; i++ {
		select {
		case response := <-projectsListCh:
			if response != nil {
				log.Debugf("a hub responded with project list: %+v", response)
				gp.mergeProjectList(&projects, response)
			}
		case err := <-errCh:
			errs.Errors[err.Host] = err.Err
		}
	}

	getResponse := GetResponse{
		endPoint: gp.endPoint,
		id:       gp.projectID,
		list:     &projects,
	}

	fed.SetLastError(gp.endPoint, &errs)

	gp.responseCh <- &getResponse
	return nil
}

func (gp *GetProjects) mergeProjectList(origList, newList *hubapi.ProjectList) {
	origList.TotalCount += newList.TotalCount
	origList.Items = append(origList.Items, newList.Items...)
	//	origList.Meta.Allow = append(origList.Meta.Allow, newList.Meta.Allow...)
	//	origList.Meta.Links = append(origList.Meta.Links, newList.Meta.Links...)
}

// GetResponse returns the response to the get projects query
func (gp *GetProjects) GetResponse() ActionResponseInterface {
	return <-gp.responseCh
}
