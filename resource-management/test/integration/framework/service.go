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

package framework

import (
	"os"
	"strings"
	"sync"
	"time"

	"k8s.io/klog/v2"

	apiapp "global-resource-service/resource-management/cmds/service-api/app"
	localMetrics "global-resource-service/resource-management/pkg/common-lib/metrics"
	"global-resource-service/resource-management/pkg/store/redis"
	simapp "global-resource-service/resource-management/test/resourceRegionMgrSimulator/app"
	"global-resource-service/resource-management/test/resourceRegionMgrSimulator/data"
)

// ServiceMain to start gcs service and simulator service for testing.
func ServiceMain(tests func() int) {

	//flush redis store to ensure all testing started with clean store
	redis.NewRedisClient("localhost", "7379", true)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		masterIp := "localhost"
		masterPort := "8080"
		urls := "localhost:9119"
		redisPort := "7379"
		enableMetrics := false
		grs_err := Start_ServiceAPI(masterIp, masterPort, urls, redisPort, enableMetrics)
		if grs_err != nil {
			klog.Infof("Starting resource management service failed with error: %v", grs_err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		regionName := "Beijing"
		rpNum := 10
		nodesPerRP := 2600
		masterPort := "9119"
		dataPattern := "Daily"
		waittime := 0
		rpDownnum := 0
		sim_err := Start_Simulator(regionName, rpNum, nodesPerRP, masterPort, dataPattern, waittime, rpDownnum)
		if sim_err != nil {
			klog.Infof("Starting simulator rest service failed with error: %v", sim_err)
		}
	}()

	//sleep 30s to get redis write to store
	time.Sleep(30 * time.Second)

	result := tests()
	//stop() // Don't defer this. See os.Exit documentation.
	os.Exit(result)
}

func Start_ServiceAPI(masterIp string, masterPort string, urls string, redisPort string, enableMetrics bool) error {
	c := &apiapp.Config{}
	c.MasterIp = masterIp
	c.MasterPort = masterPort
	c.ResourceUrls = strings.Split(urls, ",")
	c.RedisPort = redisPort
	localMetrics.SetEnableResourceManagementMeasurement(enableMetrics)
	klog.Infof("Starting resource management service")
	if err := apiapp.Run(c); err != nil {
		return err
	}
	return nil
}

func Start_Simulator(regionName string, rpNum int, nodesPerRP int, masterPort string, dataPattern string, waittime int, rpDownNum int) error {

	// Get the simulator arguments
	c := &simapp.RegionConfig{}

	c.RegionName = regionName
	c.RpNum = rpNum
	c.NodesPerRP = nodesPerRP
	c.MasterPort = masterPort
	c.DataPattern = dataPattern
	c.WaitTimeForDataChangePattern = waittime
	c.RPDownNumber = rpDownNum

	data.Init(c.RegionName, c.RpNum, c.NodesPerRP)

	// Generate update changes of Default Pattern
	// - simulate random RP down
	// OR
	// Generate update changes of Daily Pattern
	// - simulate 10 changes each minute
	data.MakeDataUpdate(c.DataPattern, c.WaitTimeForDataChangePattern, c.RPDownNumber)

	// Run simulater RSET API server
	klog.Infof("Starting simulater RSET API service")
	if err := simapp.Run(c); err != nil {
		return err
	}

	return nil
}
