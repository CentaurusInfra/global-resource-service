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
	"global-resource-service/resource-management/pkg/clientSdk/rmsclient"
	"global-resource-service/resource-management/pkg/common-lib/types"
	"global-resource-service/resource-management/pkg/common-lib/types/location"
	"global-resource-service/resource-management/pkg/store/redis"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/klog/v2"
)

func TestSingleNodeQuery(t *testing.T) {

	cfg := rmsclient.Config{}
	cfg.ServiceUrl = "localhost:8080"
	cfg.RequestTimeout = 3 * time.Minute
	client := rmsclient.NewRmsClient(cfg)
	// get one nodes from redis for single node model
	redisIp := "localhost"
	store := redis.NewRedisClient(redisIp, "7379", false)
	requiredNum := 1
	startTime := time.Now().UTC()
	klog.Infof("Requesting nodes from redis server")
	logicalNodes := store.BatchLogicalNodesInquiry(requiredNum)
	endTime := time.Since(startTime)
	klog.Infof("Total %v nodes required from redis server: %v, Total nodes got from redis: %v in duration: %v, detailes: %v\n", requiredNum, redisIp, len(logicalNodes), endTime, logicalNodes)

	start := time.Now().UTC()

	nodeId := logicalNodes[0].Id
	regionName := location.Region(logicalNodes[0].GeoInfo.Region).String()
	rpName := location.ResourcePartition(logicalNodes[0].GeoInfo.ResourcePartition).String()

	respNode, err := client.Query(nodeId, regionName, rpName)

	duration := time.Since(start)
	if err != nil {
		klog.Errorf("Failed to query node status for node ID: %s. error %v", nodeId, err)
	}
	klog.Infof("Request node (nodeId: %s, regionName: %s, rpName: %s), get node (nodeId: %s, regionName: %s, rpName: %s) in duration: %v", nodeId, regionName, rpName, respNode.Id, location.Region(respNode.GeoInfo.Region).String(), location.ResourcePartition(respNode.GeoInfo.ResourcePartition).String(), duration)
	assert.Equal(t, nodeId, respNode.Id)
	assert.Equal(t, regionName, location.Region(respNode.GeoInfo.Region).String())
	assert.Equal(t, rpName, location.ResourcePartition(respNode.GeoInfo.ResourcePartition).String())
	assert.Less(t, duration, time.Second)
}

func TestBatchNodeQuery(t *testing.T) {

	cfg := rmsclient.Config{}
	cfg.ServiceUrl = "localhost:8080"
	cfg.RequestTimeout = 3 * time.Minute
	client := rmsclient.NewRmsClient(cfg)
	// get one nodes from redis for single node model
	redisIp := "localhost"
	store := redis.NewRedisClient(redisIp, "7379", false)
	requiredNum := 100
	startTime := time.Now().UTC()
	klog.Infof("Requesting nodes from redis server")
	logicalNodes := store.BatchLogicalNodesInquiry(requiredNum)
	endTime := time.Since(startTime)
	klog.Infof("Total %v nodes required from redis server: %v, Total nodes got from redis: %v in duration: %v, detailes: %v\n", requiredNum, redisIp, len(logicalNodes), endTime, logicalNodes)

	var wgNode sync.WaitGroup

	start := time.Now().UTC()

	for i := 0; i < len(logicalNodes); i++ {
		wgNode.Add(1)
		respNode, nodeErr := queryNodeStatus(&wgNode, client, logicalNodes[i])
		if nodeErr != nil {
			klog.Errorf("Failed to query node status for node ID: %s. error %v", logicalNodes[i].Id, nodeErr)
		}
		assert.Equal(t, logicalNodes[i].Id, respNode.Id)
		assert.Equal(t, location.Region(logicalNodes[i].GeoInfo.Region).String(), location.Region(respNode.GeoInfo.Region).String())
		assert.Equal(t, location.ResourcePartition(logicalNodes[i].GeoInfo.ResourcePartition).String(), location.ResourcePartition(respNode.GeoInfo.ResourcePartition).String())
	}
	wgNode.Wait()

	duration := time.Since(start)
	assert.Less(t, duration, time.Second)
}

func queryNodeStatus(wg *sync.WaitGroup, client rmsclient.RmsInterface, node *types.LogicalNode) (respNode *types.LogicalNode, err error) {
	defer wg.Done()
	nodeId := node.Id
	regionName := location.Region(node.GeoInfo.Region).String()
	rpName := location.ResourcePartition(node.GeoInfo.ResourcePartition).String()

	respNode, err = client.Query(nodeId, regionName, rpName)

	return respNode, err
}
