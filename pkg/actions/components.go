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

// GetComponents handles retrieving components
// from all the hubs known to the federator
type GetComponents struct {
	endPoint    api.EndpointType
	componentID string
	responseCh  chan *GetResponse
}

// NewGetComponents creates a new GetComponents object
func NewGetComponents(id string, ep api.EndpointType) *GetComponents {
	return &GetComponents{componentID: id, endPoint: ep, responseCh: make(chan *GetResponse)}
}

// Execute will tell the provided federator to retrieve components
func (gc *GetComponents) Execute(fed FederatorInterface) error {
	var wg sync.WaitGroup
	var components hubapi.ComponentList
	var errs api.LastError

	hubs := fed.GetHubs()
	hubCount := len(hubs)
	componentsListCh := make(chan *hubapi.ComponentList, hubCount)
	errCh := make(chan HubError, hubCount)
	errs.Errors = make(map[string]*hubclient.HubClientError)

	wg.Add(hubCount)
	for hubURL, client := range hubs {
		go func(client *hub.Client, url string, id string) {
			defer wg.Done()
			if len(id) > 0 {
				link := hubapi.ResourceLink{Href: fmt.Sprintf("https://%s/api/components/%s", url, id)}
				log.Debugf("querying component %s", link.Href)
				c, err := client.GetComponent(link)
				log.Debugf("response to component query from %s: %+v", link.Href, c)
				if err != nil {
					hubErr := err.(*hubclient.HubClientError)
					errCh <- HubError{Host: url, Err: hubErr}
				} else {
					list := &hubapi.ComponentList{
						TotalCount: 1,
						Items:      []hubapi.Component{*c},
					}
					componentsListCh <- list
				}
			} else {
				log.Debugf("querying all components")
				list, err := client.ListAllComponents()
				if err != nil {
					log.Warningf("failed to get components from %s: %v", url, err)
					hubErr := err.(*hubclient.HubClientError)
					errCh <- HubError{Host: url, Err: hubErr}
				} else {
					componentsListCh <- list
				}
			}
		}(client, hubURL, gc.componentID)
	}

	wg.Wait()
	for i := 0; i < hubCount; i++ {
		select {
		case response := <-componentsListCh:
			if response != nil {
				log.Debugf("a hub responded with component list: %+v", response)
				gc.mergeComponentList(&components, response)
			}
		case err := <-errCh:
			errs.Errors[err.Host] = err.Err
		}
	}

	getResponse := GetResponse{
		endPoint: gc.endPoint,
		id:       gc.componentID,
		list:     &components,
	}

	fed.SetLastError(gc.endPoint, &errs)

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
