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
// Mutex for locking
var mu sync.Mutex

// lamport time for given client
var lamport uint64 = 0

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

	// join event increments lamport by one
	mu.Lock()
	lamport += 1
	mu.Unlock()

	joinMessage := &proto.Message{
		Id: "",
		Text: user.Name + " joined Chitty-Chat at Lamport time ",
		Lamport: lamport,
	}

	// Creates the stream, that is return when a user joins the server
	stream, err := client.Join(context.Background(), user)
	// Publish the join message
	_, joinMessageErr := client.Publish(context.Background(), joinMessage)

	if err != nil {
		log.Fatalf("Connection failed: %v", err)
	}
	if joinMessageErr != nil {
		log.Fatalf("Error occured when publishing join message: %v", joinMessageErr)
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
			mu.Lock()
			lamport = max(lamport, msg.Lamport) + 1
			mu.Unlock()

			// If an error occurs, the goroutine and the for loop must terminate. 
			// Error is passed to the local sError variable
			if err != nil {
				sError = fmt.Errorf("Error occured when reading message: %v", err)
				break;
			}
			// If id == "", it is a join message
			if msg.Id == "" {
				log.Printf("[%s: %d] %s", user.Id, lamport, msg.Text)
			}else{
				log.Printf("[%s: %d] %s: %s", user.Id, lamport, msg.Id, msg.Text)
			}
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
	
	// Show welcome message
	welcome()

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
			msgContent := strings.TrimSpace(scanner.Text())
			if !validateMsg(msgContent) {
				fmt.Println("Please type a valid message. A valid message is a UTF-8 encoded string consisting of max 128 characters.")
				continue 
			}
			mu.Lock()
			lamport += 1
			mu.Unlock()
			msg := &proto.Message{
				Id: id,
				Text: msgContent,
				Lamport: lamport,
			}
			// Check if said message is a command
			if strings.Contains(msg.Text, "\\leave"){
				_ , errLeave := client.Leave(context.Background(), &proto.Id{Id: msg.Id, Lamport: msg.Lamport})
				if errLeave != nil{
					log.Fatalf("Error occured when trying to leave: %v", errLeave)
				}
				wait.Done()
				break
			} else if strings.Contains(msg.Text, "\\help"){
				fmt.Println("------------------------------------")
				fmt.Println("Following commands are available:")
				fmt.Println("\\leave - Exits Chitty-Chat.")
				fmt.Println("\\help - Shows this menu again.")
				fmt.Println("------------------------------------")
			} else{
				// Call the broadcast message and distibute the message through all active useres
				_, err := client.Publish(context.Background(), msg)
				if err != nil {
					log.Fatalf("Error sending message: %v", err)
					break
				}
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

func welcome(){
	fmt.Println("Welcome to Chitty-chat! =^.^=")
	fmt.Println("------------------------------------")
	fmt.Println("Following commands are available:")
	fmt.Println("\\leave - Exits Chitty-Chat.")
	fmt.Println("\\help - Shows this menu again.")
	fmt.Println("------------------------------------")
}

func max(x, y uint64) uint64{
	if x >= y{
		return x
	}else {
		return y
	}
}

func validateMsg(x string) bool {
	return len(x) <= 128
}