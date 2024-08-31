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
	received := strings.TrimSpace(rr.Body.String(),)
	if received != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			received, expected)
	}
}