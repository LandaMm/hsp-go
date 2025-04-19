package hsp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
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
	Header      []byte
	Payload     []byte
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

func NewPacketDuplex(conn net.Conn) *PacketDuplex {
	return &PacketDuplex{
		conn,
	}
}

func (r *PacketDuplex) ReadPacket() (*Packet, error) {
	rpkt := &RawPacket{}

	err := binary.Read(r.conn, binary.BigEndian, &rpkt.Magic)
	if err != nil {
		return nil, err
	}

	if rpkt.Magic != Magic {
		return nil, errors.New("Magic bytes are invalid")
	}

	err = binary.Read(r.conn, binary.BigEndian, &rpkt.Version)
	if err != nil {
		return nil, err
	}

	err = binary.Read(r.conn, binary.BigEndian, &rpkt.Flags)
	if err != nil {
		return nil, err
	}

	err = binary.Read(r.conn, binary.BigEndian, &rpkt.HeaderSize)
	if err != nil {
		return nil, err
	}

	err = binary.Read(r.conn, binary.BigEndian, &rpkt.PayloadSize)
	if err != nil {
		return nil, err
	}

	rpkt.Header = make([]byte, rpkt.HeaderSize)
	if _, err := io.ReadFull(r.conn, rpkt.Header); err != nil {
		return nil, err
	}

	rpkt.Payload = make([]byte, rpkt.PayloadSize)
	if _, err := io.ReadFull(r.conn, rpkt.Payload); err != nil {
		return nil, err
	}

	pkt := &Packet{
		Version: int(rpkt.Version),
		Flags:   int(rpkt.Flags),
		Headers: make(map[string]string),
		Payload: rpkt.Payload,
	}

	ParseHeaders(rpkt.Header, &pkt.Headers)

	return pkt, nil
}

func (r *PacketDuplex) WritePacket(packet *Packet) (int, error) {
	buf := new(bytes.Buffer)

	if err := binary.Write(buf, binary.BigEndian, Magic); err != nil {
		return 0, errors.New(fmt.Sprintf("Failed to write magic into packet: %s", err.Error()))
	}

	if err := binary.Write(buf, binary.BigEndian, uint8(packet.Version)); err != nil {
		return 0, errors.New(fmt.Sprintf("Failed to write version into packet: %s", err.Error()))
	}

	if err := binary.Write(buf, binary.BigEndian, uint8(packet.Flags)); err != nil {
		return 0, errors.New(fmt.Sprintf("Failed to write flags into packet: %s", err.Error()))
	}

	rawHeaders := SerializeHeaders(&packet.Headers)
	headerSize := len(rawHeaders)
	payloadSize := len(packet.Payload)

	if err := binary.Write(buf, binary.BigEndian, uint16(headerSize)); err != nil {
		return 0, errors.New(fmt.Sprintf("Failed to write header size into packet: %s", err.Error()))
	}

	if err := binary.Write(buf, binary.BigEndian, uint32(payloadSize)); err != nil {
		return 0, errors.New(fmt.Sprintf("Failed to write payload size into packet: %s", err.Error()))
	}

	if _, err := buf.Write(rawHeaders[:headerSize]); err != nil {
		return 0, errors.New(fmt.Sprintf("Failed to write raw headers: %s", err.Error()))
	}

	if _, err := buf.Write(packet.Payload[:payloadSize]); err != nil {
		return 0, errors.New(fmt.Sprintf("Failed to write payload: %s", err.Error()))
	}

	n, err := r.conn.Write(buf.Bytes())
	if err != nil {
		return 0, errors.New(fmt.Sprintf("Failed to send packet over connection: %s", err.Error()))
	}

	return n, nil
}
