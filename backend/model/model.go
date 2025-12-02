package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Response untuk respons API standar
type Response struct {
	Status   string      `json:"status"`
	Message  string      `json:"message,omitempty"`
	Response string      `json:"response,omitempty"`
	Location string      `json:"location,omitempty"`
	Data     interface{} `json:"data,omitempty"`
}

// User menyimpan informasi pengguna
type User struct {
	ID            primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Email         string             `json:"email" bson:"email"`
	Password      string             `json:"-" bson:"password"`
	FullName      string             `json:"full_name" bson:"full_name"`
	Institution   string             `json:"institution" bson:"institution"`
	ResearchField string             `json:"research_field" bson:"research_field"`
	CreatedAt     time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at" bson:"updated_at"`
}

// Variables untuk variabel penelitian
type Variables struct {
	Independent []string `json:"independent" bson:"independent"`
	Dependent   []string `json:"dependent" bson:"dependent"`
	Control     []string `json:"control,omitempty" bson:"control,omitempty"`
	Moderating  []string `json:"moderating,omitempty" bson:"moderating,omitempty"`
	Mediating   []string `json:"mediating,omitempty" bson:"mediating,omitempty"`
}

// Project menyimpan informasi proyek penelitian
type Project struct {
	ID           primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	UserID       primitive.ObjectID `json:"user_id" bson:"user_id"`
	Title        string             `json:"title" bson:"title"`
	Description  string             `json:"description" bson:"description"`
	ResearchType string             `json:"research_type" bson:"research_type"`
	Hypothesis   string             `json:"hypothesis" bson:"hypothesis"`
	Variables    Variables          `json:"variables" bson:"variables"`
	Status       string             `json:"status" bson:"status"`
	CreatedAt    time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at" bson:"updated_at"`
}

// DataSummary untuk ringkasan data upload
type DataSummary struct {
	Rows         int                    `json:"rows" bson:"rows"`
	Columns      int                    `json:"columns" bson:"columns"`
	ColumnNames  []string               `json:"column_names" bson:"column_names"`
	ColumnTypes  map[string]string      `json:"column_types" bson:"column_types"`
	MissingCount map[string]int         `json:"missing_count" bson:"missing_count"`
	Statistics   map[string]interface{} `json:"statistics,omitempty" bson:"statistics,omitempty"`
}

// Upload menyimpan informasi file yang diupload
type Upload struct {
	ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	ProjectID   primitive.ObjectID `json:"project_id" bson:"project_id"`
	FileName    string             `json:"file_name" bson:"file_name"`
	FileType    string             `json:"file_type" bson:"file_type"`
	FileSize    int64              `json:"file_size" bson:"file_size"`
	StorageURL  string             `json:"storage_url" bson:"storage_url"`
	DataSummary DataSummary        `json:"data_summary" bson:"data_summary"`
	UploadedAt  time.Time          `json:"uploaded_at" bson:"uploaded_at"`
}

// Recommendation untuk rekomendasi metode analisis
type Recommendation struct {
	Method      string `json:"method" bson:"method"`
	Category    string `json:"category" bson:"category"`
	Reasoning   string `json:"reasoning" bson:"reasoning"`
	Priority    int    `json:"priority" bson:"priority"`
	Assumptions string `json:"assumptions" bson:"assumptions"`
}

// MethodResult untuk hasil analisis
type MethodResult struct {
	Method        string                 `json:"method" bson:"method"`
	RawOutput     map[string]interface{} `json:"raw_output" bson:"raw_output"`
	Interpretation string                `json:"interpretation" bson:"interpretation"`
	EffectSize    string                 `json:"effect_size,omitempty" bson:"effect_size,omitempty"`
	Conclusion    string                 `json:"conclusion" bson:"conclusion"`
}

// Figure untuk gambar/chart hasil analisis
type Figure struct {
	ID         string `json:"id" bson:"id"`
	Title      string `json:"title" bson:"title"`
	Type       string `json:"type" bson:"type"`
	StorageURL string `json:"storage_url" bson:"storage_url"`
}

// Analysis menyimpan informasi analisis
type Analysis struct {
	ID              primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	ProjectID       primitive.ObjectID `json:"project_id" bson:"project_id"`
	UploadID        primitive.ObjectID `json:"upload_id" bson:"upload_id"`
	Iteration       int                `json:"iteration" bson:"iteration"`
	Status          string             `json:"status" bson:"status"`
	Recommendations []Recommendation   `json:"recommendations" bson:"recommendations"`
	SelectedMethods []string           `json:"selected_methods" bson:"selected_methods"`
	Results         []MethodResult     `json:"results" bson:"results"`
	Figures         []Figure           `json:"figures" bson:"figures"`
	Summary         string             `json:"summary" bson:"summary"`
	UserFeedback    string             `json:"user_feedback" bson:"user_feedback"`
	CreatedAt       time.Time          `json:"created_at" bson:"created_at"`
	CompletedAt     *time.Time         `json:"completed_at,omitempty" bson:"completed_at,omitempty"`
	Error           string             `json:"error,omitempty" bson:"error,omitempty"`
}

// AuditLog untuk logging aktivitas
type AuditLog struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	UserID    primitive.ObjectID `json:"user_id" bson:"user_id"`
	Action    string             `json:"action" bson:"action"`
	Resource  string             `json:"resource" bson:"resource"`
	Details   string             `json:"details" bson:"details"`
	IP        string             `json:"ip" bson:"ip"`
	UserAgent string             `json:"user_agent" bson:"user_agent"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
}

// LoginRequest untuk request login
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterRequest untuk request registrasi
type RegisterRequest struct {
	Email         string `json:"email"`
	Password      string `json:"password"`
	FullName      string `json:"full_name"`
	Institution   string `json:"institution"`
	ResearchField string `json:"research_field"`
}

// ProjectRequest untuk request project
type ProjectRequest struct {
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	ResearchType string    `json:"research_type"`
	Hypothesis   string    `json:"hypothesis"`
	Variables    Variables `json:"variables"`
}

// RecommendRequest untuk request rekomendasi
type RecommendRequest struct {
	UploadID string   `json:"upload_id"`
	Context  string   `json:"context,omitempty"`
	Specific []string `json:"specific,omitempty"`
}

// ProcessRequest untuk request proses analisis
type ProcessRequest struct {
	AnalysisID      string   `json:"analysis_id"`
	SelectedMethods []string `json:"selected_methods"`
}

// RefineRequest untuk request refinement
type RefineRequest struct {
	Feedback    string   `json:"feedback"`
	NewMethods  []string `json:"new_methods,omitempty"`
	Adjustments string   `json:"adjustments,omitempty"`
}

// ExportRequest untuk request export
type ExportRequest struct {
	Format   string   `json:"format"`
	Sections []string `json:"sections,omitempty"`
}