/*
Copyright 2022 Authors of Global Resource Service.

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

package service_api

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"global-resource-service/resource-management/pkg/clientSdk/rmsclient"
	utilruntime "global-resource-service/resource-management/pkg/clientSdk/util/runtime"
	"global-resource-service/resource-management/pkg/common-lib/types/event"

	"github.com/stretchr/testify/assert"
	"k8s.io/klog/v2"
)

func TestRegisterClient(t *testing.T) {
	klog.Infof("List resources from service ...")

	cfg := rmsclient.Config{}
	cfg.ServiceUrl = "localhost:8080"
	cfg.ClientFriendlyName = "testclient"
	cfg.ClientRegion = "Beijing"
	cfg.InitialRequestTotalMachines = 20000
	cfg.RegionIdToWatch = "-1"

	cfg.RequestTimeout = 30 * time.Minute
	client := rmsclient.NewRmsClient(cfg)

	clientId := registerClient(client)

	assert.NotNil(t, clientId, "Expecting not nil client id")
	assert.False(t, clientId == "", "Expecting non empty client id")

}

func TestListNodes(t *testing.T) {
	klog.Infof("List resources from service ...")

	cfg := rmsclient.Config{}
	cfg.ServiceUrl = "localhost:8080"
	cfg.ClientFriendlyName = "testclient"
	cfg.ClientRegion = "Beijing"
	cfg.InitialRequestTotalMachines = 20000
	cfg.RegionIdToWatch = "-1"

	cfg.RequestTimeout = 30 * time.Minute
	client := rmsclient.NewRmsClient(cfg)

	listOpts := rmsclient.ListOptions{}
	listOpts.Limit = 25000

	clientId := registerClient(client)
	client.Id = clientId

	nodeList, crv, err := client.List(clientId, listOpts)
	if err != nil {
		klog.Errorf("failed list resource from service. error %v", err)
	}
	assert.Nil(t, err, "Expecting no error")
	assert.NotNil(t, crv, "Expecting crv is not null")
	assert.LessOrEqual(t, cfg.InitialRequestTotalMachines, len(nodeList))
	assert.Equal(t, 10, len(crv))
}

func TestWatchNodes(t *testing.T) {
	cfg := rmsclient.Config{}
	cfg.ServiceUrl = "localhost:8080"
	cfg.ClientFriendlyName = "testclient"
	cfg.ClientRegion = "Beijing"
	cfg.InitialRequestTotalMachines = 20000
	cfg.RegionIdToWatch = "-1"

	cfg.RequestTimeout = 30 * time.Minute
	client := rmsclient.NewRmsClient(cfg)

	listOpts := rmsclient.ListOptions{}
	listOpts.Limit = 25000

	clientId := registerClient(client)
	client.Id = clientId

	nodeList, crv, err := client.List(clientId, listOpts)
	if err != nil {
		klog.Errorf("failed list resource from service. error %v", err)
	}
	assert.Nil(t, err, "Expecting no error")
	assert.NotNil(t, crv, "Expecting crv is not null")
	assert.LessOrEqual(t, cfg.InitialRequestTotalMachines, len(nodeList))
	assert.Equal(t, 10, len(crv))

	watcher, werr := client.Watch(clientId, crv)

	if werr != nil {
		assert.Fail(t, "Encountered error while building watch connection.", "Encountered error while building watch connection. Error %v", werr)
		return
	}

	watchCh := watcher.ResultChan()
	watchCount := 0
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer utilruntime.HandleCrash()
		// retrieve updates from watcher
		for {
			select {
			case record, ok := <-watchCh:
				if !ok {
					// End of results.
					klog.Infof("End of results")
					return
				}
				switch record.Type {
				case event.Added:
					watchCount++
				case event.Modified:
					watchCount++
				case event.Deleted:
					watchCount++

				}
				if watchCount > 0 {
					newRV, _ := strconv.Atoi(record.Node.ResourceVersion)
					assert.NotNil(t, newRV, "Expecting event watched successfully")
					return
				}
			case <-time.After(7 * time.Minute):
				assert.Fail(t, "Failed to get any watch events within 7 minutes")
				return
			}
		}
	}()
	wg.Wait()

}

func registerClient(client rmsclient.RmsInterface) string {
	klog.Infof("Register client to service  ...")
	registrationResp, err := client.Register()
	if err != nil {
		klog.Errorf("failed register client to service. error %v", err)
	}
	return registrationResp.ClientId
}
