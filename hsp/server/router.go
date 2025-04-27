package server

import (
	"log"

	"github.com/LandaMm/hsp-go/hsp"
)

type RouteHandler func(req *hsp.Request) *hsp.Response

type Router struct {
	routes map[string]RouteHandler
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

func (r *Router) Handle(conn *hsp.Connection) error {
	defer conn.Close()

	packet, err := conn.Read()
	if err != nil {
		_, _ = conn.Write(hsp.NewErrorResponse(err).ToPacket())
		return err
	}

	if route, ok := packet.Headers["route"]; ok {
		req := hsp.NewRequest(conn, packet)

		if handler, ok := r.routes[route]; ok {
			res := handler(req)
			_, err := conn.Write(res.ToPacket())
			return err
		} else if fallback, ok := r.routes["*"]; ok {
			res := fallback(req)
			_, err := conn.Write(res.ToPacket())
			return err
		}
	}

	_, err = conn.Write(hsp.NewStatusResponse(hsp.STATUS_NOTFOUND).ToPacket())
	return err
}
