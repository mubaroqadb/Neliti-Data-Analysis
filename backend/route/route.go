package route

import (
	"fmt"
	"net/http"

	"github.com/research-data-analysis/config"
	"github.com/research-data-analysis/controller"
	"github.com/research-data-analysis/helper/at"
)

// URL adalah router utama untuk semua endpoint
func URL(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	if config.SetAccessControlHeaders(w, r) {
		return // Preflight request
	}
	
	// Load configuration di awal
	cfg := config.LoadConfig()
	
	// Health check endpoint untuk monitoring
	if r.Method == "GET" && r.URL.Path == "/health" {
		if err := cfg.ConfigurationHealthCheck(); err != nil {
			http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "healthy", "environment": "` + cfg.App.Environment + `"`))
		return
	}
	
	// Config info endpoint untuk debugging (hanya di development)
	if r.Method == "GET" && r.URL.Path == "/config" {
		if !cfg.App.Debug {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		config.PrintConfigInfo()
		return
	}
	
	method := r.Method
	path := r.URL.Path

	// Route matching
	switch {
	// Root endpoint
	case method == "GET" && path == "/":
		controller.GetHome(w, r)

	// Authentication endpoints
	case method == "POST" && path == "/auth/register":
		controller.Register(w, r)
	case method == "POST" && path == "/auth/login":
		controller.Login(w, r)
	case method == "GET" && path == "/auth/profile":
		controller.GetProfile(w, r)

	// Project endpoints
	case method == "POST" && path == "/api/project":
		controller.CreateProject(w, r)
	case method == "GET" && path == "/api/project":
		controller.GetAllProjects(w, r)
	case method == "GET" && at.URLParam(path, "/api/project/:id"):
		projectID := at.GetURLParam(path, "/api/project/:id", "id")
		controller.GetProject(w, r, projectID)
	case method == "PUT" && at.URLParam(path, "/api/project/:id"):
		projectID := at.GetURLParam(path, "/api/project/:id", "id")
		controller.UpdateProject(w, r, projectID)
	case method == "DELETE" && at.URLParam(path, "/api/project/:id"):
		projectID := at.GetURLParam(path, "/api/project/:id", "id")
		controller.DeleteProject(w, r, projectID)

	// Upload endpoints
	case method == "POST" && at.URLParam(path, "/api/upload/:projectId"):
		projectID := at.GetURLParam(path, "/api/upload/:projectId", "projectId")
		controller.UploadData(w, r, projectID)
	case method == "GET" && at.URLParam(path, "/api/preview/:uploadId"):
		uploadID := at.GetURLParam(path, "/api/preview/:uploadId", "uploadId")
		controller.GetDataPreview(w, r, uploadID)
	case method == "GET" && at.URLParam(path, "/api/stats/:uploadId"):
		uploadID := at.GetURLParam(path, "/api/stats/:uploadId", "uploadId")
		controller.GetDataStats(w, r, uploadID)

	// Analysis endpoints
	case method == "POST" && at.URLParam(path, "/api/recommend/:projectId"):
		projectID := at.GetURLParam(path, "/api/recommend/:projectId", "projectId")
		controller.GetRecommendations(w, r, projectID)
	case method == "POST" && at.URLParam(path, "/api/process/:analysisId"):
		analysisID := at.GetURLParam(path, "/api/process/:analysisId", "analysisId")
		controller.ProcessAnalysis(w, r, analysisID)
	case method == "GET" && at.URLParam(path, "/api/results/:analysisId"):
		analysisID := at.GetURLParam(path, "/api/results/:analysisId", "analysisId")
		controller.GetAnalysis(w, r, analysisID)
	case method == "POST" && at.URLParam(path, "/api/refine/:analysisId"):
		analysisID := at.GetURLParam(path, "/api/refine/:analysisId", "analysisId")
		controller.RefineAnalysis(w, r, analysisID)

	// Export endpoint
	case method == "GET" && at.URLParam(path, "/api/export/:analysisId"):
		analysisID := at.GetURLParam(path, "/api/export/:analysisId", "analysisId")
		controller.ExportResults(w, r, analysisID)

	// 404 Not Found
	default:
		NotFound(w, r)
	}
}

// NotFound handler untuk 404 responses
func NotFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(`{"error": "Route not found", "method": "` + r.Method + `", "path": "` + r.URL.Path + `"`))
}