package server

// NetworkEventType is to distinguish the types of events for specific handling
// of said events
type NetworkEventType int

// Network event types
const (
	EventClientConnect NetworkEventType = iota
	EventClientDisconnect
	EventClientMessage
)

// NetworkEvent represents the data that is to be handled by a
// client network
type NetworkEvent interface {
	Type() NetworkEventType
}

// ClientConnectionEvent represents a client connection
type ClientConnectionEvent struct {
	client *Client
}

// Type satisfies the NetworkEvent interface
func (c ClientConnectionEvent) Type() NetworkEventType {
	return EventClientConnect
}

// ClientDisconnectionEvent represents a client connection
type ClientDisconnectionEvent struct {
	clientID string
}

// Type satisfies the NetworkEvent interface
func (c ClientDisconnectionEvent) Type() NetworkEventType {
	return EventClientDisconnect
}

// ClientMessageEvent represents a client connection
type ClientMessageEvent struct {
	ClientID,
	Topic,
	MessageID,
	Message string
}

// Type satisfies the NetworkEvent interface
func (c ClientMessageEvent) Type() NetworkEventType {
	return EventClientMessage
}
