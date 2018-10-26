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

// GetComponentsRequestType defines the type of component request
type GetComponentsRequestType int

const (
	ComponentsGetAll GetComponentsRequestType = iota
	ComponentsGetOne
	ComponentsGetMany
)

// GetComponentsResponse defines the response for a GetComponents request
type GetComponentsResponse struct {
	requestType   GetComponentsRequestType
	componentID   string
	allComponents *hubapi.ComponentList
}

// ReplaceSource will replace the source URL in the component list metadata
// with the federator information
func (resp *GetComponentsResponse) ReplaceSource(ip string) {
	if resp.requestType == ComponentsGetOne {
		resp.allComponents.Items[0].Meta.Href = fmt.Sprintf("https://%s/api/components/%s", ip, resp.componentID)
	} else {
		if resp.requestType == ComponentsGetMany {
			resp.allComponents.Meta.Href = fmt.Sprintf("https://%s/api/all-components", ip)
		} else {
			resp.allComponents.Meta.Href = fmt.Sprintf("https://%s/api/components", ip)
		}
		if len(resp.componentID) > 0 {
			resp.allComponents.Meta.Href += fmt.Sprintf("/%s", resp.componentID)
		}
	}
}

// GetResult returns the component list
func (resp *GetComponentsResponse) GetResult() interface{} {
	if resp.requestType == ComponentsGetOne {
		return resp.allComponents.Items[0]
	}
	return resp.allComponents
}

// GetComponents handles retrieving components
// from all the hubs known to the federator
type GetComponents struct {
	requestType GetComponentsRequestType
	componentID string
	responseCh  chan *GetComponentsResponse
}

// NewGetComponents creates a new GetComponents object
func NewGetComponents(rt GetComponentsRequestType, id string) *GetComponents {
	return &GetComponents{requestType: rt, componentID: id, responseCh: make(chan *GetComponentsResponse)}
}

// Execute will tell the provided federator to retrieve components
func (gc *GetComponents) Execute(fed FederatorInterface) error {
	var wg sync.WaitGroup
	var components hubapi.ComponentList

	hubs := fed.GetHubs()
	hubCount := len(hubs)
	componentsListCh := make(chan *hubapi.ComponentList, hubCount)

	wg.Add(hubCount)
	for hubURL, client := range hubs {
		go func(client *hub.Client, url string, id string, rt GetComponentsRequestType) {
			defer wg.Done()
			if rt == ComponentsGetAll {
				log.Debugf("querying all components")
				list, err := client.ListAllComponents()
				if err != nil {
					log.Warningf("failed to get components from %s: %v", url, err)
					componentsListCh <- nil
				} else {
					componentsListCh <- list
				}
			} else {
				link := hubapi.ResourceLink{Href: fmt.Sprintf("https://%s/api/components/%s", url, id)}
				log.Debugf("querying component %s", link.Href)
				c, err := client.GetComponent(link)
				log.Debugf("response to component query from %s: %+v", link.Href, c)
				if err != nil {
					componentsListCh <- nil
				} else {
					list := &hubapi.ComponentList{
						TotalCount: 1,
						Items:      []hubapi.Component{*c},
					}
					componentsListCh <- list
				}
			}
		}(client, hubURL, gc.componentID, gc.requestType)
	}

	wg.Wait()
	for i := 0; i < hubCount; i++ {
		response := <-componentsListCh
		if response != nil {
			log.Debugf("a hub responded with component list: %+v", response)
			gc.mergeComponentList(&components, response)
		}
	}

	getResponse := GetComponentsResponse{
		requestType:   gc.requestType,
		componentID:   gc.componentID,
		allComponents: &components,
	}

	gc.responseCh <- &getResponse
	return nil
}

func (gc *GetComponents) mergeComponentList(origList, newList *hubapi.ComponentList) {
	origList.TotalCount += newList.TotalCount
	origList.Items = append(origList.Items, newList.Items...)
	origList.Meta.Allow = append(origList.Meta.Allow, newList.Meta.Allow...)
	origList.Meta.Links = append(origList.Meta.Links, newList.Meta.Links...)
}

// GetResponse returns the response to the get components query
func (gc *GetComponents) GetResponse() ActionResponseInterface {
	return <-gc.responseCh
}
