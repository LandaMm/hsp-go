package client

import (
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net"

	"github.com/LandaMm/hsp-go/hsp"
)

type Client struct {
	Headers map[string]string
}

func NewClient(headers map[string]string) *Client {
	return &Client{
		Headers: headers,
	}
}

func (c *Client) BuildHeaders(address *hsp.Adddress, df *hsp.DataFormat) map[string]string {
	headers := make(map[string]string)

	if len(c.Headers) > 0 {
		maps.Copy(headers, c.Headers)
	}

	headers[hsp.H_ROUTE] = address.Route
	headers[hsp.H_DATA_FORMAT] = df.String()

	return headers
}

func (c *Client) SingleHit(addr *hsp.Adddress, pkt *hsp.Packet) (*hsp.Packet, error) {
	rawConn, err := net.Dial("tcp", addr.String())
	if err != nil {
		return nil, err
	}

	defer rawConn.Close()

	keys, err := hsp.GenerateKeyPair()
	if err != nil {
		return nil, err
	}

	n, err := rawConn.Write(keys.Public[:])
	if err != nil {
		return nil, err
	}

	if n != 32 {
		return nil, fmt.Errorf("failed to send 32 bytes of key (%d sent instead)", n)
	}

	serverKey := make([]byte, 32)
	if _, err := io.ReadFull(rawConn, serverKey); err != nil {
		return nil, err
	}

	sharedKey, err := hsp.DeriveSharedKey(keys.Private, [32]byte(serverKey))
	if err != nil {
		return nil, err
	}

	conn := hsp.NewConnection(rawConn, keys, sharedKey)

	if _, err := conn.Write(pkt); err != nil {
		return nil, err
	}

	return conn.Read()
}

func (c *Client) SendText(address, text string) (*hsp.Response, error) {
	addr, err := hsp.ParseAddress(address)

	if err != nil {
		return nil, err
	}

	payload := []byte(text)

	hdrs := c.BuildHeaders(addr, hsp.TextDataFormat())

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

	hdrs := c.BuildHeaders(addr, hsp.JsonDataFormat())

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

	hdrs := c.BuildHeaders(addr, hsp.BytesDataFormat())

	pkt := hsp.BuildPacket(hdrs, data)

	rpkt, err := c.SingleHit(addr, pkt)
	if err != nil {
		return nil, err
	}

	return hsp.NewPacketResponse(rpkt), nil
}
