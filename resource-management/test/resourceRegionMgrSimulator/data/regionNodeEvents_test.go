package data

import (
	"github.com/stretchr/testify/assert"
	"global-resource-service/resource-management/pkg/common-lib/types"
	"global-resource-service/resource-management/pkg/common-lib/types/location"
	"testing"
	"time"
)

func TestGetRegionNodeModifiedEventsCRV(t *testing.T) {
	// create nodes
	rpNum := 10
	nodesPerRP := 50000
	start := time.Now()
	Init("Beijing", rpNum, nodesPerRP)
	// 2.811475159s
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
	// 29.219756ms -> 3.352µs
	duration = time.Since(start)
	assert.NotNil(t, modifiedEvents)
	assert.Equal(t, 10, len(modifiedEvents))
	t.Logf("Time to get %d update events: %v", count, duration)
	assert.True(t, count >= uint64(atEachMin10))
}
