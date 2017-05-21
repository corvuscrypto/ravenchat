package server

import "testing"

func TestNetworkSearch(T *testing.T) {
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
