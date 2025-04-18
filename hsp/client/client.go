package client

import (
	"net"
	"strings"

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

func (c *Client) SendText(address, text string) (*hsp.Response, error) {
	parts := strings.SplitN(address, "/", 1)
	
	var route string
	if len(parts) == 1 {
		route = "/"
	} else {
		route = "/" + strings.Join(parts[1:], "")
	}

	headers := make(map[string]string)
	headers[hsp.H_ROUTE] = route
	headers[hsp.H_DATA_FORMAT] = hsp.DF_TEXT

	payload := []byte(text)

	pkt := hsp.BuildPacket(headers, payload)

	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	c.Duplex = hsp.NewPacketDuplex(conn)
	if _, err := c.Duplex.WritePacket(pkt); err != nil {
		return nil, err
	}

	pkt, err = c.Duplex.ReadPacket()
	if err != nil {
		return nil, err
	}

	return hsp.NewPacketResponse(pkt), nil
}

