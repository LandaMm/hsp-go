package hsp

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"slices"
	"strconv"
	"strings"
)

type Request struct {
	conn   net.Conn
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

func (req *Request) GetRawPacket() *Packet {
	return req.packet
}

func (req *Request) GetDataFormat() (*DataFormat, error) {
	// TODO: use predefined header names
	format, ok := req.packet.Headers["data-format"]
	if !ok {
		return nil, errors.New("Data format header is not provided in request")
	}

	return ParseDataFormat(format)
}

func (req *Request) GetRequestKind() string {
	_, ok := req.GetHeader(H_XSTREAM)
	if ok {
		return "stream"
	}

	return "single-hit"
}

func (req *Request) GetStreamInfo() (*StreamInfo, error) {
	stream, ok := req.GetHeader(H_XSTREAM)
	if !ok {
		return nil, errors.New("No X-STREAM header presented in request")
	}

	parts := strings.Split(stream, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("Invalid value of X-STREAM header: '%s'", stream)
	}

	totalS, bufsizeS := parts[0], parts[1]

	total, err := strconv.ParseUint(totalS, 10, 64)
	if err != nil {
		return nil, err
	}

	bufsize, err := strconv.ParseUint(bufsizeS, 10, 16)
	if err != nil {
		return nil, err
	}
	buf := uint16(bufsize)

	return &StreamInfo{
		TotalBytes: total,
		BufferSize: buf,
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

func (req *Request) ExtractJson(out any) error {
	df, err := req.GetDataFormat()
	if err != nil {
		return err
	}

	if !slices.Contains([]string{DF_JSON}, df.Format) {
		return errors.New(fmt.Sprintf("Data format '%s' cannot be extracted as json", df.Format))
	}

	return json.Unmarshal(req.packet.Payload, out)
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
