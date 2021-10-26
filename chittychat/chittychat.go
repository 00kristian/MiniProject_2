package chittychat

import (
	"golang.org/x/net/context"
	"log"
)

type Server struct{	
}

func (s *Server) Publish(ctx context.Context, message *Message) (*Empty, error){
	log.Printf("text: %s", message.Text)
	return &Empty{}, nil
}
func (s *Server) Broadcast(ctx context.Context, message *Message) (*Message, error){
	log.Printf("text: %s", message.Text)
	return &Message{Text: "Something2"}, nil
}
func (s *Server) Join(ctx context.Context, username *Username) (*Message, error){
	log.Printf("text: %s", username.Name)
	return &Message{Text: "Something3"}, nil
}
func (s *Server) Leave(ctx context.Context, username *Username) (*Message, error){
	log.Printf("text: %s", username.Name)
	return &Message{Text: "Something4"}, nil
}
func (s *Server) mustEmbedUnimplementedChittyChatServer(){}