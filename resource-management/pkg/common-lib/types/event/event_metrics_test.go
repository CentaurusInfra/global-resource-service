package event

import (
	"github.com/google/uuid"
	"global-resource-service/resource-management/pkg/common-lib/metrics"
	"strconv"
	"testing"
	"time"

	"global-resource-service/resource-management/pkg/common-lib/types"
	"global-resource-service/resource-management/pkg/common-lib/types/location"
)

var defaultLocBeijing_RP1 = location.NewLocation(location.Beijing, location.ResourcePartition1)
var rvToGenerate = 0

func Test_PrintLatencyReport(t *testing.T) {
	n := createRandomNode(rvToGenerate+1, defaultLocBeijing_RP1)
	ne := NewNodeEvent(n, Added)

	time.Sleep(100 * time.Millisecond)
	ne.SetCheckpoint(metrics.Aggregator_Received)
	ne.SetCheckpoint(metrics.Distributor_Received)
	ne.SetCheckpoint(metrics.Distributor_Sending)
	ne.SetCheckpoint(metrics.Distributor_Sent)
	ne.SetCheckpoint(metrics.Serializer_Sent)
	AddLatencyMetricsAllCheckpoints(ne)
	PrintLatencyReport()
}

func createRandomNode(rv int, loc *location.Location) *types.LogicalNode {
	id := uuid.New()
	return &types.LogicalNode{
		Id:              id.String(),
		ResourceVersion: strconv.Itoa(rv),
		GeoInfo: types.NodeGeoInfo{
			Region:            types.RegionName(loc.GetRegion()),
			ResourcePartition: types.ResourcePartitionName(loc.GetResourcePartition()),
		},
		LastUpdatedTime: time.Now().UTC(),
	}
}
