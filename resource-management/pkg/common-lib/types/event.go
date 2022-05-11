package types

const (
	Event_AddNode    = "Add Node"
	Event_UpdateNode = "Update Node"
	Event_DeleteNode = "Delete Node"
	Event_Bookmark   = "Bookmark"
)

type NodeEvent struct {
	eventType string
	node      *Node
}

func NewNodeEvent(node *Node, eventType string) *NodeEvent {
	return &NodeEvent{
		eventType: eventType,
		node:      node,
	}
}

func (e *NodeEvent) GetNode() *Node {
	return e.node
}

func (e *NodeEvent) GetEventType() string {
	return e.eventType
}
