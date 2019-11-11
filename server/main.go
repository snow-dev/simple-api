package main

import (
	blogpb "../proto"

	"fmt"
	"log"
	"net"
)

type BlogServiceServer struct {}

func main() {
	// Configure 'log' package to give file name and line number on eg. log.Fatal
	// Pipe flags to another (log.LtsdFlags = log.Ldate | log.Ltime)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	fmt.Println("Starting server on port :50051...")

	// Start our listener, 50051 is the default gRPC port
	listener, err := net.Listen("tcp", ":50051")
	// Handle errors if any.
	if err != nil {
		log.Fatal("Unable tolisten on port :50051: %v", err)
	}

	// Set options, here we can configure tings like TLS support
	//opts := []grpc.ServerOption{}
	// Create new gRPC server wth (blank) options
	//s := grpc.NewServer(opts...)


	
}
