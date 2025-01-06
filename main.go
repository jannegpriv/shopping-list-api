package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

type Item struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

// DB is an interface that captures the database operations we need
type DB interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
	Ping() error
	Close() error
}

var db DB

func main() {
	// Load environment variables - only in development
	if os.Getenv("ENV") != "production" {
		if err := godotenv.Load(); err != nil {
			log.Println("No .env file found")
		}
	}

	// Database connection
	dbUser := getEnv("DB_USER", "root")
	dbPass := getEnv("DB_PASSWORD", "password")
	dbHost := getEnv("DB_HOST", "mysql")
	dbName := getEnv("DB_NAME", "shopping_list")

	dsn := dbUser + ":" + dbPass + "@tcp(" + dbHost + ")/" + dbName + "?parseTime=true"
	log.Printf("Connecting to database: %s@tcp(%s)/%s", dbUser, dbHost, dbName)

	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatal("Database connection failed:", err)
	}
	log.Println("Successfully connected to database")

	// Create tables if they don't exist
	createTables()

	// Initialize router
	r := mux.NewRouter()

	// Health check endpoint for Kubernetes
	r.HandleFunc("/health", MmHealthCheck).Methods("GET")

	// API routes
	r.HandleFunc("/items", getItems).Methods("GET")
	r.HandleFunc("/items", createItem).Methods("POST")
	r.HandleFunc("/items/{id}", getItem).Methods("GET")
	r.HandleFunc("/items/{id}", updateItem).Methods("PUT")
	r.HandleFunc("/items/{id}", deleteItem).Methods("DELETE")

	// Start server
	port := getEnv("PORT", "8080")
	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func MmHealthCheck(w http.ResponseWriter, r *http.Request) {
	if err := db.Ping(); err != nil {
		respondWithError(w, http.StatusServiceUnavailable, "Database connection failed")
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func createTables() {
	query := `
		CREATE TABLE IF NOT EXISTS items (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			quantity INT NOT NULL,
			price DECIMAL(10,2) NOT NULL
		)
	`
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

func getItems(w http.ResponseWriter, r *http.Request) {
	var items []Item
	rows, err := db.Query("SELECT id, name, quantity, price FROM items")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if rows == nil {
		respondWithJSON(w, http.StatusOK, items)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.ID, &item.Name, &item.Quantity, &item.Price); err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, item)
	}

	respondWithJSON(w, http.StatusOK, items)
}

func createItem(w http.ResponseWriter, r *http.Request) {
	var item Item
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	result, err := db.Exec("INSERT INTO items (name, quantity, price) VALUES (?, ?, ?)",
		item.Name, item.Quantity, item.Price)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	item.ID = int(id)
	respondWithJSON(w, http.StatusCreated, item)
}

func getItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var item Item
	err := db.QueryRow("SELECT id, name, quantity, price FROM items WHERE id = ?",
		vars["id"]).Scan(&item.ID, &item.Name, &item.Quantity, &item.Price)
	if err == sql.ErrNoRows {
		respondWithError(w, http.StatusNotFound, "Item not found")
		return
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, item)
}

func updateItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var item Item
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	_, err := db.Exec("UPDATE items SET name = ?, quantity = ?, price = ? WHERE id = ?",
		item.Name, item.Quantity, item.Price, vars["id"])
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, item)
}

func deleteItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	_, err := db.Exec("DELETE FROM items WHERE id = ?", vars["id"])
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
