package server

import (
	"errors"
	"log"
	"net"

	"github.com/LandaMm/hsp-go/hsp"
)

type RouteHandler func(req *hsp.Request) *hsp.Response

type Router struct {
	Routes map[string]RouteHandler
}

func NewRouter() *Router {
	return &Router{
		Routes: make(map[string]RouteHandler),
	}
}

func (r *Router) AddRoute(pathname string, handler RouteHandler) {
	if _, ok := r.Routes[pathname]; ok {
		log.Printf("WARN: Rewriting existing route '%s'\n", pathname)
	}
	r.Routes[pathname] = handler
}

func (r *Router) Handle(conn net.Conn) error {
	defer conn.Close()

	log.Printf("Got new connection from %s\n", conn.RemoteAddr().String())

	dupl := hsp.NewPacketDuplex(conn)

	// TODO: Ability to keep connection alive
	packet, err := dupl.ReadPacket()
	if err != nil {
		_, _ = dupl.WritePacket(hsp.NewErrorResponse(err).ToPacket())
		return err
	}

	if route, ok := packet.Headers["route"]; ok {
		log.Printf("[ROUTER] New connection to '%s'", route)
		if handler, ok := r.Routes[route]; ok {
			req := hsp.NewRequest(conn, packet)
			res := handler(req)
			_, err := dupl.WritePacket(res.ToPacket())
			return err
		}
	}
	return errors.New("Not Found")
}
