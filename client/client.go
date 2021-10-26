package main

import (
	"log"

	"github.com/00kristian/MiniProject_2/chittychat"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func main(){
	conn, err := grpc.Dial(":8080", grpc.WithInsecure())
	if err != nil{
		log.Fatalf("Could not connect: %s", err)
	}

	defer conn.Close()

	c := chittychat.NewChittyChatClient(conn)

	message := chittychat.Message{
		Text: "Hello from client",
	}
	response, err := c.Publish(context.Background(), &message) 

	if err != nil{
		log.Fatalf("Error when calling publish: %s", err)
	}

	log.Printf("Response from server: %s", response)

}