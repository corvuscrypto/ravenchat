package server

import "net"

type Server struct {
	ID       string
	Listener net.Listener
	networks []*ClientNetwork
}

func (s *Server) newClient(client *Client) {

}
