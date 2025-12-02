package controller

import (
	"net/http"

	"github.com/research-data-analysis/helper/at"
	"github.com/research-data-analysis/model"
)

// GetHome handler untuk endpoint root
func GetHome(w http.ResponseWriter, r *http.Request) {
	response := model.Response{
		Status:  "success",
		Message: "Research Data Analysis API v1.0",
		Data: map[string]interface{}{
			"version":   "1.0.0",
			"endpoints": []string{
				"POST /auth/register",
				"POST /auth/login",
				"GET /auth/profile",
				"POST /api/project",
				"GET /api/project",
				"PUT /api/project",
				"DELETE /api/project",
				"POST /api/upload/:projectId",
				"GET /api/preview/:uploadId",
				"GET /api/stats/:uploadId",
				"POST /api/recommend/:projectId",
				"POST /api/process",
				"GET /api/results/:analysisId",
				"POST /api/refine/:analysisId",
				"GET /api/export/:analysisId",
			},
		},
	}
	at.WriteJSON(w, http.StatusOK, response)
}

// NotFound handler untuk endpoint yang tidak ditemukan
func NotFound(w http.ResponseWriter, r *http.Request) {
	response := model.Response{
		Status:  "error",
		Message: "Endpoint not found",
		Data: map[string]string{
			"path":   r.URL.Path,
			"method": r.Method,
		},
	}
	at.WriteJSON(w, http.StatusNotFound, response)
}

// MethodNotAllowed handler untuk method yang tidak diizinkan
func MethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	response := model.Response{
		Status:  "error",
		Message: "Method not allowed",
		Data: map[string]string{
			"method": r.Method,
		},
	}
	at.WriteJSON(w, http.StatusMethodNotAllowed, response)
}