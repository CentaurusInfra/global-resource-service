package distributor

import (
	"resource-management/pkg/common-lib/types"
	"resource-management/pkg/common-lib/types/event"
)

type Interface interface {
	RegisterClient(requestedHostNum int) (string, bool, error)
	ListNodesForClient(clientId string) ([]*types.Node, types.ResourceVersionMap, error)
	Watch(clientId string, rvs types.ResourceVersionMap, watchChan chan *event.NodeEvent, stopCh chan struct{}) error
	ProcessEvents(events []*event.NodeEvent) (bool, types.ResourceVersionMap)
}
