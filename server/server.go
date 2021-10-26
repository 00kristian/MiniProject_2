package main

import (
	"log"
	"net"
	"github.com/00kristian/MiniProject_2/chittychat"
	"google.golang.org/grpc"
)

func main(){
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Failed to listen on port 8080: %v", err)
	}
	
	s := chittychat.Server{}
	server := grpc.NewServer()

	chittychat.RegisterChittyChatServer(server, &s)
	

	if err := server.Serve(lis); err != nil {
		log.Fatalf("Failed to serve on port 8080: %v", err)
	}
}
