package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/LandaMm/hsp-go/server"
)

func FileUploadRoute(req *server.Request) *server.Response {
	log.Println("[MAIN] File Upload request:", req)
	bytes, err := req.ExtractBytes()
	if err != nil {
		return server.NewErrorResponse(err)
	}

	filename := "received.bin"
	err = os.WriteFile(filename, bytes, 0644)
	if err != nil {
		return server.NewErrorResponse(err)
	}

	log.Println("Received new request from client:", req.Conn().RemoteAddr().String())

	res := server.NewTextResponse("Hello, World!")
	res.AddHeader("filename", filename)

	return res
}

func main() {
	srv := server.NewServer("localhost:3000")
	fmt.Printf("Server created on address: %s\n", srv.Addr)

	handler := make(chan net.Conn, 1)

	router := server.NewRouter()

	router.AddRoute("/file-upload", FileUploadRoute)

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

	if err := srv.Start(); err != nil {
		log.Fatalln("Failed to start server")
	}
}
