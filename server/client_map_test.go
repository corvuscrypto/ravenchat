package server

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
}
