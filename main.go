package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// User represents a user in our system
type User struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Created string `json:"created"`
}

// Response represents a standard API response
type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// In-memory storage for demo purposes
var users []User
var nextID = 1

// Health check endpoint
func healthHandler(w http.ResponseWriter, r *http.Request) {
	response := Response{
		Status:  "success",
		Message: "API is healthy",
		Data: map[string]interface{}{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"version":   "1.0.0",
			"service":   "go-backend-api",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Get all users
func getUsersHandler(w http.ResponseWriter, r *http.Request) {
	response := Response{
		Status:  "success",
		Message: "Users retrieved successfully",
		Data:    users,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Get user by ID
func getUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	for _, user := range users {
		if fmt.Sprintf("%d", user.ID) == userID {
			response := Response{
				Status:  "success",
				Message: "User found",
				Data:    user,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
	response := Response{
		Status:  "error",
		Message: "User not found",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Create new user
func createUserHandler(w http.ResponseWriter, r *http.Request) {
	var newUser struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response := Response{
			Status:  "error",
			Message: "Invalid JSON payload",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	if newUser.Name == "" || newUser.Email == "" {
		w.WriteHeader(http.StatusBadRequest)
		response := Response{
			Status:  "error",
			Message: "Name and email are required",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	user := User{
		ID:      nextID,
		Name:    newUser.Name,
		Email:   newUser.Email,
		Created: time.Now().UTC().Format(time.RFC3339),
	}

	users = append(users, user)
	nextID++

	w.WriteHeader(http.StatusCreated)
	response := Response{
		Status:  "success",
		Message: "User created successfully",
		Data:    user,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Delete user
func deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	for i, user := range users {
		if fmt.Sprintf("%d", user.ID) == userID {
			users = append(users[:i], users[i+1:]...)
			response := Response{
				Status:  "success",
				Message: "User deleted successfully",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
	response := Response{
		Status:  "error",
		Message: "User not found",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	// Initialize with some sample data
	users = []User{
		{ID: 1, Name: "John Doe", Email: "john@example.com", Created: time.Now().UTC().Format(time.RFC3339)},
		{ID: 2, Name: "Jane Smith", Email: "jane@example.com", Created: time.Now().UTC().Format(time.RFC3339)},
	}
	nextID = 3

	router := mux.NewRouter()

	// API routes
	api := router.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/health", healthHandler).Methods("GET")
	api.HandleFunc("/users", getUsersHandler).Methods("GET")
	api.HandleFunc("/users/{id:[0-9]+}", getUserHandler).Methods("GET")
	api.HandleFunc("/users", createUserHandler).Methods("POST")
	api.HandleFunc("/users/{id:[0-9]+}", deleteUserHandler).Methods("DELETE")

	// CORS middleware
	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)(router)

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Printf("Health check available at: http://localhost:%s/api/v1/health", port)

	if err := http.ListenAndServe(":"+port, corsHandler); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
