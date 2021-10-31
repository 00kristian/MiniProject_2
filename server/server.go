package main

import (
	"context"
	"log"
	"net"
	"sync"

	"github.com/00kristian/MiniProject_2/proto"
	"google.golang.org/grpc"
)

type Connection struct {
	stream proto.Chat_JoinServer
	user *proto.User
	error chan error 
}

type Server struct {
	proto.UnimplementedChatServer
	connections map[string]*Connection
}

// Implementation of the Join rpc - alllows user to join the server
func (s *Server) Join(user *proto.User, stream proto.Chat_JoinServer) error{
	// Create a connection to server	
	conn := &Connection{
		stream: stream,
		user: user,
		error: make(chan error),
	}
	
	// Check if the user is already in the map. If the user is just turn on the connection, instead of makin a new one
	if val, ok := s.connections[user.Name]; ok {
		val.user.Active = true
		return <-conn.error
	}

	// Make the user active
	conn.user.Active = true

	// Add the connection to the map of connections
	s.connections[conn.user.Name] = conn

	// Return whatever error that is in the conn error field
	return <- conn.error
}

func (s *Server) Broadcast(ctx context.Context, msg *proto.Message) (*proto.Empty, error){
	// Allows counting of go routines. Go routines can be added to the wait group, and it is possible to decrease the counter when a go routine finishes its job.
	// This makes it possible to block the method from exiting before all go routines finishes their job.
	wait := sync.WaitGroup{}

	// Dummy channel for us to know when our all our go routines are done
	done := make(chan int)

	
	//Loop through all connections
	for name, conn := range s.connections {
		//Increments the counter of the wait group - increments by one for each connection
		wait.Add(1)
		
		// Go routine that spawn an anonymous function
		go func (msg *proto.Message, conn *Connection){
			// When the method exits, decrement the wait group by one
			defer wait.Done()

			// Check if user is active, and act if the user is
			if conn.user.Active {
				// Send message to the client which is attached to given connection
				err := conn.stream.Send(msg)
				// Log which message is sent on the server
				log.Printf("Sending %s's to everyone", name)
				
				// If an error occurs - print the error and terminate the conneciton making the user go offline
				if err != nil {
					log.Fatalf("Error sending message %s - Error: %v", name, err)
					conn.user.Active = false
					// Pass the error to the error chan for the connection
					conn.error <- err
				}
			}
		}(msg, conn)
	}

	// Go routine that spawns anonymous function that ensures that the wait group waits for the go routines to exit
	go func(){
		wait.Wait()
		// Closes our done dummy channel
		close(done)
	}()
	
	// Acts as a blocker - code will not proceed from this until our done channel has been closed. That happens after all our go routines are done.
	<- done
	
	// Method can exit, and nothing with no error
	return &proto.Empty{}, nil
}


func main(){
	// Create the map of connections
	connections := make(map[string]*Connection)

	// Reference to our server with our connections
	server := &Server{proto.UnimplementedChatServer{},connections}

	// Startup of the grpc server
	grpcServer := grpc.NewServer()

	// Create a listener. This listener listen on our port 8080
	listener, err := net.Listen("tcp", ":8080")

	//Check if error occured  when trying to listen on port
	if err != nil {
		log.Fatalf("Error creating server: %v", err)
	}

	//Print to show that server has started
	log.Print("Started server on port: 8080")

	// Register our Chat server on out grpc server, and pass our service which is the server type
	proto.RegisterChatServer(grpcServer, server)

	// Serve incomming connetions to the listener
	grpcServer.Serve(listener)
}
