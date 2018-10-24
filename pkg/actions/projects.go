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

// GetProjectsRequestType defines the type of project request
type GetProjectsRequestType int

const (
	ProjectsGetAll GetProjectsRequestType = iota
	ProjectsGetOne
	ProjectsGetMany
)

// GetProjectsResponse defines the response for a GetProjects request
type GetProjectsResponse struct {
	requestType GetProjectsRequestType
	projectID   string
	allProjects *hubapi.ProjectList
}

// ReplaceSource will replace the source URL in the project list metadata
// with the federator information
func (resp *GetProjectsResponse) ReplaceSource(ip string) {
	if resp.requestType == ProjectsGetOne {
		resp.allProjects.Items[0].Meta.Href = fmt.Sprintf("https://%s/api/projects/%s", ip, resp.projectID)
	} else {
		/*
			if resp.requestType == ProjectsGetMany {
				resp.allProjects.Meta.Href = fmt.Sprintf("https://%s/api/all-projects", ip)
			} else {
				resp.allProjects.Meta.Href = fmt.Sprintf("https://%s/api/projects", ip)
			}
			if len(resp.projectID) > 0 {
				resp.allProjects.Meta.Href += fmt.Sprintf("/%s", resp.projectID)
			}
		*/
	}
}

// GetResult returns the project list
func (resp *GetProjectsResponse) GetResult() interface{} {
	if resp.requestType == ProjectsGetOne {
		return resp.allProjects.Items[0]
	}
	return resp.allProjects
}

// GetProjects handles retrieving projects
// from all the hubs known to the federator
type GetProjects struct {
	requestType GetProjectsRequestType
	projectID   string
	responseCh  chan *GetProjectsResponse
}

// NewGetProjects creates a new GetProjects object
func NewGetProjects(rt GetProjectsRequestType, id string) *GetProjects {
	return &GetProjects{requestType: rt, projectID: id, responseCh: make(chan *GetProjectsResponse)}
}

// Execute will tell the provided federator to retrieve projects
func (gp *GetProjects) Execute(fed FederatorInterface) error {
	var wg sync.WaitGroup
	var projects hubapi.ProjectList

	hubs := fed.GetHubs()
	log.Debugf("GetProjects federator hubs: %+v", hubs)
	hubCount := len(hubs)
	projectsListCh := make(chan *hubapi.ProjectList, hubCount)

	wg.Add(hubCount)
	for hubURL, client := range hubs {
		go func(client *hub.Client, url string, id string, rt GetProjectsRequestType) {
			defer wg.Done()
			if rt == ProjectsGetAll {
				log.Debugf("querying all projects")
				list, err := client.ListAllProjects()
				if err != nil {
					log.Warningf("failed to get projects from %s: %v", url, err)
					projectsListCh <- nil
				} else {
					projectsListCh <- list
				}
			} else {
				link := hubapi.ResourceLink{Href: fmt.Sprintf("https://%s/api/projects/%s", url, id)}
				log.Debugf("querying project %s", link.Href)
				cl, err := client.GetProject(link)
				log.Debugf("response to project query from %s: %+v", link.Href, cl)
				if err != nil {
					projectsListCh <- nil
				} else {
					list := &hubapi.ProjectList{
						TotalCount: 1,
						Items:      []hubapi.Project{*cl},
					}
					projectsListCh <- list
				}
			}
		}(client, hubURL, gp.projectID, gp.requestType)
	}

	wg.Wait()
	for i := 0; i < hubCount; i++ {
		response := <-projectsListCh
		if response != nil {
			log.Debugf("a hub responded with project list: %+v", response)
			gp.mergeProjectList(&projects, response)
		}
	}

	getResponse := GetProjectsResponse{
		requestType: gp.requestType,
		projectID:   gp.projectID,
		allProjects: &projects,
	}

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
