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
	"net/http"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/blackducksoftware/armada/pkg/actions"
	"github.com/blackducksoftware/armada/pkg/api"
	"github.com/blackducksoftware/armada/pkg/hub"
	"github.com/blackducksoftware/armada/pkg/util"
	"github.com/blackducksoftware/armada/pkg/webapi"
	"github.com/blackducksoftware/armada/pkg/webapi/responders"
	httpresponder "github.com/blackducksoftware/armada/pkg/webapi/responders/http"
	mockresponder "github.com/blackducksoftware/armada/pkg/webapi/responders/mock"

	"github.com/blackducksoftware/hub-client-go/hubapi"
	"github.com/blackducksoftware/hub-client-go/hubclient"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/gorilla/mux"

	log "github.com/sirupsen/logrus"
)

const (
	actionChannelSize = 100
)

type hubError struct {
	Host string
	Err  *hubclient.HubClientError
}

// Federator handles federating queries across multiple hubs
type Federator struct {
	responder  responders.ResponderInterface
	router     *mux.Router
	hubCreator *HubClientCreator

	// model
	config     *FederatorConfig
	hubs       map[string]*hub.Client
	lastErrors map[api.EndpointType]map[string]*api.LastError

	// channels
	stop    chan struct{}
	actions chan actions.ActionInterface
}

// NewFederator creates a new Federator object
func NewFederator(configPath string) (*Federator, error) {
	var responder responders.ResponderInterface

	config, err := GetConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config file %s: %v", configPath, err)
	}
	if config == nil {
		return nil, fmt.Errorf("expected non-nil config from path %s, but got nil", configPath)
	}

	level, err := config.GetLogLevel()
	if err != nil {
		return nil, fmt.Errorf("error setting log level: %v", err)
	}
	log.SetLevel(level)

	ip, err := util.GetOutboundIP()
	if err != nil {
		return nil, fmt.Errorf("failed to find local ip address: %v", err)
	}

	router := mux.NewRouter().StrictSlash(true)
	if config.UseMockMode {
		responder = mockresponder.NewResponder()
	} else {
		responder = httpresponder.NewResponder(ip)
	}
	webapi.SetupHTTPServer(responder, router)

	hubCreator, err := NewHubClientCreator(config.HubPasswordEnvVar, config.HubDefaults)
	if err != nil {
		return nil, fmt.Errorf("failed to recreate HubClientCreator: %v", err)
	}

	prometheus.Unregister(prometheus.NewProcessCollector(os.Getpid(), ""))
	prometheus.Unregister(prometheus.NewGoCollector())

	fed := &Federator{
		responder:  responder,
		router:     router,
		hubCreator: hubCreator,
		config:     config,
		hubs:       map[string]*hub.Client{},
		stop:       make(chan struct{}),
		actions:    make(chan actions.ActionInterface, actionChannelSize),
		lastErrors: map[api.EndpointType]map[string]*api.LastError{},
	}

	return fed, nil
}

// Run will start the federator listening for requests
func (fed *Federator) Run(stopCh chan struct{}) {

	// dump events into 'actions' queue
	go func() {
		for {
			select {
			case a := <-fed.responder.GetRequestCh():
				log.Debugf("received action: %+v", a)
				fed.actions <- a
			case d := <-fed.hubCreator.didFinishHubCreation:
				fed.actions <- d
			}
		}
	}()

	// process actions
	go func() {
		for {
			a := <-fed.actions
			log.Debugf("processing action %s", reflect.TypeOf(a))
			start := time.Now()
			a.Execute(fed)
			stop := time.Now()
			log.Debugf("finished processing action -- %s", stop.Sub(start))
		}
	}()

	log.Infof("starting HTTP server on port %d", fed.config.Port)
	go func() {
		addr := fmt.Sprintf(":%d", fed.config.Port)
		http.ListenAndServe(addr, fed.router)
	}()
	<-stopCh
}

// CreateHubClients will create hub clients for the provided hubs
func (fed *Federator) CreateHubClients(hubList *api.HubList) {
	fed.hubCreator.CreateClients(hubList)
}

// AddHub will add a hub to the list of know hubs
func (fed *Federator) AddHub(url string, client *hub.Client) {
	if _, ok := fed.hubs[url]; ok {
		log.Warningf("cannot add hub %s: already present", url)
	}
	fed.hubs[url] = client
}

// DeleteHub will remove a hub from the list of know hubs
func (fed *Federator) DeleteHub(url string) {
	client, ok := fed.hubs[url]
	if !ok {
		log.Warningf("received request to delete hub %s, but it does not exist")
	} else {
		client.Stop()
		delete(fed.hubs, url)
	}
}

// GetHubs returns the hubs known to the federator
func (fed *Federator) GetHubs() map[string]*hub.Client {
	return fed.hubs
}

func (fed *Federator) setLastError(method string, endPoint api.EndpointType, lastError *api.LastError) {
	if _, ok := fed.lastErrors[endPoint]; !ok {
		fed.lastErrors[endPoint] = make(map[string]*api.LastError)
	}
	fed.lastErrors[endPoint][method] = lastError
}

// GetLastError will retrieve the last error information for a provided endpoint
func (fed *Federator) GetLastError(method string, endPoint api.EndpointType) *api.LastError {
	return fed.lastErrors[endPoint][strings.ToUpper(method)]
}

// SendGetRequest will retrieve information from the hubs
func (fed *Federator) SendGetRequest(endpoint api.EndpointType, funcs api.GetFuncsType, objID string, result interface{}) {
	var wg sync.WaitGroup
	var errs api.LastError

	hubCount := len(fed.hubs)
	resultCh := make(chan interface{}, hubCount)
	errCh := make(chan hubError, hubCount)
	errs.Errors = make(map[string]*hubclient.HubClientError)

	wg.Add(hubCount)
	for hubURL, client := range fed.hubs {
		go func(client *hub.Client, url string, id string, ep api.EndpointType, funcs api.GetFuncsType) {
			defer wg.Done()
			strEp := actions.ConvertAllEndpoint(ep)
			if len(id) > 0 {
				link := hubapi.ResourceLink{Href: fmt.Sprintf("https://%s/api/%s/%s", url, strEp, id)}
				log.Debugf("querying %s %s", strEp, link.Href)
				hubFunc := fed.getHubClientFunc(funcs.Get, client)
				callResp := hubFunc.Call([]reflect.Value{reflect.ValueOf(link)})
				resp := callResp[0].Interface()
				err := callResp[1].Interface()
				log.Debugf("response to %s query from %s: %+v", strEp, link.Href, resp)
				if err != nil {
					hubErr := err.(*hubclient.HubClientError)
					errCh <- hubError{Host: url, Err: hubErr}
				} else {
					list := funcs.SingleToList(resp)
					resultCh <- &list
				}
			} else {
				log.Debugf("querying all %s", strEp)
				hubFunc := fed.getHubClientFunc(funcs.GetAll, client)
				callResp := hubFunc.Call([]reflect.Value{})
				resp := callResp[0].Interface()
				err := callResp[1].Interface()
				if err != nil {
					log.Warningf("failed to get %s from %s: %v", strEp, url, err)
					hubErr := err.(*hubclient.HubClientError)
					errCh <- hubError{Host: url, Err: hubErr}
				} else {
					resultCh <- &resp
				}
			}
		}(client, hubURL, objID, endpoint, funcs)
	}

	wg.Wait()
	for i := 0; i < hubCount; i++ {
		select {
		case response := <-resultCh:
			if response != nil {
				value := reflect.ValueOf(response).Elem().Interface()
				log.Debugf("a hub responsed to a get request to endpoint %s with: %+v", actions.ConvertAllEndpoint(endpoint), value)
				fed.mergeHubList(result, value)
			}
		case err := <-errCh:
			errs.Errors[err.Host] = err.Err
		}
	}

	fed.setLastError(http.MethodGet, endpoint, &errs)
}

func (fed *Federator) getHubClientFunc(name string, hubClient *hub.Client) reflect.Value {
	return reflect.ValueOf(hubClient).MethodByName(name)
}

func (fed *Federator) mergeHubList(orig interface{}, new interface{}) {
	log.Debugf("mergeHubList orig: %+v", orig)
	log.Debugf("mergeHubList new: %+v", new)

	// Merge TotalCount first
	origCount := fed.getHubListField(orig, "TotalCount")
	newCount := fed.getHubListField(new, "TotalCount")
	origCount.SetUint(origCount.Uint() + newCount.Uint())

	// Merge Items
	origItems := fed.getHubListField(orig, "Items")
	newItems := fed.getHubListField(new, "Items")
	origItems.Set(reflect.AppendSlice(origItems, newItems))

	// Merge Meta if it exists
	origMeta := fed.getHubListField(orig, "Meta")
	if origMeta.IsValid() {
		newMeta := fed.getHubListField(new, "Meta")
		origAllow := origMeta.FieldByName("Allow")
		newAllow := newMeta.FieldByName("Allow")
		origMeta.FieldByName("Allow").Set(reflect.AppendSlice(origAllow, newAllow))

		origLinks := origMeta.FieldByName("Links")
		newLinks := newMeta.FieldByName("Links")
		origMeta.FieldByName("Links").Set(reflect.AppendSlice(origLinks, newLinks))
	}
}

func (fed *Federator) getHubListField(list interface{}, field string) reflect.Value {
	return reflect.ValueOf(list).Elem().FieldByName(field)
}

// SendCreateRequest create information in the hubs
func (fed *Federator) SendCreateRequest(endpoint api.EndpointType, createFunc string, request interface{}) {
	var wg sync.WaitGroup
	var errs api.LastError

	hubCount := len(fed.hubs)
	resultCh := make(chan interface{}, hubCount)
	errCh := make(chan hubError, hubCount)
	errs.Errors = make(map[string]*hubclient.HubClientError)

	wg.Add(hubCount)
	for hubURL, client := range fed.hubs {
		go func(client *hub.Client, url string, ep api.EndpointType, createF string, req interface{}) {
			defer wg.Done()
			strEp := actions.ConvertAllEndpoint(ep)
			log.Debugf("creating %s %+v", strEp, req)
			hubFunc := fed.getHubClientFunc(createF, client)
			callResp := hubFunc.Call([]reflect.Value{reflect.ValueOf(req)})
			resp := callResp[0].Interface()
			err := callResp[1].Interface()
			if err != nil {
				log.Warningf("failed to create %s in %s: %v", strEp, url, err)
				hubErr := err.(*hubclient.HubClientError)
				errCh <- hubError{Host: url, Err: hubErr}
			} else {
				resultCh <- &resp
			}
		}(client, hubURL, endpoint, createFunc, request)
	}

	wg.Wait()
	for i := 0; i < hubCount; i++ {
		select {
		case response := <-resultCh:
			if response != nil {
				value := reflect.ValueOf(response).Elem().Interface()
				log.Debugf("a hub responded to create request to endpoint %s with: %+v", string(endpoint), value)
			}
		case err := <-errCh:
			errs.Errors[err.Host] = err.Err
		}
	}

	fed.setLastError(http.MethodPost, endpoint, &errs)
}

// SendDeleteRequest will delete information from the hubs
func (fed *Federator) SendDeleteRequest(endpoint api.EndpointType, deleteFunc string, id string) {
	var wg sync.WaitGroup
	var errs api.LastError

	hubCount := len(fed.hubs)
	errCh := make(chan hubError, hubCount)
	errs.Errors = make(map[string]*hubclient.HubClientError)

	wg.Add(hubCount)
	for hubURL, client := range fed.hubs {
		go func(client *hub.Client, url string, id string, ep api.EndpointType, deleteF string) {
			strEp := actions.ConvertAllEndpoint(ep)
			defer wg.Done()
			log.Debugf("deleting %s %s from hub %s", strEp, id, url)
			loc := fmt.Sprintf("https://%s/api/%s/%s", url, strEp, id)
			hubFunc := fed.getHubClientFunc(deleteF, client)
			callResp := hubFunc.Call([]reflect.Value{reflect.ValueOf(loc)})
			err := callResp[0].Interface()
			if err != nil {
				log.Warningf("failed to delete %s %s in %s: %v", strEp, id, url, err)
				hubErr := err.(*hubclient.HubClientError)
				errCh <- hubError{Host: url, Err: hubErr}
			} else {
				errCh <- hubError{Host: url, Err: nil}
			}
		}(client, hubURL, id, endpoint, deleteFunc)
	}
	wg.Wait()

	for i := 0; i < hubCount; i++ {
		err := <-errCh
		if err.Err != nil {
			errs.Errors[err.Host] = err.Err
		}
	}

	fed.setLastError(http.MethodDelete, endpoint, &errs)
}
