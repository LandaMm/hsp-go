package client

import (
	"net"

	"github.com/LandaMm/hsp-go/hsp"
)

type Client struct {
	Duplex *hsp.PacketDuplex
}

func NewClient() *Client {
	return &Client{
		Duplex: nil,
	}
}

func (c *Client) SendRequest(req *hsp.Request, address string) (*hsp.Response, error) {
	// TODO: Parse pathname
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	c.Duplex = hsp.NewPacketDuplex(conn)
}

