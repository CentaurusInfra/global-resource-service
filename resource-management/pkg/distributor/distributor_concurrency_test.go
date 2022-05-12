package distributor

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"resource-management/pkg/common-lib/types"
	"resource-management/pkg/common-lib/types/event"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestSinglePRMutipleClients_Workflow(t *testing.T) {
	testCases := []struct {
		name           string
		nodeNum        int
		clientNum      int
		hostPerClient  int
		updateEventNum int
	}{
		{
			name:           "Test 10K nodes with 5 clients each has 500 hosts, each got 1K update events",
			nodeNum:        10000,
			clientNum:      5,
			hostPerClient:  500,
			updateEventNum: 1000,
		},
		{
			name:           "Test 10K nodes with 5 clients each has 500 , each got 10K update events",
			nodeNum:        10000,
			clientNum:      5,
			hostPerClient:  500,
			updateEventNum: 10000,
		},
		{
			name:           "Test 10K nodes with 5 clients each has 500 , each got 100K update events",
			nodeNum:        10000,
			clientNum:      5,
			hostPerClient:  500,
			updateEventNum: 100000,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			distributor := setUp()
			defer tearDown(distributor)

			// initialize node store with tt.nodeNum nodes
			eventsAdd := generateAddNodeEvent(tt.nodeNum)

			start := time.Now()
			result, rvMap := distributor.ProcessEvents(eventsAdd)
			duration := time.Since(start)

			assert.True(t, result)
			assert.NotNil(t, rvMap)
			assert.Equal(t, tt.nodeNum, distributor.defaultNodeStore.GetTotalHostNum())

			// register clients
			clientIds := make([]string, tt.clientNum)
			for i := 0; i < tt.clientNum; i++ {
				start = time.Now()
				clientId, result, err := distributor.RegisterClient(tt.hostPerClient)
				duration += time.Since(start)

				assert.True(t, result, "Expecting register client successfully")
				assert.NotNil(t, clientId, "Expecting not nil client id")
				assert.False(t, clientId == "", "Expecting non empty client id")
				assert.Nil(t, err, "Expecting nil error")
				clientIds[i] = clientId
			}

			// client list nodes
			latestRVsByClient := make([]types.ResourceVersionMap, tt.clientNum)
			nodesByClient := make([][]*types.Node, tt.clientNum)
			for i := 0; i < tt.clientNum; i++ {
				clientId := clientIds[i]

				start = time.Now()
				nodes, latestRVs, err := distributor.ListNodesForClient(clientId)
				duration += time.Since(start)

				assert.Nil(t, err)
				assert.NotNil(t, latestRVs)
				assert.True(t, len(nodes) >= tt.hostPerClient)
				// fmt.Printf("Client %d %s latest rvs: %v.Total hosts: %d\n", i, clientId, latestRVs, len(nodes))
				latestRVsByClient[i] = latestRVs
				nodesByClient[i] = nodes

				// check each node event
				nodeIds := make(map[string]bool)
				for _, node := range nodes {
					assert.NotNil(t, node.GetLocation())
					assert.True(t, latestRVs[*node.GetLocation()] >= node.GetResourceVersion())
					if _, isOK := nodeIds[node.GetId()]; isOK {
						assert.Fail(t, "List nodes cannot have more than ")
					} else {
						nodeIds[node.GetId()] = true
					}
				}
				assert.Equal(t, len(nodes), len(nodeIds))
			}

			// clients watch nodes
			stopCh := make(chan struct{})
			allWaitGroup := new(sync.WaitGroup)
			start = time.Now()
			for i := 0; i < tt.clientNum; i++ {
				watchCh := make(chan *event.NodeEvent)
				err := distributor.Watch(clientIds[i], latestRVsByClient[i], watchCh, stopCh)
				if err != nil {
					assert.Fail(t, "Encountered error while building watch connection.", "Encountered error while building watch connection. Error %v", err)
					return
				}
				allWaitGroup.Add(1)

				go func(expectedEventCount int, watchCh chan *event.NodeEvent, wg *sync.WaitGroup) {
					eventCount := 0

					for e := range watchCh {
						assert.Equal(t, event.Modified, e.Type)
						eventCount++

						if eventCount >= expectedEventCount {
							wg.Done()
							return
						}
					}
				}(tt.updateEventNum, watchCh, allWaitGroup)
			}

			// update nodes
			for i := 0; i < tt.clientNum; i++ {
				go func(expectedEventCount int, nodes []*types.Node, clientId string) {
					for j := 0; j < expectedEventCount/len(nodes)+2; j++ {
						updateNodeEvents := make([]*event.NodeEvent, len(nodes))
						for k := 0; k < len(nodes); k++ {
							rvToGenerate += 1
							updateNodeEvents[k] = event.NewNodeEvent(
								types.NewNode(nodes[k].GetId(), strconv.Itoa(rvToGenerate), "", nodes[k].GetLocation()),
								event.Modified)
						}
						result, rvMap := distributor.ProcessEvents(updateNodeEvents)
						assert.True(t, result)
						assert.NotNil(t, rvMap)
						//fmt.Printf("Successfully processed %d update node events. RV map returned: %v. ClientId %s\n", len(nodes), rvMap, clientId)
					}
				}(tt.updateEventNum, nodesByClient[i], clientIds[i])
			}

			// wait for watch done
			allWaitGroup.Wait()
			duration += time.Since(start)
			fmt.Printf("Test %s succeed! Total duration %v\n", tt.name, duration)
		})
	}
}
