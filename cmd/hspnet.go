package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/LandaMm/hsp-go/hsp"
	"github.com/LandaMm/hsp-go/hsp/client"
	"github.com/LandaMm/hsp-go/hsp/server"
	"github.com/chzyer/readline"
)

func Index(req *hsp.Request) *hsp.Response {
	pkt := req.GetRawPacket()

	fmt.Printf("%s %s\n", strings.ToUpper(req.GetRequestKind()), req.GetRoute())
	fmt.Println("Headers:")

	for k, v := range pkt.Headers {
		fmt.Printf("\t%s: %s\n", k, v)
	}

	df, err := req.GetDataFormat()
	if err != nil {
		h, ok := req.GetHeader(hsp.H_DATA_FORMAT)
		if !ok {
			fmt.Println("ERR: Data-Format header is not present")
			return hsp.NewErrorResponse(err)
		}
		fmt.Printf("Invalid Data Format: %s\n", h)
		return hsp.NewErrorResponse(err)
	}

	fmt.Printf("Payload (%s):\n", df.String())

	switch df.Format {
	case "text":
		body, err := req.ExtractText()
		if err != nil {
			fmt.Println("ERR: Failed to extract text from payload:", err)
			return hsp.NewErrorResponse(err)
		}

		fmt.Println(body)
		break
	case "json":
		var out map[string]any
		err := req.ExtractJson(&out)
		if err != nil {
			fmt.Println("ERR: Failed to extract json from payload:", err)
			return hsp.NewErrorResponse(err)
		}

		fmt.Println(json.MarshalIndent(out, "", "  "))
		break
	case "bytes":
		bts, err := req.ExtractBytes()
		if err != nil {
			fmt.Println("ERR: Failed to extract bytes from payload:", err)
			return hsp.NewErrorResponse(err)
		}

		fmt.Println(bts)
		break;
	default:
		fmt.Printf("ERR: Unsupported data format: %s\n", df.String())
	}

	return hsp.NewStatusResponse(hsp.STATUS_SUCCESS)
}

func StartServer(addr *hsp.Adddress) {
	srv := server.NewServer(*addr)
	fmt.Printf("Server created on address: %s\n", srv.Addr.String())

	handler := make(chan net.Conn, 1)

	router := server.NewRouter()

	router.AddRoute("*", Index)

	srv.SetListener(handler)

	go func() {
		for {
			conn := <-handler
			if err := router.Handle(conn); err != nil {
				fmt.Println("ERR: Couldn't handle connection:", err.Error())
			}
		}
	}()

	sigs := make(chan os.Signal, 1)

	go func() {
		s := <-sigs
		if s == syscall.SIGINT || s == syscall.SIGTERM {
			fmt.Println("Gracefully shutting down the server")
			if err := srv.Stop(); err != nil {
				fmt.Println("Failed to close the server:", err)
			}
		}
	}()

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	if err := srv.Start(); err != nil {
		fmt.Println("ERR: Failed to start server:", err)
	}
}

func StartSession(addr *hsp.Adddress) {
	url := addr.String() + addr.Route
	fmt.Println("Starting session on", url)

	c := client.NewClient()

	rl, err := readline.New("> ")
	if err != nil {
		fmt.Println("ERR: Failed to open readline session:", err)
	}

	defer rl.Close()

	for {
		line, err := rl.Readline()
		if err != nil {
			break
		}

		c.SendText(url, line)
	}
}

func main() {
	var listening bool
	flag.BoolVar(&listening, "server", false, "start a simple server")

	var host string
	var service string
	var address string

	flag.StringVar(&host, "host", "localhost", "specify server host (default: localhost)")
	flag.StringVar(&service, "port", "998", "specify server port (default: 998)")
	flag.StringVar(&address, "addr", "localhost:998", "specify target address (default: :998)")

	flag.Parse()

	if listening {
		a := fmt.Sprintf("%s:%s", host, service)
		addr, err := hsp.ParseAddress(a)
		if err != nil {
			fmt.Printf("ERR: Invalid address %s: %v\n", a, err)
			return
		}

		StartServer(addr)
		return
	}

	addr, err := hsp.ParseAddress(address)
	if err != nil {
		fmt.Printf("ERR: Invalid address %s: %v\n", address, err)
		return
	}
	StartSession(addr)
}



