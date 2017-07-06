package server

import (
	"net"

	"github.com/corvuscrypto/ravenchat/network"
)

type Server struct {
	ID       string
	Listener net.Listener
	networks []*network.ClientNetwork
}

func (s *Server) newClient(client *network.Client) {

}
