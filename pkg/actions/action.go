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
	"github.com/blackducksoftware/armada/pkg/api"
	"github.com/blackducksoftware/armada/pkg/hub"
)

// ActionInterface defines an interface for actions of the federator
type ActionInterface interface {
	// Apply the action to the provided federator
	Execute(FederatorInterface) error
	GetResponse() ActionResponseInterface
}

// ActionResponseInterface defines the interface for action responses
type ActionResponseInterface interface {
	ReplaceSource(string)
	GetResult() interface{}
}

// FederatorInterface defines the interface for federator objects
type FederatorInterface interface {
	DeleteHub(string)
	CreateHubClients(*api.HubList)
	AddHub(string, *hub.Client)
	GetHubs() map[string]*hub.Client
}

// EmptyResponse is a generic response to a request which
// provides no response that satisfies the ActionResponseInterface
type EmptyResponse struct{}

// ReplaceSource does nothing for the generic response
func (cr *EmptyResponse) ReplaceSource(string) {}

// GetResult returns nothing for the generic response
func (cr *EmptyResponse) GetResult() interface{} {
	return nil
}
