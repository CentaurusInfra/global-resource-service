package distributor

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"strconv"
	"sync"

	"global-resource-service/resource-management/pkg/common-lib/types"
	"global-resource-service/resource-management/pkg/common-lib/types/event"
	"global-resource-service/resource-management/pkg/common-lib/types/location"
	"global-resource-service/resource-management/pkg/distributor/cache"
	"global-resource-service/resource-management/pkg/distributor/storage"
)

type ResourceDistributor struct {
	defaultNodeStore *storage.NodeStore

	// clientId to node event queue
	nodeEventQueueMap   map[string]*cache.NodeEventQueue
	eventProcessingLock sync.RWMutex

	// clientId to virtual node store map
	clientToStores map[string][]*storage.VirtualNodeStore
	allocateLock   sync.Mutex
}

var _distributor *ResourceDistributor = nil
var once sync.Once

const (
	MinimalRequestHostNum = 50
)

var virutalStoreNumPerResourcePartition = 200 // 10K per resource partition, 50 hosts per virtual node store

func GetResourceDistributor() *ResourceDistributor {
	once.Do(func() {
		_distributor = &ResourceDistributor{
			defaultNodeStore:  createNodeStore(),
			nodeEventQueueMap: make(map[string]*cache.NodeEventQueue),
			clientToStores:    make(map[string][]*storage.VirtualNodeStore),
		}
	})
	return _distributor
}

// TODO - get virtual node number, region num, partition num from external
func createNodeStore() *storage.NodeStore {
	return storage.NewNodeStore(virutalStoreNumPerResourcePartition, location.GetRegionNum(), location.GetRPNum())
}

func (dis *ResourceDistributor) RegisterClient(requestedHostNum int) (string, bool, error) {
	clientId := uuid.New().String()
	result, err := dis.allocateNodesToClient(clientId, requestedHostNum)
	if err != nil {
		fmt.Printf("Error register client. Error %v\n", err)
	}
	fmt.Printf("Registered client id: %s\n", clientId)
	return clientId, result, err
}

func (dis *ResourceDistributor) allocateNodesToClient(clientId string, requestedHostNum int) (bool, error) {
	dis.allocateLock.Lock()
	defer dis.allocateLock.Unlock()
	if requestedHostNum <= MinimalRequestHostNum {
		return false, types.Error_HostRequestLessThanMiniaml
	} else if requestedHostNum > dis.defaultNodeStore.GetTotalHostNum() {
		return false, types.Error_HostRequestExceedLimit
	} else if !dis.defaultNodeStore.CheckFreeCapacity(requestedHostNum) {
		return false, types.Error_HostRequestExceedCapacity
	}

	// check client id existence
	if _, isOK := dis.nodeEventQueueMap[clientId]; isOK {
		return false, types.Error_ClientIdExisted
	}
	if _, isOK := dis.clientToStores[clientId]; isOK {
		return false, types.Error_ClientIdExisted
	}

	// allocate virtual nodes to client
	// get all virtual stores that are unassigned
	allStores := dis.defaultNodeStore.GetVirtualStores()
	freeStores := make(map[*storage.VirtualNodeStore]bool)
	for _, vs := range *allStores {
		if vs.GetAssignedClient() == "" && vs.GetHostNum() > 0 {
			freeStores[vs] = true
		}
	}
	if len(freeStores) == 0 {
		return false, errors.New("No available hosts")
	}

	// Get sorted virtual node stores based on ordering criteria
	storesToSelectInorder := dis.getSortedVirtualStores(freeStores)
	selectedStores := make([]*storage.VirtualNodeStore, 0)
	assignedHostCount := 0
	hostAssignIsOK := false
	for i := 0; i < len(storesToSelectInorder); i++ {
		selectedStores = append(selectedStores, storesToSelectInorder[i])
		assignedHostCount += (*storesToSelectInorder[i]).GetHostNum()
		if assignedHostCount >= requestedHostNum {
			hostAssignIsOK = true
			break
		}
	}
	if !hostAssignIsOK {
		return false, errors.New("Not enough hosts")
	}

	// Create event queue for client
	eventQueue := cache.NewNodeEventQueue(clientId)
	dis.nodeEventQueueMap[clientId] = eventQueue
	dis.addBookmarkEvent(selectedStores, eventQueue)

	// Assign virtual node stores to client
	for _, store := range selectedStores {
		store.AssignToClient(clientId, eventQueue)
	}
	dis.clientToStores[clientId] = selectedStores

	return true, nil
}

func (dis *ResourceDistributor) addBookmarkEvent(stores []*storage.VirtualNodeStore, eventQueue *cache.NodeEventQueue) {
	locations := make(map[location.Location]bool)

	for _, store := range stores {
		loc := store.GetLocation()
		if _, isOK := locations[loc]; !isOK {
			locations[loc] = true

			node := types.NewNode("", strconv.FormatUint(store.GetOneNode().GetResourceVersion(), 10), "", &loc)
			bookmarkEvent := event.NewNodeEvent(node, event.Bookmark)
			eventQueue.EnqueueEvent(bookmarkEvent)
		}
	}
}

// TODO: sort virtual node stores based on ordering criteria
// Do not sort by host number since this can lead to over assignment more and more
func (dis *ResourceDistributor) getSortedVirtualStores(stores map[*storage.VirtualNodeStore]bool) []*storage.VirtualNodeStore {
	sortedStores := make([]*storage.VirtualNodeStore, len(stores))
	index := 0
	for vs, isOK := range stores {
		if isOK {
			sortedStores[index] = vs
			index++
		}
	}

	return sortedStores
}

func (dis *ResourceDistributor) ListNodesForClient(clientId string) ([]*types.Node, types.ResourceVersionMap, error) {
	if clientId == "" {
		return nil, nil, errors.New("Empty clientId")
	}
	dis.allocateLock.Lock()
	defer dis.allocateLock.Unlock()
	assignedStores, isOK := dis.clientToStores[clientId]
	if !isOK {
		return nil, nil, errors.New(fmt.Sprintf("Client %s not registered.", clientId))
	}

	dis.eventProcessingLock.RLock()
	nodesByStore := make([][]*types.Node, len(assignedStores))
	rvMapByStore := make([]types.ResourceVersionMap, len(assignedStores))
	hostCount := 0
	for i := 0; i < len(assignedStores); i++ {
		nodesByStore[i], rvMapByStore[i] = assignedStores[i].SnapShot()
		hostCount += len(nodesByStore[i])
	}
	dis.eventProcessingLock.RUnlock()

	// combine to single array of nodeEvent
	nodes := make([]*types.Node, hostCount)
	index := 0
	for i := 0; i < len(nodesByStore); i++ {
		for j := 0; j < len(nodesByStore[i]); j++ {
			nodes[index] = nodesByStore[i][j]
			index++
		}
	}

	// combine to single ResourceVersionMap
	finalRVs := make(types.ResourceVersionMap)
	for i := 0; i < len(rvMapByStore); i++ {
		currentRVs := rvMapByStore[i]
		for loc, rv := range currentRVs {
			if oldRV, isOK := finalRVs[loc]; isOK {
				if oldRV < rv {
					finalRVs[loc] = rv
				}
			} else {
				finalRVs[loc] = rv
			}
		}
	}

	return nodes, finalRVs, nil
}

func (dis *ResourceDistributor) Watch(clientId string, rvs types.ResourceVersionMap, watchChan chan *event.NodeEvent, stopCh chan struct{}) error {
	var nodeEventQueue *cache.NodeEventQueue
	var isOK bool
	if nodeEventQueue, isOK = dis.nodeEventQueueMap[clientId]; !isOK || nodeEventQueue == nil {
		return errors.New(fmt.Sprintf("Client %s not registered", clientId))
	}
	if rvs == nil {
		return errors.New("Invalid resource versions: nil")
	}
	if watchChan == nil {
		return errors.New("Watch channel not provided")
	}
	if stopCh == nil {
		return errors.New("Stop watch channel not provided")
	}

	return nodeEventQueue.Watch(rvs, watchChan, stopCh)
}

func (dis *ResourceDistributor) ProcessEvents(events []*event.NodeEvent) (bool, types.ResourceVersionMap) {
	result, rvMap := dis.defaultNodeStore.ProcessNodeEvents(events)

	return result, rvMap
}
