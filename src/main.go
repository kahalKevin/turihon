package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"

	"model"
)

var clients = make(map[string]model.User) // connected clients
var broadcast = make(chan model.Message)           // broadcast channel

// Configure the upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	// Create a simple file server
	fs := http.FileServer(http.Dir("../public"))
	http.Handle("/", fs)

	// Configure websocket route
	http.HandleFunc("/ws", handleConnections)

	// Start listening for incoming chat messages
	go handleMessages()

	// Start the server on localhost port 8000 and log any errors
	log.Println("http server started on :8000")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to a websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	// Make sure we close the connection when the function returns
	defer ws.Close()
	newUuid := uuid.NewV4()
	contactExceptSelf := getAllClientIds(clients)
	contactExceptSelf = removeStringSlices(contactExceptSelf, newUuid.String())
	newUser := model.NewUser(newUuid.String(), ws, time.Now(), contactExceptSelf) 
	// Register our new client
	clients[newUser.Id] = newUser
	contact := model.Message{
		TypeMsg:	"init",
		From   :	newUser.Id,
		Contact:	newUser.Contact}
	// Send init data to client, with some contact
	newUser.Connection.WriteJSON(contact)
	go spreadNewUserToConnectedClient(newUser)
	for {
		var msg model.Message
		// Read in a new message as JSON and map it to a Message object
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("1 error: %v", err)
			delete(clients, newUser.Id)
			go removeDisconnectedClient(newUser)
			break
		}
		// Send the newly received message to the broadcast channel
		broadcast <- msg
	}
}

func getAllClientIds(AllId map[string]model.User) model.Ids{
	var _Ids model.Ids
	for id := range AllId {
		_Ids = append(_Ids, id)
	}
	return _Ids
}

func spreadNewUserToConnectedClient(userToSpread model.User){
	newContact := model.Message{
		TypeMsg:	"add",
		From   :	userToSpread.Id}
	for clientId := range clients{
		if (!stringIsInSlice(userToSpread.Id, clients[clientId].Contact) && (clientId != userToSpread.Id)){
			clients[clientId].Connection.WriteJSON(newContact)
			tempClient := clients[clientId]
			tempClient.Contact = append(tempClient.Contact, userToSpread.Id)
			clients[clientId] = tempClient
		}
	}
}

func removeDisconnectedClient(disconnectedClient model.User){
	runawayContact := model.Message{
		TypeMsg:	"remove",
		From   :	disconnectedClient.Id}
	for clientId := range clients{
		if stringIsInSlice(disconnectedClient.Id, clients[clientId].Contact){
			clients[clientId].Connection.WriteJSON(runawayContact)
			tempClient := clients[clientId]
			tempClient.Contact = removeStringSlices(tempClient.Contact, disconnectedClient.Id)
			clients[clientId] = tempClient
		}
	}	
	log.Printf("the removal is finished!")
}

func stringIsInSlice(a string, list []string) bool {
    for _, b := range list {
        if b == a {
            return true
        }
    }
    return false
}

func removeStringSlices(s []string, r string) []string {
    for i, v := range s {
        if v == r {
            return append(s[:i], s[i+1:]...)
        }
    }
    return s
}

func handleMessages() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcast
		err := clients[msg.Target].Connection.WriteJSON(msg)
		if err != nil {
			log.Printf("2 error: %v", err)
			clients[msg.Target].Connection.Close()
			delete(clients, msg.Target)
			go removeDisconnectedClient(clients[msg.Target])
		}
	}
}
