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

// GetCodeLocationsRequestType defines the type of code location request
type GetCodeLocationsRequestType int

const (
	CodeLocationsGetAll GetCodeLocationsRequestType = iota
	CodeLocationsGetOne
	CodeLocationsGetMany
)

// GetCodeLocationsResponse defines the response for a GetCodeLocations request
type GetCodeLocationsResponse struct {
	requestType      GetCodeLocationsRequestType
	codeLocationID   string
	allCodeLocations *hubapi.CodeLocationList
}

// ReplaceSource will replace the source URL in the code location list metadata
// with the federator information
func (resp *GetCodeLocationsResponse) ReplaceSource(ip string) {
	if resp.requestType == CodeLocationsGetOne {
		resp.allCodeLocations.Items[0].Meta.Href = fmt.Sprintf("https://%s/api/codelocations/%s", ip, resp.codeLocationID)
	} else {
		if resp.requestType == CodeLocationsGetMany {
			resp.allCodeLocations.Meta.Href = fmt.Sprintf("https://%s/api/all-codelocations", ip)
		} else {
			resp.allCodeLocations.Meta.Href = fmt.Sprintf("https://%s/api/codelocations", ip)
		}
		if len(resp.codeLocationID) > 0 {
			resp.allCodeLocations.Meta.Href += fmt.Sprintf("/%s", resp.codeLocationID)
		}
	}
}

// GetResult returns the code location list
func (resp *GetCodeLocationsResponse) GetResult() interface{} {
	if resp.requestType == CodeLocationsGetOne {
		return resp.allCodeLocations.Items[0]
	}
	return resp.allCodeLocations
}

// GetCodeLocations handles retrieving code locations
// from all the hubs known to the federator
type GetCodeLocations struct {
	requestType    GetCodeLocationsRequestType
	codeLocationID string
	responseCh     chan *GetCodeLocationsResponse
}

// NewGetCodeLocations creates a new GetCodeLocations object
func NewGetCodeLocations(rt GetCodeLocationsRequestType, id string) *GetCodeLocations {
	return &GetCodeLocations{requestType: rt, codeLocationID: id, responseCh: make(chan *GetCodeLocationsResponse)}
}

// Execute will tell the provided federator to retrieve code locations
func (gcl *GetCodeLocations) Execute(fed FederatorInterface) error {
	var wg sync.WaitGroup
	var codeLocations hubapi.CodeLocationList

	hubs := fed.GetHubs()
	log.Debugf("GetCodeLocations federator hubs: %+v", hubs)
	hubCount := len(hubs)
	codeLocationsListCh := make(chan *hubapi.CodeLocationList, hubCount)

	wg.Add(hubCount)
	for hubURL, client := range hubs {
		go func(client *hub.Client, url string, id string, rt GetCodeLocationsRequestType) {
			defer wg.Done()
			if rt == CodeLocationsGetAll {
				log.Debugf("querying all code locations")
				list, err := client.ListAllCodeLocations()
				if err != nil {
					log.Warningf("failed to get code locations from %s: %v", url, err)
					codeLocationsListCh <- nil
				} else {
					codeLocationsListCh <- list
				}
			} else {
				link := hubapi.ResourceLink{Href: fmt.Sprintf("https://%s/api/codelocations/%s", url, id)}
				log.Debugf("querying code location %s", link.Href)
				cl, err := client.GetCodeLocation(link)
				log.Debugf("response to code location query from %s: %+v", link.Href, cl)
				if err != nil {
					codeLocationsListCh <- nil
				} else {
					list := &hubapi.CodeLocationList{
						TotalCount: 1,
						Items:      []hubapi.CodeLocation{*cl},
					}
					codeLocationsListCh <- list
				}
			}
		}(client, hubURL, gcl.codeLocationID, gcl.requestType)
	}

	wg.Wait()
	for i := 0; i < hubCount; i++ {
		response := <-codeLocationsListCh
		if response != nil {
			log.Debugf("a hub responded with codelocation list: %+v", response)
			gcl.mergeCodeLocationList(&codeLocations, response)
		}
	}

	getResponse := GetCodeLocationsResponse{
		requestType:      gcl.requestType,
		codeLocationID:   gcl.codeLocationID,
		allCodeLocations: &codeLocations,
	}

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
