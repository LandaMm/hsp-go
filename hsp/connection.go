package hsp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
)

type Connection struct {
	Conn      net.Conn
	Keys      *KeyPair
	SharedKey [32]byte
}

func NewConnection(conn net.Conn, keys *KeyPair, sharedKey [32]byte) *Connection {
	return &Connection{
		Conn:      conn,
		Keys:      keys,
		SharedKey: sharedKey,
	}
}

func (c *Connection) Close() error {
	return c.Conn.Close()
}

func (c *Connection) Read() (*Packet, error) {
	rpkt := &RawPacket{}

	err := binary.Read(c.Conn, binary.BigEndian, &rpkt.Magic)
	if err != nil {
		return nil, err
	}

	if rpkt.Magic != Magic {
		return nil, errors.New("Magic bytes are invalid")
	}

	err = binary.Read(c.Conn, binary.BigEndian, &rpkt.Version)
	if err != nil {
		return nil, err
	}

	err = binary.Read(c.Conn, binary.BigEndian, &rpkt.Flags)
	if err != nil {
		return nil, err
	}

	err = binary.Read(c.Conn, binary.BigEndian, &rpkt.HeaderSize)
	if err != nil {
		return nil, err
	}

	err = binary.Read(c.Conn, binary.BigEndian, &rpkt.PayloadSize)
	if err != nil {
		return nil, err
	}

	rpkt.Nonce = make([]byte, 12)
	if _, err := io.ReadFull(c.Conn, rpkt.Nonce); err != nil {
		return nil, err
	}

	data := make([]byte, uint32(rpkt.HeaderSize)+rpkt.PayloadSize)
	if _, err := io.ReadFull(c.Conn, data); err != nil {
		return nil, err
	}

	rpkt.Mac = make([]byte, 16)
	if _, err := io.ReadFull(c.Conn, rpkt.Mac); err != nil {
		return nil, err
	}

	decrypted, err := Decrypt(c.SharedKey[:], rpkt.Nonce, append(data, rpkt.Mac...))
	if err != nil {
		return nil, err
	}

	rpkt.Header = decrypted[:rpkt.HeaderSize]
	rpkt.Payload = decrypted[rpkt.HeaderSize : uint32(rpkt.HeaderSize)+rpkt.PayloadSize]

	pkt := &Packet{
		Version: int(rpkt.Version),
		Flags:   int(rpkt.Flags),
		Headers: make(map[string]string),
		Payload: rpkt.Payload,
	}

	ParseHeaders(rpkt.Header, &pkt.Headers)

	return pkt, nil
}

func (c *Connection) Write(packet *Packet) (n int, err error) {
	buf := new(bytes.Buffer)

	if err := binary.Write(buf, binary.BigEndian, Magic); err != nil {
		return 0, fmt.Errorf("failed to write magic into packet: %s", err.Error())
	}

	if err := binary.Write(buf, binary.BigEndian, uint8(packet.Version)); err != nil {
		return 0, fmt.Errorf("failed to write version into packet: %s", err.Error())
	}

	if err := binary.Write(buf, binary.BigEndian, uint8(packet.Flags)); err != nil {
		return 0, fmt.Errorf("failed to write flags into packet: %s", err.Error())
	}

	rawHeaders := SerializeHeaders(&packet.Headers)

	data := append(rawHeaders, packet.Payload...)

	encrypted, nonce, err := Encrypt(c.SharedKey[:], data)
	if err != nil {
		return 0, err
	}

	mac := encrypted[len(encrypted)-16:]

	headerSize := len(rawHeaders)
	payloadSize := len(packet.Payload)

	if err := binary.Write(buf, binary.BigEndian, uint16(headerSize)); err != nil {
		return 0, errors.New(fmt.Sprintf("Failed to write header size into packet: %s", err.Error()))
	}

	if err := binary.Write(buf, binary.BigEndian, uint32(payloadSize)); err != nil {
		return 0, errors.New(fmt.Sprintf("Failed to write payload size into packet: %s", err.Error()))
	}

	if _, err := buf.Write(nonce[:12]); err != nil {
		return 0, errors.New(fmt.Sprintf("Failed to write nonce: %s", err.Error()))
	}

	if _, err := buf.Write(encrypted[:len(encrypted)-16]); err != nil {
		return 0, errors.New(fmt.Sprintf("Failed to write encrypted data: %s", err.Error()))
	}

	if _, err := buf.Write(mac); err != nil {
		return 0, errors.New(fmt.Sprintf("Failed to write mac: %s", err.Error()))
	}

	n, err = c.Conn.Write(buf.Bytes())
	if err != nil {
		return 0, errors.New(fmt.Sprintf("Failed to send packet over connection: %s", err.Error()))
	}

	return n, nil
}
