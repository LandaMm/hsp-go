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

func (c *Client) SendText(address, text string) (*hsp.Response, error) {
	addr, err := hsp.ParseAddress(address)

	if err != nil {
		return nil, err
	}

	df := hsp.DataFormat{
		Format: hsp.DF_TEXT,
		Encoding: hsp.E_UTF8,
	}

	headers := make(map[string]string)
	headers[hsp.H_ROUTE] = addr.Route
	headers[hsp.H_DATA_FORMAT] = df.String()

	payload := []byte(text)

	pkt := hsp.BuildPacket(headers, payload)

	conn, err := net.Dial("tcp", addr.String())
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

