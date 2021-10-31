package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/00kristian/MiniProject_2/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// Global variable for our client
var client proto.ChatClient
// Global wait group
var wait *sync.WaitGroup

// Init func to initialize the wait group
func init(){
	wait = &sync.WaitGroup{}
}

// Lets the client join into the server
func join(id string, name string) error {
	// Stream error to be used if something fails
	var sError error

	user := &proto.User{
		Id: id,
		Name: name,
		Active: true,
	}

	// Creates the stream, that is return when a user joins the server
	stream, err := client.Join(context.Background(), user)

	if err != nil {
		log.Fatalf("Connection failed: %v", err)
	}

	// Increments the wait group by one
	wait.Add(1)

	// Go rotuine that spawns an anonymous function, that takes the newly created stream as input.
	// The proto.Chat_JoinClient allows us to define the behaviour in order to implement the streaming part of our client
	go func(str proto.Chat_JoinClient) {
		// Decrements the wait group when mehtod exits
		defer wait.Done()
		
		// Infinite for loop 
		for{
			// Wait until a message is recieved in the stream
			msg, err := str.Recv()
			
			// If an error occurs, the goroutine and the for loop must terminate. 
			// Error is passed to the local sError variable
			if err != nil {
				sError = fmt.Errorf("Error occured when reading message: %v", err)
				break;
			}

			log.Printf("%s: %s", msg.Id, msg.Text)
		}
	}(stream)

	return sError
}

func main(){
	// Reader to read user input
	reader := bufio.NewReader(os.Stdin)

	// Dummy channel to ensure all go routines are finished
	done := make(chan int)

	// Reads and parse name into id and name, which is used to connect
	fmt.Print("Please enter you name: ")
	temp, _ := reader.ReadString('\n')
	name := strings.TrimSpace(temp)
	id := name

	// Connect to our server - no https, so connect with grpc.WithInsecure()
	conn, err := grpc.Dial(":8080", grpc.WithInsecure())
	if err != nil{
		log.Fatalf("Could not connect: %s", err)
	}

	// When method is done close the connection
	defer conn.Close()

	// Creates the client on our connection
	client = proto.NewChatClient(conn)
	// Join the server with the given name and id
	join(id, name)
	
	//Increment wait gorup before go routine
	wait.Add(1)

	// Go routine that spawns an anonymous function
	go func(){
		defer wait.Done()

		// Create scanner in order to scan user messages
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan(){
			msg := &proto.Message{
				Id: id,
				Text: scanner.Text(),
			}
			
			// Call the broadcast message and distibute the message through all active useres
			_, err := client.Broadcast(context.Background(), msg)

			if err != nil {
				log.Fatalf("Error sending message: %v", err)
				break
			}
		}
	}()

	// Go routine that spawns anonymous function that ensures that the wait group waits for the go routines to exit
	go func(){
		wait.Wait()
		// Closes our done dummy channel
		close(done)
	}()
	
	// Acts as a blocker - code will not proceed from this until our done channel has been closed. That happens after all our go routines are done.
	<- done
}