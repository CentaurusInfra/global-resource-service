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

	// get update nodes
	rvs := make(types.TransitResourceVersionMap)
	for i := 0; i < location.GetRegionNum(); i++ {
		for j := 0; j < location.GetRPNum(); j++ {
			rvLoc := types.RvLocation{
				Region:    location.Region(i),
				Partition: location.ResourcePartition(j),
			}
			rvs[rvLoc] = uint64(nodesPerRP + 1)
		}
	}
	start = time.Now()
	modifiedEvents, count := GetRegionNodeModifiedEventsCRV(rvs)
	// 29.219756ms
	duration = time.Since(start)
	assert.NotNil(t, modifiedEvents)
	assert.Equal(t, 10, len(modifiedEvents))
	assert.Equal(t, uint64(0), count)
	t.Logf("Time to get %d update events: %v", count, duration)
}
