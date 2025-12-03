package controller

import (
    "encoding/json"
    "fmt"
    "net/http"
    "time"

    "github.com/research-data-analysis/config"
    "github.com/research-data-analysis/helper/at"
    "github.com/research-data-analysis/helper/atdb"
    "github.com/research-data-analysis/helper/vertexai"
    "github.com/research-data-analysis/model"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

// GetRecommendations handler untuk mendapat rekomendasi berdasarkan mode analysis
func GetRecommendations(w http.ResponseWriter, r *http.Request, projectIDStr string) {
    userID, err := getUserIDFromToken(r)
    if err != nil || userID == primitive.NilObjectID {
        at.WriteJSON(w, http.StatusUnauthorized, model.Response{
            Status:   "error",
            Message:  "Unauthorized",
        })
        return
    }

    projectID, err := primitive.ObjectIDFromHex(projectIDStr)
    if err != nil {
        at.WriteJSON(w, http.StatusBadRequest, model.Response{
            Status:   "error",
            Message:  "Invalid project ID",
        })
        return
    }

    // Verify dan ambil project
    project, err := atdb.GetOneDoc[model.Project](
        config.Mongoconn,
        "projects",
        bson.M{"_id": projectID, "user_id": userID},
    )
    if err != nil {
        at.WriteJSON(w, http.StatusNotFound, model.Response{
            Status:   "error",
            Message:  "Project not found or unauthorized",
        })
        return
    }

    var req model.RecommendationRequest
    json.NewDecoder(r.Body).Decode(&req)

    // Ambil upload data jika ada
    var uploadData model.Upload
    if req.UploadID != "" {
        uploadID, _ := primitive.ObjectIDFromHex(req.UploadID)
        uploadData, _ = atdb.GetOneDoc[model.Upload](config.Mongoconn, "uploads", bson.M{"_id": uploadID})
    } else {
        // Ambil upload terbaik dari project
        uploads, _ := atdb.GetAllDocWithSort[model.Upload](
            config.Mongoconn,
            "uploads",
            bson.M{"project_id": projectID},
            bson.D{{Key: "uploaded_at", Value: -1}},
        )
        if len(uploads) > 0 {
            uploadData = uploads[0]
        }
    }

    // Build context untuk Vertex AI
    context := buildResearchContext(project, uploadData, req.Context)

    // Panggil Vertex AI untuk rekomendasi
    aiResponse, err := vertexai.GenerateResearchRecommendations(context)
    if err != nil {
        at.WriteJSON(w, http.StatusInternalServerError, model.Response{
            Status:   "error",
            Message:  "Failed to generate recommendations: " + err.Error(),
        })
        return
    }

    // Parse response dari AI
    recommendations, err := parseRecommendations(aiResponse)
    if err != nil {
        // Return raw response jika parsing gagal
        at.WriteJSON(w, http.StatusOK, model.Response{
            Status:   "success",
            Message:  "Recommendations generated (raw format)",
            Data: map[string]interface{}{
                "raw_response": aiResponse,
            },
        })
        return
    }

    // Buat analysis record
    analysis := model.Analysis{
        ProjectID:     projectID,
        UploadID:      uploadData.ID,
        Iteration:     1,
        Status:        "pending",
        Recommendations: recommendations,
        CreatedAt:     time.Now(),
    }

    // Check if there's existing analysis
    existingAnalyses, _ := atdb.GetAllDoc[model.Analysis](
        config.Mongoconn,
        "analyses",
        bson.M{"project_id": projectID},
    )
    if len(existingAnalyses) > 0 {
        analysis.Iteration = len(existingAnalyses) + 1
    }

    analysisID, err := atdb.InsertOneDoc(config.Mongoconn, "analyses", analysis)
    if err != nil {
        at.WriteJSON(w, http.StatusInternalServerError, model.Response{
            Status:   "error",
            Message:  "Failed to save recommendations",
        })
        return
    }

    analysis.ID = analysisID
    logAudit(userID, "recommended", "analyses", "Generated recommendations for project: "+project.Title, r)

    at.WriteJSON(w, http.StatusOK, model.Response{
        Status:   "success",
        Message:  "Recommendations generated successfully",
        Data: map[string]interface{}{
            "analysis_id":     analysisID.Hex(),
            "recommendations": recommendations,
        },
    })
}

// ProcessAnalysis handler untuk memproses analysis dalam mode yang berbeda
func ProcessAnalysis(w http.ResponseWriter, r *http.Request) {
    userID, err := getUserIDFromToken(r)
    if err != nil || userID == primitive.NilObjectID {
        at.WriteJSON(w, http.StatusUnauthorized, model.Response{
            Status:   "error",
            Message:  "Unauthorized",
        })
        return
    }

    var req model.ProcessRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        at.WriteJSON(w, http.StatusBadRequest, model.Response{
            Status:   "error",
            Message:  "Invalid request body",
        })
        return
    }

    analysisID, err := primitive.ObjectIDFromHex(req.AnalysisID)
    if err != nil {
        at.WriteJSON(w, http.StatusBadRequest, model.Response{
            Status:   "error",
            Message:  "Invalid analysis ID",
        })
        return
    }

    // Get analysis
    analysis, err := atdb.GetOneDoc[model.Analysis](config.Mongoconn, "analyses", bson.M{"_id": analysisID})
    if err != nil {
        at.WriteJSON(w, http.StatusNotFound, model.Response{
            Status:   "error",
            Message:  "Analysis not found",
        })
        return
    }

    // Verify ownership
    project, err := atdb.GetOneDoc[model.Project](
        config.Mongoconn,
        "projects",
        bson.M{"_id": analysis.ProjectID, "user_id": userID},
    )
    if err != nil {
        at.WriteJSON(w, http.StatusForbidden, model.Response{
            Status:   "error",
            Message:  "Access denied",
        })
        return
    }

    if len(req.SelectedMethods) == 0 {
        at.WriteJSON(w, http.StatusBadRequest, model.Response{
            Status:   "error",
            Message:  "No methods selected",
        })
        return
    }

    // Update status to processing
    atdb.UpdateOneDoc(config.Mongoconn, "analyses", bson.M{"_id": analysisID}, bson.M{
        "status":           "processing",
        "selected_methods": req.SelectedMethods,
    })
    atdb.UpdateOneDoc(config.Mongoconn, "projects", bson.M{"_id": project.ID}, bson.M{
        "status":   "analyzing",
        "updated_at": time.Now(),
    })

    // Get upload data for context
    uploadData, _ := atdb.GetOneDoc[model.Upload](config.Mongoconn, "uploads", bson.M{"_id": analysis.UploadID})

    // Process each selected method (simulate analysis)
    var results []model.MethodResult
    for _, method := range req.SelectedMethods {
        result := processMethod(method, project, uploadData)
        results = append(results, result)
    }

    // Generate summary using AI
    summaryContext := buildSummaryContext(project, results)
    summaryResponse, err := vertexai.GenerateResearchSummary(summaryContext)
    
    summary := ""
    if err == nil {
        summary = extractSummaryFromAI(summaryResponse)
    }

    // Update analysis with results
    now := time.Now()
    atdb.UpdateOneDoc(config.Mongoconn, "analyses", bson.M{"_id": analysisID}, bson.M{
        "status":     "completed",
        "results":    results,
        "summary":    summary,
        "completed_at": now,
    })

    // Update project status
    atdb.UpdateOneDoc(config.Mongoconn, "projects", bson.M{"_id": project.ID}, bson.M{
        "status":     "completed",
        "updated_at": now,
    })

    logAudit(userID, "process", "analyses", fmt.Sprintf("Processed analysis for project: %s with methods: %v", project.Title, req.SelectedMethods), r)

    at.WriteJSON(w, http.StatusOK, model.Response{
        Status:   "success",
        Message:  "Analysis completed successfully",
        Data: map[string]interface{}{
            "analysis_id": analysisID.Hex(),
            "results":     results,
            "summary":     summary,
        },
    })
}

// GetAnalysisResults handler untuk mendapat hasil analysis
func GetAnalysisResults(w http.ResponseWriter, r *http.Request, analysisIDStr string) {
    userID, err := getUserIDFromToken(r)
    if err != nil || userID == primitive.NilObjectID {
        at.WriteJSON(w, http.StatusUnauthorized, model.Response{
            Status:   "error",
            Message:  "Unauthorized",
        })
        return
    }

    analysisID, err := primitive.ObjectIDFromHex(analysisIDStr)
    if err != nil {
        at.WriteJSON(w, http.StatusBadRequest, model.Response{
            Status:   "error",
            Message:  "Invalid analysis ID",
        })
        return
    }

    analysis, err := atdb.GetOneDoc[model.Analysis](config.Mongoconn, "analyses", bson.M{"_id": analysisID})
    if err != nil {
        at.WriteJSON(w, http.StatusNotFound, model.Response{
            Status:   "error",
            Message:  "Analysis not found",
        })
        return
    }

    // Verify ownership
    _, err = atdb.GetOneDoc[model.Project](
        config.Mongoconn,
        "projects",
        bson.M{"_id": analysis.ProjectID, "user_id": userID},
    )
    if err != nil {
        at.WriteJSON(w, http.StatusForbidden, model.Response{
            Status:   "error",
            Message:  "Access denied",
        })
        return
    }

    at.WriteJSON(w, http.StatusOK, model.Response{
        Status:   "success",
        Message:  "Analysis results retrieved successfully",
        Data:     analysis,
    })
}

// RefineAnalysis handler untuk refine analysis berdasarkan feedback
func RefineAnalysis(w http.ResponseWriter, r *http.Request, analysisIDStr string) {
    userID, err := getUserIDFromToken(r)
    if err != nil || userID == primitive.NilObjectID {
        at.WriteJSON(w, http.StatusUnauthorized, model.Response{
            Status:   "error",
            Message:  "Unauthorized",
        })
        return
    }

    analysisID, err := primitive.ObjectIDFromHex(analysisIDStr)
    if err != nil {
        at.WriteJSON(w, http.StatusBadRequest, model.Response{
            Status:   "error",
            Message:  "Invalid analysis ID",
        })
        return
    }

    var req model.RefineRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        at.WriteJSON(w, http.StatusBadRequest, model.Response{
            Status:   "error",
            Message:  "Invalid request body",
        })
        return
    }

    analysis, err := atdb.GetOneDoc[model.Analysis](config.Mongoconn, "analyses", bson.M{"_id": analysisID})
    if err != nil {
        at.WriteJSON(w, http.StatusNotFound, model.Response{
            Status:   "error",
            Message:  "Analysis not found",
        })
        return
    }

    // Verify ownership
    project, err := atdb.GetOneDoc[model.Project](
        config.Mongoconn,
        "projects",
        bson.M{"_id": analysis.ProjectID, "user_id": userID},
    )
    if err != nil {
        at.WriteJSON(w, http.StatusForbidden, model.Response{
            Status:   "error",
            Message:  "Access denied",
        })
        return
    }

    // Store user feedback
    atdb.UpdateOneDoc(config.Mongoconn, "analyses", bson.M{"_id": analysisID}, bson.M{
        "user_feedback": req.Feedback,
    })

    // Generate new recommendations based on feedback
    uploadData, _ := atdb.GetOneDoc[model.Upload](config.Mongoconn, "uploads", bson.M{"_id": analysis.UploadID})

    context := buildResearchContext(project, uploadData, "")
    context += fmt.Sprintf("\n\nPrevious Analysis Feedback: %s\nAdjustments requested: %s", req.Feedback, req.Adjustments)

    if len(req.NewMethods) > 0 {
        context += fmt.Sprintf("\nUser specifically wants to try: %v", req.NewMethods)
    }

    aiResponse, err := vertexai.GenerateResearchRecommendations(context)
    if err != nil {
        at.WriteJSON(w, http.StatusInternalServerError, model.Response{
            Status:   "error",
            Message:  "Failed to generate refined recommendations",
        })
        return
    }

    recommendations, _ := parseRecommendations(aiResponse)

    // Create new iteration
    newAnalysis := model.Analysis{
        ProjectID:     project.ID,
        UploadID:      analysis.UploadID,
        Iteration:     analysis.Iteration + 1,
        Status:        "pending",
        Recommendations: recommendations,
        UserFeedback:  req.Feedback,
        CreatedAt:     time.Now(),
    }

    newAnalysisID, err := atdb.InsertOneDoc(config.Mongoconn, "analyses", newAnalysis)
    if err != nil {
        at.WriteJSON(w, http.StatusInternalServerError, model.Response{
            Status:   "error",
            Message:  "Failed to save refined analysis",
        })
        return
    }

    logAudit(userID, "refine", "analyses", fmt.Sprintf("Refined analysis for project: %s, iteration: %d", project.Title, newAnalysis.Iteration), r)

    at.WriteJSON(w, http.StatusOK, model.Response{
        Status:   "success",
        Message:  "Analysis refined successfully",
        Data: map[string]interface{}{
            "analysis_id":  newAnalysisID.Hex(),
            "iteration":    newAnalysis.Iteration,
            "recommendations": recommendations,
        },
    })
}

// Helper functions

func buildResearchContext(project model.Project, upload model.Upload, additionalContext string) string {
    context := fmt.Sprintf(`
Judul Penelitian: %s
Deskripsi: %s
Jenis Penelitian: %s
Hipotesis: %s
Variabel Independen: %v
Variabel Dependen: %v
`, project.Title, project.Description, project.ResearchType, project.Hypothesis,
        project.Variables.Independent, project.Variables.Dependent)

    if upload.ID != primitive.NilObjectID {
        context += fmt.Sprintf(`
Data yang tersedia:
- Jumlah baris: %d
- Jumlah kolom: %d
- Nama kolom: %v
- Tipe data: %v
`, upload.DataSummary.Rows, upload.DataSummary.Columns, 
            upload.DataSummary.ColumnNames, upload.DataSummary.ColumnTypes)
    }

    if additionalContext != "" {
        context += "\nKonteks tambahan: " + additionalContext
    }

    return context
}

func parseRecommendations(aiResponse string) ([]model.Recommendation, error) {
    var result struct {
        Recommendations []model.Recommendation `json:"recommendations"`
    }

    // Try to extract JSON from response
    start := -1
    end := -1
    braceCount := 0

    for i, c := range aiResponse {
        if c == '{' {
            if start == -1 {
                start = i
            }
            braceCount++
        } else if c == '}' {
            braceCount--
            if braceCount == 0 {
                end = i + 1
                break
            }
        }
    }

    if start != -1 && end != -1 {
        jsonStr := aiResponse[start:end]
        if err := json.Unmarshal([]byte(jsonStr), &result); err == nil {
            return result.Recommendations, nil
        }
    }

    return nil, fmt.Errorf("could not parse recommendations from AI response")
}

func processMethod(method string, project model.Project, upload model.Upload) model.MethodResult {
    // Simulate statistical analysis based on method
    // In real implementation, this would call actual statistical libraries
    
    result := model.MethodResult{
        Method: method,
        RawOutput: map[string]interface{}{
            "note": "Analysis results would be computed here",
        },
    }

    // Generate interpretation using AI
    interpretationPrompt := fmt.Sprintf("Method: %s, Project: %s, Data columns: %v", 
        method, project.Title, upload.DataSummary.ColumnNames)

    interpretation, err := vertexai.GenerateAnalysisInterpretation(method, interpretationPrompt)
    if err == nil {
        result.Interpretation = interpretation
    } else {
        result.Interpretation = fmt.Sprintf("Analysis %s masih dalam tahap pengembangan. Silakan interpretasikan hasil analisis berdasarkan data statistik yang tersedia.", method)
    }

    result.Conclusion = "Kesimpulan akan dapat ditentukan berdasarkan hasil interpretasi hasil analisis."

    return result
}

func buildSummaryContext(project model.Project, results []model.MethodResult) string {
    context := fmt.Sprintf("Judul Penelitian: %s\nHipotesis: %s\n\nHasil Analisis:\n",
        project.Title, project.Hypothesis)

    for _, r := range results {
        context += fmt.Sprintf("- %s: %s\n", r.Method, r.Interpretation)
    }

    return context
}

func extractSummaryFromAI(response string) string {
    var result struct {
        ExecutiveSummary string `json:"executive_summary"`
    }

    // Try to extract JSON
    start := -1
    end := -1
    braceCount := 0

    for i, c := range response {
        if c == '{' {
            if start == -1 {
                start = i
            }
            braceCount++
        } else if c == '}' {
            braceCount--
            if braceCount == 0 {
                end = i + 1
                break
            }
        }
    }

    if start != -1 && end != -1 {
        jsonStr := response[start:end]
        if err := json.Unmarshal([]byte(jsonStr), &result); err == nil {
            return result.ExecutiveSummary
        }
    }

    return response
}

func getUserIDFromToken(r *http.Request) (primitive.ObjectID, error) {
    // Get token from Authorization header
    token := r.Header.Get("Authorization")
    if token == "" {
        token = r.Header.Get("Login")
    }

    if token == "" {
        return primitive.NilObjectID, fmt.Errorf("no token found")
    }

    // Validate token and get user ID
    // This would use the actual JWT validation logic
    // For now, return a mock user ID for testing
    return primitive.ObjectID{}, fmt.Errorf("token validation not implemented")
}

func logAudit(userID primitive.ObjectID, action, collection, description string, r *http.Request) {
    // Log audit trail
    // Implementation would store audit logs in database
}
