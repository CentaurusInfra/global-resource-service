package types

import (
	"github.com/google/uuid"
	"runtime"
	"strconv"
	"testing"
	"time"

	"global-resource-service/resource-management/pkg/common-lib/metrics"
)

var defaultLocBeijing_RP1 = NewLocation(Beijing, ResourcePartition1)
var rvToGenerate = 0

func Test_PrintLatencyReport(t *testing.T) {
	ne := createNodeEvent()

	time.Sleep(100 * time.Millisecond)
	ne.SetCheckpoint(metrics.Aggregator_Received)
	ne.SetCheckpoint(metrics.Distributor_Received)
	ne.SetCheckpoint(metrics.Distributor_Sending)
	ne.SetCheckpoint(metrics.Distributor_Sent)
	ne.SetCheckpoint(metrics.Serializer_Encoded)
	ne.SetCheckpoint(metrics.Serializer_Sent)
	AddLatencyMetricsAllCheckpoints(ne)
	PrintLatencyReport()
}

func createRandomNode(rv int, loc *Location) *LogicalNode {
	id := uuid.New()
	return &LogicalNode{
		Id:              id.String(),
		ResourceVersion: strconv.Itoa(rv),
		GeoInfo: NodeGeoInfo{
			Region:            RegionName(loc.GetRegion()),
			ResourcePartition: ResourcePartitionName(loc.GetResourcePartition()),
		},
		LastUpdatedTime: NewTime(time.Now().UTC()),
	}
}

func Test_MemoryUsageOfLatencyReport(t *testing.T) {
	count := 1000000
	// Get memory usage for 1M node events
	metrics.ResourceManagementMeasurement_Enabled = false
	nodes := make([]*NodeEvent, count)
	for i := 0; i < count; i++ {
		nodes[i] = createNodeEvent()
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	t.Logf("Alloc = %v, TotalAlloc = %v, Sys = %v, NumGC = %v", m.Alloc, m.TotalAlloc, m.Sys, m.NumGC)

	// Enable metrics
	metrics.ResourceManagementMeasurement_Enabled = true
	for i := 0; i < count; i++ {
		nodes[i].SetCheckpoint(metrics.Aggregator_Received)
		nodes[i].SetCheckpoint(metrics.Distributor_Received)
		nodes[i].SetCheckpoint(metrics.Distributor_Sending)
		nodes[i].SetCheckpoint(metrics.Distributor_Sent)
		nodes[i].SetCheckpoint(metrics.Serializer_Encoded)
		nodes[i].SetCheckpoint(metrics.Serializer_Sent)
		AddLatencyMetricsAllCheckpoints(nodes[i])
	}
	PrintLatencyReport()

	runtime.ReadMemStats(&m)
	t.Logf("Alloc = %v, TotalAlloc = %v, Sys = %v, NumGC = %v", m.Alloc, m.TotalAlloc, m.Sys, m.NumGC)
}

func createNodeEvent() *NodeEvent {
	n := createRandomNode(rvToGenerate+1, defaultLocBeijing_RP1)
	ne := NewNodeEvent(n, Added)
	return ne
}
