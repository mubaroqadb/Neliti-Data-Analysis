package controller

import (
	"net/http"
	"./helper/at"
	"./model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Response handler untuk responses konsisten
func Response(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	at.WriteJSON(w, status, data)
}

// GetAllProjects handler untuk mengambil semua projects user
func GetAllProjects(w http.ResponseWriter, r *http.Request) {
	// Implementation for getting all projects
	// This would use the actual authentication and database logic
	response := model.Response{
		Status:  "success",
		Message: "Projects retrieved successfully",
		Data:    []interface{}{},
	}
	Response(w, http.StatusOK, response)
}

// GetProject handler untuk mengambil satu project berdasarkan ID
func GetProject(w http.ResponseWriter, r *http.Request, projectIDStr string) {
	// Implementation for getting a specific project
	response := model.Response{
		Status:  "success",
		Message: "Project retrieved successfully",
		Data:    nil,
	}
	Response(w, http.StatusOK, response)
}

// CreateProject handler untuk membuat project baru
func CreateProject(w http.ResponseWriter, r *http.Request) {
	// Implementation for creating a new project
	response := model.Response{
		Status:  "success",
		Message: "Project created successfully",
		Data:    nil,
	}
	Response(w, http.StatusCreated, response)
}

// UpdateProject handler untuk update project
func UpdateProject(w http.ResponseWriter, r *http.Request, projectIDStr string) {
	// Implementation for updating a project
	response := model.Response{
		Status:  "success",
		Message: "Project updated successfully",
		Data:    nil,
	}
	Response(w, http.StatusOK, response)
}

// DeleteProject handler untuk hapus project
func DeleteProject(w http.ResponseWriter, r *http.Request, projectIDStr string) {
	// Implementation for deleting a project
	response := model.Response{
		Status:  "success",
		Message: "Project deleted successfully",
		Data:    nil,
	}
	Response(w, http.StatusOK, response)
}

// HealthCheck handler untuk health check endpoint
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	// Implementation for health check
	response := map[string]interface{}{
		"status":  "healthy",
		"message": "Service is running",
		"time":    primitive.NewObjectID().Timestamp(),
	}
	Response(w, http.StatusOK, response)
}