package chittychat

import (
	"golang.org/x/net/context"
	"log"
)

type Server struct{	
}

func (s *Server) Publish(ctx context.Context, message *Message) (*Message, error){
	log.Printf("text: %s", message.Text)
	return &Message{Text: "Something"}, nil
}