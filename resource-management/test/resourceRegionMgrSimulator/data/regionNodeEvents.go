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

package data

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/google/uuid"
	"k8s.io/klog/v2"

	"global-resource-service/resource-management/pkg/common-lib/types"
	"global-resource-service/resource-management/pkg/common-lib/types/event"
	"global-resource-service/resource-management/pkg/common-lib/types/location"

	simulatorTypes "global-resource-service/resource-management/test/resourceRegionMgrSimulator/types"
)

// The following varables are used to create Region Node Event List in multiply RPs of one region
//
// RegionNodeEventsList      - for initpull
// SnapshotNodeListEvents    - for subsequent pull and
//                                 post CRV to discard all old node events
var RegionNodeEventsList simulatorTypes.RegionNodeEvents

var snapshotNodeListEvents simulatorTypes.RegionNodeEvents
var RegionId, RpNum, NodesPerRP int

// The constants are for repeatly generate new modified events
// Outage pattern - one RP down

// Daily Patttern - 10 modified changes per minute
const atEachMin10 = 10

// Initialize two events list
// RegionNodeEventsList - for initpull
//
func Init(regionName string, rpNum, nodesPerRP int) {
	RegionNodeEventsList = generateAddedNodeEvents(regionName, rpNum, nodesPerRP)
	RegionId = int(location.GetRegionFromRegionName(regionName))
	RpNum = rpNum
	NodesPerRP = nodesPerRP
}

// Generate region node update event changes to
// add them into RegionNodeEventsList
//
func MakeDataUpdate(data_pattern string, wait_time_for_make_rp_down int) {
	go func(data_pattern string, wait_time_for_make_rp_down int) {
		switch data_pattern {
		case "Outage":
			for {
				// Generate one RP down event during specfied interval
				time.Sleep(time.Duration(wait_time_for_make_rp_down) * time.Minute)
				makeOneRPDown()
				klog.V(3).Info("Generating one RP down event is completed")

				time.Sleep(120 * time.Minute)
				klog.V(6).Info("Simulate to delay 2 hours")
			}
		case "Daily":
			for {
				// At each minute mark, generate 10 modified node events
				time.Sleep(1 * time.Minute)
				makeDataUpdate(atEachMin10)

				klog.V(3).Info("At each minute mark, generating 10 modified and added node events is completed")
			}
		default:
			klog.V(3).Infof("Current Simulator Data Pattern (%v) is supported", data_pattern)
			return
		}
	}(data_pattern, wait_time_for_make_rp_down)
}

///////////////////////////////////////////////
// The following functions are for handler.
//////////////////////////////////////////////

// Return region node added events in BATCH LENGTH from all RPs
// TO DO: paginate support
//
func GetRegionNodeAddedEvents(batchLength uint64) (simulatorTypes.RegionNodeEvents, uint64) {
	klog.V(6).Infof("Total (%v) Added events are to be pulled", RpNum*NodesPerRP)
	return RegionNodeEventsList, uint64(RpNum * NodesPerRP)

}

// Return region node modified events with CRVs in BATCH LENGTH from all RPs
// TO DO: paginate support
//
func GetRegionNodeModifiedEventsCRV(rvs types.TransitResourceVersionMap) (simulatorTypes.RegionNodeEvents, uint64) {
	snapshotNodeListEvents = RegionNodeEventsList
	pulledNodeListEvents := make(simulatorTypes.RegionNodeEvents, RpNum)

	var count uint64 = 0
	for j := 0; j < RpNum; j++ {
		pulledNodeListEventsPerRP := make([]*event.NodeEvent, NodesPerRP)
		indexPerRP := 0
		for i := 0; i < NodesPerRP; i++ {
			region := snapshotNodeListEvents[j][i].Node.GeoInfo.Region
			rp := snapshotNodeListEvents[j][i].Node.GeoInfo.ResourcePartition
			loc := types.RvLocation{Region: location.Region(region), Partition: location.ResourcePartition(rp)}

			nodeRV := snapshotNodeListEvents[j][i].Node.GetResourceVersionInt64()
			requestedRV := rvs[loc]
			if nodeRV > requestedRV {
				count += 1
				pulledNodeListEventsPerRP[indexPerRP] = snapshotNodeListEvents[j][i]
				indexPerRP += 1
			}
		}

		pulledNodeListEvents[j] = pulledNodeListEventsPerRP[:indexPerRP]
	}

	klog.V(9).Infof("Total (%v) Modified events are to be pulled", count)
	return pulledNodeListEvents, count
}

////////////////////////////////////////
// The below are all helper functions
////////////////////////////////////////

// This function is used to initialize the region node added event
//
func generateAddedNodeEvents(regionName string, rpNum, nodesPerRP int) simulatorTypes.RegionNodeEvents {
	regionId := location.GetRegionFromRegionName(regionName)
	eventsAdd := make(simulatorTypes.RegionNodeEvents, rpNum)

	for j := 0; j < rpNum; j++ {
		rpName := location.ResourcePartitions[j]
		loc := location.NewLocation(regionId, rpName)

		// Initialize the resource version starting from 0 for each RP
		var rvToGenerateRPs = 0
		eventsAdd[j] = make([]*event.NodeEvent, nodesPerRP)
		for i := 0; i < nodesPerRP; i++ {
			rvToGenerateRPs += 1

			node := createRandomNode(rvToGenerateRPs, loc)
			nodeEvent := event.NewNodeEvent(node, event.Added)
			eventsAdd[j][i] = nodeEvent
		}

	}
	return eventsAdd
}

//This function simulates one random RP down
func makeOneRPDown() {
	selectedRP := int(rand.Intn(RpNum))
	klog.V(3).Infof("Generating all node down events in selected RP (%v) is starting", selectedRP)

	eventsPerRP := RegionNodeEventsList[selectedRP]

	// Search the nodes in the RP to get the highestRV
	var highestRVForRP uint64 = 0
	length := len(eventsPerRP)
	for k := 0; k < length; k++ {
		currentResourceVersion := eventsPerRP[k].Node.GetResourceVersionInt64()
		if highestRVForRP < currentResourceVersion {
			highestRVForRP = currentResourceVersion
		}
	}

	// Make the modified changes for all nodes of selected RP
	rvToGenerateRPs := highestRVForRP + 1
	for i := 0; i < NodesPerRP; i++ {

		// reset the version of node with the current rvToGenerateRPs
		node := eventsPerRP[i].Node
		node.ResourceVersion = strconv.FormatUint(rvToGenerateRPs, 10)

		// record the time to change resource version in resource partition
		node.LastUpdatedTime = time.Now().UTC()

		newEvent := event.NewNodeEvent(node, event.Modified)
		RegionNodeEventsList[selectedRP][i] = newEvent

		rvToGenerateRPs++
	}
}

// This function is used to create region node modified event by go routine
//
func makeDataUpdate(changesThreshold int) {
	// Calculate how many node changes per RP
	var nodeChangesPerRP = 1
	if changesThreshold >= 2*RpNum {
		nodeChangesPerRP = changesThreshold / RpNum
	}

	// Make data update for each RP
	for j := 0; j < RpNum; j++ {
		eventsPerRP := RegionNodeEventsList[j]

		// Search the nodes in the RP to get the highestRV
		var highestRVForRP uint64 = 0
		length := len(eventsPerRP)
		for k := 0; k < length; k++ {
			currentResourceVersion := eventsPerRP[k].Node.GetResourceVersionInt64()
			if highestRVForRP < currentResourceVersion {
				highestRVForRP = currentResourceVersion
			}
		}

		// Pick up 'nodeChangesPerRP' nodes and make changes and assign the node RV to highestRV + 1
		count := 0
		rvToGenerateRPs := highestRVForRP + 1
		for count < nodeChangesPerRP {
			// Randonly create data update per RP node events list
			i := int(rand.Intn(length))
			node := eventsPerRP[i].Node

			// special case: Consider 5000 changes per RP for 500 nodes per RP
			// each node has 10 changes within this cycle
			currentResourceVersion := node.GetResourceVersionInt64()
			if currentResourceVersion < rvToGenerateRPs {
				node.ResourceVersion = strconv.FormatUint(rvToGenerateRPs, 10)
			} else {
				node.ResourceVersion = strconv.FormatUint(currentResourceVersion+1, 10)
			}
			// record the time to change resource version in resource partition
			node.LastUpdatedTime = time.Now().UTC()

			newEvent := event.NewNodeEvent(node, event.Modified)
			RegionNodeEventsList[j][i] = newEvent

			count++
		}
	}

	klog.V(6).Infof("Actually total (%v) new modified events are created.", changesThreshold)
}

// Create logical node with random UUID
//
func createRandomNode(rv int, loc *location.Location) *types.LogicalNode {
	id := uuid.New()

	return &types.LogicalNode{
		Id:              id.String(),
		ResourceVersion: strconv.Itoa(rv),
		GeoInfo: types.NodeGeoInfo{
			Region:            types.RegionName(loc.GetRegion()),
			ResourcePartition: types.ResourcePartitionName(loc.GetResourcePartition()),
			DataCenter:        types.DataCenterName(strconv.Itoa(int(rand.Intn(10000))) + "-DataCenter"),
			AvailabilityZone:  types.AvailabilityZoneName("AZ-" + strconv.Itoa(int(rand.Intn(6)))),
			FaultDomain:       types.FaultDomainName(id.String() + "-FaultDomain"),
		},
		Taints: types.NodeTaints{
			NoSchedule: false,
			NoExecute:  false,
		},
		SpecialHardwareTypes: types.NodeSpecialHardWareTypeInfo{
			HasGpu:  true,
			HasFPGA: true,
		},
		AllocatableResource: types.NodeResource{
			MilliCPU:         int64(rand.Intn(200) + 20),
			Memory:           int64(rand.Intn(2000)),
			EphemeralStorage: int64(rand.Intn(2000000)),
			AllowedPodNumber: int(rand.Intn(20000000)),
			ScalarResources: map[types.ResourceName]int64{
				"GPU":  int64(rand.Intn(200)),
				"FPGA": int64(rand.Intn(200)),
			},
		},
		Conditions:      111,
		Reserved:        false,
		MachineType:     types.NodeMachineType(id.String() + "-highend"),
		LastUpdatedTime: time.Now().UTC(),
	}
}
