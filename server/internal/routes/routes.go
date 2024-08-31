package routes

import (
	"fmt"
	"log"
	"net/http"
)

func homepage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the homepage!")
}

// HandleRequests is the main function of the routes package. It sets up the routes for the server.
func HandleRequests() {
	log.Println("Starting server...")
	http.HandleFunc("/", homepage)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
