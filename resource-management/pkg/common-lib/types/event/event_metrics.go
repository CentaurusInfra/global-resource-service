package event

import (
	"fmt"
	"k8s.io/klog/v2"
	"sync"

	"global-resource-service/resource-management/pkg/common-lib/metrics"
)

type LatencyMetricsAllCheckpoints struct {
	Aggregator_Received  *metrics.LatencyMetrics
	Distributor_Received *metrics.LatencyMetrics
	Distributor_Sending  *metrics.LatencyMetrics
	Distributor_Sent     *metrics.LatencyMetrics
	Serializer_Encoded   *metrics.LatencyMetrics
	Serializer_Sent      *metrics.LatencyMetrics
}

var latencyNewNodeEvents *LatencyMetricsAllCheckpoints
var latencyUpdateNodeEvents *LatencyMetricsAllCheckpoints
var latencyMetricsNewEventLock sync.RWMutex
var latencyMetricsUpdateEventLock sync.RWMutex

func init() {
	latencyNewNodeEvents = new(LatencyMetricsAllCheckpoints)
	latencyNewNodeEvents.Aggregator_Received = metrics.NewLatencyMetrics(int(metrics.Aggregator_Received))
	latencyNewNodeEvents.Distributor_Received = metrics.NewLatencyMetrics(int(metrics.Distributor_Received))
	latencyNewNodeEvents.Distributor_Sending = metrics.NewLatencyMetrics(int(metrics.Distributor_Sending))
	latencyNewNodeEvents.Distributor_Sent = metrics.NewLatencyMetrics(int(metrics.Distributor_Sent))
	latencyNewNodeEvents.Serializer_Encoded = metrics.NewLatencyMetrics(int(metrics.Serializer_Encoded))
	latencyNewNodeEvents.Serializer_Sent = metrics.NewLatencyMetrics(int(metrics.Serializer_Sent))

	latencyUpdateNodeEvents = new(LatencyMetricsAllCheckpoints)
	latencyUpdateNodeEvents.Aggregator_Received = metrics.NewLatencyMetrics(int(metrics.Aggregator_Received))
	latencyUpdateNodeEvents.Distributor_Received = metrics.NewLatencyMetrics(int(metrics.Distributor_Received))
	latencyUpdateNodeEvents.Distributor_Sending = metrics.NewLatencyMetrics(int(metrics.Distributor_Sending))
	latencyUpdateNodeEvents.Distributor_Sent = metrics.NewLatencyMetrics(int(metrics.Distributor_Sent))
	latencyUpdateNodeEvents.Serializer_Encoded = metrics.NewLatencyMetrics(int(metrics.Serializer_Encoded))
	latencyUpdateNodeEvents.Serializer_Sent = metrics.NewLatencyMetrics(int(metrics.Serializer_Sent))
}

func AddLatencyMetricsAllCheckpoints(e *NodeEvent) {
	if !metrics.ResourceManagementMeasurement_Enabled {
		return
	}
	if e == nil {
		klog.Error("Nil event")
	}
	checkpointsPerEvent := e.GetCheckpoints()
	if checkpointsPerEvent == nil {
		klog.Errorf("Event (%v, Id %s, RV %s) does not have checkpoint stamped", e.Type, e.Node.Id, e.Node.ResourceVersion)
	}
	lastUpdatedTime := e.Node.LastUpdatedTime

	agg_received_time := checkpointsPerEvent[metrics.Aggregator_Received]
	dis_received_time := checkpointsPerEvent[metrics.Distributor_Received]
	dis_sending_time := checkpointsPerEvent[metrics.Distributor_Sending]
	dis_sent_time := checkpointsPerEvent[metrics.Distributor_Sent]
	serializer_encoded_time := checkpointsPerEvent[metrics.Serializer_Encoded]
	serializer_sent_time := checkpointsPerEvent[metrics.Serializer_Sent]

	var latencyToUpdate *LatencyMetricsAllCheckpoints
	if e.Type == Added {
		latencyMetricsNewEventLock.Lock()
		defer latencyMetricsNewEventLock.Unlock()

		latencyToUpdate = latencyNewNodeEvents
	} else { // not differentiate update and delete for now
		latencyMetricsUpdateEventLock.Lock()
		defer latencyMetricsUpdateEventLock.Unlock()

		latencyToUpdate = latencyUpdateNodeEvents
	}
	errMsg := fmt.Sprintf("Event (%v, Id %s, RV %s) ", e.Type, e.Node.Id, e.Node.ResourceVersion) + "does not have %v stamped"
	if !agg_received_time.IsZero() {
		latencyToUpdate.Aggregator_Received.AddLatencyMetrics(agg_received_time.Sub(lastUpdatedTime))
	} else {
		klog.Errorf(errMsg, metrics.Aggregator_Received)
	}
	if !dis_received_time.IsZero() {
		latencyToUpdate.Distributor_Received.AddLatencyMetrics(dis_received_time.Sub(lastUpdatedTime))
	} else {
		klog.Errorf(errMsg, metrics.Distributor_Received)
	}
	if !dis_sending_time.IsZero() {
		latencyToUpdate.Distributor_Sending.AddLatencyMetrics(dis_sending_time.Sub(lastUpdatedTime))
	} else {
		klog.Errorf(errMsg, metrics.Distributor_Sending)
	}
	if !dis_sent_time.IsZero() {
		latencyToUpdate.Distributor_Sent.AddLatencyMetrics(dis_sent_time.Sub(lastUpdatedTime))
	} else {
		klog.Errorf(errMsg, metrics.Distributor_Sent)
	}
	if !serializer_encoded_time.IsZero() {
		latencyToUpdate.Serializer_Encoded.AddLatencyMetrics(serializer_encoded_time.Sub(lastUpdatedTime))
	} else {
		klog.Errorf(errMsg, metrics.Serializer_Encoded)
	}
	if !serializer_sent_time.IsZero() {
		latencyToUpdate.Serializer_Sent.AddLatencyMetrics(serializer_sent_time.Sub(lastUpdatedTime))
	} else {
		klog.Errorf(errMsg, metrics.Serializer_Sent)
	}
	klog.V(6).Infof("[Metrics][Detail][%v] node %v RV %s: %s: %v, %s: %v, %s: %v, %s: %v, %s: %v, %s: %v",
		e.Type, e.Node.Id, e.Node.ResourceVersion,
		metrics.Aggregator_Received_Name, agg_received_time.Sub(lastUpdatedTime),
		metrics.Distributor_Received_Name, dis_received_time.Sub(lastUpdatedTime),
		metrics.Distributor_Sending_Name, dis_sending_time.Sub(lastUpdatedTime),
		metrics.Distributor_Sent_Name, dis_sent_time.Sub(lastUpdatedTime),
		metrics.Serializer_Encoded_Name, serializer_encoded_time.Sub(lastUpdatedTime),
		metrics.Serializer_Sent_Name, serializer_sent_time.Sub(lastUpdatedTime))
}

func PrintLatencyReport() {
	printLatencyReport(latencyMetricsNewEventLock, latencyNewNodeEvents, Added)
	printLatencyReport(latencyMetricsUpdateEventLock, latencyUpdateNodeEvents, Modified)
}

func printLatencyReport(eventLock sync.RWMutex, latencyRecorder *LatencyMetricsAllCheckpoints, eventType EventType) {
	eventLock.RLock()
	agg_received_summary := latencyRecorder.Aggregator_Received.GetSummary()
	dis_received_summary := latencyRecorder.Distributor_Received.GetSummary()
	dis_sending_summary := latencyRecorder.Distributor_Sending.GetSummary()
	dis_sent_summary := latencyRecorder.Distributor_Sent.GetSummary()
	serializer_encoded_summary := latencyRecorder.Serializer_Encoded.GetSummary()
	serializer_sent_summary := latencyRecorder.Serializer_Sent.GetSummary()

	eventLock.RUnlock()
	metrics_Message := "[Metrics][%s][%s] perc50 %v, perc90 %v, perc99 %v. Total count %v"
	klog.Infof(metrics_Message, eventType, metrics.Aggregator_Received_Name, agg_received_summary.P50, agg_received_summary.P90, agg_received_summary.P99, agg_received_summary.TotalCount)
	klog.Infof(metrics_Message, eventType, metrics.Distributor_Received_Name, dis_received_summary.P50, dis_received_summary.P90, dis_received_summary.P99, dis_received_summary.TotalCount)
	klog.Infof(metrics_Message, eventType, metrics.Distributor_Sending_Name, dis_sending_summary.P50, dis_sending_summary.P90, dis_sending_summary.P99, dis_sending_summary.TotalCount)
	klog.Infof(metrics_Message, eventType, metrics.Distributor_Sent_Name, dis_sent_summary.P50, dis_sent_summary.P90, dis_sent_summary.P99, dis_sent_summary.TotalCount)
	klog.Infof(metrics_Message, eventType, metrics.Serializer_Encoded_Name, serializer_encoded_summary.P50, serializer_encoded_summary.P90, serializer_encoded_summary.P99, serializer_encoded_summary.TotalCount)
	klog.Infof(metrics_Message, eventType, metrics.Serializer_Sent_Name, serializer_sent_summary.P50, serializer_sent_summary.P90, serializer_sent_summary.P99, serializer_sent_summary.TotalCount)
}
