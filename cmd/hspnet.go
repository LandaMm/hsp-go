package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/LandaMm/hsp-go/hsp"
	"github.com/LandaMm/hsp-go/hsp/client"
	"github.com/LandaMm/hsp-go/hsp/server"
	"github.com/chzyer/readline"
)

type Header struct {
	Key   string
	Value string
}

type HeaderList struct {
	Headers []Header
}

func (hl *HeaderList) Map() map[string]string {
	headerMap := make(map[string]string)
	for _, header := range hl.Headers {
		headerMap[header.Key] = header.Value
	}
	return headerMap
}

func (hl *HeaderList) Set(arg string) error {
	var key, value string
	if _, err := fmt.Sscanf(arg, "%s %s", &key, &value); err != nil {
		return err
	}
	hl.Headers = append(hl.Headers, Header{
		Key:   key,
		Value: value,
	})
	return nil
}

func (hl *HeaderList) String() string {
	return fmt.Sprintf("%d headers", len(hl.Headers))
}

func PrintPacket(pkt *hsp.Packet) error {
	fmt.Printf("REQUEST %s\n", pkt.Headers[hsp.H_ROUTE])
	fmt.Println("Headers:")

	for k, v := range pkt.Headers {
		fmt.Printf("\t%s: %s\n", k, v)
	}

	h, ok := pkt.Headers[hsp.H_DATA_FORMAT]
	if !ok {
		return fmt.Errorf("data format header is not present")
	}

	df, err := hsp.ParseDataFormat(h)
	if err != nil {
		return fmt.Errorf("invalid data format: %s", h)
	}

	fmt.Printf("Payload (%s):\n", df.String())

	switch df.Format {
	case hsp.DF_TEXT:
		fmt.Println(string(pkt.Payload))
	case hsp.DF_JSON:
		var out any
		err := json.Unmarshal(pkt.Payload, &out)
		if err != nil {
			return err
		}

		raw, err := json.MarshalIndent(out, "", "  ")
		if err != nil {
			return err
		}

		fmt.Println(string(raw))
	case hsp.DF_BYTES:
		fmt.Printf("%d bytes\n", len(pkt.Payload))
	default:
		fmt.Printf("ERR: Unsupported data format: %s\n", df.String())
	}

	return nil
}

func Index(req *hsp.Request) *hsp.Response {
	pkt := req.GetRawPacket()

	err := PrintPacket(pkt)
	if err != nil {
		fmt.Println("ERR: Couldn't print out the packet:", err)
	}

	rsp := hsp.NewStatusResponse(hsp.STATUS_SUCCESS)

	reqF, err := req.GetDataFormat()
	if err != nil {
		return hsp.NewErrorResponse(err)
	}

	rsp.Format = *reqF
	rsp.Payload = req.GetRawPacket().Payload

	return rsp
}

func StartServer(addr *hsp.Adddress) {
	srv := server.NewServer(*addr)
	fmt.Printf("Server created on address: %s\n", srv.Addr.String())

	handler := make(chan *hsp.Connection, 1)

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

func StartSession(options *client.ClientOptions) {
	c := client.NewClient(options)

	rl, err := readline.New("> ")
	if err != nil {
		fmt.Println("ERR: Failed to open readline session:", err)
	}

	defer func() {
		if err := rl.Close(); err != nil {
			fmt.Println("ERR: Failed to close readline session:", err)
		}
	}()

	for {
		rl.SetPrompt("Route > ")
		route, err := rl.Readline()
		if err != nil {
			break
		}

		rl.SetPrompt("Data > ")
		line, err := rl.Readline()
		if err != nil {
			break
		}

		var rsp *hsp.Response
		var rerr error

		if strings.HasPrefix(line, "/file") {
			what := strings.TrimLeft(line, "/file")
			isJson := false
			var filename string
			if strings.HasPrefix(what, ":json ") {
				isJson = true
				filename = strings.TrimLeft(what, ":json ")
			} else {
				filename = strings.TrimLeft(what, " ")
			}

			file, err := os.Open(filename)
			if err != nil {
				fmt.Printf("ERR: Failed to open file '%s': %v\n", filename, err)
				continue
			}

			if !isJson {
				buf, err := io.ReadAll(file)
				if err != nil {
					fmt.Printf("ERR: Failed to read from file '%s': %v\n", filename, err)
					continue
				}
				rsp, err = c.SendBytes(route, buf)
			} else {
				var data any

				decoder := json.NewDecoder(file)
				err = decoder.Decode(&data)
				if err == nil {
					rsp, err = c.SendJson(route, data)
				}
			}
		} else if strings.HasPrefix(line, "/json ") {
			var data any
			err = json.Unmarshal([]byte(strings.TrimLeft(line, "/json ")), &data)
			if err != nil {
				fmt.Println("ERR: Invalid JSON for request:", err)
			} else {
				rsp, err = c.SendJson(route, data)
			}
		} else {
			rsp, rerr = c.SendText(route, line)
		}

		if rerr != nil {
			fmt.Println("ERR: Failed to send text to server:", err)
			continue
		}

		if err = PrintPacket(rsp.ToPacket()); err != nil {
			fmt.Println("ERR: Couldn't print out response:", err)
		}
	}
}

func main() {
	var listening bool
	flag.BoolVar(&listening, "server", false, "start a simple server")

	var host string
	var service string
	var address string

	var headerList HeaderList
	var auth string

	flag.StringVar(&host, "host", "localhost", "specify server host")
	flag.StringVar(&service, "port", "998", "specify server port")
	flag.StringVar(&address, "addr", "localhost:998", "specify target address")

	flag.StringVar(&auth, "auth", "", "provide auth credentials")

	flag.Var(&headerList, "H", "provide additional header")

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

	options := &client.ClientOptions{
		Headers: headerList.Map(),
		Auth:    auth,
		BaseURL: address,
	}

	StartSession(options)
}
