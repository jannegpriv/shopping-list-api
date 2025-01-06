package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthCheck(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(healthCheck)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := `{"status":"ok"}`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func TestCreateItem(t *testing.T) {
	// Test valid item
	validItem := Item{
		Name:     "Test Item",
		Quantity: 1,
		Price:    9.99,
	}
	body, _ := json.Marshal(validItem)

	req := httptest.NewRequest("POST", "/items", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// Create a mock DB handler that doesn't actually connect to the database
	mockDB := &DBHandler{} // This will use nil db, but we're not testing DB operations
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		createItem(w, r, mockDB)
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Test invalid JSON
	req = httptest.NewRequest("POST", "/items", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code for invalid JSON: got %v want %v",
			status, http.StatusBadRequest)
	}
}

func TestGetItems(t *testing.T) {
	req := httptest.NewRequest("GET", "/items", nil)
	rr := httptest.NewRecorder()

	mockDB := &DBHandler{} // This will use nil db, but we're not testing DB operations
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		getItems(w, r, mockDB)
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}
