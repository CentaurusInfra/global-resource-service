/*
Copyright The Kubernetes Authors.
Copyright 2022 Authors of Global Resource Service - file modified.
Copyright 2020 Authors of Arktos - file modified.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by client-gen. DO NOT EDIT.

package rmsclient

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"k8s.io/klog/v2"

	"global-resource-service/resource-management/pkg/clientSdk/rest"
	"global-resource-service/resource-management/pkg/clientSdk/watch"
	"global-resource-service/resource-management/pkg/common-lib/types"
	apiTypes "global-resource-service/resource-management/pkg/service-api/types"
)

// client config that the client can setup for
type Config struct {
	ServiceUrl                  string
	RequestTimeout              time.Duration
	ClientFriendlyName          string
	ClientRegion                string
	RegionIdToWatch             string
	InitialRequestTotalMachines int
	InitialRequestRegions []string
	CaptureDetailedLog    bool
}

// ListOptions contains optional settings for List nodes
type ListOptions struct {
	// Limit is equailent to URL query parameter ?limit=500
	Limit int
}

// RmsInterface has methods to work with Resource management service resources.
// below are just 630 related interface definitions
type RmsInterface interface {
	Register() (*apiTypes.ClientRegistrationResponse, error)
	List(string, ListOptions) ([]*types.LogicalNode, types.TransitResourceVersionMap, error)
	Watch(string, types.TransitResourceVersionMap) (watch.Interface, error)
	Query(string, string, string) (*types.LogicalNode, error)
}

// rmsClient implements RmsInterface
type rmsClient struct {
	config Config
	// REST client to RMS service
	restClient rest.Interface
	// ClientId to be set by the Register to the RMS service
	Id string
}

// NewRmsClient returns a refence to the rsmClient object
func NewRmsClient(cfg Config) *rmsClient {
	httpclient := http.Client{Timeout: cfg.RequestTimeout}
	url, err := rest.DefaultServerURL(cfg.ServiceUrl, "", false)

	if err != nil {
		klog.Errorf("failed to get the default URL. error %v", err)
		return nil
	}

	c, err := rest.NewRESTClient(url, rest.ClientContentConfig{}, nil, &httpclient)
	if err != nil {
		klog.Errorf("failed to get the RESTClient. error %v", err)
		return nil
	}

	return &rmsClient{
		config:     cfg,
		restClient: c,
		Id:         "",
	}
}

func (c *rmsClient) Register() (*apiTypes.ClientRegistrationResponse, error) {
	req := c.restClient.Post()

	cq := apiTypes.ClientRegistrationRequest{
		ClientInfo: types.ClientInfoType{
			ClientName: c.config.ClientFriendlyName,
			Region:     c.config.ClientRegion},
		InitialRequestedResource: types.ResourceRequest{
			TotalMachines:    c.config.InitialRequestTotalMachines,
			RequestInRegions: nil},
	}

	body, err := json.Marshal(cq)
	if err != nil {
		return nil, err
	}

	// construct the request oject
	req = req.Body(body)
	req = req.Resource("clients")
	req = req.Timeout(c.config.RequestTimeout)

	resp, err := req.DoRaw()

	if err != nil {
		return nil, err
	}

	ret := apiTypes.ClientRegistrationResponse{}

	err = json.Unmarshal(resp, &ret)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

// List takes label and field selectors, and returns the list of Nodes that match those selectors.
func (c *rmsClient) List(clientId string, opts ListOptions) ([]*types.LogicalNode, types.TransitResourceVersionMap, error) {
	req := c.restClient.Get()
	req = req.Resource("resource")
	req = req.Name(c.Id)
	req = req.Timeout(c.config.RequestTimeout)
	req = req.Param("limit", strconv.Itoa(opts.Limit))

	respRet, err := req.DoRaw()
	if err != nil {
		return nil, nil, err
	}

	resp := apiTypes.ListNodeResponse{}

	err = json.Unmarshal(respRet, &resp)

	if err != nil {
		return nil, nil, err
	}

	actualCrv := resp.ResourceVersions

	return resp.NodeList, actualCrv, nil

}

// Watch returns a watch.Interface that watches the requested rmsClient.
func (c *rmsClient) Watch(clientId string, versionMap types.TransitResourceVersionMap) (watch.Interface, error) {
	req := c.restClient.Post()
	req = req.Resource("resource")
	req = req.Name(c.Id)
	req = req.Timeout(c.config.RequestTimeout)
	req = req.Param("watch", "true")

	crv := apiTypes.WatchRequest{ResourceVersions: versionMap}

	body, err := json.Marshal(crv)
	if err != nil {
		return nil, err
	}
	req = req.Body(body)

	watcher, err := req.Watch()
	if err != nil {
		return nil, err
	}

	return watcher, nil
}

// Query Nodes, and returns Nodes that match those selectors.
func (c *rmsClient) Query(nodeId string, regionName string, rpName string) (*types.LogicalNode, error) {
	req := c.restClient.Get()
	resourcePath := "nodes"
	req = req.Resource(resourcePath)
	req = req.Timeout(c.config.RequestTimeout)
	req = req.Param("nodeId", nodeId)
	req = req.Param("region", regionName)
	req = req.Param("resourcePartition", rpName)

	respRet, err := req.DoRaw()
	if err != nil {
		return nil, err
	}

	resp := apiTypes.NodeResponse{}

	err = json.Unmarshal(respRet, &resp)

	if err != nil {
		return nil, err
	}

	respNode := resp.Node

	return &respNode, nil

}
