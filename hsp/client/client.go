package client

import (
	"encoding/json"
	"maps"
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

func (c *Client) BuildHeaders(address *hsp.Adddress, df *hsp.DataFormat, mergeHeaders *map[string]string) map[string]string {
	headers := make(map[string]string)

	if mergeHeaders != nil {
		maps.Copy(headers, *mergeHeaders)
	}

	headers[hsp.H_ROUTE] = address.Route
	headers[hsp.H_DATA_FORMAT] = df.String()

	return headers
}

func (c *Client) SingleHit(addr *hsp.Adddress, pkt *hsp.Packet) (*hsp.Packet, error) {
	conn, err := net.Dial("tcp", addr.String())
	if err != nil {
		return nil, err
	}

	c.Duplex = hsp.NewPacketDuplex(conn)
	if _, err := c.Duplex.WritePacket(pkt); err != nil {
		return nil, err
	}

	return c.Duplex.ReadPacket()
}

func (c *Client) SendText(address, text string) (*hsp.Response, error) {
	addr, err := hsp.ParseAddress(address)

	if err != nil {
		return nil, err
	}

	payload := []byte(text)

	hdrs := c.BuildHeaders(addr, hsp.TextDataFormat(), nil)

	pkt := hsp.BuildPacket(hdrs, payload)

	rpkt, err := c.SingleHit(addr, pkt)
	if err != nil {
		return nil, err
	}

	return hsp.NewPacketResponse(rpkt), nil
}

func (c *Client) SendJson(address string, data any) (*hsp.Response, error) {
	addr, err := hsp.ParseAddress(address)

	if err != nil {
		return nil, err
	}

	payload, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	hdrs := c.BuildHeaders(addr, hsp.JsonDataFormat(), nil)

	pkt := hsp.BuildPacket(hdrs, payload)

	rpkt, err := c.SingleHit(addr, pkt)
	if err != nil {
		return nil, err
	}

	return hsp.NewPacketResponse(rpkt), nil
}

func (c *Client) SendBytes(address string, data []byte) (*hsp.Response, error) {
	addr, err := hsp.ParseAddress(address)

	if err != nil {
		return nil, err
	}

	hdrs := c.BuildHeaders(addr, hsp.BytesDataFormat(), nil)

	pkt := hsp.BuildPacket(hdrs, data)

	rpkt, err := c.SingleHit(addr, pkt)
	if err != nil {
		return nil, err
	}

	return hsp.NewPacketResponse(rpkt), nil
}
