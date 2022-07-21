package event

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"runtime"
	"strconv"
	"testing"
	"time"

	"global-resource-service/resource-management/pkg/common-lib/metrics"
	"global-resource-service/resource-management/pkg/common-lib/types"
	"global-resource-service/resource-management/pkg/common-lib/types/location"
)

var defaultLocBeijing_RP1 = location.NewLocation(location.Beijing, location.ResourcePartition1)
var rvToGenerate = 0

func Test_PrintLatencyReport(t *testing.T) {
	for i := 0; i < 10000; i++ {
		ne := createNodeEvent()

		//time.Sleep(100 * time.Millisecond)
		ne.SetCheckpoint(metrics.Aggregator_Received)
		ne.SetCheckpoint(metrics.Distributor_Received)
		ne.SetCheckpoint(metrics.Distributor_Sending)
		ne.SetCheckpoint(metrics.Distributor_Sent)
		ne.SetCheckpoint(metrics.Serializer_Encoded)
		ne.SetCheckpoint(metrics.Serializer_Sent)
		AddLatencyMetricsAllCheckpoints(ne)

		assert.Equal(t, i+1, latencyNewNodeEvents.Aggregator_Received.GetSummary().TotalCount)
		assert.Equal(t, i+1, latencyNewNodeEvents.Distributor_Received.GetSummary().TotalCount)
		assert.Equal(t, i+1, latencyNewNodeEvents.Distributor_Sending.GetSummary().TotalCount)
		assert.Equal(t, i+1, latencyNewNodeEvents.Distributor_Sent.GetSummary().TotalCount)
		assert.Equal(t, i+1, latencyNewNodeEvents.Serializer_Encoded.GetSummary().TotalCount)
		assert.Equal(t, i+1, latencyNewNodeEvents.Serializer_Sent.GetSummary().TotalCount)
	}

	for i := 0; i < 5; i++ {
		updateEvent := updateNodeEvent()
		time.Sleep(50 * time.Millisecond)
		updateEvent.SetCheckpoint(metrics.Aggregator_Received)
		updateEvent.SetCheckpoint(metrics.Distributor_Received)
		updateEvent.SetCheckpoint(metrics.Distributor_Sending)
		updateEvent.SetCheckpoint(metrics.Distributor_Sent)
		updateEvent.SetCheckpoint(metrics.Serializer_Encoded)
		updateEvent.SetCheckpoint(metrics.Serializer_Sent)
		AddLatencyMetricsAllCheckpoints(updateEvent)

		assert.Equal(t, i+1, latencyUpdateNodeEvents.Aggregator_Received.GetSummary().TotalCount)
		assert.Equal(t, i+1, latencyUpdateNodeEvents.Distributor_Received.GetSummary().TotalCount)
		assert.Equal(t, i+1, latencyUpdateNodeEvents.Distributor_Sending.GetSummary().TotalCount)
		assert.Equal(t, i+1, latencyUpdateNodeEvents.Distributor_Sent.GetSummary().TotalCount)
		assert.Equal(t, i+1, latencyUpdateNodeEvents.Serializer_Encoded.GetSummary().TotalCount)
		assert.Equal(t, i+1, latencyUpdateNodeEvents.Serializer_Sent.GetSummary().TotalCount)
	}
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

func updateNodeEvent() *NodeEvent {
	n := createRandomNode(rvToGenerate+1, defaultLocBeijing_RP1)
	ne := NewNodeEvent(n, Modified)
	return ne
}
