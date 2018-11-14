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

package webapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/blackducksoftware/armada/pkg/api"
	"github.com/blackducksoftware/armada/pkg/util"
	"github.com/blackducksoftware/armada/pkg/webapi/responders"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/gorilla/mux"

	log "github.com/sirupsen/logrus"
)

// SetupHTTPServer .....
func SetupHTTPServer(responder responders.ResponderInterface, router *mux.Router) {
	http.Handle("/metrics", prometheus.Handler())

	// state of the program
	router.HandleFunc("/model", responder.ModelHandler)

	router.HandleFunc("/sethubs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PUT" {
			log.Debugf("http request: PUT sethubs")
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				responder.Error(w, r, err, 400)
				return
			}
			var setHubs api.HubList
			err = json.Unmarshal(body, &setHubs)
			if err != nil {
				responder.Error(w, r, err, 400)
				return
			}
			responder.SetHubs(&setHubs)
		} else {
			responder.NotFound(w, r)
		}
	})

	/*
		http.HandleFunc("/findproject", func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" {
				log.Debugf("http request: POST findproject")
				body, err := ioutil.ReadAll(r.Body)
				if err != nil {
					responder.Error(w, r, err, 400)
					return
				}
				var request APIProjectSearchRequest
				err = json.Unmarshal(body, &request)
				if err != nil {
					responder.Error(w, r, err, 400)
					return
				}
				projects := responder.FindProject(request)
				jsonBytes, err := json.MarshalIndent(projects, "", "  ")
				if err != nil {
					responder.Error(w, r, err, 500)
				} else {
					header := w.Header()
					header.Set(http.CanonicalHeaderKey("content-type"), "application/json")
					fmt.Fprint(w, string(jsonBytes))
				}
			} else {
				responder.NotFound(w, r)
			}
		})
	*/

	router.HandleFunc("/stackdump", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			log.Debugf("http request: GET stackdump")
			runtimeStack := util.DumpRuntimeStack()
			pprofStack, grsCount := util.DumpPProfStack()
			heap, heapCount := util.DumpHeap()
			dict := map[string]interface{}{
				"runtime":    runtimeStack,
				"pprof":      pprofStack,
				"pprofCount": grsCount,
				"heap":       heap,
				"heapCount":  heapCount,
			}
			fmt.Printf("runtime:\n%s\n\n", runtimeStack)
			fmt.Printf("pprof: %d\n%s\n\n", grsCount, pprofStack)
			fmt.Printf("heap: %d\n%s\n\n", heapCount, heap)
			//			log.Printf()
			//			log.Debugf("runtime: %s", )
			jsonBytes, err := json.MarshalIndent(dict, "", "  ")
			if err != nil {
				responder.Error(w, r, err, 500)
			} else {
				header := w.Header()
				header.Set(http.CanonicalHeaderKey("content-type"), "application/json")
				fmt.Fprint(w, string(jsonBytes))
			}
		} else {
			responder.NotFound(w, r)
		}
	})

	/*
		router.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" {
				log.Debugf("http request: POST config")
				body, err := ioutil.ReadAll(r.Body)
				if err != nil {
					responder.Error(w, r, err, 400)
					return
				}
				var request APIUpdateConfigRequest
				err = json.Unmarshal(body, &request)
				if err != nil {
					responder.Error(w, r, err, 400)
					return
				}
				responder.UpdateConfig(&request)
				fmt.Fprint(w, "")
			} else {
				responder.NotFound(w, r)
			}
		})
	*/

	router.HandleFunc("/api/components", responder.ComponentsHandler)
	router.HandleFunc("/api/components/{componentId}", responder.ComponentsHandler)
	router.HandleFunc("/api/all-components/{componentId}", responder.AllComponentsHandler)
	router.HandleFunc("/api/projects", responder.ProjectsHandler)
	router.HandleFunc("/api/projects/{projectId}", responder.ProjectsHandler)
	router.HandleFunc("/api/all-projects/{projectId}", responder.AllProjectsHandler)
	router.HandleFunc("/api/codelocations", responder.CodeLocationsHandler)
	router.HandleFunc("/api/codelocations/{codeLocationId}", responder.CodeLocationsHandler)
	router.HandleFunc("/api/all-codelocations/{codeLocationId}", responder.AllCodeLocationsHandler)
	router.HandleFunc("/api/policy-rules", responder.PolicyRulesHandler)
	router.HandleFunc("/api/policy-rules/{policyRuleId}", responder.PolicyRulesHandler)
	router.HandleFunc("/api/all-policy-rules/{policyRuleId}", responder.AllPolicyRulesHandler)
	router.HandleFunc("/api/users", responder.UsersHandler)
	router.HandleFunc("/api/users/{userId}", responder.UsersHandler)
	router.HandleFunc("/api/all-users/{userId}", responder.AllUsersHandler)
	router.Path("/api/lasterror").Queries("endpoint", "{endpoint}", "method", "{method}").HandlerFunc(responder.LastErrorHandler)
}
