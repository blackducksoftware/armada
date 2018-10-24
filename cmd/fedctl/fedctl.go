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

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/blackducksoftware/armada/pkg/api"
)

func main() {
	url := "http://localhost:10000/sethubs"
	setHubs := api.HubList{
		Items: []api.Hub{
			{
				Host: "hub-hub.10.1.176.76.xip.io",
				Port: 443,
				Config: &api.HubConfig{
					User:     "sysadmin",
					Password: "blackduck",
				},
			},
			{
				Host: "hub2-hub2.10.1.176.76.xip.io",
				Port: 443,
				Config: &api.HubConfig{
					User:     "sysadmin",
					Password: "blackduck",
				},
			},
		},
	}

	jsonBytes, err := json.MarshalIndent(setHubs, "", "  ")
	if err != nil {
		fmt.Printf("error encoding set hub request: %+v", err)
		os.Exit(1)
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		panic(err)
	}
	req.Header.Set(http.CanonicalHeaderKey("content-type"), "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	fmt.Println("response Status:", resp.Status)
}
