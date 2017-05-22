package server

import (
	"math"
	"time"
)

// RegionArea is the area (in coordinate degrees)
const RegionArea = 1.0

// EventBufferSize limits the number of events that each network can hold.
// Theoretically in the maximum load scenario (with (180*180) networks) our
// max buffered events will be (180*180*256) = ~8.3 Million events which amounts
// to 67.2 Million Bytes at a minimum. This is reasonably manageable even in an unlikely case
const EventBufferSize = 1 << 8

// Client is the representation of chat clients
type Client struct {
	ID string
	Lat,
	Long float64
}

// ClientWorld represents the entirety of server-managed networks and clients.
// It is the first line in event handling
type ClientWorld struct {
	Networks  []*ClientNetwork
	eventChan chan NetworkEvent
}

func (c *ClientWorld) mergeNetworks(networks []*ClientNetwork, lat, long float64) {
	var rootNetRegion = networks[0].root.findClientRegion(lat, long)
	for _, network := range networks[1:] {
		region := network.root.findClientRegion(lat, long)
		// merge the clients to the new root region
		for id, client := range region.clients {
			rootNetRegion.clients[id] = client
		}

		// Now splice out the node
		if region.Up != nil {
			rootNetRegion.Up = region.Up
			region.Up.Down = rootNetRegion
			region.Up = nil
		}
		if region.Left != nil {
			rootNetRegion.Left = region.Left
			region.Left.Right = rootNetRegion
			region.Left = nil
		}
		if region.Down != nil {
			rootNetRegion.Down = region.Down
			region.Down.Up = rootNetRegion
			region.Down = nil
		}
		if region.Right != nil {
			rootNetRegion.Right = region.Right
			region.Right.Left = rootNetRegion
			region.Right = nil
		}
	}
}

func (c *ClientWorld) handleClientConnect(client *Client) {
	//Code to handle client connect
	var connectedNetworks []*ClientNetwork
	for _, network := range c.Networks {
		if network.possiblyContains(client) {
			if network.AddClient(client) {
				connectedNetworks = append(connectedNetworks, network)
			}
		}
	}
	if len(connectedNetworks) > 1 {
		c.mergeNetworks(connectedNetworks, math.Floor(client.Lat), math.Floor(client.Long))
	}
}

func (c *ClientWorld) handleClientDisconnect(clientID string) {
	//Code to handle client disconnect
}

// SendEvent is a convenience function
func (c *ClientWorld) SendEvent(event NetworkEvent) {
	c.eventChan <- event
}

func (c *ClientWorld) waitForEvents() {
	for {
		event := <-c.eventChan
		switch t := event.Type(); t {
		case EventClientConnect:
			c.handleClientConnect(event.(ClientConnectionEvent).client)
		case EventClientDisconnect:
			c.handleClientDisconnect(event.(ClientDisconnectionEvent).clientID)
		}
	}
}

//NewClient creates a new client and returns its reference
func NewClient(lat, long float64) (client *Client) {
	client = new(Client)
	client.Lat = lat
	client.Long = long
	return
}

//ClientNetwork is the representation of the Network of all client regions
type ClientNetwork struct {
	root        *clientRegion
	allRegions  []*clientRegion
	latRange    [2]float64
	longRange   [2]float64
	messageChan chan ClientMessageEvent
}

func (c *ClientNetwork) possiblyContains(client *Client) bool {
	return (client.Lat >= c.latRange[0] && client.Lat < c.latRange[1] &&
		client.Long >= c.longRange[0] && client.Long < c.longRange[1])
}

// To avoid the use of mux's use a channel to shuttle events into the network
// This way we can perform operations without having to use locks
func (c *ClientNetwork) waitForMessages() {
	for {
		event := <-c.messageChan
		var aggregateMessages = []ClientMessageEvent{event}
		var done bool
		for !done {
			select {
			case e := <-c.messageChan:
				aggregateMessages = append(aggregateMessages, e)
			case <-time.Tick(1 * time.Microsecond):
				done = true
			}
		}
	}
}

// AddClient adds a client to the network in the appropriate region
func (c *ClientNetwork) AddClient(client *Client) (connected bool) {
	lat := math.Floor(client.Lat)
	long := math.Floor(client.Long)
	// we want to also track possible connecting regions in the case that
	// we need to add a region. Key as follows:
	// 0: Up
	// 1: Left
	// 2: Down
	// 3: Right
	var possibleRegionConnects = [4]*clientRegion{nil, nil, nil, nil}
	for _, region := range c.allRegions {
		if region.Lat == lat && region.Long == long {
			region.AddClient(client)
			connected = true
			return
		}
		if (region.Lat-lat) == RegionArea && (region.Long-long) == 0 {
			possibleRegionConnects[0] = region
			connected = true
		}
		if (region.Lat-lat) == -RegionArea && (region.Long-long) == 0 {
			possibleRegionConnects[3] = region
			connected = true
		}
		if (region.Lat-lat) == 0 && (region.Long-long) == -RegionArea {
			possibleRegionConnects[1] = region
			connected = true
		}
		if (region.Lat-lat) == 0 && (region.Long-long) == RegionArea {
			possibleRegionConnects[3] = region
			connected = true
		}
	}

	if !connected {
		return
	}

	newRegion := newClientRegion(client.Lat, client.Long)
	newRegion.AddClient(client)

	for i, r := range possibleRegionConnects {
		if r == nil {
			continue
		}
		switch i {
		case 0:
			newRegion.Up = r
			r.Down = newRegion
		case 1:
			newRegion.Left = r
			r.Right = newRegion
		case 2:
			newRegion.Down = r
			r.Up = newRegion
		case 3:
			newRegion.Right = r
			r.Left = newRegion
		}
	}

	// Append the region to the region array
	c.allRegions = append(c.allRegions, newRegion)

	// Also update our rectangular boundary
	if newRegion.Lat == c.latRange[1] {
		c.latRange[1]++
	}
	if newRegion.Lat < c.latRange[0] {
		c.latRange[0]--
	}
	if newRegion.Long == c.longRange[1] {
		c.longRange[1]++
	}
	if newRegion.Long == c.longRange[0] {
		c.longRange[0]--
	}
	return
}

// NewClientNetwork creates a new network of client regions
func NewClientNetwork(root *clientRegion) (network *ClientNetwork) {
	network = new(ClientNetwork)
	network.root = root
	network.allRegions = []*clientRegion{root}

	network.latRange = [2]float64{root.Lat, root.Lat + RegionArea}
	network.longRange = [2]float64{root.Long, root.Long + RegionArea}

	network.messageChan = make(chan ClientMessageEvent, EventBufferSize)

	go network.waitForMessages()

	return
}

type clientRegion struct {
	Up,
	Left,
	Down,
	Right *clientRegion
	clients map[string]*Client
	isRoot,
	visited bool
	Lat,
	Long float64
}

func (c *clientRegion) isConnectedToRoot(previousConnection bool) bool {
	if c.visited {
		return previousConnection
	}
	c.visited = true
	if c.isRoot {
		return true
	}
	// Graph search order is Up Left Down Right
	if c.Up != nil && c.Up.isConnectedToRoot(previousConnection) {
		return true
	}
	if c.Left != nil && c.Left.isConnectedToRoot(previousConnection) {
		return true
	}
	if c.Down != nil && c.Down.isConnectedToRoot(previousConnection) {
		return true
	}
	if c.Right != nil && c.Right.isConnectedToRoot(previousConnection) {
		return true
	}
	return false
}

func (c *clientRegion) findClientRegion(lat, long float64) *clientRegion {
	if c.visited {
		return nil
	}
	c.visited = true
	if c.Lat == lat && c.Long == long {
		return c
	}
	// Graph search order is Up Left Down Right
	if c.Up != nil {
		if n := c.Up.findClientRegion(lat, long); n != nil {
			return n
		}
	}
	if c.Left != nil {
		if n := c.Left.findClientRegion(lat, long); n != nil {
			return n
		}
	}
	if c.Down != nil {
		if n := c.Down.findClientRegion(lat, long); n != nil {
			return n
		}
	}
	if c.Right != nil {
		if n := c.Right.findClientRegion(lat, long); n != nil {
			return n
		}
	}
	return nil
}

//AddClient adds a client to a region. In the case that we already have the client in the region we ignore.
func (c *clientRegion) AddClient(client *Client) {
	if _, ok := c.clients[client.ID]; ok {
		return
	}
	c.clients[client.ID] = client
}

func newClientRegion(lat, long float64) (region *clientRegion) {
	region = new(clientRegion)
	region.clients = make(map[string]*Client)
	region.Lat = lat
	region.Long = long
	return
}
