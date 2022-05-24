package store

import (
	"global-resource-service/resource-management/pkg/common-lib/types"
	"global-resource-service/resource-management/pkg/common-lib/types/location"
)

type Interface interface {
	PersistNodes([]*types.LogicalNode) bool
	PersistNodeStoreStatus(*NodeStoreStatus) bool
	PersistVirtualNodesAssignments(*VirtualNodeAssignment) bool
}

type NodeStoreStatus struct {
	// # of regions
	RegionNum int

	// # of max resource partition in each region
	PartitionMaxNum int

	// virutal node number per resource partition
	VirtualNodeNumPerRP int

	// Latest resource version map
	CurrentResourceVerions types.ResourceVersionMap
}

type VirtualNodeAssignment struct {
	ClientId     string
	VirtualNodes []*VirtualNodeConfig
}

type VirtualNodeConfig struct {
	Lowerbound float64
	Upperbound float64
	Location   location.Location
}
