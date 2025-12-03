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
        config.GetMongoDB(),
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
        uploadData, _ = atdb.GetOneDoc[model.Upload](config.GetMongoDB(), "uploads", bson.M{"_id": uploadID})
    } else {
        // Ambil upload terbaik dari project
        uploads, _ := atdb.GetAllDocWithSort[model.Upload](
            config.GetMongoDB(),
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
        project.Name, project.Description, uploadData.FileName)

    // Generate rekomendasi menggunakan Vertex AI
    recommendations, err := vertexai.GenerateResearchRecommendations(context)
    if err != nil {
        at.WriteJSON(w, http.StatusInternalServerError, model.Response{
            Status:   "error",
            Message:  "Failed to generate recommendations",
        })
        return
    }

    // Simpan hasil analisis
    analysis := model.Analysis{
        ProjectID:   projectID,
        UploadID:    uploadData.ID,
        Method:      "AI_Recommendation",
        Results:     recommendations,
        CreatedAt:   time.Now(),
        CreatedBy:   userID,
        Status:      "completed",
    }

    analysisID, err := atdb.InsertOneDoc(config.GetMongoDB(), "analyses", analysis)
    if err != nil {
        at.WriteJSON(w, http.StatusInternalServerError, model.Response{
            Status:   "error",
            Message:  "Failed to save analysis",
        })
        return
    }

    // Update status project
    project.LastAnalysisAt = time.Now()
    project.AnalysisCount++
    project.UpdatedAt = time.Now()
    
    err = atdb.UpdateOneDoc(
        config.GetMongoDB(),
        "projects",
        bson.M{"_id": projectID},
        bson.M{"$set": bson.M{
            "last_analysis_at": project.LastAnalysisAt,
            "analysis_count":   project.AnalysisCount,
            "updated_at":       project.UpdatedAt,
        }},
    )

    // Return hasil
    at.WriteJSON(w, http.StatusOK, model.Response{
        Status:  "success",
        Message: "Recommendations generated successfully",
        Data: map[string]interface{}{
            "recommendations": recommendations,
            "analysis_id":    analysisID,
            "project_id":     projectIDStr,
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

    // Ambil data analysis
    analysis, err := atdb.GetOneDoc[model.Analysis](config.GetMongoDB(), "analyses", bson.M{"_id": analysisID})
    if err != nil {
        at.WriteJSON(w, http.StatusNotFound, model.Response{
            Status:   "error",
            Message:  "Analysis not found",
        })
        return
    }

    // Ambil project data
    project, err := atdb.GetOneDoc[model.Project](config.GetMongoDB(), "projects", bson.M{"_id": analysis.ProjectID})
    if err != nil {
        at.WriteJSON(w, http.StatusNotFound, model.Response{
            Status:   "error",
            Message:  "Project not found",
        })
        return
    }

    // Update status analysis
    err = atdb.UpdateOneDoc(
        config.GetMongoDB(),
        "analyses",
        bson.M{"_id": analysisID},
        bson.M{"$set": bson.M{
            "status":       "processing",
            "updated_at":   time.Now(),
        }},
    )
    
    err = atdb.UpdateOneDoc(
        config.GetMongoDB(),
        "projects",
        bson.M{"_id": project.ID},
        bson.M{"$set": bson.M{
            "last_activity_at": time.Now(),
            "updated_at":       time.Now(),
        }},
    )

    // Ambil upload data
    uploadData, _ := atdb.GetOneDoc[model.Upload](config.GetMongoDB(), "uploads", bson.M{"_id": analysis.UploadID})

    // Simulate processing (untuk production, implementasi actual analysis logic)
    time.Sleep(2 * time.Second)
    
    // Update hasil analysis
    results := fmt.Sprintf("Analysis completed for project: %s\nFile: %s\nMethod: %s",
        project.Name, uploadData.FileName, analysis.Method)

    // Generate interpretation using Vertex AI
    interpretation, err := vertexai.GenerateAnalysisInterpretation(analysis.Method, results)
    if err != nil {
        interpretation = "Analysis completed but interpretation failed"
    }

    // Update analysis dengan hasil final
    err = atdb.UpdateOneDoc(
        config.GetMongoDB(),
        "analyses",
        bson.M{"_id": analysisID},
        bson.M{"$set": bson.M{
            "results":        results,
            "interpretation": interpretation,
            "status":         "completed",
            "completed_at":   time.Now(),
            "updated_at":     time.Now(),
        }},
    )
    
    err = atdb.UpdateOneDoc(
        config.GetMongoDB(),
        "projects",
        bson.M{"_id": project.ID},
        bson.M{"$set": bson.M{
            "last_activity_at": time.Now(),
            "updated_at":       time.Now(),
        }},
    )

    // Return hasil
    at.WriteJSON(w, http.StatusOK, model.Response{
        Status:  "success",
        Message: "Analysis processed successfully",
        Data: map[string]interface{}{
            "analysis_id":    analysisID,
            "project_id":     project.ID,
            "results":        results,
            "interpretation": interpretation,
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

    analysis, err := atdb.GetOneDoc[model.Analysis](config.GetMongoDB(), "analyses", bson.M{"_id": analysisID})
    if err != nil {
        at.WriteJSON(w, http.StatusNotFound, model.Response{
            Status:   "error",
            Message:  "Analysis not found",
        })
        return
    }

    // Ambil project data
    project, err := atdb.GetOneDoc[model.Project](config.GetMongoDB(), "projects", bson.M{"_id": analysis.ProjectID})
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
        uploadData, _ = atdb.GetOneDoc[model.Upload](config.GetMongoDB(), "uploads", bson.M{"_id": analysis.UploadID})
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

    // Verify project exists and belongs to user
    project, err := atdb.GetOneDoc[model.Project](config.GetMongoDB(), "projects", bson.M{"_id": projectID, "user_id": userID})
    if err != nil {
        at.WriteJSON(w, http.StatusNotFound, model.Response{
            Status:   "error",
            Message:  "Project not found or unauthorized",
        })
        return
    }

    // Ambil semua analysis untuk project
    analyses, err := atdb.GetAllDoc[model.Analysis](config.GetMongoDB(), "analyses", bson.M{"project_id": projectID})
    if err != nil {
        at.WriteJSON(w, http.StatusInternalServerError, model.Response{
            Status:   "error",
            Message:  "Failed to retrieve analyses",
        })
        return
    }

    // Update project last activity
    atdb.UpdateOneDoc(
        config.GetMongoDB(),
        "projects",
        bson.M{"_id": projectID},
        bson.M{"$set": bson.M{
            "last_activity_at": time.Now(),
            "updated_at":       time.Now(),
        }},
    )

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
        updateFields["notes"] = updateData.Notes
    }

    err = atdb.UpdateOneDoc(
        config.GetMongoDB(),
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

    // Ambil upload data jika ada
    analysis, err := atdb.GetOneDoc[model.Analysis](config.GetMongoDB(), "analyses", bson.M{"_id": analysisID})
    if err == nil && !analysis.UploadID.IsZero() {
        atdb.UpdateOneDoc(
            config.GetMongoDB(),
            "uploads",
            bson.M{"_id": analysis.UploadID},
            bson.M{"$set": bson.M{"updated_at": time.Now()}},
        )
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

    // Ambil analysis untuk verifikasi
    analysis, err := atdb.GetOneDoc[model.Analysis](config.GetMongoDB(), "analyses", bson.M{"_id": analysisID})
    if err != nil {
        at.WriteJSON(w, http.StatusNotFound, model.Response{
            Status:   "error",
            Message:  "Analysis not found",
        })
        return
    }

    // Delete analysis (dalam production, mungkin soft delete)
    // Untuk sekarang, kita update status menjadi deleted
    err = atdb.UpdateOneDoc(
        config.GetMongoDB(),
        "analyses",
        bson.M{"_id": analysisID},
        bson.M{"$set": bson.M{
            "status":     "deleted",
            "deleted_at": time.Now(),
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

    // Update project analysis count
    project, err := atdb.GetOneDoc[model.Project](config.GetMongoDB(), "projects", bson.M{"_id": analysis.ProjectID})
    if err == nil {
        if project.AnalysisCount > 0 {
            project.AnalysisCount--
        }
        project.UpdatedAt = time.Now()
        
        atdb.UpdateOneDoc(
            config.GetMongoDB(),
            "projects",
            bson.M{"_id": analysis.ProjectID},
            bson.M{"$set": bson.M{
                "analysis_count": project.AnalysisCount,
                "updated_at":     project.UpdatedAt,
            }},
        )
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
    
    err = json.NewDecoder(r.Body).Decode(&refinementRequest)
    if err != nil {
        at.WriteJSON(w, http.StatusBadRequest, model.Response{
            Status:   "error",
            Message:  "Invalid request body",
        })
        return
    }

    // Ambil original analysis
    originalAnalysis, err := atdb.GetOneDoc[model.Analysis](config.GetMongoDB(), "analyses", bson.M{"_id": analysisID})
    if err != nil {
        at.WriteJSON(w, http.StatusNotFound, model.Response{
            Status:   "error",
            Message:  "Analysis not found",
        })
        return
    }

    // Create refined analysis dengan AI
    refinedPrompt := fmt.Sprintf("Please refine the following analysis based on these instructions:\n\nInstructions: %s\n\nOriginal Analysis: %s\n\nPlease provide a refined version that addresses the instructions.",
        refinementRequest.Instructions, originalAnalysis.Results)

    refinedResults, err := vertexai.GenerateContent(refinedPrompt)
    if err != nil {
        at.WriteJSON(w, http.StatusInternalServerError, model.Response{
            Status:   "error",
            Message:  "Failed to refine analysis",
        })
        return
    }

    // Simpan refined analysis sebagai analysis baru
    newAnalysis := model.Analysis{
        ProjectID:   originalAnalysis.ProjectID,
        UploadID:    originalAnalysis.UploadID,
        Method:      originalAnalysis.Method + "_Refined",
        Results:     refinedResults,
        CreatedAt:   time.Now(),
        CreatedBy:   userID,
        Status:      "completed",
        Notes:       "Refined version: " + refinementRequest.Instructions,
    }

    newAnalysisID, err := atdb.InsertOneDoc(config.GetMongoDB(), "analyses", newAnalysis)
    if err != nil {
        at.WriteJSON(w, http.StatusInternalServerError, model.Response{
            Status:   "error",
            Message:  "Failed to save refined analysis",
        })
        return
    }

    // Update project
    project, err := atdb.GetOneDoc[model.Project](config.GetMongoDB(), "projects", bson.M{"_id": originalAnalysis.ProjectID})
    if err == nil {
        project.LastAnalysisAt = time.Now()
        project.AnalysisCount++
        project.UpdatedAt = time.Now()
        
        atdb.UpdateOneDoc(
            config.GetMongoDB(),
            "projects",
            bson.M{"_id": originalAnalysis.ProjectID},
            bson.M{"$set": bson.M{
                "last_analysis_at": project.LastAnalysisAt,
                "analysis_count":   project.AnalysisCount,
                "updated_at":       project.UpdatedAt,
            }},
        )
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

    // Ambil semua analysis untuk project
    analyses, err := atdb.GetAllDoc[model.Analysis](config.GetMongoDB(), "analyses", bson.M{"project_id": projectID, "status": "completed"})
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
    project, err := atdb.GetOneDoc[model.Project](config.GetMongoDB(), "projects", bson.M{"_id": projectID, "user_id": userID})
    if err != nil {
        at.WriteJSON(w, http.StatusNotFound, model.Response{
            Status:   "error",
            Message:  "Project not found or unauthorized",
        })
        return
    }

    // Prepare context untuk summary
    analysisContext := fmt.Sprintf("Project: %s\nDescription: %s\n\nAnalyses Summary:\n", project.Name, project.Description)
    
    for _, analysis := range analyses {
        analysisContext += fmt.Sprintf("- Method: %s, Status: %s, Created: %s\nResults: %s\n\n",
            analysis.Method, analysis.Status, analysis.CreatedAt.Format("2006-01-02 15:04:05"), analysis.Results)
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
        ProjectID:   projectID,
        Method:      "Project_Summary",
        Results:     summary,
        CreatedAt:   time.Now(),
        CreatedBy:   userID,
        Status:      "completed",
        Notes:       fmt.Sprintf("Auto-generated summary from %d analyses", len(analyses)),
    }

    summaryID, err := atdb.InsertOneDoc(config.GetMongoDB(), "analyses", summaryAnalysis)
    if err != nil {
        at.WriteJSON(w, http.StatusInternalServerError, model.Response{
            Status:   "error",
            Message:  "Failed to save summary analysis",
        })
        return
    }

    // Update project
    project.LastAnalysisAt = time.Now()
    project.AnalysisCount++
    project.UpdatedAt = time.Now()
    
    atdb.UpdateOneDoc(
        config.GetMongoDB(),
        "projects",
        bson.M{"_id": projectID},
        bson.M{"$set": bson.M{
            "last_analysis_at": project.LastAnalysisAt,
            "analysis_count":   project.AnalysisCount,
            "updated_at":       project.UpdatedAt,
        }},
    )

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