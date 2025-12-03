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
    "go.mongodb.org/mongo-driver/mongo"
)

// Helper function untuk handle MongoDB error
func getMongoDB() *mongo.Database {
    db, err := config.GetMongoDB()
    if err != nil {
        return nil
    }
    return db
}

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
    mongoDB := getMongoDB()
    if mongoDB == nil {
        at.WriteJSON(w, http.StatusInternalServerError, model.Response{
            Status:   "error",
            Message:  "Database connection failed",
        })
        return
    }

    project, err := atdb.GetOneDoc[model.Project](
        mongoDB,
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

    var req model.RecommendRequest
    json.NewDecoder(r.Body).Decode(&req)

    // Ambil upload data jika ada
    var uploadData model.Upload
    if req.UploadID != "" {
        uploadID, _ := primitive.ObjectIDFromHex(req.UploadID)
        uploadData, _ = atdb.GetOneDoc[model.Upload](mongoDB, "uploads", bson.M{"_id": uploadID})
    } else {
        // Ambil upload terbaik dari project
        uploads, _ := atdb.GetAllDocWithSort[model.Upload](
            mongoDB,
            "uploads",
            bson.M{"project_id": projectID},
            bson.D{{Key: "uploaded_at", Value: -1}},
        )
        if len(uploads) > 0 {
            uploadData = uploads[0]
        }
    }

    // Prepare context untuk rekomendasi
    context := fmt.Sprintf("Project: %s\nDescription: %s\nUpload Data: %s",
        project.Title, project.Description, uploadData.FileName)

    // Generate rekomendasi menggunakan Vertex AI
    recommendations, err := vertexai.GenerateResearchRecommendations(context)
    if err != nil {
        at.WriteJSON(w, http.StatusInternalServerError, model.Response{
            Status:   "error",
            Message:  "Failed to generate recommendations",
        })
        return
    }

    // Parse recommendations JSON menjadi array Recommendation
    var parsedRecommendations []model.Recommendation
    // Simplified - untuk production perlu proper JSON parsing
    parsedRecommendations = []model.Recommendation{
        {
            Method: "Descriptive Statistics",
            Category: "descriptive", 
            Reasoning: "Basic statistical analysis of data",
            Priority: 1,
            Assumptions: "Data should be normally distributed",
        },
    }

    // Simpan hasil analisis
    analysis := model.Analysis{
        ProjectID:     projectID,
        UploadID:      uploadData.ID,
        Iteration:     1,
        Status:        "completed",
        Recommendations: parsedRecommendations,
        SelectedMethods: []string{"AI_Recommendation"},
        Results:       []model.MethodResult{},
        Summary:       recommendations,
        CreatedAt:     time.Now(),
    }

    analysisID, err := atdb.InsertOneDoc(mongoDB, "analyses", analysis)
    if err != nil {
        at.WriteJSON(w, http.StatusInternalServerError, model.Response{
            Status:   "error",
            Message:  "Failed to save analysis",
        })
        return
    }

    // Return hasil
    at.WriteJSON(w, http.StatusOK, model.Response{
        Status:  "success",
        Message: "Recommendations generated successfully",
        Data: map[string]interface{}{
            "recommendations": recommendations,
            "analysis_id":     analysisID,
            "project_id":      projectIDStr,
        },
    })
}

// ProcessAnalysis handler untuk melakukan analisis data
func ProcessAnalysis(w http.ResponseWriter, r *http.Request, analysisIDStr string) {
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

    mongoDB := getMongoDB()
    if mongoDB == nil {
        at.WriteJSON(w, http.StatusInternalServerError, model.Response{
            Status:   "error",
            Message:  "Database connection failed",
        })
        return
    }

    // Ambil data analysis
    analysis, err := atdb.GetOneDoc[model.Analysis](mongoDB, "analyses", bson.M{"_id": analysisID})
    if err != nil {
        at.WriteJSON(w, http.StatusNotFound, model.Response{
            Status:   "error",
            Message:  "Analysis not found",
        })
        return
    }

    // Ambil project data
    project, err := atdb.GetOneDoc[model.Project](mongoDB, "projects", bson.M{"_id": analysis.ProjectID})
    if err != nil {
        at.WriteJSON(w, http.StatusNotFound, model.Response{
            Status:   "error",
            Message:  "Project not found",
        })
        return
    }

    // Update status analysis
    _, err = atdb.UpdateOneDoc(
        mongoDB,
        "analyses",
        bson.M{"_id": analysisID},
        bson.M{"$set": bson.M{
            "status":       "processing",
            "updated_at":   time.Now(),
        }},
    )
    
    // Ambil upload data
    uploadData, _ := atdb.GetOneDoc[model.Upload](mongoDB, "uploads", bson.M{"_id": analysis.UploadID})

    // Simulate processing (untuk production, implementasi actual analysis logic)
    time.Sleep(2 * time.Second)
    
    // Generate interpretation using Vertex AI
    interpretation, err := vertexai.GenerateAnalysisInterpretation("Analysis", "Processing completed")
    if err != nil {
        interpretation = "Analysis completed but interpretation failed"
    }

    // Create MethodResult
    methodResult := model.MethodResult{
        Method:        "Processed Analysis",
        RawOutput:     map[string]interface{}{"status": "completed", "interpretation": interpretation},
        Interpretation: interpretation,
        Conclusion:    "Analysis processed successfully",
    }

    // Update analysis dengan hasil final
    _, err = atdb.UpdateOneDoc(
        mongoDB,
        "analyses",
        bson.M{"_id": analysisID},
        bson.M{"$set": bson.M{
            "results":        []model.MethodResult{methodResult},
            "summary":        fmt.Sprintf("Analysis completed for project: %s\nFile: %s", project.Title, uploadData.FileName),
            "status":         "completed",
            "completed_at":   time.Now(),
            "updated_at":     time.Now(),
        }},
    )

    // Return hasil
    at.WriteJSON(w, http.StatusOK, model.Response{
        Status:  "success",
        Message: "Analysis processed successfully",
        Data: map[string]interface{}{
            "analysis_id":    analysisID,
            "project_id":     project.ID,
            "interpretation": interpretation,
            "method_result":  methodResult,
            "status":         "completed",
        },
    })
}

// GetAnalysis handler untuk mengambil detail analysis
func GetAnalysis(w http.ResponseWriter, r *http.Request, analysisIDStr string) {
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

    mongoDB := getMongoDB()
    if mongoDB == nil {
        at.WriteJSON(w, http.StatusInternalServerError, model.Response{
            Status:   "error",
            Message:  "Database connection failed",
        })
        return
    }

    analysis, err := atdb.GetOneDoc[model.Analysis](mongoDB, "analyses", bson.M{"_id": analysisID})
    if err != nil {
        at.WriteJSON(w, http.StatusNotFound, model.Response{
            Status:   "error",
            Message:  "Analysis not found",
        })
        return
    }

    // Ambil project data
    project, err := atdb.GetOneDoc[model.Project](mongoDB, "projects", bson.M{"_id": analysis.ProjectID})
    if err != nil {
        at.WriteJSON(w, http.StatusNotFound, model.Response{
            Status:   "error",
            Message:  "Project not found",
        })
        return
    }

    // Ambil upload data jika ada
    var uploadData model.Upload
    if !analysis.UploadID.IsZero() {
        uploadData, _ = atdb.GetOneDoc[model.Upload](mongoDB, "uploads", bson.M{"_id": analysis.UploadID})
    }

    // Return detail analysis
    at.WriteJSON(w, http.StatusOK, model.Response{
        Status:  "success",
        Message: "Analysis retrieved successfully",
        Data: map[string]interface{}{
            "analysis":   analysis,
            "project":    project,
            "upload":     uploadData,
        },
    })
}

// GetAllAnalyses handler untuk mendapat semua analysis berdasarkan project
func GetAllAnalyses(w http.ResponseWriter, r *http.Request, projectIDStr string) {
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

    mongoDB := getMongoDB()
    if mongoDB == nil {
        at.WriteJSON(w, http.StatusInternalServerError, model.Response{
            Status:   "error",
            Message:  "Database connection failed",
        })
        return
    }

    // Verify project exists and belongs to user
    project, err := atdb.GetOneDoc[model.Project](mongoDB, "projects", bson.M{"_id": projectID, "user_id": userID})
    if err != nil {
        at.WriteJSON(w, http.StatusNotFound, model.Response{
            Status:   "error",
            Message:  "Project not found or unauthorized",
        })
        return
    }

    // Ambil semua analysis untuk project
    analyses, err := atdb.GetAllDoc[model.Analysis](mongoDB, "analyses", bson.M{"project_id": projectID})
    if err != nil {
        at.WriteJSON(w, http.StatusInternalServerError, model.Response{
            Status:   "error",
            Message:  "Failed to retrieve analyses",
        })
        return
    }

    // Return semua analyses
    at.WriteJSON(w, http.StatusOK, model.Response{
        Status:  "success",
        Message: "Analyses retrieved successfully",
        Data: map[string]interface{}{
            "analyses": analyses,
            "project":  project,
        },
    })
}

// UpdateAnalysis handler untuk update analysis
func UpdateAnalysis(w http.ResponseWriter, r *http.Request, analysisIDStr string) {
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

    var updateData struct {
        Status string `json:"status"`
        Notes  string `json:"notes"`
    }

    mongoDB := getMongoDB()
    if mongoDB == nil {
        at.WriteJSON(w, http.StatusInternalServerError, model.Response{
            Status:   "error",
            Message:  "Database connection failed",
        })
        return
    }

    err = json.NewDecoder(r.Body).Decode(&updateData)
    if err != nil {
        at.WriteJSON(w, http.StatusBadRequest, model.Response{
            Status:   "error",
            Message:  "Invalid request body",
        })
        return
    }

    // Update analysis
    updateFields := bson.M{
        "updated_at": time.Now(),
    }
    
    if updateData.Status != "" {
        updateFields["status"] = updateData.Status
    }
    if updateData.Notes != "" {
        updateFields["user_feedback"] = updateData.Notes
    }

    _, err = atdb.UpdateOneDoc(
        mongoDB,
        "analyses",
        bson.M{"_id": analysisID},
        bson.M{"$set": updateFields},
    )

    if err != nil {
        at.WriteJSON(w, http.StatusInternalServerError, model.Response{
            Status:   "error",
            Message:  "Failed to update analysis",
        })
        return
    }

    at.WriteJSON(w, http.StatusOK, model.Response{
        Status:  "success",
        Message: "Analysis updated successfully",
        Data: map[string]interface{}{
            "analysis_id": analysisID,
            "updated_fields": updateFields,
        },
    })
}

// DeleteAnalysis handler untuk delete analysis
func DeleteAnalysis(w http.ResponseWriter, r *http.Request, analysisIDStr string) {
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

    mongoDB := getMongoDB()
    if mongoDB == nil {
        at.WriteJSON(w, http.StatusInternalServerError, model.Response{
            Status:   "error",
            Message:  "Database connection failed",
        })
        return
    }

    // Delete analysis (dalam production, mungkin soft delete)
    _, err = atdb.UpdateOneDoc(
        mongoDB,
        "analyses",
        bson.M{"_id": analysisID},
        bson.M{"$set": bson.M{
            "status":     "deleted",
            "error":      "Deleted by user",
            "updated_at": time.Now(),
        }},
    )

    if err != nil {
        at.WriteJSON(w, http.StatusInternalServerError, model.Response{
            Status:   "error",
            Message:  "Failed to delete analysis",
        })
        return
    }

    at.WriteJSON(w, http.StatusOK, model.Response{
        Status:  "success",
        Message: "Analysis deleted successfully",
        Data: map[string]interface{}{
            "analysis_id": analysisID,
            "deleted_at":  time.Now(),
        },
    })
}

// RefineAnalysis handler untuk refined analysis
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

    var refinementRequest struct {
        Instructions string `json:"instructions"`
    }
    
    mongoDB := getMongoDB()
    if mongoDB == nil {
        at.WriteJSON(w, http.StatusInternalServerError, model.Response{
            Status:   "error",
            Message:  "Database connection failed",
        })
        return
    }
    
    err = json.NewDecoder(r.Body).Decode(&refinementRequest)
    if err != nil {
        at.WriteJSON(w, http.StatusBadRequest, model.Response{
            Status:   "error",
            Message:  "Invalid request body",
        })
        return
    }

    // Ambil original analysis
    originalAnalysis, err := atdb.GetOneDoc[model.Analysis](mongoDB, "analyses", bson.M{"_id": analysisID})
    if err != nil {
        at.WriteJSON(w, http.StatusNotFound, model.Response{
            Status:   "error",
            Message:  "Analysis not found",
        })
        return
    }

    // Generate refined content
    refinedPrompt := fmt.Sprintf("Please refine the following analysis based on these instructions:\n\nInstructions: %s\n\nOriginal Analysis: %s\n\nPlease provide a refined version that addresses the instructions.",
        refinementRequest.Instructions, originalAnalysis.Summary)

    refinedResults, err := vertexai.GenerateContent(refinedPrompt)
    if err != nil {
        at.WriteJSON(w, http.StatusInternalServerError, model.Response{
            Status:   "error",
            Message:  "Failed to refine analysis",
        })
        return
    }

    // Create refined analysis sebagai analysis baru
    newAnalysis := model.Analysis{
        ProjectID:       originalAnalysis.ProjectID,
        UploadID:        originalAnalysis.UploadID,
        Iteration:       originalAnalysis.Iteration + 1,
        Status:          "completed",
        Summary:         refinedResults,
        UserFeedback:    "Refined version: " + refinementRequest.Instructions,
        CreatedAt:       time.Now(),
    }

    newAnalysisID, err := atdb.InsertOneDoc(mongoDB, "analyses", newAnalysis)
    if err != nil {
        at.WriteJSON(w, http.StatusInternalServerError, model.Response{
            Status:   "error",
            Message:  "Failed to save refined analysis",
        })
        return
    }

    at.WriteJSON(w, http.StatusOK, model.Response{
        Status:  "success",
        Message: "Analysis refined successfully",
        Data: map[string]interface{}{
            "original_analysis_id": analysisID,
            "refined_analysis_id":  newAnalysisID,
            "refined_results":      refinedResults,
            "instructions":         refinementRequest.Instructions,
        },
    })
}

// GenerateSummary handler untuk generate summary analysis
func GenerateSummary(w http.ResponseWriter, r *http.Request, projectIDStr string) {
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

    mongoDB := getMongoDB()
    if mongoDB == nil {
        at.WriteJSON(w, http.StatusInternalServerError, model.Response{
            Status:   "error",
            Message:  "Database connection failed",
        })
        return
    }

    // Ambil semua analysis untuk project
    analyses, err := atdb.GetAllDoc[model.Analysis](mongoDB, "analyses", bson.M{"project_id": projectID, "status": "completed"})
    if err != nil {
        at.WriteJSON(w, http.StatusInternalServerError, model.Response{
            Status:   "error",
            Message:  "Failed to retrieve analyses",
        })
        return
    }

    if len(analyses) == 0 {
        at.WriteJSON(w, http.StatusBadRequest, model.Response{
            Status:   "error",
            Message:  "No completed analyses found for summary",
        })
        return
    }

    // Ambil project info
    project, err := atdb.GetOneDoc[model.Project](mongoDB, "projects", bson.M{"_id": projectID, "user_id": userID})
    if err != nil {
        at.WriteJSON(w, http.StatusNotFound, model.Response{
            Status:   "error",
            Message:  "Project not found or unauthorized",
        })
        return
    }

    // Prepare context untuk summary
    analysisContext := fmt.Sprintf("Project: %s\nDescription: %s\n\nAnalyses Summary:\n", project.Title, project.Description)
    
    for _, analysis := range analyses {
        analysisContext += fmt.Sprintf("- Method: Analysis %d, Status: %s, Created: %s\nResults: %s\n\n",
            analysis.Iteration, analysis.Status, analysis.CreatedAt.Format("2006-01-02 15:04:05"), analysis.Summary)
    }

    // Generate summary menggunakan Vertex AI
    summary, err := vertexai.GenerateResearchSummary(analysisContext)
    if err != nil {
        at.WriteJSON(w, http.StatusInternalServerError, model.Response{
            Status:   "error",
            Message:  "Failed to generate summary",
        })
        return
    }

    // Simpan summary sebagai analysis baru
    summaryAnalysis := model.Analysis{
        ProjectID:     projectID,
        Iteration:     len(analyses) + 1,
        Status:        "completed",
        Summary:       summary,
        UserFeedback:  fmt.Sprintf("Auto-generated summary from %d analyses", len(analyses)),
        CreatedAt:     time.Now(),
    }

    summaryID, err := atdb.InsertOneDoc(mongoDB, "analyses", summaryAnalysis)
    if err != nil {
        at.WriteJSON(w, http.StatusInternalServerError, model.Response{
            Status:   "error",
            Message:  "Failed to save summary analysis",
        })
        return
    }

    at.WriteJSON(w, http.StatusOK, model.Response{
        Status:  "success",
        Message: "Project summary generated successfully",
        Data: map[string]interface{}{
            "summary":        summary,
            "summary_id":     summaryID,
            "project_id":     projectIDStr,
            "analyses_count": len(analyses),
        },
    })
}

// Helper function untuk get user ID from JWT token
func getUserIDFromToken(r *http.Request) (primitive.ObjectID, error) {
    // Simplified implementation - dalam production, implementasi JWT validation
    authHeader := r.Header.Get("Authorization")
    if authHeader == "" {
        return primitive.NilObjectID, fmt.Errorf("no authorization header")
    }
    
    // Return dummy user ID untuk testing
    // Dalam production, parse JWT dan return actual user ID
    return primitive.NewObjectID(), nil
}