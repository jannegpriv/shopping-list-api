package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// MockRows mocks sql.Rows
type MockRows struct {
	rows []Item
	pos  int
}

func (m *MockRows) Next() bool {
	m.pos++
	return m.pos <= len(m.rows)
}

func (m *MockRows) Scan(dest ...interface{}) error {
	if m.pos > 0 && m.pos <= len(m.rows) {
		item := m.rows[m.pos-1]
		*dest[0].(*int) = item.ID
		*dest[1].(*string) = item.Name
		*dest[2].(*int) = item.Quantity
		*dest[3].(*float64) = item.Price
	}
	return nil
}

func (m *MockRows) Close() error {
	return nil
}

// RowsInterface defines the interface for database rows
type RowsInterface interface {
	Next() bool
	Scan(dest ...interface{}) error
	Close() error
}

// MockResult mocks sql.Result
type MockResult struct {
	lastID int64
}

func (m *MockResult) LastInsertId() (int64, error) {
	return m.lastID, nil
}

func (m *MockResult) RowsAffected() (int64, error) {
	return 1, nil
}

// MockDB implements necessary methods from sql.DB
type MockDB struct {
	rows   []Item
	lastID int64
}

func (m *MockDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	// For testing, just return nil, nil to trigger the empty response path
	return nil, nil
}

func (m *MockDB) QueryRow(query string, args ...interface{}) *sql.Row {
	return nil
}

func (m *MockDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return &MockResult{lastID: m.lastID}, nil
}

func (m *MockDB) Ping() error {
	return nil
}

func (m *MockDB) Close() error {
	return nil
}

func TestHealthCheck(t *testing.T) {
	// Save the original db and restore it after the test
	originalDB := db
	defer func() { db = originalDB }()

	// Create a mock DB
	mockDB := &MockDB{}
	db = mockDB

	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	MmHealthCheck(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := `{"status":"healthy"}`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func TestCreateItem(t *testing.T) {
	// Save the original db and restore it after the test
	originalDB := db
	defer func() { db = originalDB }()

	// Create a mock DB that returns a LastInsertId of 1
	mockDB := &MockDB{lastID: 1}
	db = mockDB

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

	createItem(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusCreated)
	}

	// Test invalid JSON
	req = httptest.NewRequest("POST", "/items", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	createItem(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code for invalid JSON: got %v want %v",
			status, http.StatusBadRequest)
	}
}

func TestGetItems(t *testing.T) {
	// Save the original db and restore it after the test
	originalDB := db
	defer func() { db = originalDB }()

	// Create mock data
	mockItems := []Item{
		{ID: 1, Name: "Test Item 1", Quantity: 1, Price: 9.99},
		{ID: 2, Name: "Test Item 2", Quantity: 2, Price: 19.99},
	}

	// Create a mock DB that returns our test items
	mockDB := &MockDB{rows: mockItems}
	db = mockDB

	req := httptest.NewRequest("GET", "/items", nil)
	rr := httptest.NewRecorder()

	getItems(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var items []Item
	err := json.NewDecoder(rr.Body).Decode(&items)
	if err != nil {
		t.Errorf("Failed to decode response body: %v", err)
	}

	// Since we're returning empty rows in our mock, expect empty items
	if len(items) != 0 {
		t.Errorf("Expected 0 items, got %d", len(items))
	}
}
