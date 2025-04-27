package hsp

import (
	"bytes"
	"fmt"
	"net"
)

const (
	Magic uint32 = 0xDEADBEEF
)

const (
	PacketVersion int = 1
)

type RawPacket struct {
	Magic       uint32
	Version     uint8
	Flags       uint8
	HeaderSize  uint16
	PayloadSize uint32
	Nonce       []byte
	Header      []byte
	Payload     []byte
	Mac         []byte
}

type Packet struct {
	Version int
	Flags   int
	Headers map[string]string
	Payload []byte
}

type PacketDuplex struct {
	conn net.Conn
}

func BuildPacket(headers map[string]string, payload []byte) *Packet {
	return &Packet{
		Version: PacketVersion,
		Flags:   0, // TODO:
		Headers: headers,
		Payload: payload,
	}
}

func ParseHeaders(rawHeaders []byte, headers *map[string]string) error {
	i := 0
	for i < len(rawHeaders) {
		if rawHeaders[i] == '\n' {
			break
		}
		var key string
		for rawHeaders[i] != ':' {
			if rawHeaders[i] != ' ' {
				key += string(rawHeaders[i])
			}
			i++
		}
		i++
		var value string
		for rawHeaders[i] != '\n' {
			if rawHeaders[i] != ' ' {
				value += string(rawHeaders[i])
			}
			i++
		}
		i++
		(*headers)[key] = value
	}
	return nil
}

func SerializeHeaders(headers *map[string]string) []byte {
	buf := new(bytes.Buffer)
	for k, v := range *headers {
		fmt.Fprintf(buf, "%s:%s\n", k, v)
	}
	fmt.Fprintf(buf, "\n")
	return buf.Bytes()
}
