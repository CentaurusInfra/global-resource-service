package cache

import (
	"errors"
	"fmt"
	"resource-management/pkg/common-lib/types"
	"resource-management/pkg/common-lib/types/event"
	"resource-management/pkg/common-lib/types/location"
	"sort"
	"sync"
)

// TODO - read from config
const LengthOfNodeEventQueue = 10000

type nodeEventQueueByLoc struct {
	circularEventQueue []*event.NodeEvent
	// circular event queue start position and end position
	startPos int
	endPos   int

	// mutex for event queue operation
	eqLock sync.RWMutex
}

func newNodeQueueByLoc() *nodeEventQueueByLoc {
	return &nodeEventQueueByLoc{
		circularEventQueue: make([]*event.NodeEvent, LengthOfNodeEventQueue),
		startPos:           0,
		endPos:             0,
		eqLock:             sync.RWMutex{},
	}
}

func (qloc *nodeEventQueueByLoc) enqueueEvent(e *event.NodeEvent) {
	qloc.eqLock.Lock()
	defer qloc.eqLock.Unlock()

	if qloc.endPos == qloc.startPos+LengthOfNodeEventQueue {
		// cache is full - remove the oldest element
		qloc.startPos++
	}

	qloc.circularEventQueue[qloc.endPos%LengthOfNodeEventQueue] = e
	qloc.endPos++
}

func (qloc *nodeEventQueueByLoc) dequeueEvents(startIndex int) ([]*event.NodeEvent, error) {
	qloc.eqLock.RLock()
	defer qloc.eqLock.RUnlock()

	if qloc.startPos == qloc.endPos || qloc.startPos > startIndex || startIndex > qloc.endPos { // queue is empty or out of range
		return nil, errors.New(fmt.Sprintf("Event queue start pos %d, end pos %d, invalid start index %d", qloc.startPos, qloc.endPos, startIndex))
	}

	length := qloc.endPos - qloc.startPos
	result := make([]*event.NodeEvent, length)
	for i := 0; i < length; i++ {
		result[i] = qloc.circularEventQueue[(startIndex+i)%LengthOfNodeEventQueue]
	}

	qloc.startPos = qloc.endPos

	return result, nil
}

func (qloc *nodeEventQueueByLoc) getEventIndexSinceResourceVersion(resourceVersion uint64) (int, error) {
	qloc.eqLock.RLock()
	defer qloc.eqLock.RUnlock()
	if qloc.endPos-qloc.startPos == 0 {
		return -1, errors.New(fmt.Sprintf("Empty event queue"))
	}
	oldestRV := qloc.circularEventQueue[qloc.startPos%LengthOfNodeEventQueue].GetNode().GetResourceVersion()
	if oldestRV > resourceVersion {
		return -1, errors.New(fmt.Sprintf("Loc %s events oldest resource Version %d is newer than requested resource version %d",
			qloc.circularEventQueue[qloc.startPos%LengthOfNodeEventQueue].GetNode().GetLocation(),
			oldestRV, resourceVersion))
	}

	index := sort.Search(qloc.endPos-qloc.startPos, func(i int) bool {
		return qloc.circularEventQueue[(qloc.startPos+i)%LengthOfNodeEventQueue].GetNode().GetResourceVersion() > resourceVersion
	})
	if index >= qloc.endPos || index < qloc.startPos {
		return -1, errors.New(fmt.Sprintf("Event queue start pos %d, end pos %d, found invalid start index %d", qloc.startPos, qloc.endPos, index))
	}
	return index, nil
}

type NodeEventQueue struct {
	// corresponding client id
	clientId  string
	watchChan chan *event.NodeEvent

	eventQueueByLoc map[location.Location]*nodeEventQueueByLoc
	locationLock    sync.RWMutex
}

func NewNodeEventQueue(clientId string) *NodeEventQueue {
	queue := &NodeEventQueue{
		clientId:        clientId,
		eventQueueByLoc: make(map[location.Location]*nodeEventQueueByLoc),
	}

	return queue
}

func (eq *NodeEventQueue) EnqueueEvent(e *event.NodeEvent) {
	if eq.watchChan != nil {
		go func() {
			eq.watchChan <- e
		}()
	}

	eq.locationLock.Lock()
	defer eq.locationLock.Unlock()
	queueByLoc, isOK := eq.eventQueueByLoc[*e.GetNode().GetLocation()]
	if !isOK {
		queueByLoc = newNodeQueueByLoc()
		eq.eventQueueByLoc[*e.GetNode().GetLocation()] = queueByLoc
	}
	queueByLoc.enqueueEvent(e)
}

func (eq *NodeEventQueue) Watch(rvs types.ResourceVersionMap, clientWatchChan chan *event.NodeEvent, stopCh chan struct{}) error {
	if eq.watchChan != nil {
		return errors.New("Currently only support one watcher per node event queue.")
	}

	// get events already in queues
	events, err := eq.getAllEventsSinceResourceVersion(rvs)
	if err != nil {
		return err
	}

	eq.watchChan = make(chan *event.NodeEvent)
	// writing event to channel
	go func(downstreamCh chan *event.NodeEvent, initEvents []*event.NodeEvent, stopCh chan struct{}, upstreamCh chan *event.NodeEvent) {
		if downstreamCh == nil {
			return
		}
		// send init events
		for i := 0; i < len(initEvents); i++ {
			downstreamCh <- initEvents[i]
		}

		// continue to watch
		for {
			select {
			case <-stopCh:
				eq.watchChan = nil
				fmt.Printf("Watch stopped due to client request")
				return
			case event, ok := <-upstreamCh:
				if !ok {
					break
				}
				downstreamCh <- event
			}
		}

	}(clientWatchChan, events, stopCh, eq.watchChan)

	return nil
}

func (eq *NodeEventQueue) getAllEventsSinceResourceVersion(rvs types.ResourceVersionMap) ([]*event.NodeEvent, error) {
	locStartPostitions := make(map[location.Location]int)

	for loc, rv := range rvs {
		qByLoc, isOK := eq.eventQueueByLoc[loc]
		if isOK {
			startIndex, err := qByLoc.getEventIndexSinceResourceVersion(rv)
			if err != nil {
				return nil, err
			}
			locStartPostitions[loc] = startIndex
		}
	}

	nodeEvents := make([]*event.NodeEvent, 0)
	for loc, qByLoc := range eq.eventQueueByLoc {
		startIndex, isOK := locStartPostitions[loc]
		var events []*event.NodeEvent
		var err error
		if isOK {
			events, err = qByLoc.dequeueEvents(startIndex)
		} else {
			events, err = qByLoc.dequeueEvents(qByLoc.startPos)
		}
		if err != nil {
			return nil, err
		}

		if len(events) > 0 {
			nodeEvents = append(nodeEvents, events...)
		}
	}

	return nodeEvents, nil
}