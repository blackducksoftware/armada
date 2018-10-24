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

	"github.com/blackducksoftware/armada/pkg/api"
	"github.com/blackducksoftware/armada/pkg/hub"
)

// SetHubs will set the list of hubs the federator
type SetHubs struct {
	Hubs *api.HubList
}

// Execute will set the hubs in the provided federator
func (sh *SetHubs) Execute(fed FederatorInterface) error {
	newHubURLs := map[string]bool{}
	hubsToCreate := api.HubList{}

	hubs := fed.GetHubs()
	for _, newHub := range sh.Hubs.Items {
		hubURL := fmt.Sprintf("https://%s:%d", newHub.Host, newHub.Port)
		newHubURLs[hubURL] = true
		if _, ok := hubs[hubURL]; !ok {
			hubsToCreate.Items = append(hubsToCreate.Items, newHub)
		}
	}

	// 1. create new hubs
	// TODO handle retries and failures intelligently
	fed.CreateHubClients(&hubsToCreate)

	// 2. delete removed hubs
	for hubURL := range hubs {
		if _, ok := newHubURLs[hubURL]; !ok {
			fed.DeleteHub(hubURL)
		}
	}

	return nil
}

// GetResponse has nothing to return
func (sh *SetHubs) GetResponse() ActionResponseInterface {
	return nil
}

// HubCreationResult contains the client for a hub
type HubCreationResult struct {
	Hub *hub.Client
}

// Execute will add the hub to the federator's list of known hubs
func (hcr *HubCreationResult) Execute(fed FederatorInterface) error {
	fed.AddHub(hcr.Hub.Host(), hcr.Hub)
	return nil
}

// GetResponse has nothing to return
func (hcr *HubCreationResult) GetResponse() ActionResponseInterface {
	return nil
}
