package server

import (
	"log"
	"net"
	"sync"

	"github.com/LandaMm/hsp-go/hsp"
)

type Server struct {
	Addr hsp.Adddress
	routePrefix string // TODO: Support route prefix, e.g listening on localhost/api
	Running bool
	ConnChan chan net.Conn
	listener net.Listener
	mu sync.Mutex
}

func NewServer(addr hsp.Adddress) *Server {
	return &Server{
		Addr: addr,
		routePrefix: addr.Route,
		Running: false,
	}
}

func (s *Server) SetListener(ln chan net.Conn) {
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
		log.Println("DEBUG:", "Waiting for new connection to accept")
		conn, err := ln.Accept()
		if err != nil {
			if !s.IsRunning() {
				break;
			}
			return err
		}

		if s.ConnChan != nil {
			s.ConnChan <- conn
		}
	}

	log.Println("DEBUG:", "Finished listening for connections")

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


