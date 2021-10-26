package server

import (
	"log"
	"net"

	"google.golang.org/grpc"
)

func main(){
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Failed to listen on port 8080: %v", err)
	}

	server := grpc.NewServer()

	if err := server.Serve(lis); err != nil {
		log.Fatalf("Failed to serve on port 8080: %v", err)
	}
}
