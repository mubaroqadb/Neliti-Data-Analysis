package controller

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/research-data-analysis/helper/at"
	"github.com/research-data-analysis/helper/atdb"
	"github.com/research-data-analysis/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ExportResults handler untuk mengekspor hasil analisis
func ExportResults(w http.ResponseWriter, r *http.Request, analysisIDStr string) {
	userID, err := getUserIDFromToken(r)
	if err != nil || userID == primitive.NilObjectID {
		at.WriteJSON(w, http.StatusUnauthorized, model.Response{
			Status:  "error",
			Message: "Unauthorized",
		})
		return
	}

	analysisID, err := primitive.ObjectIDFromHex(analysisIDStr)
	if err != nil {
		at.WriteJSON(w, http.StatusBadRequest, model.Response{
			Status:  "error",
			Message: "Invalid analysis ID",
		})
		return
	}

	// Get export format from query parameter
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "pdf"
	}

	mongoDB := getMongoDB()
	if mongoDB == nil {
		at.WriteJSON(w, http.StatusInternalServerError, model.Response{
			Status:  "error",
			Message: "Database connection failed",
		})
		return
	}

	analysis, err := atdb.GetOneDoc[model.Analysis](mongoDB, "analyses", bson.M{"_id": analysisID})
	if err != nil {
		at.WriteJSON(w, http.StatusNotFound, model.Response{
			Status:  "error",
			Message: "Analysis not found",
		})
		return
	}

	// Verify ownership
	project, err := atdb.GetOneDoc[model.Project](
		mongoDB,
		"projects",
		bson.M{"_id": analysis.ProjectID, "user_id": userID},
	)
	if err != nil {
		at.WriteJSON(w, http.StatusForbidden, model.Response{
			Status:  "error",
			Message: "Access denied",
		})
		return
	}

	if analysis.Status != "completed" {
		at.WriteJSON(w, http.StatusBadRequest, model.Response{
			Status:  "error",
			Message: "Analysis not yet completed",
		})
		return
	}

	switch format {
	case "pdf":
		exportPDF(w, project, analysis)
	case "csv":
		exportCSV(w, project, analysis)
	case "json":
		exportJSON(w, project, analysis)
	default:
		at.WriteJSON(w, http.StatusBadRequest, model.Response{
			Status:  "error",
			Message: "Invalid format. Supported: pdf, csv, json",
		})
	}
}

func exportPDF(w http.ResponseWriter, project model.Project, analysis model.Analysis) {
	// Generate PDF content (simplified - in production use a PDF library like gofpdf)
	content := fmt.Sprintf(`
LAPORAN ANALISIS PENELITIAN
===========================

Judul: %s
Deskripsi: %s
Jenis Penelitian: %s
Hipotesis: %s

VARIABEL PENELITIAN
-------------------
Independen: %v
Dependen: %v

HASIL ANALISIS
--------------
`, project.Title, project.Description, project.ResearchType, project.Hypothesis,
		project.Variables.Independent, project.Variables.Dependent)

	for i, result := range analysis.Results {
		content += fmt.Sprintf(`
%d. %s
   Interpretasi: %s
   Kesimpulan: %s
`, i+1, result.Method, result.Interpretation, result.Conclusion)
	}

	content += fmt.Sprintf(`

RINGKASAN
---------
%s

---
Dibuat pada: %v
Iterasi ke: %d
`, analysis.Summary, analysis.CompletedAt, analysis.Iteration)

	// Send as text/plain for now (in production, generate actual PDF)
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"analysis_report_%s.pdf\"", analysis.ID.Hex()))
	w.Write([]byte(content))
}

func exportCSV(w http.ResponseWriter, project model.Project, analysis model.Analysis) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Header
	writer.Write([]string{"Metode", "Interpretasi", "Kesimpulan", "Effect Size"})

	// Data
	for _, result := range analysis.Results {
		writer.Write([]string{
			result.Method,
			result.Interpretation,
			result.Conclusion,
			result.EffectSize,
		})
	}

	writer.Flush()

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"analysis_results_%s.csv\"", analysis.ID.Hex()))
	w.Write(buf.Bytes())
}

func exportJSON(w http.ResponseWriter, project model.Project, analysis model.Analysis) {
	exportData := map[string]interface{}{
		"project": map[string]interface{}{
			"id":            project.ID.Hex(),
			"title":         project.Title,
			"description":   project.Description,
			"research_type": project.ResearchType,
			"hypothesis":    project.Hypothesis,
			"variables":     project.Variables,
		},
		"analysis": map[string]interface{}{
			"id":           analysis.ID.Hex(),
			"iteration":    analysis.Iteration,
			"status":       analysis.Status,
			"results":      analysis.Results,
			"summary":      analysis.Summary,
			"created_at":   analysis.CreatedAt,
			"completed_at": analysis.CompletedAt,
		},
	}

	jsonData, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		at.WriteJSON(w, http.StatusInternalServerError, model.Response{
			Status:  "error",
			Message: "Failed to generate JSON export",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"analysis_export_%s.json\"", analysis.ID.Hex()))
	w.Write(jsonData)
}
