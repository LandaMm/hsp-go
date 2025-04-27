package server

import (
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/LandaMm/hsp-go/hsp"
)

type Server struct {
	Addr        hsp.Adddress
	routePrefix string // TODO: Support route prefix, e.g listening on localhost/api
	Running     bool
	ConnChan    chan *hsp.Connection
	listener    net.Listener
	mu          sync.Mutex
}

func NewServer(addr hsp.Adddress) *Server {
	return &Server{
		Addr:        addr,
		routePrefix: addr.Route,
		Running:     false,
	}
}

func (s *Server) SetListener(ln chan *hsp.Connection) {
	s.ConnChan = ln
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.Addr.String())
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.listener = ln
	s.Running = true
	s.mu.Unlock()

	for s.IsRunning() {
		conn, err := ln.Accept()
		if err != nil {
			if !s.IsRunning() {
				break
			}
			return err
		}

		keys, err := hsp.GenerateKeyPair()
		if err != nil {
			return err
		}

		// Receive client's public key
		clientKey := make([]byte, 32)
		n, err := io.ReadFull(conn, clientKey)
		if err != nil {
			return err
		}

		if n != 32 {
			return fmt.Errorf("Received invalid client's key with %d bytes (expected 32 bytes)", n)
		}

		// Send our public key to client
		n, err = conn.Write(keys.Public[:])
		if err != nil {
			return err
		}

		if n != 32 {
			return fmt.Errorf("Couldn't send 32 bytes of public key (%d sent instead)", n)
		}

		sharedKey, err := hsp.DeriveSharedKey(keys.Private, [32]byte(clientKey))
		if err != nil {
			return err
		}

		if s.ConnChan != nil {
			connection := hsp.NewConnection(conn, keys, sharedKey)
			s.ConnChan <- connection
		} else {
			conn.Close()
		}
	}

	s.mu.Lock()
	s.Running = false
	s.listener = nil
	s.mu.Unlock()

	return nil
}

func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Running = false
	if s.listener != nil {
		return s.listener.Close()
	}
	if s.ConnChan != nil {
		close(s.ConnChan)
	}

	return nil
}

func (s *Server) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Running
}
