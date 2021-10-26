package server

import (
	"log"
	"net"
	
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding/proto"
)

func main(){
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Failed to listen on port 8080: %v", err)
	}
	
	s := protobuf.Server{}
	server := grpc.NewServer()

	protobuf.RegisterChittyChatServer(server, &s)
	

	if err := server.Serve(lis); err != nil {
		log.Fatalf("Failed to serve on port 8080: %v", err)
	}
}
