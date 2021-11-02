package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/00kristian/MiniProject_2/proto"
	"google.golang.org/grpc"
)

// Mutex for locking lamport
var mu sync.Mutex

// Lamport time for server
var lamport uint64 = 0

type Connection struct {
	stream proto.Chat_JoinServer
	user *proto.User
	error chan error 
}

type Server struct {
	// Has to be implemented, otherwise the grpc cannot register
	proto.UnimplementedChatServer
	connections map[string]*Connection
}

func (s *Server) Leave(ctx context.Context, Id *proto.Id) (*proto.Empty, error){
	mu.Lock()
	lamport = max(lamport, Id.Lamport) + 1
	mu.Unlock()
	temp := s.connections[Id.Id]
	temp.user.Active = false
	s.connections[Id.Id] = temp
	leaveMessage := &proto.Message{
		Id: "",
		Text: Id.Id + " left Chitty-Chat at Lamport time " + fmt.Sprintf("%d", lamport),
	}
	s.Broadcast(ctx, leaveMessage)
	return &proto.Empty{}, nil
}

// Implementation of the Publish rpc - Allows users to publish messages to be broadcasted
func (s *Server) Publish(ctx context.Context, msg *proto.Message) (*proto.Empty,error){
	mu.Lock()
	lamport = max(lamport, msg.Lamport) + 1
	mu.Unlock()
	
	// if id == "", it is a join or leave message
	if msg.Id == ""{
		updatedMsg := &proto.Message{
			Id: msg.Id,
			Text: msg.Text + fmt.Sprintf("%d", lamport),
			Lamport: lamport,
		}
		log.Printf("[Server: %d] A message was published with following content: %s", lamport, updatedMsg.Text)
		s.Broadcast(ctx, updatedMsg)
	}else{
		log.Printf("[Server: %d] A message was published by %s with following content: %s", lamport, msg.Id, msg.Text)
		s.Broadcast(ctx, msg)
	}

	return &proto.Empty{}, nil
}

// Implementation of the Join rpc - alllows user to join the server
func (s *Server) Join(user *proto.User, stream proto.Chat_JoinServer) error{
	// Create a connection to server	
	conn := &Connection{
		stream: stream,
		user: user,
		error: make(chan error),
	}
	
	// Make the user active
	conn.user.Active = true

	// Add the connection to the map of connections
	s.connections[conn.user.Id] = conn

	// Return whatever error that is in the conn error field
	return <- conn.error
}

func (s *Server) Broadcast(ctx context.Context, msg *proto.Message) (*proto.Empty, error){
	// Allows counting of go routines. Go routines can be added to the wait group, and it is possible to decrease the counter when a go routine finishes its job.
	// This makes it possible to block the method from exiting before all go routines finishes their job.
	wait := sync.WaitGroup{}

	// Dummy channel for us to know when our all our go routines are done
	done := make(chan int)

	// Channel to pick up id's of active users
	//aUsers := make(chan string, 10)
	log.Printf("[Server: %d] Broadcasting message to active users:", lamport)
	//Loop through all connections
	for name, conn := range s.connections {
		//Increments the counter of the wait group - increments by one for each connection
		wait.Add(1)
		// Gather active users
		
		// Go routine that spawn an anonymous function
		go func (msg *proto.Message, conn *Connection){
			// When the method exits, decrement the wait group by one
			defer wait.Done()
			
			
			// Check if user is active, and act if the user is
			if conn.user.Active {
				mu.Lock()
				lamport += 1
				updatedMsg := &proto.Message{
					Id: msg.Id,
					Text: msg.Text,
					Lamport: lamport,
				}
				mu.Unlock()
				log.Printf("[Server: %d] Sending message to %s.", lamport, conn.user.Id)
				// Send message to the client which is attached to given connection
				err := conn.stream.Send(updatedMsg)
				
				//aUsers <- conn.user.Id
				
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
		// Closes our active user channel in order to loop over it
		//close(aUsers)
		// Closes our done dummy channel
		close(done)
	}()

	// Create logstring to be printed for a message to be broadcasted
	
	// logString := "" 
	// if msg.Id == ""{
	// 	logString = "Broadcasting message to active users:"
	// }else{
	// 	logString = fmt.Sprintf("Broadcasting %s's message to active users:", msg.Id)
	// }
	// for id := range aUsers{
	// 	logString += " " + id
	// }
	// log.Print(logString + "\n\n")

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
	log.Printf("[Server: %d] Started server on port: 8080", lamport)

	// Register our Chat server on out grpc server, and pass our service which is the server type
	proto.RegisterChatServer(grpcServer, server)

	// Serve incomming connetions to the listener
	grpcServer.Serve(listener)
}

func max(x, y uint64) uint64{
	if x >= y{
		return x
	}else {
		return y
	}
}
