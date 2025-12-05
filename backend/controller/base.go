package controller

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/research-data-analysis/helper/at"
	"github.com/research-data-analysis/helper/atdb"
	"github.com/research-data-analysis/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Response handler untuk responses konsisten
func Response(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	at.WriteJSON(w, status, data)
}

// GetAllProjects handler untuk mendapat semua project user
func GetAllProjects(w http.ResponseWriter, r *http.Request) {
	// Get user ID from token
	userID, err := getUserIDFromToken(r)
	if err != nil || userID.IsZero() {
		Response(w, http.StatusUnauthorized, model.Response{
			Status:  "error",
			Message: "Unauthorized",
		})
		return
	}

	// Get MongoDB connection
	mongoDB := getMongoDB()
	if mongoDB == nil {
		Response(w, http.StatusInternalServerError, model.Response{
			Status:  "error",
			Message: "Database connection failed",
		})
		return
	}

	// Get all projects for this user
	projects, err := atdb.GetAllDoc[model.Project](mongoDB, "projects", bson.M{"user_id": userID})
	if err != nil {
		Response(w, http.StatusInternalServerError, model.Response{
			Status:  "error",
			Message: "Failed to retrieve projects",
		})
		return
	}

	Response(w, http.StatusOK, model.Response{
		Status:  "success",
		Message: "Projects retrieved successfully",
		Data:    projects,
	})
}

// GetProject handler untuk mendapat project berdasarkan ID
func GetProject(w http.ResponseWriter, r *http.Request, projectIDStr string) {
	// Get user ID from token
	userID, err := getUserIDFromToken(r)
	if err != nil || userID.IsZero() {
		Response(w, http.StatusUnauthorized, model.Response{
			Status:  "error",
			Message: "Unauthorized",
		})
		return
	}

	// Convert projectID string to ObjectID
	projectID, err := primitive.ObjectIDFromHex(projectIDStr)
	if err != nil {
		Response(w, http.StatusBadRequest, model.Response{
			Status:  "error",
			Message: "Invalid project ID",
		})
		return
	}

	// Get MongoDB connection
	mongoDB := getMongoDB()
	if mongoDB == nil {
		Response(w, http.StatusInternalServerError, model.Response{
			Status:  "error",
			Message: "Database connection failed",
		})
		return
	}

	// Get the project
	project, err := atdb.GetOneDoc[model.Project](mongoDB, "projects", bson.M{"_id": projectID, "user_id": userID})
	if err != nil {
		Response(w, http.StatusNotFound, model.Response{
			Status:  "error",
			Message: "Project not found",
		})
		return
	}

	Response(w, http.StatusOK, model.Response{
		Status:  "success",
		Message: "Project retrieved successfully",
		Data:    project,
	})
}

// CreateProject handler untuk membuat project baru
func CreateProject(w http.ResponseWriter, r *http.Request) {
	// Get user ID from token
	userID, err := getUserIDFromToken(r)
	if err != nil || userID.IsZero() {
		Response(w, http.StatusUnauthorized, model.Response{
			Status:  "error",
			Message: "Unauthorized",
		})
		return
	}

	// Parse request body
	var projectReq model.ProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&projectReq); err != nil {
		Response(w, http.StatusBadRequest, model.Response{
			Status:  "error",
			Message: "Invalid request body",
		})
		return
	}

	// Get MongoDB connection
	mongoDB := getMongoDB()
	if mongoDB == nil {
		Response(w, http.StatusInternalServerError, model.Response{
			Status:  "error",
			Message: "Database connection failed",
		})
		return
	}

	// Create new project
	newProject := model.Project{
		UserID:       userID,
		Title:        projectReq.Title,
		Description:  projectReq.Description,
		ResearchType: projectReq.ResearchType,
		Hypothesis:   projectReq.Hypothesis,
		Variables:    projectReq.Variables,
		Status:       "draft",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Insert project into database
	projectID, err := atdb.InsertOneDoc(mongoDB, "projects", newProject)
	if err != nil {
		Response(w, http.StatusInternalServerError, model.Response{
			Status:  "error",
			Message: "Failed to create project",
		})
		return
	}

	// Set the ID for response
	newProject.ID = projectID

	Response(w, http.StatusCreated, model.Response{
		Status:  "success",
		Message: "Project created successfully",
		Data:    newProject,
	})
}

// UpdateProject handler untuk update project
func UpdateProject(w http.ResponseWriter, r *http.Request, projectIDStr string) {
	// Get user ID from token
	userID, err := getUserIDFromToken(r)
	if err != nil || userID.IsZero() {
		Response(w, http.StatusUnauthorized, model.Response{
			Status:  "error",
			Message: "Unauthorized",
		})
		return
	}

	// Convert projectID string to ObjectID
	projectID, err := primitive.ObjectIDFromHex(projectIDStr)
	if err != nil {
		Response(w, http.StatusBadRequest, model.Response{
			Status:  "error",
			Message: "Invalid project ID",
		})
		return
	}

	// Parse request body
	var projectReq model.ProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&projectReq); err != nil {
		Response(w, http.StatusBadRequest, model.Response{
			Status:  "error",
			Message: "Invalid request body",
		})
		return
	}

	// Get MongoDB connection
	mongoDB := getMongoDB()
	if mongoDB == nil {
		Response(w, http.StatusInternalServerError, model.Response{
			Status:  "error",
			Message: "Database connection failed",
		})
		return
	}

	// Check if project exists and belongs to user
	_, err = atdb.GetOneDoc[model.Project](mongoDB, "projects", bson.M{"_id": projectID, "user_id": userID})
	if err != nil {
		Response(w, http.StatusNotFound, model.Response{
			Status:  "error",
			Message: "Project not found",
		})
		return
	}

	// Update project
	updateData := bson.M{
		"title":         projectReq.Title,
		"description":   projectReq.Description,
		"research_type": projectReq.ResearchType,
		"hypothesis":    projectReq.Hypothesis,
		"variables":     projectReq.Variables,
		"updated_at":    time.Now(),
	}

	_, err = atdb.UpdateOneDoc(mongoDB, "projects", bson.M{"_id": projectID}, updateData)
	if err != nil {
		Response(w, http.StatusInternalServerError, model.Response{
			Status:  "error",
			Message: "Failed to update project",
		})
		return
	}

	// Get updated project
	updatedProject, _ := atdb.GetOneDoc[model.Project](mongoDB, "projects", bson.M{"_id": projectID})

	Response(w, http.StatusOK, model.Response{
		Status:  "success",
		Message: "Project updated successfully",
		Data:    updatedProject,
	})
}

// DeleteProject handler untuk delete project
func DeleteProject(w http.ResponseWriter, r *http.Request, projectIDStr string) {
	// Get user ID from token
	userID, err := getUserIDFromToken(r)
	if err != nil || userID.IsZero() {
		Response(w, http.StatusUnauthorized, model.Response{
			Status:  "error",
			Message: "Unauthorized",
		})
		return
	}

	// Convert projectID string to ObjectID
	projectID, err := primitive.ObjectIDFromHex(projectIDStr)
	if err != nil {
		Response(w, http.StatusBadRequest, model.Response{
			Status:  "error",
			Message: "Invalid project ID",
		})
		return
	}

	// Get MongoDB connection
	mongoDB := getMongoDB()
	if mongoDB == nil {
		Response(w, http.StatusInternalServerError, model.Response{
			Status:  "error",
			Message: "Database connection failed",
		})
		return
	}

	// Check if project exists and belongs to user
	_, err = atdb.GetOneDoc[model.Project](mongoDB, "projects", bson.M{"_id": projectID, "user_id": userID})
	if err != nil {
		Response(w, http.StatusNotFound, model.Response{
			Status:  "error",
			Message: "Project not found",
		})
		return
	}

	// Delete project
	_, err = atdb.DeleteOneDoc(mongoDB, "projects", bson.M{"_id": projectID})
	if err != nil {
		Response(w, http.StatusInternalServerError, model.Response{
			Status:  "error",
			Message: "Failed to delete project",
		})
		return
	}

	Response(w, http.StatusOK, model.Response{
		Status:  "success",
		Message: "Project deleted successfully",
		Data: map[string]interface{}{
			"deleted_project_id": projectIDStr,
			"deleted_at":         time.Now(),
		},
	})
}

// HealthCheck handler untuk health check endpoint
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	// Implementation for health check
	response := map[string]interface{}{
		"status":    "healthy",
		"message":   "Service is running",
		"timestamp": primitive.NewObjectID().Timestamp(),
	}
	Response(w, http.StatusOK, response)
}

// GetHome handler untuk welcome message di root endpoint
func GetHome(w http.ResponseWriter, r *http.Request) {
	// Implementation for home endpoint
	response := model.Response{
		Status:  "success",
		Message: "Welcome to Data Analysis API",
		Data: map[string]interface{}{
			"version": "1.0.0",
			"endpoints": []string{
				"/health", "/projects", "/auth/register",
				"/auth/login", "/profile", "/upload", "/data/preview",
				"/data/stats", "/uploads", "/upload/{id}",
			},
		},
	}
	Response(w, http.StatusOK, response)
}

// Register handler untuk user registration
func Register(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var registerReq model.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&registerReq); err != nil {
		Response(w, http.StatusBadRequest, model.Response{
			Status:  "error",
			Message: "Invalid request body",
		})
		return
	}

	// Get MongoDB connection
	mongoDB := getMongoDB()
	if mongoDB == nil {
		Response(w, http.StatusInternalServerError, model.Response{
			Status:  "error",
			Message: "Database connection failed",
		})
		return
	}

	// Check if user already exists
	existingUser, _ := atdb.GetOneDoc[model.User](mongoDB, "users", bson.M{"email": registerReq.Email})
	if existingUser.Email != "" {
		Response(w, http.StatusConflict, model.Response{
			Status:  "error",
			Message: "User with this email already exists",
		})
		return
	}

	// Hash password (in production, use proper password hashing)
	hashedPassword := registerReq.Password // In production, hash this

	// Create new user
	newUser := model.User{
		Email:         registerReq.Email,
		Password:      hashedPassword,
		FullName:      registerReq.FullName,
		Institution:   registerReq.Institution,
		ResearchField: registerReq.ResearchField,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Insert user into database
	userID, err := atdb.InsertOneDoc(mongoDB, "users", newUser)
	if err != nil {
		Response(w, http.StatusInternalServerError, model.Response{
			Status:  "error",
			Message: "Failed to register user",
		})
		return
	}

	// Set the ID for response
	newUser.ID = userID

	// Remove password from response
	newUser.Password = ""

	Response(w, http.StatusCreated, model.Response{
		Status:  "success",
		Message: "User registered successfully",
		Data:    newUser,
	})
}

// Login handler untuk user login
func Login(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var loginReq model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		Response(w, http.StatusBadRequest, model.Response{
			Status:  "error",
			Message: "Invalid request body",
		})
		return
	}

	// Get MongoDB connection
	mongoDB := getMongoDB()
	if mongoDB == nil {
		Response(w, http.StatusInternalServerError, model.Response{
			Status:  "error",
			Message: "Database connection failed",
		})
		return
	}

	// Find user by email
	user, err := atdb.GetOneDoc[model.User](mongoDB, "users", bson.M{"email": loginReq.Email})
	if err != nil || user.Email == "" {
		Response(w, http.StatusUnauthorized, model.Response{
			Status:  "error",
			Message: "Invalid email or password",
		})
		return
	}

	// Check password (in production, use proper password verification)
	if user.Password != loginReq.Password { // In production, compare hashed passwords
		Response(w, http.StatusUnauthorized, model.Response{
			Status:  "error",
			Message: "Invalid email or password",
		})
		return
	}

	// Generate token (in production, use proper JWT/PASETO generation)
	token := "generated-token-placeholder" // In production, generate actual token

	// Remove password from response
	user.Password = ""

	Response(w, http.StatusOK, model.Response{
		Status:  "success",
		Message: "User logged in successfully",
		Data: map[string]interface{}{
			"token": token,
			"user":  user,
		},
	})
}

// GetProfile handler untuk mendapat user profile
func GetProfile(w http.ResponseWriter, r *http.Request) {
	// Get user ID from token
	userID, err := getUserIDFromToken(r)
	if err != nil || userID.IsZero() {
		Response(w, http.StatusUnauthorized, model.Response{
			Status:  "error",
			Message: "Unauthorized",
		})
		return
	}

	// Get MongoDB connection
	mongoDB := getMongoDB()
	if mongoDB == nil {
		Response(w, http.StatusInternalServerError, model.Response{
			Status:  "error",
			Message: "Database connection failed",
		})
		return
	}

	// Get user profile
	user, err := atdb.GetOneDoc[model.User](mongoDB, "users", bson.M{"_id": userID})
	if err != nil {
		Response(w, http.StatusNotFound, model.Response{
			Status:  "error",
			Message: "User not found",
		})
		return
	}

	// Remove password from response
	user.Password = ""

	Response(w, http.StatusOK, model.Response{
		Status:  "success",
		Message: "Profile retrieved successfully",
		Data:    user,
	})
}

// UploadData handler untuk file upload
func UploadData(w http.ResponseWriter, r *http.Request, projectIDStr string) {
	// Get user ID from token
	userID, err := getUserIDFromToken(r)
	if err != nil || userID.IsZero() {
		Response(w, http.StatusUnauthorized, model.Response{
			Status:  "error",
			Message: "Unauthorized",
		})
		return
	}

	// Convert projectID string to ObjectID
	projectID, err := primitive.ObjectIDFromHex(projectIDStr)
	if err != nil {
		Response(w, http.StatusBadRequest, model.Response{
			Status:  "error",
			Message: "Invalid project ID",
		})
		return
	}

	// Get MongoDB connection
	mongoDB := getMongoDB()
	if mongoDB == nil {
		Response(w, http.StatusInternalServerError, model.Response{
			Status:  "error",
			Message: "Database connection failed",
		})
		return
	}

	// Check if project exists and belongs to user
	_, err = atdb.GetOneDoc[model.Project](mongoDB, "projects", bson.M{"_id": projectID, "user_id": userID})
	if err != nil {
		Response(w, http.StatusNotFound, model.Response{
			Status:  "error",
			Message: "Project not found",
		})
		return
	}

	// Parse multipart form for file upload
	if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB max
		Response(w, http.StatusBadRequest, model.Response{
			Status:  "error",
			Message: "Failed to parse form",
		})
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		Response(w, http.StatusBadRequest, model.Response{
			Status:  "error",
			Message: "No file uploaded",
		})
		return
	}
	defer file.Close()

	// In production, you would:
	// 1. Save the file to Google Cloud Storage
	// 2. Process the file to extract data summary
	// 3. Store the file metadata in MongoDB

	// For now, create a placeholder upload record
	newUpload := model.Upload{
		ProjectID:  projectID,
		FileName:   handler.Filename,
		FileType:   handler.Header.Get("Content-Type"),
		FileSize:   handler.Size,
		StorageURL: "placeholder-storage-url", // In production, actual GCS URL
		DataSummary: model.DataSummary{
			Rows:        100,                                              // Placeholder
			Columns:     5,                                                // Placeholder
			ColumnNames: []string{"col1", "col2", "col3", "col4", "col5"}, // Placeholder
		},
		UploadedAt: time.Now(),
	}

	// Insert upload record into database
	uploadID, err := atdb.InsertOneDoc(mongoDB, "uploads", newUpload)
	if err != nil {
		Response(w, http.StatusInternalServerError, model.Response{
			Status:  "error",
			Message: "Failed to save upload record",
		})
		return
	}

	// Set the ID for response
	newUpload.ID = uploadID

	Response(w, http.StatusOK, model.Response{
		Status:  "success",
		Message: "File uploaded successfully",
		Data:    newUpload,
	})
}

// GetDataPreview handler untuk preview uploaded data
func GetDataPreview(w http.ResponseWriter, r *http.Request, uploadIDStr string) {
	// Get user ID from token
	userID, err := getUserIDFromToken(r)
	if err != nil || userID.IsZero() {
		Response(w, http.StatusUnauthorized, model.Response{
			Status:  "error",
			Message: "Unauthorized",
		})
		return
	}

	// Convert uploadID string to ObjectID
	uploadID, err := primitive.ObjectIDFromHex(uploadIDStr)
	if err != nil {
		Response(w, http.StatusBadRequest, model.Response{
			Status:  "error",
			Message: "Invalid upload ID",
		})
		return
	}

	// Get MongoDB connection
	mongoDB := getMongoDB()
	if mongoDB == nil {
		Response(w, http.StatusInternalServerError, model.Response{
			Status:  "error",
			Message: "Database connection failed",
		})
		return
	}

	// Get upload record
	upload, err := atdb.GetOneDoc[model.Upload](mongoDB, "uploads", bson.M{"_id": uploadID})
	if err != nil {
		Response(w, http.StatusNotFound, model.Response{
			Status:  "error",
			Message: "Upload not found",
		})
		return
	}

	// Check if user has access to this upload (through project ownership)
	_, err = atdb.GetOneDoc[model.Project](mongoDB, "projects", bson.M{"_id": upload.ProjectID, "user_id": userID})
	if err != nil {
		Response(w, http.StatusForbidden, model.Response{
			Status:  "error",
			Message: "Access denied",
		})
		return
	}

	// In production, you would:
	// 1. Fetch actual data from the stored file
	// 2. Parse and return a sample of rows
	// For now, return placeholder data

	Response(w, http.StatusOK, model.Response{
		Status:  "success",
		Message: "Data preview retrieved",
		Data: map[string]interface{}{
			"upload_id": uploadIDStr,
			"file_name": upload.FileName,
			"columns":   upload.DataSummary.ColumnNames,
			"sample_rows": []map[string]interface{}{
				{"col1": 1, "col2": "value1", "col3": 10.5, "col4": "cat1", "col5": true},
				{"col1": 2, "col2": "value2", "col3": 15.2, "col4": "cat2", "col5": false},
			},
			"total_rows": upload.DataSummary.Rows,
		},
	})
}

// GetDataStats handler untuk get data statistics
func GetDataStats(w http.ResponseWriter, r *http.Request, uploadIDStr string) {
	// Get user ID from token
	userID, err := getUserIDFromToken(r)
	if err != nil || userID.IsZero() {
		Response(w, http.StatusUnauthorized, model.Response{
			Status:  "error",
			Message: "Unauthorized",
		})
		return
	}

	// Convert uploadID string to ObjectID
	uploadID, err := primitive.ObjectIDFromHex(uploadIDStr)
	if err != nil {
		Response(w, http.StatusBadRequest, model.Response{
			Status:  "error",
			Message: "Invalid upload ID",
		})
		return
	}

	// Get MongoDB connection
	mongoDB := getMongoDB()
	if mongoDB == nil {
		Response(w, http.StatusInternalServerError, model.Response{
			Status:  "error",
			Message: "Database connection failed",
		})
		return
	}

	// Get upload record
	upload, err := atdb.GetOneDoc[model.Upload](mongoDB, "uploads", bson.M{"_id": uploadID})
	if err != nil {
		Response(w, http.StatusNotFound, model.Response{
			Status:  "error",
			Message: "Upload not found",
		})
		return
	}

	// Check if user has access to this upload (through project ownership)
	_, err = atdb.GetOneDoc[model.Project](mongoDB, "projects", bson.M{"_id": upload.ProjectID, "user_id": userID})
	if err != nil {
		Response(w, http.StatusForbidden, model.Response{
			Status:  "error",
			Message: "Access denied",
		})
		return
	}

	// In production, you would calculate actual statistics from the data
	// For now, return placeholder statistics
	statistics := map[string]interface{}{
		"mean": 5.0,
		"std":  2.5,
		"min":  1.0,
		"max":  10.0,
	}

	// Use existing statistics if available
	if upload.DataSummary.Statistics != nil {
		statistics = upload.DataSummary.Statistics
	}

	Response(w, http.StatusOK, model.Response{
		Status:  "success",
		Message: "Data statistics retrieved",
		Data: map[string]interface{}{
			"upload_id":  uploadIDStr,
			"file_name":  upload.FileName,
			"total_rows": upload.DataSummary.Rows,
			"total_cols": upload.DataSummary.Columns,
			"file_size":  upload.FileSize,
			"statistics": statistics,
		},
	})
}

// GetUploads handler untuk get all uploads
func GetUploads(w http.ResponseWriter, r *http.Request) {
	// Get user ID from token
	userID, err := getUserIDFromToken(r)
	if err != nil || userID.IsZero() {
		Response(w, http.StatusUnauthorized, model.Response{
			Status:  "error",
			Message: "Unauthorized",
		})
		return
	}

	// Get MongoDB connection
	mongoDB := getMongoDB()
	if mongoDB == nil {
		Response(w, http.StatusInternalServerError, model.Response{
			Status:  "error",
			Message: "Database connection failed",
		})
		return
	}

	// Get all projects for this user
	projects, err := atdb.GetAllDoc[model.Project](mongoDB, "projects", bson.M{"user_id": userID})
	if err != nil {
		Response(w, http.StatusInternalServerError, model.Response{
			Status:  "error",
			Message: "Failed to retrieve projects",
		})
		return
	}

	// Get project IDs
	var projectIDs []primitive.ObjectID
	for _, project := range projects {
		projectIDs = append(projectIDs, project.ID)
	}

	// Get all uploads for these projects
	var uploads []model.Upload
	if len(projectIDs) > 0 {
		// Create filter for project IDs
		var projectIDFilters []bson.M
		for _, id := range projectIDs {
			projectIDFilters = append(projectIDFilters, bson.M{"project_id": id})
		}

		// Combine with $or if there are multiple project IDs
		var filter bson.M
		if len(projectIDFilters) == 1 {
			filter = projectIDFilters[0]
		} else {
			filter = bson.M{"$or": projectIDFilters}
		}

		uploads, err = atdb.GetAllDoc[model.Upload](mongoDB, "uploads", filter)
		if err != nil {
			Response(w, http.StatusInternalServerError, model.Response{
				Status:  "error",
				Message: "Failed to retrieve uploads",
			})
			return
		}
	}

	Response(w, http.StatusOK, model.Response{
		Status:  "success",
		Message: "Uploads retrieved successfully",
		Data:    uploads,
	})
}

// GetUpload handler untuk get specific upload by ID
func GetUpload(w http.ResponseWriter, r *http.Request, uploadIDStr string) {
	// Get user ID from token
	userID, err := getUserIDFromToken(r)
	if err != nil || userID.IsZero() {
		Response(w, http.StatusUnauthorized, model.Response{
			Status:  "error",
			Message: "Unauthorized",
		})
		return
	}

	// Convert uploadID string to ObjectID
	uploadID, err := primitive.ObjectIDFromHex(uploadIDStr)
	if err != nil {
		Response(w, http.StatusBadRequest, model.Response{
			Status:  "error",
			Message: "Invalid upload ID",
		})
		return
	}

	// Get MongoDB connection
	mongoDB := getMongoDB()
	if mongoDB == nil {
		Response(w, http.StatusInternalServerError, model.Response{
			Status:  "error",
			Message: "Database connection failed",
		})
		return
	}

	// Get upload record
	upload, err := atdb.GetOneDoc[model.Upload](mongoDB, "uploads", bson.M{"_id": uploadID})
	if err != nil {
		Response(w, http.StatusNotFound, model.Response{
			Status:  "error",
			Message: "Upload not found",
		})
		return
	}

	// Check if user has access to this upload (through project ownership)
	_, err = atdb.GetOneDoc[model.Project](mongoDB, "projects", bson.M{"_id": upload.ProjectID, "user_id": userID})
	if err != nil {
		Response(w, http.StatusForbidden, model.Response{
			Status:  "error",
			Message: "Access denied",
		})
		return
	}

	Response(w, http.StatusOK, model.Response{
		Status:  "success",
		Message: "Upload retrieved successfully",
		Data:    upload,
	})
}

// DeleteUpload handler untuk delete upload
func DeleteUpload(w http.ResponseWriter, r *http.Request, uploadIDStr string) {
	// Get user ID from token
	userID, err := getUserIDFromToken(r)
	if err != nil || userID.IsZero() {
		Response(w, http.StatusUnauthorized, model.Response{
			Status:  "error",
			Message: "Unauthorized",
		})
		return
	}

	// Convert uploadID string to ObjectID
	uploadID, err := primitive.ObjectIDFromHex(uploadIDStr)
	if err != nil {
		Response(w, http.StatusBadRequest, model.Response{
			Status:  "error",
			Message: "Invalid upload ID",
		})
		return
	}

	// Get MongoDB connection
	mongoDB := getMongoDB()
	if mongoDB == nil {
		Response(w, http.StatusInternalServerError, model.Response{
			Status:  "error",
			Message: "Database connection failed",
		})
		return
	}

	// Get upload record
	upload, err := atdb.GetOneDoc[model.Upload](mongoDB, "uploads", bson.M{"_id": uploadID})
	if err != nil {
		Response(w, http.StatusNotFound, model.Response{
			Status:  "error",
			Message: "Upload not found",
		})
		return
	}

	// Check if user has access to this upload (through project ownership)
	_, err = atdb.GetOneDoc[model.Project](mongoDB, "projects", bson.M{"_id": upload.ProjectID, "user_id": userID})
	if err != nil {
		Response(w, http.StatusForbidden, model.Response{
			Status:  "error",
			Message: "Access denied",
		})
		return
	}

	// In production, you would also delete the file from Google Cloud Storage

	// Delete upload record from database
	_, err = atdb.DeleteOneDoc(mongoDB, "uploads", bson.M{"_id": uploadID})
	if err != nil {
		Response(w, http.StatusInternalServerError, model.Response{
			Status:  "error",
			Message: "Failed to delete upload",
		})
		return
	}

	Response(w, http.StatusOK, model.Response{
		Status:  "success",
		Message: "Upload deleted successfully",
		Data: map[string]interface{}{
			"deleted_upload_id": uploadIDStr,
			"deleted_at":        time.Now(),
		},
	})
}
