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
	//	"time"

	"github.com/blackducksoftware/armada/pkg/api"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

/*
// HubConfig defines the config used when contacting a hub
type HubConfig struct {
	User                         string
	PasswordEnvVar               string
	ClientTimeoutMilliseconds    int
	Port                         int
	FetchAllProjectsPauseSeconds int
}

// ClientTimeout converts the milliseconds to a duration
func (config *HubConfig) ClientTimeout() time.Duration {
	return time.Duration(config.ClientTimeoutMilliseconds) * time.Millisecond
}

// FetchAllProjectsPause converts the minutes to a duration
func (config *HubConfig) FetchAllProjectsPause() time.Duration {
	return time.Duration(config.FetchAllProjectsPauseSeconds) * time.Second
}
*/

// FederatorConfig defines the configuration for the federator
type FederatorConfig struct {
	HubDefaults       *api.HubConfig `json:"hubDefaults,omitempty"`
	HubPasswordEnvVar string         `json:"hubPasswordEnvVar,omitempty"`
	UseMockMode       bool           `json:"useMockMode,omitempty"`
	Port              int            `json:"port,omitempty"`
	LogLevel          string         `json:"logLevel,omitempty"`
}

// GetLogLevel returns the log level
func (config *FederatorConfig) GetLogLevel() (log.Level, error) {
	return log.ParseLevel(config.LogLevel)
}

// GetConfig returns a configuration object to configure Federator
func GetConfig(configPath string) (*FederatorConfig, error) {
	config := FederatorConfig{}

	viper.SetConfigFile(configPath)

	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	if viper.Get("HubDefaults") != nil {
		config.HubDefaults = &api.HubConfig{}
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	return &config, nil
}
