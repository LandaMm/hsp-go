package server

import (
	"fmt"
	"log"
	"net"

	"github.com/LandaMm/hsp-go/hsp"
)

type RouteHandler func(req *hsp.Request) *hsp.Response
type StreamHandler func(req *hsp.Request, stream chan []byte)

type Router struct {
	routes    map[string]RouteHandler
	streamers map[string]StreamHandler
	streamMaxSize uint64
	streamBufferSize uint16
}

func NewRouter() *Router {
	return &Router{
		routes: make(map[string]RouteHandler),
		streamers: make(map[string]StreamHandler),
	}
}

func (r *Router) AddRoute(pathname string, handler RouteHandler) {
	if _, ok := r.routes[pathname]; ok {
		log.Printf("WARN: Rewriting existing route '%s'\n", pathname)
	}
	r.routes[pathname] = handler
}

func (r *Router) AddStreamer(pathname string, handler StreamHandler) {
	if _, ok := r.streamers[pathname]; ok {
		log.Printf("WARN: Rewriting existing streamer '%s'\n", pathname)
	}
	r.streamers[pathname] = handler
}

func (r *Router) SetStreamMaxSize(size uint64) {
	r.streamMaxSize = size
}

func (r *Router) SetStreamBufferSize(size uint16) {
	r.streamBufferSize = size
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

		switch req.GetRequestKind() {
		case "single-hit":
			if handler, ok := r.routes[route]; ok {
				res := handler(req)
				_, err := dupl.WritePacket(res.ToPacket())
				return err
			} else if fallback, ok := r.routes["*"]; ok {
				res := fallback(req)
				_, err := dupl.WritePacket(res.ToPacket())
				return err
			}
		case "stream":
			if handler, ok := r.streamers[route]; ok {
				info, err := req.GetStreamInfo()
				if err != nil {
					_, err = dupl.WritePacket(hsp.NewErrorResponse(err).ToPacket())
					return err
				}

				streamSize := uint64(min(info.TotalBytes, r.streamMaxSize))
				bufferSize := uint16(min(info.BufferSize, r.streamBufferSize))

				res := hsp.NewStatusResponse(hsp.STATUS_SUCCESS)
				res.AddHeader(hsp.H_XSTREAM, fmt.Sprintf("%d:%d", streamSize, bufferSize))
				res.AddHeader(hsp.H_XSTREAM_KEY, "0") // TODO: generate id

				_, err = dupl.WritePacket(res.ToPacket())
				if err != nil {
					return err
				}

				req := hsp.NewRequest(conn, res.ToPacket())
				bc := make(chan []byte)

				go func() {
					handler(req, bc)
				}()

				buf := make([]byte, bufferSize)
				var totalReceived uint64
				totalReceived = 0
				for totalReceived < streamSize {
					n, err := conn.Read(buf)
					if err != nil || n <= 0 {
						break
					}
					if n > 0 {
						totalReceived += uint64(n)
					}
				}

				res = hsp.NewStatusResponse(hsp.STATUS_SUCCESS)
				res.AddHeader(hsp.H_XSTREAM, fmt.Sprintf("%d:0", streamSize - totalReceived))
				res.AddHeader(hsp.H_XSTREAM_KEY, "0") // TODO: generate id
				_, err = dupl.WritePacket(res.ToPacket())

				conn.Close()
				close(bc)

				return err
			}
		default:
			return fmt.Errorf("Unsupported request kind: %s", req.GetRequestKind())
		}
	}

	_, err = dupl.WritePacket(hsp.NewStatusResponse(hsp.STATUS_NOTFOUND).ToPacket())
	return err
}
