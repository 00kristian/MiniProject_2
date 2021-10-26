package proto

import (
	"context"
	"log"
)

type Server struct{
	
}

func (s *Server) Publish(ctx context.Context, message *Message) (*Message, error){

}