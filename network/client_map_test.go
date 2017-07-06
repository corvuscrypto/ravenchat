package network

import "testing"

func TestNetworkSetup(T *testing.T) {
	// make a root region
	region1 := newClientRegion(40, 38)
	// Make a new network with root as region1
	network := NewClientNetwork(region1)
	client1 := NewClient(40, 38)
	network.AddClient(client1)

	// our region1 should now have only one client
	if numClients := len(region1.clients); numClients != 1 {
		T.Errorf("Unexpected number of clients in region. Expected %d, Got %d", 1, numClients)
	}
}

func TestAddClientNewRegion(T *testing.T) {
	// make a root region
	region1 := newClientRegion(40, 38)
	// Make a new network with root as region1
	network := NewClientNetwork(region1)
	client1 := NewClient(40, 38)
	client2 := NewClient(40, 39)
	network.AddClient(client1)
	network.AddClient(client2)
	foundRegion := network.root.findClientRegion(40, 39)

	//Ensure we found the right region
	if foundRegion == nil || !(foundRegion.Lat == 40 && foundRegion.Long == 39) {
		T.Errorf("Didn't find the expected region.")
	}

	//Ensure we have a client in that region
	if numClients := len(foundRegion.clients); numClients != 1 {
		T.Errorf("Unexpected number of clients in region. Expected %d, Got %d", 1, numClients)
	}

	//ensure we have 2 regions
	if numRegions := len(network.allRegions); numRegions != 2 {
		T.Errorf("Unexpected number of clients in region. Expected %d, Got %d", 2, numRegions)
	}

	//Ensure that the bounds have updated on the network instance
	if !(network.latRange == [2]float64{40, 41} && network.longRange == [2]float64{38, 40}) {
		T.Errorf("Unexpected rectangular boundary.")
	}
}

func TestNetworkMerge(T *testing.T) {
	world := new(ClientWorld)

	client1 := NewClient(38.8, 40.2)
	client1.ID = "A"

	client2 := NewClient(38.8, 40.4)
	client2.ID = "B"

	net1 := NewClientNetwork(newClientRegion(38, 40))
	net1.AddClient(client1)

	net2 := NewClientNetwork(newClientRegion(37, 40))
	net2.AddClient(client2)

	addedRegion := net2.root.findClientRegion(38, 40)

	// We have to manually mark all regions unvisited after the findClientRegion call
	net2.markAllRegionsUnvisited()

	world.mergeNetworks([]*ClientNetwork{net1, net2}, 38, 40)
	//Ensure the net2 region was spliced out and can be GC'd
	if addedRegion.Up != nil || addedRegion.Left != nil || addedRegion.Down != nil || addedRegion.Right != nil {
		T.Errorf("Secondary Network's region was not spliced out")
	}
	//Ensure the clients were merged into the first net
	if !(net1.root.clients["A"] == client1 && net1.root.clients["B"] == client2) {
		T.Errorf("Merge did not successfully occur")
	}
}
