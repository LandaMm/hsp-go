package server

import (
	"errors"
	"fmt"
	"net"
	"slices"
	"strings"
)

type Request struct {
	conn net.Conn
	packet *Packet
}

func NewRequest(conn net.Conn, packet *Packet) *Request {
	return &Request{
		conn, packet,
	}
}

func (req *Request) Conn() net.Conn {
	return req.conn
}

func (req *Request) GetHeader(key string) (string, bool) {
	value, ok := req.packet.Headers[key]
	return value, ok
}

func (req *Request) GetDataFormat() (*DataFormat, error) {
	// TODO: use predefined header names
	format, ok := req.packet.Headers["data-format"]
	if !ok {
		return nil, errors.New("Data format header is not provided in request")
	}

	parts := strings.Split(format, ":")
	if len(parts) != 2 {
		if format == "bytes" {
			return &DataFormat{
				Format: DF_BYTES,
			}, nil
		}
		return nil, errors.New("Invalid data format header")
	}

	f, ok := DATA_FORMATS[parts[0]]
	if !ok {
		return nil, errors.New(fmt.Sprintf("Unknown data format: %s", parts[0]))
	}

	encoding, ok := ENCODINGS[parts[1]]
	if !ok {
		return nil, errors.New(fmt.Sprintf("Unknown payload encoding: %s", parts[1]))
	}

	return &DataFormat{
		Format: f,
		Encoding: encoding,
	}, nil
}

func (req *Request) ExtractText() (string, error) {
	df, err := req.GetDataFormat()
	if err != nil {
		return "", err
	}

	if !slices.Contains([]string{DF_TEXT, DF_JSON}, df.Format) {
		return "", errors.New(fmt.Sprintf("Data format '%s' cannot be extracted as text", df.Format))
	}

	return string(req.packet.Payload), nil
}

func (req *Request) ExtractBytes() ([]byte, error) {
	df, err := req.GetDataFormat()
	if err != nil {
		return nil, err
	}

	if df.Format != "bytes" {
		return nil, errors.New(fmt.Sprintf("Data format '%s' is invalid for extracting bytes", df.Format))
	}

	return req.packet.Payload, nil
}

