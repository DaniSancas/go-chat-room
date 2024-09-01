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
}
