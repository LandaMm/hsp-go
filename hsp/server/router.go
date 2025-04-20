package server

import (
	"log"
	"net"

	"github.com/LandaMm/hsp-go/hsp"
)

type RouteHandler func(req *hsp.Request) *hsp.Response

type Router struct {
	routes    map[string]RouteHandler
}

func NewRouter() *Router {
	return &Router{
		routes: make(map[string]RouteHandler),
	}
}

func (r *Router) AddRoute(pathname string, handler RouteHandler) {
	if _, ok := r.routes[pathname]; ok {
		log.Printf("WARN: Rewriting existing route '%s'\n", pathname)
	}
	r.routes[pathname] = handler
}

func (r *Router) Handle(conn net.Conn) error {
	defer conn.Close()

	dupl := hsp.NewPacketDuplex(conn)

	// TODO: Ability to keep connection alive
	packet, err := dupl.ReadPacket()
	if err != nil {
		_, _ = dupl.WritePacket(hsp.NewErrorResponse(err).ToPacket())
		return err
	}

	if route, ok := packet.Headers["route"]; ok {
		req := hsp.NewRequest(conn, packet)

		if handler, ok := r.routes[route]; ok {
			res := handler(req)
			_, err := dupl.WritePacket(res.ToPacket())
			return err
		} else if fallback, ok := r.routes["*"]; ok {
			res := fallback(req)
			_, err := dupl.WritePacket(res.ToPacket())
			return err
		}
	}

	_, err = dupl.WritePacket(hsp.NewStatusResponse(hsp.STATUS_NOTFOUND).ToPacket())
	return err
}
