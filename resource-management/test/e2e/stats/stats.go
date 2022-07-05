package stats

import (
	"time"

	"k8s.io/klog/v2"
)

const LongWatchThreshold = time.Duration(1000 * time.Millisecond)

type RegisterClientStats struct {
	RegisterClientDuration time.Duration
}

func NewRegisterClientStats() *RegisterClientStats {
	return &RegisterClientStats{}
}

func (rs *RegisterClientStats) PrintStats() {
	klog.Infof("RegisterClientDuration: %v", rs.RegisterClientDuration)
}

type ListStats struct {
	ListDuration        time.Duration
	NumberOfNodesListed int
}

func NewListStats() *ListStats {
	return &ListStats{}
}

func (ls *ListStats) PrintStats() {
	klog.Infof("ListDuration: %v. Number of nodes listed: %v", ls.ListDuration, ls.NumberOfNodesListed)
}

type WatchStats struct {
	WatchDuration            time.Duration
	NumberOfProlongedItems   int
	NumberOfAddedNodes       int
	NumberOfUpdatedNodes     int
	NumberOfDeletedNodes     int
	NumberOfProlongedWatches int
}

func NewWatchStats() *WatchStats {
	return &WatchStats{}
}

func (ws *WatchStats) PrintStats() {
	klog.Infof("Watch session last: %v", ws.WatchDuration)
	klog.Infof("Number of nodes Added: %v", ws.NumberOfAddedNodes)
	klog.Infof("Number of nodes Updated: %v", ws.NumberOfUpdatedNodes)
	klog.Infof("Number of nodes Deleted: %v", ws.NumberOfDeletedNodes)
	klog.Infof("Number of nodes watch prolonged than %v: %v", LongWatchThreshold, ws.NumberOfProlongedItems)
}
