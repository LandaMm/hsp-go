package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/LandaMm/hsp-go/hsp"
	"github.com/LandaMm/hsp-go/hsp/client"
	"github.com/LandaMm/hsp-go/hsp/server"
)

func PingPongRoute(req *hsp.Request) *hsp.Response {
	log.Println("Ping pong request:", req.GetRawPacket())
	df, err := req.GetDataFormat()
	if err != nil {
		return hsp.NewErrorResponse(err)
	}

	log.Println("Data format of the request:", df)

	text, err := req.ExtractText()
	if err != nil {
		return hsp.NewErrorResponse(err)
	}

	log.Println("Received text from req:", text)

	// return hsp.NewStatusResponse(hsp.STATUS_SUCCESS)
	res, err := hsp.NewJsonResponse(map[string]any{
		"success": true,
		"message": text,
	})

	if err != nil {
		return hsp.NewErrorResponse(err)
	}

	return res
}

func FileUploadRoute(req *hsp.Request) *hsp.Response {
	log.Println("[MAIN] File Upload request:", req)
	bytes, err := req.ExtractBytes()
	if err != nil {
		return hsp.NewStatusResponse(hsp.STATUS_INTERNALERR)
	}

	filename := "received.bin"
	err = os.WriteFile(filename, bytes, 0644)
	if err != nil {
		log.Fatalln("Failed to write packet payload into a file:", err)
		return hsp.NewStatusResponse(hsp.STATUS_INTERNALERR)
	}

	log.Println("Received new request from client:", req.Conn().RemoteAddr().String())

	res := hsp.NewTextResponse("Hello, World!")
	res.AddHeader("filename", filename)

	return res
}

func main() {
	addr, err := hsp.ParseAddress("127.0.0.1")
	if err != nil {
		panic(err)
	}
	srv := server.NewServer(*addr)
	fmt.Printf("Server created on address: %s\n", srv.Addr.String())

	handler := make(chan net.Conn, 1)

	router := server.NewRouter()

	router.AddRoute("/file-upload", FileUploadRoute)
	router.AddRoute("/ping-pong", PingPongRoute)

	srv.SetListener(handler)

	go func() {
		for {
			conn := <-handler
			if err := router.Handle(conn); err != nil {
				log.Println("[MAIN] Error handling connection:", err.Error())
			}
		}
	}()

	sigs := make(chan os.Signal, 1)

	go func() {
		s := <-sigs
		if s == syscall.SIGINT || s == syscall.SIGTERM {
			log.Println("Gracefully shutting down the server")
			if err := srv.Stop(); err != nil {
				log.Fatalln("Failed to close the server:", err)
			}
		}
	}()

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		time.Sleep(3 * time.Second)
		c := client.NewClient()
		rsp, err := c.SendText("localhost/ping-pong", "Hello, guys!")
		if err != nil {
			log.Fatalf("[CLIENT] Failed to received response from ping-pong: %s\n", err)
		}
		log.Println("[CLIENT] received response from remote server:", rsp)
		bts, err := json.Marshal(rsp)
		if err != nil {
			log.Fatalln("[CLIENT] Failed to serialize response from server")
		}

		log.Println("[CLIENT] Serialized response from server:", string(bts))
	}()

	if err := srv.Start(); err != nil {
		log.Fatalln("Failed to start server:", err)
	}	
}
