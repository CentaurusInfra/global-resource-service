package storage

import (
	"sync"

	"global-resource-service/resource-management/pkg/common-lib/interfaces/store"
	"global-resource-service/resource-management/pkg/common-lib/types"
)

type DistributorPersistHelper struct {
	nodesToSave    map[string]*types.LogicalNode
	nodeUpdateLock sync.RWMutex

	persistHelper store.Interface
}

var _persistHelper *DistributorPersistHelper = nil
var once sync.Once

func GetDistributorPersistHelper() *DistributorPersistHelper {
	once.Do(func() {
		_persistHelper = &DistributorPersistHelper{
			nodesToSave: make(map[string]*types.LogicalNode),
		}
	})
	return _persistHelper
}

func (c *DistributorPersistHelper) SetPersistHelper(persistTool store.Interface) {
	c.persistHelper = persistTool
}

func (c *DistributorPersistHelper) UpdateNode(newNode *types.LogicalNode) {
	c.nodeUpdateLock.Lock()
	defer c.nodeUpdateLock.Unlock()
	c.nodesToSave[newNode.GetKey()] = newNode
}

func (c *DistributorPersistHelper) PersistNodesAndStoreConfigs(nodeStoreStatus *store.NodeStoreStatus) bool {
	c.nodeUpdateLock.Lock()
	defer c.nodeUpdateLock.Unlock()

	// persist nodes
	resultPersistNodes := c.persistNodes()

	// persist virtual nodes location and latest resource version map
	resultPersistRVs := c.persistStoreStatus(nodeStoreStatus)

	if resultPersistNodes && resultPersistRVs { // flush cache
		c.nodesToSave = make(map[string]*types.LogicalNode)
	}

	return resultPersistNodes && resultPersistRVs
}

func (c *DistributorPersistHelper) PersistVirtualNodesAssignment(assignment *store.VirtualNodeAssignment) bool {
	return c.persistHelper.PersistVirtualNodesAssignments(assignment)
}

func (c *DistributorPersistHelper) persistNodes() bool {
	nodes := make([]*types.LogicalNode, len(c.nodesToSave))
	index := 0
	for _, node := range c.nodesToSave {
		nodes[index] = node
		index++
	}
	return c.persistHelper.PersistNodes(nodes)
}

func (c *DistributorPersistHelper) persistStoreStatus(nodeStoreStatus *store.NodeStoreStatus) bool {
	return c.persistHelper.PersistNodeStoreStatus(nodeStoreStatus)
}
