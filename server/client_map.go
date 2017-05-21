package server

import (
	"math"
	"sync"
)

// RegionArea is the area (in coordinate degrees)
const RegionArea = 1.0

// Client is the representation of chat clients
type Client struct {
	ID string
	Lat,
	Long float64
}

//ClientNetwork is the representation of the Network of all client regions
type ClientNetwork struct {
	root            *clientRegion
	allRegions      []*clientRegion
	latRange        [2]float64
	longRange       [2]float64
	modificationMux *sync.Mutex
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
		if region.lat == lat && region.long == long {
			region.AddClient(client)
			break
		}
		if (region.lat-lat) == RegionArea && (region.long-long) == 0 {
			possibleRegionConnects[0] = region
			connected = true
		}
		if (region.lat-lat) == -RegionArea && (region.long-long) == 0 {
			possibleRegionConnects[3] = region
			connected = true
		}
		if (region.lat-lat) == 0 && (region.long-long) == -RegionArea {
			possibleRegionConnects[1] = region
			connected = true
		}
		if (region.lat-lat) == 0 && (region.long-long) == RegionArea {
			possibleRegionConnects[3] = region
			connected = true
		}
	}

	if !connected {
		return
	}

	newRegion := newClientRegion()
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
	return
}

// NewClientNetwork creates a new network of client regions
func NewClientNetwork(root *clientRegion) (network *ClientNetwork) {
	network.root = root
	network.allRegions = []*clientRegion{root}

	network.latRange = [2]float64{root.lat, root.lat + 1}
	network.longRange = [2]float64{root.long, root.long + 1}

	network.modificationMux = new(sync.Mutex)
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
	lat,
	long float64
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
	if c.lat == lat && c.long == long {
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

func newClientRegion() (region *clientRegion) {
	region.clients = make(map[string]*Client)
	return
}
