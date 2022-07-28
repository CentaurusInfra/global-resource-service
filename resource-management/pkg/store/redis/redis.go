package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"global-resource-service/resource-management/pkg/common-lib/serializer"
	"global-resource-service/resource-management/pkg/common-lib/serializer/protobuf"
	"time"

	"k8s.io/klog/v2"

	"global-resource-service/resource-management/pkg/common-lib/interfaces/store"
	"global-resource-service/resource-management/pkg/common-lib/types"

	"github.com/go-redis/redis/v8"
)

type Goredis struct {
	client     *redis.Client
	ctx        context.Context
	serializer serializer.Serializer
}

// Initialize Redis Client
// TODO: with configured parameters for the store
func NewRedisClient() *Goredis {
	client := redis.NewClient(&redis.Options{
		Addr:         "localhost:6379",
		PoolSize:     1000,
		PoolTimeout:  2 * time.Minute,
		IdleTimeout:  10 * time.Minute,
		ReadTimeout:  2 * time.Minute,
		WriteTimeout: 1 * time.Minute,
		Password:     "", //no password set
		DB:           0,  // use default DB
	})

	ctx := context.Background()

	return &Goredis{
		client:     client,
		ctx:        ctx,
		serializer: protobuf.NewSerializer("foo"),
	}
}

// To Test Persist Simple String
//
func (gr *Goredis) setString(myKey, myValue string) bool {
	if len(myKey) == 0 || len(myValue) == 0 {
		klog.Errorf("The Key or Value is blank")
		return false
	}

	err := gr.client.Set(gr.ctx, myKey, []byte(myValue), 0).Err()

	if err != nil {
		klog.Errorf("Error to persist String key and Value to Redis Store. error %v", err)
		return false
	}

	return true
}

// To Test Get Simple String
//
func (gr *Goredis) getString(myKey string) string {
	var myValue string

	if len(myKey) == 0 {
		klog.Errorf("The Key is blank")
		return ""
	}

	value, err := gr.client.Get(gr.ctx, myKey).Bytes()

	if err != nil {
		klog.Errorf("Error to get String Key from Redis Store. error %v", err)
		return ""
	}

	if err != redis.Nil {
		myValue = string(value)
	}

	return myValue
}

// Use Redis data type - Set to store Logical Nodes
// One key has one record
//
// Note: Need re-visit these codes to see whether using function pointer is much better
//
// TODO: Error handling for loop persistence failure in the middle
//
func (gr *Goredis) PersistNodes(logicalNodes []*types.LogicalNode) bool {
	if logicalNodes == nil {
		klog.Errorf("The array of Logical Nodes is nil")
		return false
	}

	for _, logicalNode := range logicalNodes {
		logicalNodeKey := logicalNode.GetKey()
		logicalNodeBytes, err := gr.serializer.Marshal(logicalNode)

		if err != nil {
			klog.Errorf("Error from JSON Marshal for Logical Nodes. error %v", err)
			return false
		}

		err = gr.client.Set(gr.ctx, logicalNodeKey, logicalNodeBytes, 0).Err()

		if err != nil {
			klog.Errorf("Error to persist Logical Nodes to Redis Store. error %v", err)
			return false
		}
	}

	return true
}

// Use Redis data type - String to store Node Store Status
//
func (gr *Goredis) PersistNodeStoreStatus(nodeStoreStatus *store.NodeStoreStatus) bool {
	nodeStoreStatusBytes, err := json.Marshal(nodeStoreStatus)

	if err != nil {
		klog.Errorf("Error from JSON Marshal for Node Store Status. error %v", err)
		return false
	}

	err = gr.client.Set(gr.ctx, nodeStoreStatus.GetKey(), nodeStoreStatusBytes, 0).Err()

	if err != nil {
		klog.Errorf("Error to persist Node Store Status to Redis Store. error %v", err)
		return false
	}

	return true
}

// Use Redis data type - String to store Virtual Node Assignment
//
func (gr *Goredis) PersistVirtualNodesAssignments(virtualNodeAssignment *store.VirtualNodeAssignment) bool {
	virtualNodeAssignmentBytes, err := json.Marshal(virtualNodeAssignment)

	if err != nil {
		klog.Errorf("Error from JSON Marshal for Virtual Node Assignment:", err)
		return false
	}

	err = gr.client.Set(gr.ctx, virtualNodeAssignment.GetKey(), virtualNodeAssignmentBytes, 0).Err()

	if err != nil {
		klog.Errorf("Error to persist Virtual Node Assignment to Redis Store. error %v", err)
		return false
	}

	return true
}

// Get all Logical Nodes based on PreserveNode_KeyPrefix = "MinNode"
//
// Note: Need re-visit these codes to see whether using function pointer is much better
//
func (gr *Goredis) GetNodes() []*types.LogicalNode {
	keys := gr.client.Keys(gr.ctx, types.PreserveNode_KeyPrefix+"*").Val()

	logicalNodes := make([]*types.LogicalNode, len(keys))

	var logicalNode *types.LogicalNode

	for i, logicalNodeKey := range keys {
		value, err := gr.client.Get(gr.ctx, logicalNodeKey).Bytes()

		if err != nil {
			klog.Errorf("Error to get LogicalNode from Redis Store. error %v", err)
			return nil
		}

		if err != redis.Nil {
			_, err = gr.serializer.Decode(value, &logicalNode)

			if err != nil {
				klog.Errorf("Error from JSON Unmarshal for LogicalNode. error %v", err)
				return nil
			}

			logicalNodes[i] = logicalNode

		}
	}

	return logicalNodes
}

// Get Node Store Status
//
func (gr *Goredis) GetNodeStoreStatus() *store.NodeStoreStatus {
	var nodeStoreStatus *store.NodeStoreStatus

	value, err := gr.client.Get(gr.ctx, nodeStoreStatus.GetKey()).Bytes()

	if err != nil {
		klog.Errorf("Error to get NodeStoreStatus from Redis Store. error %v", err)
		return nil
	}

	if err != redis.Nil {
		err = json.Unmarshal(value, &nodeStoreStatus)

		if err != nil {
			klog.Errorf("Error from JSON Unmarshal for NodeStoreStatus. error %v", err)
			return nil
		}
	}

	return nodeStoreStatus
}

// Get Virtual Nodes Assignments
//
func (gr *Goredis) GetVirtualNodesAssignments() *store.VirtualNodeAssignment {
	var virtualNodeAssignment *store.VirtualNodeAssignment

	value, err := gr.client.Get(gr.ctx, virtualNodeAssignment.GetKey()).Bytes()

	if err != nil {
		klog.Errorf("Error to get VirtualNodeAssignment from Redis Store. error %v", err)
		return nil
	}

	if err != redis.Nil {
		err = json.Unmarshal(value, &virtualNodeAssignment)

		if err != nil {
			klog.Errorf("Error from JSON Unmarshal for VirtualNodeAssignment. error %v", err)
			return nil
		}
	}

	return virtualNodeAssignment
}

func (gr *Goredis) PersistClient(clientId string, client *types.Client) error {
	ci, err := json.Marshal(client)

	if err != nil {
		klog.Errorf("Error marshalling client. error %v", err)
		return err
	}

	err = gr.client.Set(gr.ctx, clientId, ci, 0).Err()

	if err != nil {
		klog.Errorf("Error persisting client to Redis Store. error %v", err)
		return err
	}

	return nil
}

func (gr *Goredis) GetClient(clientId string) (*types.Client, error) {
	ci := &types.Client{}

	value, err := gr.client.Get(gr.ctx, clientId).Bytes()

	if err != nil {
		klog.Errorf("Error getting client from Redis Store. error %v", err)
		return nil, err
	}

	if err == redis.Nil {
		klog.Errorf("Client %s, is not found in store", clientId)
		return nil, fmt.Errorf("client not found")
	}

	err = json.Unmarshal(value, ci)

	if err != nil {
		klog.Errorf("Error unmarshal client type. error %v", err)
		return nil, err
	}

	return ci, nil
}

func (gr *Goredis) UpdateClient(clientId string, client *types.Client) error {
	return fmt.Errorf("not implemented")
}

func (gr *Goredis) GetClients() ([]*types.Client, error) {
	return nil, fmt.Errorf("not implemented")
}
