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
	"github.com/stretchr/testify/assert"
	"global-resource-service/resource-management/pkg/common-lib/types"
	"global-resource-service/resource-management/pkg/common-lib/types/location"
	"sync"
	"testing"
	"time"
)

var oldValue_maxPullUpdateEventsSize int
var singleTestLock = sync.Mutex{}

func setUp() {
	singleTestLock.Lock()
	oldValue_maxPullUpdateEventsSize = maxPullUpdateEventsSize
}

func tearDown() {
	maxPullUpdateEventsSize = oldValue_maxPullUpdateEventsSize
	singleTestLock.Unlock()
}

func TestGetRegionNodeModifiedEventsCRV(t *testing.T) {
	setUp()
	defer tearDown()

	maxPullUpdateEventsSize = 50000

	// create nodes
	rpNum := 10
	nodesPerRP := 50000
	start := time.Now()
	Init("Beijing", rpNum, nodesPerRP)
	// 2.827539846s
	duration := time.Since(start)
	assert.Equal(t, rpNum, len(RegionNodeEventsList))
	assert.Equal(t, nodesPerRP, len(RegionNodeEventsList[0]))
	t.Logf("Time to generate %d init events: %v", rpNum*nodesPerRP, duration)

	// generate update node events
	makeDataUpdate(atEachMin10)

	// get update nodes
	rvs := make(types.TransitResourceVersionMap)
	for j := 0; j < location.GetRPNum(); j++ {
		rvLoc := types.RvLocation{
			Region:    location.Region(RegionId),
			Partition: location.ResourcePartition(j),
		}
		rvs[rvLoc] = uint64(nodesPerRP)
	}
	start = time.Now()
	modifiedEvents, count := GetRegionNodeModifiedEventsCRV(rvs)
	// 29.219756ms -> 4.096µs
	duration = time.Since(start)
	assert.NotNil(t, modifiedEvents)
	assert.Equal(t, rpNum, len(modifiedEvents))
	t.Logf("Time to get %d update events in Daily data pattern: %v", count, duration)
	assert.Equal(t, uint64(atEachMin10), count)

	//check remaining event list
	assert.Equal(t, rpNum, len(RegionNodeUpdateEventList))
	for i := 0; i < rpNum; i++ {
		assert.Nil(t, nil, RegionNodeUpdateEventList[i])
	}

	// update again
	makeDataUpdate(atEachMin10)
	makeDataUpdate(atEachMin10)
	for j := 0; j < location.GetRPNum(); j++ {
		rvLoc := types.RvLocation{
			Region:    location.Region(RegionId),
			Partition: location.ResourcePartition(j),
		}
		rvs[rvLoc] = uint64(nodesPerRP + 1)
	}
	start = time.Now()
	modifiedEvents, count = GetRegionNodeModifiedEventsCRV(rvs)
	// 3.987µs
	duration = time.Since(start)
	assert.NotNil(t, modifiedEvents)
	assert.Equal(t, rpNum, len(modifiedEvents))
	t.Logf("Time to get %d update events in Daily data pattern: %v", count, duration)
	assert.Equal(t, atEachMin10*2, int(count))

	//check remaining event list
	assert.Equal(t, rpNum, len(RegionNodeUpdateEventList))
	for i := 0; i < rpNum; i++ {
		assert.Nil(t, nil, RegionNodeUpdateEventList[i])
	}

	// generate update node events of Outage pattern
	makeOneRPDown()

	// get update nodes
	for j := 0; j < location.GetRPNum(); j++ {
		rvLoc := types.RvLocation{
			Region:    location.Region(RegionId),
			Partition: location.ResourcePartition(j),
		}
		rvs[rvLoc] = uint64(nodesPerRP)
	}
	start = time.Now()
	modifiedEvents, count = GetRegionNodeModifiedEventsCRV(rvs)
	// 38.041491ms -> 3.929264ms on AWS EC2 instance (t2.2xlarge - 8 vcpu/32G memory)
	duration = time.Since(start)
	assert.NotNil(t, modifiedEvents)
	assert.Equal(t, rpNum, len(modifiedEvents))
	t.Logf("Time to get %d update events in Outage data pattern: %v", count, duration)
	assert.Equal(t, uint64(nodesPerRP), count)

	//check remaining event list
	assert.Equal(t, rpNum, len(RegionNodeUpdateEventList))
	for i := 0; i < rpNum; i++ {
		assert.Nil(t, nil, RegionNodeUpdateEventList[i])
	}
}

func TestGetRegionNodeModifiedEventsCRV_WithEventsLimit(t *testing.T) {
	setUp()
	defer tearDown()

	maxPullUpdateEventsSize = 10000

	// create nodes
	rpNum := 10
	nodesPerRP := 25000
	start := time.Now()
	Init("Beijing", rpNum, nodesPerRP)
	// 2.827539846s
	duration := time.Since(start)
	assert.Equal(t, rpNum, len(RegionNodeEventsList))
	assert.Equal(t, nodesPerRP, len(RegionNodeEventsList[0]))
	t.Logf("Time to generate %d init events: %v", rpNum*nodesPerRP, duration)

	// generate RP down node events
	makeOneRPDown()

	// get update nodes
	rvs := make(types.TransitResourceVersionMap)
	for j := 0; j < location.GetRPNum(); j++ {
		rvLoc := types.RvLocation{
			Region:    location.Region(RegionId),
			Partition: location.ResourcePartition(j),
		}
		rvs[rvLoc] = uint64(nodesPerRP)
	}

	totalEventCount := 0
	for i := 0; i < (nodesPerRP+maxPullUpdateEventsSize-1)/maxPullUpdateEventsSize; i++ { // set a loop limit
		start = time.Now()
		modifiedEvents, count := GetRegionNodeModifiedEventsCRV(rvs)
		duration = time.Since(start)
		assert.NotNil(t, modifiedEvents)
		assert.Equal(t, rpNum, len(modifiedEvents))
		t.Logf("Time to get %d update events in Outage data pattern: %v", count, duration)
		if maxPullUpdateEventsSize > int(count) {
			assert.Equal(t, nodesPerRP-totalEventCount, int(count))
		}

		totalEventCount += int(count)
	}
	assert.Equal(t, nodesPerRP, totalEventCount)
}
