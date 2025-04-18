package hsp

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"maps"
	"strconv"
)

type Response struct {
	StatusCode int
	Format DataFormat
	Headers map[string]string
	Payload []byte
}

func NewPacketResponse(packet *Packet) *Response {
	status, sok := packet.Headers[H_STATUS]
	if !sok {
		panic(errors.New("Packet must contain status header for response"))
	}

	format, fok := packet.Headers[H_DATA_FORMAT]
	if !fok {
		panic(errors.New("Packet must contain data format header for response"))
	}

	s, err := strconv.Atoi(status)
	if err != nil {
		panic(errors.New(fmt.Sprintf("Packet's status code is invalid: %s", err.Error())))
	}

	df, err := ParseDataFormat(format)
	if err != nil {
		panic(errors.New(fmt.Sprintf("Failed to parse packet's data format: %s", err.Error())))
	}

	return &Response{
		StatusCode: s,
		Format: *df,
		Headers: packet.Headers,
		Payload: packet.Payload,
	}
}

func NewStatusResponse(status int) *Response {
	return &Response{
		StatusCode: status,
		Headers: make(map[string]string),
		Format: DataFormat{
			Format: DF_BYTES,
			Encoding: "",
		},
		Payload: make([]byte, 0),
	}
}

func NewTextResponse(text string) *Response {
	return &Response{
		StatusCode: STATUS_SUCCESS,
		Headers: make(map[string]string),
		Format: DataFormat{
			Format: DF_TEXT,
			Encoding: E_UTF8,
		},
		Payload: []byte(text),
	}
}

func NewJsonResponse(data map[string]string) (*Response, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return &Response{
		StatusCode: STATUS_SUCCESS,
		Headers: make(map[string]string),
		Format: DataFormat{
			Format: DF_JSON,
			Encoding: E_UTF8,
		},
		Payload: jsonBytes,
	}, nil
}

func (res *Response) ToPacket() *Packet {
	headers := make(map[string]string)

	maps.Copy(headers, res.Headers)

	headers[H_DATA_FORMAT] = fmt.Sprintf("%s:%s", res.Format.Format, res.Format.Encoding)
	headers[H_STATUS] = strconv.Itoa(res.StatusCode)

	return BuildPacket(headers, res.Payload)
}

func (res *Response) AddHeader(key, value string) {
	if _, ok := res.Headers[key]; ok {
		log.Printf("WARN: Rewriting already existing header: '%s'\n", key)
	}
	res.Headers[key] = value
}

func (res *Response) Write(p []byte) (int, error) {
	buf := new(bytes.Buffer)

	n, err := buf.Write(p)
	if err != nil {
		return n, err
	}

	res.Payload = buf.Bytes()

	return n, err
}

