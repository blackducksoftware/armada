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

// GetCodeLocations handles retrieving code locations
// from all the hubs known to the federator
type GetCodeLocations struct {
	endPoint       api.EndpointType
	codeLocationID string
	responseCh     chan *GetResponse
}

// NewGetCodeLocations creates a new GetCodeLocations object
func NewGetCodeLocations(id string, ep api.EndpointType) *GetCodeLocations {
	return &GetCodeLocations{codeLocationID: id, endPoint: ep, responseCh: make(chan *GetResponse)}
}

// Execute will tell the provided federator to retrieve code locations
func (gcl *GetCodeLocations) Execute(fed FederatorInterface) error {
	var wg sync.WaitGroup
	var codeLocations hubapi.CodeLocationList
	var errs api.LastError

	hubs := fed.GetHubs()
	log.Debugf("GetCodeLocations federator hubs: %+v", hubs)
	hubCount := len(hubs)
	codeLocationsListCh := make(chan *hubapi.CodeLocationList, hubCount)
	errCh := make(chan HubError, hubCount)
	errs.Errors = make(map[string]*hubclient.HubClientError)

	wg.Add(hubCount)
	for hubURL, client := range hubs {
		go func(client *hub.Client, url string, id string) {
			defer wg.Done()
			if len(id) > 0 {
				link := hubapi.ResourceLink{Href: fmt.Sprintf("https://%s/api/codelocations/%s", url, id)}
				log.Debugf("querying code location %s", link.Href)
				cl, err := client.GetCodeLocation(link)
				log.Debugf("response to code location query from %s: %+v", link.Href, cl)
				if err != nil {
					hubErr := err.(*hubclient.HubClientError)
					errCh <- HubError{Host: url, Err: hubErr}
				} else {
					list := &hubapi.CodeLocationList{
						TotalCount: 1,
						Items:      []hubapi.CodeLocation{*cl},
					}
					codeLocationsListCh <- list
				}
			} else {
				log.Debugf("querying all code locations")
				list, err := client.ListAllCodeLocations()
				if err != nil {
					log.Warningf("failed to get code locations from %s: %v", url, err)
					hubErr := err.(*hubclient.HubClientError)
					errCh <- HubError{Host: url, Err: hubErr}
				} else {
					codeLocationsListCh <- list
				}
			}
		}(client, hubURL, gcl.codeLocationID)
	}

	wg.Wait()
	for i := 0; i < hubCount; i++ {
		select {
		case response := <-codeLocationsListCh:
			if response != nil {
				log.Debugf("a hub responded with codelocation list: %+v", response)
				gcl.mergeCodeLocationList(&codeLocations, response)
			}
		case err := <-errCh:
			errs.Errors[err.Host] = err.Err
		}
	}

	getResponse := GetResponse{
		endPoint: gcl.endPoint,
		id:       gcl.codeLocationID,
		list:     &codeLocations,
	}

	fed.SetLastError(gcl.endPoint, &errs)

	gcl.responseCh <- &getResponse
	return nil
}

func (gcl *GetCodeLocations) mergeCodeLocationList(origList, newList *hubapi.CodeLocationList) {
	origList.TotalCount += newList.TotalCount
	origList.Items = append(origList.Items, newList.Items...)
	origList.Meta.Allow = append(origList.Meta.Allow, newList.Meta.Allow...)
	origList.Meta.Links = append(origList.Meta.Links, newList.Meta.Links...)
}

// GetResponse returns the response to the get code locations query
func (gcl *GetCodeLocations) GetResponse() ActionResponseInterface {
	return <-gcl.responseCh
}
