package model

import "sync"

// User is a struct that represents a user in the system. It has a username and a token.
type User struct {
	Username string;
	Token string;
}

// Users is a map of usernames to User objects. The key is the username and the value is the User object.
type Users map[string]User;

// LoggedUsers is a struct that represents the users that are currently logged in. 
// It has a mutex to ensure thread safety and a Users object to store the users.
type LoggedUsers struct {
	sync.Mutex
	Users Users
}