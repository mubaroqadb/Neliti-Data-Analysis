package config

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Config adalah struktur konfigurasi utama aplikasi
type Config struct {
	// Database Configuration
	MongoDB       *MongoDBConfig `json:"mongodb"`
	
	// Authentication Configuration  
	Auth          *AuthConfig `json:"auth"`
	
	// Google Cloud Platform Configuration
	GCP           *GCPConfig `json:"gcp"`
	
	// Server Configuration
	Server        *ServerConfig `json:"server"`
	
	// Application Configuration
	App           *AppConfig `json:"app"`
	
	// Runtime Configuration
	isProduction  bool
	mongoClient   *mongo.Client
}

// MongoDBConfig konfigurasi untuk MongoDB
type MongoDBConfig struct {
	ConnectionString string `json:"connection_string"`
	DatabaseName     string `json:"database_name"`
	ConnectTimeout   time.Duration `json:"connect_timeout"`
	MaxPoolSize      uint64 `json:"max_pool_size"`
}

// AuthConfig konfigurasi untuk authentication
type AuthConfig struct {
	PrivateKey      string `json:"private_key"`
	PublicKey       string `json:"public_key"`
	JWTSecret       string `json:"jwt_secret"`
	TokenExpiration time.Duration `json:"token_expiration"`
}

// GCPConfig konfigurasi untuk Google Cloud Platform
type GCPConfig struct {
	ProjectID       string `json:"project_id"`
	GCSBucket       string `json:"gcs_bucket"`
	VertexAIRegion  string `json:"vertexai_region"`
	ServiceAccount  string `json:"service_account"`
}

// ServerConfig konfigurasi untuk server
type ServerConfig struct {
	Port             string        `json:"port"`
	ReadTimeout      time.Duration `json:"read_timeout"`
	WriteTimeout     time.Duration `json:"write_timeout"`
	IdleTimeout      time.Duration `json:"idle_timeout"`
	ShutdownTimeout  time.Duration `json:"shutdown_timeout"`
}

// AppConfig konfigurasi aplikasi
type AppConfig struct {
	Environment      string `json:"environment"`
	Name             string `json:"name"`
	Version          string `json:"version"`
	Debug            bool   `json:"debug"`
	LogLevel         string `json:"log_level"`
	AllowedOrigins   []string `json:"allowed_origins"`
}

// Global configuration instance
var (
	appConfig *Config
	once      sync.Once
)

// LoadConfig memuat konfigurasi aplikasi dari environment variables
func LoadConfig() *Config {
	once.Do(func() {
		appConfig = loadFromEnvironment()
		if err := appConfig.Validate(); err != nil {
			log.Fatalf("Configuration validation failed: %v", err)
		}
		
		if err := appConfig.initializeConnections(); err != nil {
			log.Fatalf("Failed to initialize connections: %v", err)
		}
	})
	
	return appConfig
}

// GetConfig mengambil instance konfigurasi yang sudah dimuat
func GetConfig() *Config {
	if appConfig == nil {
		return LoadConfig()
	}
	return appConfig
}

// loadFromEnvironment memuat konfigurasi dari environment variables
func loadFromEnvironment() *Config {
	// Tentukan environment
	environment := getEnv("ENVIRONMENT", "development")
	isProduction := environment == "production"
	
	// Load default values berdasarkan environment
	defaultOrigins := getDefaultAllowedOrigins(isProduction)
	
	config := &Config{
		MongoDB: &MongoDBConfig{
			ConnectionString: getEnv("MONGOSTRING", getDefaultMongoString(environment)),
			DatabaseName:     getEnv("MONGO_DATABASE", "research_data_analysis"),
			ConnectTimeout:   10 * time.Second,
			MaxPoolSize:      100,
		},
		Auth: &AuthConfig{
			PrivateKey:      getEnv("PRIVATEKEY", ""),
			PublicKey:       getEnv("PUBLICKEY", ""),
			JWTSecret:       getEnv("JWT_SECRET", "default-jwt-secret-change-in-production"),
			TokenExpiration: 24 * time.Hour,
		},
		GCP: &GCPConfig{
			ProjectID:       getEnv("GCP_PROJECT_ID", ""),
			GCSBucket:       getEnv("GCS_BUCKET", ""),
			VertexAIRegion:  getEnv("VERTEXAI_REGION", "asia-southeast1"),
			ServiceAccount:  getEnv("GOOGLE_APPLICATION_CREDENTIALS", ""),
		},
		Server: &ServerConfig{
			Port:            getEnv("PORT", "8080"),
			ReadTimeout:     30 * time.Second,
			WriteTimeout:    30 * time.Second,
			IdleTimeout:     60 * time.Second,
			ShutdownTimeout: 10 * time.Second,
		},
		App: &AppConfig{
			Environment:  environment,
			Name:         "Research Data Analysis",
			Version:      "1.0.0",
			Debug:        !isProduction,
			LogLevel:     getEnv("LOG_LEVEL", "info"),
			AllowedOrigins: defaultOrigins,
		},
		isProduction: isProduction,
	}
	
	return config
}

// getEnv mengambil environment variable dengan default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getDefaultMongoString mendapat default MongoDB connection string
func getDefaultMongoString(environment string) string {
	if environment == "development" {
		return "mongodb://localhost:27017/research_data_analysis"
	}
	return "" // Production requires explicit MONGOSTRING
}

// getDefaultAllowedOrigins mendapat CORS origins default
func getDefaultAllowedOrigins(isProduction bool) []string {
	if isProduction {
		return []string{
			"https://research-data-analysis.github.io",
		}
	}
	return []string{
		"http://localhost:8080",
		"http://127.0.0.1:8080",
		"http://localhost:3000",
		"http://127.0.0.1:3000",
	}
}

// SetEnv fungsi legacy untuk backward compatibility
func SetEnv() {
	// Compatibility function - triggers config loading
	LoadConfig()
}

// Validate melakukan validasi konfigurasi yang diperlukan
func (c *Config) Validate() error {
	// Validasi Database
	if c.MongoDB.ConnectionString == "" {
		return fmt.Errorf("MONGOSTRING environment variable is required")
	}
	
	if c.MongoDB.DatabaseName == "" {
		return fmt.Errorf("database name cannot be empty")
	}
	
	// Validasi Authentication
	if c.Auth.PrivateKey == "" {
		return fmt.Errorf("PRIVATEKEY environment variable is required for authentication")
	}
	
	if c.Auth.PublicKey == "" {
		return fmt.Errorf("PUBLICKEY environment variable is required for authentication")
	}
	
	if c.Auth.JWTSecret == "" || c.Auth.JWTSecret == "default-jwt-secret-change-in-production" {
		return fmt.Errorf("JWT_SECRET environment variable must be set to a secure value")
	}
	
	// Validasi GCP (required for production)
	if c.isProduction {
		if c.GCP.ProjectID == "" {
			return fmt.Errorf("GCP_PROJECT_ID environment variable is required in production")
		}
	}
	
	// Validasi Server
	if c.Server.Port == "" {
		return fmt.Errorf("PORT environment variable cannot be empty")
	}
	
	// Validasi Port format
	if _, err := strconv.Atoi(c.Server.Port); err != nil {
		return fmt.Errorf("PORT must be a valid number, got: %s", c.Server.Port)
	}
	
	return nil
}

// IsProduction menentukan apakah aplikasi berjalan di production
func (c *Config) IsProduction() bool {
	return c.isProduction
}

// GetMongoDatabase mendapat MongoDB database instance
func (c *Config) GetMongoDatabase() (*mongo.Database, error) {
	if c.mongoClient == nil {
		return nil, fmt.Errorf("MongoDB client not initialized")
	}
	
	return c.mongoClient.Database(c.MongoDB.DatabaseName), nil
}

// TestConnection menguji koneksi ke database
func (c *Config) TestConnection() error {
	if c.mongoClient == nil {
		return fmt.Errorf("MongoDB client not initialized")
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	return c.mongoClient.Ping(ctx, nil)
}

// ConfigurationHealthCheck melakukan health check konfigurasi
func (c *Config) ConfigurationHealthCheck() error {
	// Test MongoDB connection
	if err := c.TestConnection(); err != nil {
		return fmt.Errorf("MongoDB health check failed: %w", err)
	}

	// Validate authentication config
	if c.Auth.PrivateKey == "" || c.Auth.PublicKey == "" {
		return fmt.Errorf("authentication keys not configured")
	}

	// Test GCP config if in production
	if c.isProduction && c.GCP.ProjectID == "" {
		return fmt.Errorf("GCP Project ID not configured for production")
	}

	return nil
}

// GetAllowedOrigins mendapat daftar origins yang diizinkan untuk CORS
func (c *Config) GetAllowedOrigins() []string {
	return c.App.AllowedOrigins
}

// initializeConnections menginisialisasi koneksi ke services
func (c *Config) initializeConnections() error {
	// Initialize MongoDB connection
	ctx, cancel := context.WithTimeout(context.Background(), c.MongoDB.ConnectTimeout)
	defer cancel()
	
	clientOptions := options.Client()
	clientOptions.SetMaxPoolSize(c.MongoDB.MaxPoolSize)
	clientOptions.SetConnectTimeout(c.MongoDB.ConnectTimeout)
	
	client, err := mongo.Connect(ctx, clientOptions.ApplyURI(c.MongoDB.ConnectionString))
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}
	
	// Test connection
	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}
	
	c.mongoClient = client
	
	log.Printf("Successfully connected to MongoDB: %s", c.MongoDB.DatabaseName)
	return nil
}

// Close menutup semua koneksi
func (c *Config) Close() error {
	if c.mongoClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		
		if err := c.mongoClient.Disconnect(ctx); err != nil {
			return fmt.Errorf("failed to disconnect MongoDB: %w", err)
		}
		log.Println("MongoDB connection closed")
	}
	
	return nil
}

// SetAccessControlHeaders mengatur CORS headers menggunakan konfigurasi
func SetAccessControlHeaders(w http.ResponseWriter, r *http.Request) bool {
	cfg := GetConfig()
	
	origin := r.Header.Get("Origin")
	allowedOrigins := cfg.GetAllowedOrigins()
	
	// Check if origin is allowed
	originAllowed := false
	for _, allowedOrigin := range allowedOrigins {
		if origin == allowedOrigin {
			originAllowed = true
			break
		}
	}
	
	if originAllowed {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	} else if !cfg.IsProduction() {
		// Allow all origins in development
		w.Header().Set("Access-Control-Allow-Origin", "*")
	}

	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Login, Secret")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Max-Age", "3600")

	// Handle preflight
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusNoContent)
		return true
	}
	return false
}

// GetEnvironment mendapat environment aplikasi (development/production)
func GetEnvironment() string {
	return GetConfig().App.Environment
}

// IsDebugMode menentukan apakah debug mode aktif
func IsDebugMode() bool {
	return GetConfig().App.Debug
}

// GetLogLevel mendapat level logging
func GetLogLevel() string {
	return GetConfig().App.LogLevel
}

// PrintConfigInfo menampilkan informasi konfigurasi (untuk debugging)
func PrintConfigInfo() {
	cfg := GetConfig()
	
	fmt.Printf("=== Configuration Info ===")
	fmt.Printf("\nEnvironment: %s", cfg.App.Environment)
	fmt.Printf("\nDebug Mode: %t", cfg.App.Debug)
	fmt.Printf("\nPort: %s", cfg.Server.Port)
	fmt.Printf("\nDatabase: %s", cfg.MongoDB.DatabaseName)
	fmt.Printf("\nGCP Project ID: %s", cfg.GCP.ProjectID)
	fmt.Printf("\nVertexAI Region: %s", cfg.GCP.VertexAIRegion)
	fmt.Printf("\nLog Level: %s", cfg.App.LogLevel)
	fmt.Printf("\nAllowed Origins: %v", cfg.App.AllowedOrigins)
	fmt.Printf("\n========================\n")
}

// Legacy compatibility functions untuk backward compatibility
func GetMongoString() string {
	return GetConfig().MongoDB.ConnectionString
}

func GetMongoDB() (*mongo.Database, error) {
	return GetConfig().GetMongoDatabase()
}

func GetPrivateKey() string {
	return GetConfig().Auth.PrivateKey
}

func GetPublicKey() string {
	return GetConfig().Auth.PublicKey
}

func GetGCSBucket() string {
	return GetConfig().GCP.GCSBucket
}

func GetGCPProjectID() string {
	return GetConfig().GCP.ProjectID
}

func GetVertexAIRegion() string {
	return GetConfig().GCP.VertexAIRegion
}

// Legacy support for GoCroot compatibility
var Mongoconn = GetConfig()