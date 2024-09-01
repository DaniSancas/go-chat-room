package routes

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DaniSancas/go-chat-room/server/internal/model"
)

func TestHomepage(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(homepage)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := "Welcome to the homepage!"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func TestLoginRequestInvalidRequestMethod(t *testing.T) {
	// Test with GET, PUT and DELETE
	var tests = []struct {
		method string
	}{
		{"GET"},
		{"PUT"},
		{"DELETE"},
	}
	for _, tt := range tests {
		testname := tt.method
		t.Run(testname, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, "/login", nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handlerFixture := Handler{
				LoggedUsers: model.LoggedUsers{
					Users: make(model.Users),
				},
			}
			handler := http.HandlerFunc(handlerFixture.login)

			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusMethodNotAllowed {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, http.StatusMethodNotAllowed)
			}

			expected := "Invalid request method"
			received := strings.TrimSpace(rr.Body.String())
			if received != expected {
				t.Errorf("handler returned unexpected body: got %v want %v",
					received, expected)
			}

			handlerFixture.LoggedUsers.RLock()
			defer handlerFixture.LoggedUsers.RUnlock()
			if len(handlerFixture.LoggedUsers.Users) != 0 {
				t.Errorf("The list of logged users should be empty")
			}
		})
	}
}

func TestLoginRequestBodyMissing(t *testing.T) {
	req, err := http.NewRequest("POST", "/login", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handlerFixture := Handler{
		LoggedUsers: model.LoggedUsers{
			Users: make(model.Users),
		},
	}
	handler := http.HandlerFunc(handlerFixture.login)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}

	expected := "Request body missing"
	received := strings.TrimSpace(rr.Body.String())
	if received != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			received, expected)
	}

	handlerFixture.LoggedUsers.RLock()
	defer handlerFixture.LoggedUsers.RUnlock()
	if len(handlerFixture.LoggedUsers.Users) != 0 {
		t.Errorf("The list of logged users should be empty")
	}
}

func TestLoginRequestCanNotDecodeBody(t *testing.T) {
	req, err := http.NewRequest("POST", "/login", strings.NewReader("invalid json"))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handlerFixture := Handler{
		LoggedUsers: model.LoggedUsers{
			Users: make(model.Users),
		},
	}
	handler := http.HandlerFunc(handlerFixture.login)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}

	expected := "Can't decode body"
	received := strings.TrimSpace(rr.Body.String())
	if received != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			received, expected)
	}

	handlerFixture.LoggedUsers.RLock()
	defer handlerFixture.LoggedUsers.RUnlock()
	if len(handlerFixture.LoggedUsers.Users) != 0 {
		t.Errorf("The list of logged users should be empty")
	}
}

func TestLoginSuccess(t *testing.T) {
	req, err := http.NewRequest("POST", "/login", strings.NewReader(`{"username": "user"}`))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handlerFixture := Handler{
		LoggedUsers: model.LoggedUsers{
			Users: make(model.Users),
		},
	}
	handler := http.HandlerFunc(handlerFixture.login)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := `{"token":"`
	received := strings.TrimSpace(rr.Body.String())
	if !strings.HasPrefix(received, expected) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			received, expected)
	}

	handlerFixture.LoggedUsers.RLock()
	defer handlerFixture.LoggedUsers.RUnlock()

	if _, ok := handlerFixture.LoggedUsers.Users["user"]; !ok {
		t.Errorf("User should be present in the list of logged users")
	}

	if len(handlerFixture.LoggedUsers.Users) != 1 {
		t.Errorf("The list of logged users should have only one user")
	}
}

func TestLoginUserAlreadyLoggedIn(t *testing.T) {
	req, err := http.NewRequest("POST", "/login", strings.NewReader(`{"username": "user"}`))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handlerFixture := Handler{
		LoggedUsers: model.LoggedUsers{
			Users: model.Users{
				"user": model.User{
					Username: "user",
					Token:    "token",
				},
			},
		},
	}
	handler := http.HandlerFunc(handlerFixture.login)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusConflict {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusConflict)
	}

	expected := "User user is already logged in"
	received := strings.TrimSpace(rr.Body.String())
	if received != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			received, expected)
	}

	handlerFixture.LoggedUsers.RLock()
	defer handlerFixture.LoggedUsers.RUnlock()
	if _, ok := handlerFixture.LoggedUsers.Users["user"]; !ok {
		t.Errorf("User should be present the list of logged users")
	}

	if len(handlerFixture.LoggedUsers.Users) != 1 {
		t.Errorf("There should be only one user in the list of logged users")
	}
}

func TestLogoutRequestInvalidRequestMethod(t *testing.T) {
	// Test with GET, PUT and DELETE
	var tests = []struct {
		method string
	}{
		{"GET"},
		{"PUT"},
		{"DELETE"},
	}
	for _, tt := range tests {
		testname := tt.method
		t.Run(testname, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, "/logout", nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handlerFixture := Handler{
				LoggedUsers: model.LoggedUsers{
					Users: make(model.Users),
				},
			}
			handler := http.HandlerFunc(handlerFixture.logout)

			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusMethodNotAllowed {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, http.StatusMethodNotAllowed)
			}

			expected := "Invalid request method"
			received := strings.TrimSpace(rr.Body.String())
			if received != expected {
				t.Errorf("handler returned unexpected body: got %v want %v",
					received, expected)
			}

			handlerFixture.LoggedUsers.RLock()
			defer handlerFixture.LoggedUsers.RUnlock()
			if len(handlerFixture.LoggedUsers.Users) != 0 {
				t.Errorf("The list of logged users should be empty")
			}
		})
	}
}

func TestLogoutRequestBodyMissing(t *testing.T) {
	req, err := http.NewRequest("POST", "/logout", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handlerFixture := Handler{
		LoggedUsers: model.LoggedUsers{
			Users: make(model.Users),
		},
	}
	handler := http.HandlerFunc(handlerFixture.logout)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}

	expected := "Request body missing"
	received := strings.TrimSpace(rr.Body.String())
	if received != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			received, expected)
	}

	handlerFixture.LoggedUsers.RLock()
	defer handlerFixture.LoggedUsers.RUnlock()
	if len(handlerFixture.LoggedUsers.Users) != 0 {
		t.Errorf("The list of logged users should be empty")
	}
}

func TestLogoutRequestCanNotDecodeBody(t *testing.T) {
	req, err := http.NewRequest("POST", "/logout", strings.NewReader("invalid json"))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handlerFixture := Handler{
		LoggedUsers: model.LoggedUsers{
			Users: make(model.Users),
		},
	}
	handler := http.HandlerFunc(handlerFixture.logout)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}

	expected := "Can't decode body"
	received := strings.TrimSpace(rr.Body.String())
	if received != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			received, expected)
	}

	handlerFixture.LoggedUsers.RLock()
	defer handlerFixture.LoggedUsers.RUnlock()
	if len(handlerFixture.LoggedUsers.Users) != 0 {

		t.Errorf("The list of logged users should be empty")
	}
}

func TestLogoutUserNotLoggedIn(t *testing.T) {
	req, err := http.NewRequest("POST", "/logout", strings.NewReader(`{"username": "user", "token": "some-token"}`))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handlerFixture := Handler{
		LoggedUsers: model.LoggedUsers{
			Users: make(model.Users),
		},
	}
	handler := http.HandlerFunc(handlerFixture.logout)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusConflict {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusConflict)
	}

	expected := "User user is not logged in"
	received := strings.TrimSpace(rr.Body.String())
	if received != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			received, expected)
	}

	handlerFixture.LoggedUsers.RLock()
	defer handlerFixture.LoggedUsers.RUnlock()
	if len(handlerFixture.LoggedUsers.Users) != 0 {
		t.Errorf("There should be only one user in the list of logged users")
	}
}

func TestLogoutSuccess(t *testing.T) {
	req, err := http.NewRequest("POST", "/logout", strings.NewReader(`{"username": "user", "token": "some-token"}`))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handlerFixture := Handler{
		LoggedUsers: model.LoggedUsers{
			Users: model.Users{
				"user": model.User{
					Username: "user",
					Token:    "some-token",
				},
			},
		},
	}
	handler := http.HandlerFunc(handlerFixture.logout)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := `{"message":"User successfully logged out"}`
	received := strings.TrimSpace(rr.Body.String())
	if received != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			received, expected)
	}

	handlerFixture.LoggedUsers.RLock()
	defer handlerFixture.LoggedUsers.RUnlock()
	if _, ok := handlerFixture.LoggedUsers.Users["user"]; ok {
		t.Errorf("User should be removed from the list of logged users")
	}

	if len(handlerFixture.LoggedUsers.Users) != 0 {
		t.Errorf("There should be no users in the list of logged users")
	}
}

func TestLogoutInvalidToken(t *testing.T) {
	req, err := http.NewRequest("POST", "/logout", strings.NewReader(`{"username": "user", "token": "invalid-token"}`))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handlerFixture := Handler{
		LoggedUsers: model.LoggedUsers{
			Users: model.Users{
				"user": model.User{
					Username: "user",
					Token:    "some-token",
				},
			},
		},
	}
	handler := http.HandlerFunc(handlerFixture.logout)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusConflict {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusConflict)
	}

	expected := "Invalid token"
	received := strings.TrimSpace(rr.Body.String())
	if received != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			received, expected)
	}

	handlerFixture.LoggedUsers.RLock()
	defer handlerFixture.LoggedUsers.RUnlock()
	if _, ok := handlerFixture.LoggedUsers.Users["user"]; !ok {
		t.Errorf("User should not be removed from the list of logged users")
	}

	if len(handlerFixture.LoggedUsers.Users) != 1 {
		t.Errorf("There should be only one user in the list of logged users")
	}
}
