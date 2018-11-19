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

package federation

import (
	"fmt"
	"os"
	"time"

	"github.com/blackducksoftware/armada/pkg/actions"
	"github.com/blackducksoftware/armada/pkg/api"
	"github.com/blackducksoftware/armada/pkg/hub"

	"github.com/blackducksoftware/hub-client-go/hubclient"

	"github.com/imdario/mergo"

	log "github.com/sirupsen/logrus"
)

// HubClientCreator handles creating new hub clients
type HubClientCreator struct {
	hubDefaultConfig     *api.HubConfig
	didFinishHubCreation chan actions.ActionInterface
}

// NewHubClientCreator creates a new HubClientCreator object
func NewHubClientCreator(envVar string, defaults *api.HubConfig) (*HubClientCreator, error) {
	if len(defaults.Password) == 0 {
		hubPassword, ok := os.LookupEnv(envVar)
		if !ok {
			return nil, fmt.Errorf("unable to get default password : environment variable %s not set", envVar)
		}
		defaults.Password = hubPassword
	}

	return &HubClientCreator{
		hubDefaultConfig:     defaults,
		didFinishHubCreation: make(chan actions.ActionInterface)}, nil
}

// CreateClients will create hub clients for the provided list of hubs
func (hc *HubClientCreator) CreateClients(hubs *api.HubList) {
	log.Infof("hubs to add: %+v", hubs)
	for _, newHub := range hubs.Items {
		log.Infof("adding hub: %+v", newHub)
		host := newHub.Host
		port := newHub.Port
		hubConfig, err := hc.generateHubConfig(newHub.Config)
		if err != nil {
			log.Errorf("unable to generate config for hub %s: %v", host, err)
			continue
		}

		user := hubConfig.User
		timeout := time.Duration(*hubConfig.ClientTimeoutMilliseconds) * time.Millisecond

		hostWithPort := fmt.Sprintf("%s:%d", host, port)
		baseURL := fmt.Sprintf("https://%s", hostWithPort)
		rawClient, err := hubclient.NewWithSession(baseURL, hubclient.HubClientDebugTimings, timeout)
		if err != nil {
			log.Errorf("unable to create client for hub %s: %v", host, err)
			continue
		}

		client := hub.NewClient(user, hubConfig.Password, hostWithPort, rawClient, 0, 0, 0)
		go func() {
			hc.didFinishHubCreation <- &actions.HubCreationResult{Hub: client}
		}()
	}
}

func (hc *HubClientCreator) generateHubConfig(hubConfig *api.HubConfig) (*api.HubConfig, error) {
	var config api.HubConfig

	if hubConfig != nil {
		config = *hubConfig
	} else {
		config = api.HubConfig{}
	}

	if err := mergo.Merge(&config, *hc.hubDefaultConfig); err != nil {
		return nil, err
	}

	log.Infof("hub generated config: %+v", config)

	return &config, nil
}
