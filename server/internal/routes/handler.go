package routes

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"github.com/DaniSancas/go-chat-room/server/internal/model"
    "github.com/google/uuid"

)

// Handler is a struct that contains the shared state of the server. 
// It is used to pass the shared state to the handlers.
type Handler struct {
	LoggedUsers model.LoggedUsers
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
		Token: token,
	}

	// If everything is ok, finally return the token
	log.Printf("User %s logged in with token %s", userLoginRequest.Username, token)
	json.NewEncoder(w).Encode(model.UserLoginResponse{Token: token})
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

	// Start server
	log.Println("Starting server...")
	http.HandleFunc("/", homepage)
	http.HandleFunc("/login", handler.login)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
