package routes

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/DaniSancas/go-chat-room/server/internal/model"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
)

// Handler is a struct that contains the shared state of the server.
// It is used to pass the shared state to the handlers.
type Handler struct {
	LoggedUsers model.LoggedUsers
}

// upgrader is a websocket upgrader that is used to upgrade an HTTP
// connection to a websocket connection.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// login is a handler function that logs in a user. It receives a POST request with a JSON body containing the username of the user.
// It generates a random token for the user and adds the user to the list of logged users.
//
// If the user is already logged in, it returns an error.
// If the request is not a POST request, it returns an error.
// If the body of the request is not a valid JSON, it returns an error.
// If everything is ok, it returns the token of the user.
func (handler *Handler) login(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != "POST" {
		responseMessage := "Invalid request method"
		log.Printf("%s: %s", responseMessage, r.Method)
		http.Error(w, responseMessage, http.StatusMethodNotAllowed)
		return
	}
	// request body can't be nil
	if r.Body == nil {
		responseMessage := "Request body missing"
		log.Print(responseMessage)
		http.Error(w, responseMessage, http.StatusBadRequest)
		return
	}

	// Parse the request body to get the user data
	var userLoginRequest model.UserLoginRequest
	err := json.NewDecoder(r.Body).Decode(&userLoginRequest)
	if err != nil {
		responseMessage := "Can't decode body"
		log.Printf("%s: %v", responseMessage, err)
		http.Error(w, "Can't decode body", http.StatusBadRequest)
		return
	}

	// Check if the user is already logged in, in which case return an error
	// Aquire lock in write mode
	handler.LoggedUsers.Lock()
	defer handler.LoggedUsers.Unlock()
	if _, ok := handler.LoggedUsers.Users[userLoginRequest.Username]; ok {
		responseMessage := fmt.Sprintf("User %s is already logged in", userLoginRequest.Username)
		log.Print(responseMessage)
		http.Error(w, responseMessage, http.StatusConflict)
		return
	}

	// Generate a random UUID for the user
	token := uuid.NewString()
	// Add the user to the logged users
	handler.LoggedUsers.Users[userLoginRequest.Username] = model.User{
		Username: userLoginRequest.Username,
		Token:    token,
	}

	// If everything is ok, finally return the token
	log.Printf("User %s logged in with token %s", userLoginRequest.Username, token)
	json.NewEncoder(w).Encode(model.UserLoginResponse{Token: token})
}

// logout is a handler function that logs out a user. It receives a POST request with a JSON body containing the username and the token of the user.
// It removes the user from the list of logged users.
//
// If the request is not a POST request, it returns an error.
// If the body of the request is not a valid JSON, it returns an error.
// If the user is not logged in, it returns an error.
// If the token is incorrect, it returns an error.
// If everything is ok, it returns a message saying that the user was successfully logged out.
func (handler *Handler) logout(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != "POST" {
		responseMessage := "Invalid request method"
		log.Printf("%s: %s", responseMessage, r.Method)
		http.Error(w, responseMessage, http.StatusMethodNotAllowed)
		return
	}
	// request body can't be nil
	if r.Body == nil {
		responseMessage := "Request body missing"
		log.Print(responseMessage)
		http.Error(w, responseMessage, http.StatusBadRequest)
		return
	}

	// Parse the request body to get the user data
	var userLogoutRequest model.UserWithTokenRequest
	err := json.NewDecoder(r.Body).Decode(&userLogoutRequest)
	if err != nil {
		responseMessage := "Can't decode body"
		log.Printf("%s: %v", responseMessage, err)
		http.Error(w, "Can't decode body", http.StatusBadRequest)
		return
	}

	// Check if the user is not logged in, in which case return an error
	// Aquire lock in write mode
	handler.LoggedUsers.Lock()
	defer handler.LoggedUsers.Unlock()
	if _, ok := handler.LoggedUsers.Users[userLogoutRequest.Username]; !ok {
		responseMessage := fmt.Sprintf("User %s is not logged in", userLogoutRequest.Username)
		log.Print(responseMessage)
		http.Error(w, responseMessage, http.StatusConflict)
		return
	}

	// In case the user is logged in, check if the token is correct
	if handler.LoggedUsers.Users[userLogoutRequest.Username].Token != userLogoutRequest.Token {
		responseMessage := "Invalid token"
		log.Print(responseMessage)
		http.Error(w, responseMessage, http.StatusConflict)
		return
	}

	// Remove the user from the logged users, closing the channel if it exists
	userToLogout := handler.LoggedUsers.Users[userLogoutRequest.Username]
	if userToLogout.Channel != nil {
		close(userToLogout.Channel)
		log.Printf("User %s disconnected from the stream", userLogoutRequest.Username)
	}
	delete(handler.LoggedUsers.Users, userLogoutRequest.Username)

	// If everything is ok, finally return the token
	log.Printf("User %s successfully logged out", userLogoutRequest.Username)
	json.NewEncoder(w).Encode(model.UserLogoutResponse{Message: "User successfully logged out"})
}

// stream is a handler function that streams messages to the user. 
// It upgrades an HTTP connection to a websocket connection, reads the username and token from the first message, and validates the user.
// If the user is not logged in or the token is incorrect, it returns an error.
// If everything is ok, it starts a goroutine to send messages to the user and handles the rest of the messages in a loop.
func (handler *Handler) stream(w http.ResponseWriter, r *http.Request) {
	websocket, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer websocket.Close()

	// Manage first message which should be the username and token to validate the user
	// read a message
	messageType, messageContent, err := websocket.ReadMessage()
	if err != nil {
		log.Println(err)
		return
	}

	// Parse the request body to get the user data
	var userWithTokenRequest model.UserWithTokenRequest
	err = json.Unmarshal(messageContent, &userWithTokenRequest)
	if err != nil {
		responseMessage := "Can't decode body"
		log.Printf("%s: %v", responseMessage, err)

		if err := websocket.WriteMessage(messageType, []byte(responseMessage)); err != nil {
			log.Println(err)
			return
		}
		return
	}

	// Check if the provided username and token are valid
	// In case the currentUser is logged in and the token is correct, create a channel and add it to the logged users map.
	// Should return true if the user is not logged in or the token is incorrect, and false otherwise.
	shouldReturn := BindChannelToUserIfExists(handler, userWithTokenRequest, websocket, messageType)
	if shouldReturn {
		return
	}

	// Start a goroutine to send messages to the user from the channel
	go func() {
		for {
			message := <-handler.LoggedUsers.Users[userWithTokenRequest.Username].Channel
			if err := websocket.WriteMessage(messageType, message); err != nil {
				log.Println(err)
				return
			}
		}
	}()

	// Handle the rest of the messages in a loop
	handler.listenForMessages(websocket)
}

// BindChannelToUserIfExists checks if the user is logged in and if the token is correct.
// If the user is logged in and the token is correct, it creates a channel for the user and adds it to the logged users map.
// It returns true if the user is not logged in or the token is incorrect, and false otherwise.
func BindChannelToUserIfExists(handler *Handler, userWithTokenRequest model.UserWithTokenRequest, websocket *websocket.Conn, messageType int) bool {
	handler.LoggedUsers.Lock()
	defer handler.LoggedUsers.Unlock()
	if _, ok := handler.LoggedUsers.Users[userWithTokenRequest.Username]; !ok {
		responseMessage := fmt.Sprintf("User %s is not logged in", userWithTokenRequest.Username)
		log.Println(responseMessage)

		if err := websocket.WriteMessage(messageType, []byte(responseMessage)); err != nil {
			log.Println(err)
			return true
		}
		return true
	}

	if handler.LoggedUsers.Users[userWithTokenRequest.Username].Token != userWithTokenRequest.Token {
		responseMessage := fmt.Sprintf("Invalid token '%s' for user %s", userWithTokenRequest.Token, userWithTokenRequest.Username)
		log.Println(responseMessage)

		if err := websocket.WriteMessage(messageType, []byte(responseMessage)); err != nil {
			log.Println(err)
			return true
		}
		return true
	}

	currentUser := handler.LoggedUsers.Users[userWithTokenRequest.Username]
	currentUser.Channel = make(chan []byte)
	handler.LoggedUsers.Users[userWithTokenRequest.Username] = currentUser
	log.Printf("User %s is now connected to the stream", userWithTokenRequest.Username)
	return false
}

func (handler *Handler) listenForMessages(conn *websocket.Conn) {
	for {
		// read a message
		messageType, messageContent, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		// print out that message
		fmt.Println(string(messageContent))

		// reponse message
		messageResponse := fmt.Sprintf("Your message is: %s", messageContent)

		if err := conn.WriteMessage(messageType, []byte(messageResponse)); err != nil {
			log.Println(err)
			return
		}
	}
}

// homepage is a handler function that returns a welcome message to the user.
func homepage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Welcome to the homepage!")
}

// HandleRequests is the main function of the routes package. It sets up the routes for the server.
func HandleRequests() {
	// Initialize shared state
	handler := Handler{
		LoggedUsers: model.LoggedUsers{
			Users: make(model.Users),
		},
	}

	// Enable CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		// Enable Debugging for testing, consider disabling in production
		Debug: false,
	})

	// Start server
	log.Println("Starting server...")
	mux := http.NewServeMux()
	mux.HandleFunc("/", homepage)
	mux.HandleFunc("/login", handler.login)
	mux.HandleFunc("/logout", handler.logout)
	mux.HandleFunc("/stream", handler.stream)
	log.Fatal(http.ListenAndServe(":8080", c.Handler(mux)))
}
