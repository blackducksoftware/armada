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

package api

import (
	"time"
)

// ModelHub describes a hub client model
type ModelHub struct {
	// can we log in to the hub?
	//      IsLoggedIn bool
	// have all the projects been sucked in?
	HasLoadedAllCodeLocations bool
	// map of project name to ... ? hub URL?
	//      Projects map[string]string
	// map of code location name to mapped project version url
	CodeLocations  map[string]*ModelCodeLocation
	Errors         []string
	Status         string
	CircuitBreaker *ModelCircuitBreaker
	Host           string
}

// ModelCodeLocation ...
type ModelCodeLocation struct {
	Stage                string
	Href                 string
	URL                  string
	MappedProjectVersion string
	UpdatedAt            string
	ComponentsHref       string
}

// ModelCircuitBreaker ...
type ModelCircuitBreaker struct {
	State               string
	NextCheckTime       *time.Time
	MaxBackoffDuration  ModelTime
	ConsecutiveFailures int
}

// ModelTime ...
type ModelTime struct {
	duration     time.Duration
	Minutes      float64
	Seconds      float64
	Milliseconds float64
}

// NewModelTime consumes a time.Duration and calculates the minutes, seconds,
// and milliseconds
func NewModelTime(duration time.Duration) *ModelTime {
	return &ModelTime{
		duration:     duration,
		Minutes:      float64(duration) / float64(time.Minute),
		Seconds:      float64(duration) / float64(time.Second),
		Milliseconds: float64(duration) / float64(time.Millisecond),
	}
}

// Hub contains the information relevant to idenfity and connect
// to a hub
type Hub struct {
	Host   string     `json:"host,omitempty"`
	Port   int        `json:"port,omitempty"`
	Config *HubConfig `json:"hubConfig,omitempty"`
}

// HubConfig defines the config used when contacting a hub
type HubConfig struct {
	User                      string `json:"user,omitempty"`
	Password                  string `json:"password,omitempty"`
	ClientTimeoutMilliseconds *int   `json:"clientTimeoutMilliseconds,omitempty"`
}

// HubList is a list of Hub items
type HubList struct {
	Items []Hub `json:"items,omitempty"`
}

// EndpointType defines and endpoint
type EndpointType string

const (
	ComponentsEndpoint       EndpointType = "components"
	AllComponentsEndpoint    EndpointType = "all-components"
	ProjectsEndpoint         EndpointType = "projects"
	AllProjectsEndpoint      EndpointType = "all-projects"
	CodeLocationsEndpoint    EndpointType = "codelocations"
	AllCodeLocationsEndpoint EndpointType = "all-codelocations"
	PolicyRulesEndpoint      EndpointType = "policy-rules"
	AllPolicyRulesEndpoint   EndpointType = "all-policy-rules"
	UsersEndpoint            EndpointType = "users"
	AllUsersEndpoint         EndpointType = "all-users"
	LastErrorEndpoint        EndpointType = "lasterror"
)

// HubFuncsType defines various functions for interacting with a hub
type HubFuncsType struct {
	Create       string
	Get          string
	GetAll       string
	Delete       string
	SingleToList func(interface{}) interface{}
}
