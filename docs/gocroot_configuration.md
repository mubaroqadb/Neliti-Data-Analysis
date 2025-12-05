# GoCroot Configuration Implementation

## Overview

Implementasi konfigurasi GoCroot yang modern dan production-ready untuk aplikasi Research Data Analysis. Konfigurasi ini mengimplementasikan pola best practices dengan environment variable management, validation, dan Cloud Run compatibility.

## Struktur Konfigurasi

### 1. Environment Variables yang Diperlukan

#### Database Configuration
```bash
MONGOSTRING=mongodb+srv://username:password@cluster.mongodb.net/research_data_analysis
MONGO_DATABASE=research_data_analysis
```

#### Authentication (PASETO + JWT)
```bash
PRIVATEKEY=-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----
PUBLICKEY=-----BEGIN PUBLIC KEY-----\n...\n-----END PUBLIC KEY-----
JWT_SECRET=your-super-secure-jwt-secret-key-here
```

#### Google Cloud Platform
```bash
GCP_PROJECT_ID=your-gcp-project-id
GCS_BUCKET=your-gcs-bucket-name
VERTEXAI_REGION=asia-southeast1
GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json
```

#### Server Configuration
```bash
PORT=8080
ENVIRONMENT=production
LOG_LEVEL=info
```

### 2. Default Values untuk Development

- **MongoDB**: `mongodb://localhost:27017/research_data_analysis`
- **PORT**: `8080`
- **Environment**: `development`
- **CORS Origins**: 
  - Development: `localhost:8080`, `127.0.0.1:8080`, `localhost:3000`
  - Production: `https://research-data-analysis.github.io`

### 3. Production-Ready Features

#### Validation
- ✅ Required environment variables check
- ✅ Connection testing untuk MongoDB
- ✅ Port format validation
- ✅ JWT secret security validation

#### Connection Management
- ✅ MongoDB connection pooling
- ✅ Connection timeout configuration
- ✅ Automatic reconnection
- ✅ Graceful shutdown

#### CORS Configuration
- ✅ Dynamic origin checking berdasarkan environment
- ✅ Production security: strict origin checking
- ✅ Development flexibility: allow localhost origins

#### Health Monitoring
- ✅ `/health` endpoint untuk Cloud Run monitoring
- ✅ Database connection health check
- ✅ Configuration validation

## Usage Examples

### Basic Configuration Loading

```go
// Load configuration
cfg := config.LoadConfig()

// Get configuration instance
cfg := config.GetConfig()

// Check if production
if cfg.IsProduction() {
    // Production specific logic
}
```

### Database Access

```go
// Get MongoDB database instance
db, err := config.GetConfig().GetMongoDatabase()
if err != nil {
    log.Fatal(err)
}

// Legacy compatibility (backward compatible)
mongoString := config.GetMongoString()
```

### Authentication Configuration

```go
// Get PASETO keys
privateKey := config.GetPrivateKey()
publicKey := config.GetPublicKey()

// Get JWT secret
jwtSecret := config.GetConfig().Auth.JWTSecret
```

### GCP Configuration

```go
// Get GCP settings
projectID := config.GetGCPProjectID()
bucket := config.GetGCSBucket()
region := config.GetVertexAIRegion()
```

## Configuration Validation

### Automatic Validation

Konfigurasi melakukan validasi otomatis saat `LoadConfig()` dipanggil:

```go
func Validate() error {
    // Database validation
    if c.MongoDB.ConnectionString == "" {
        return fmt.Errorf("MONGOSTRING environment variable is required")
    }
    
    // Authentication validation
    if c.Auth.PrivateKey == "" {
        return fmt.Errorf("PRIVATEKEY environment variable is required")
    }
    
    // Production-specific validation
    if c.isProduction && c.GCP.ProjectID == "" {
        return fmt.Errorf("GCP_PROJECT_ID required in production")
    }
}
```

### Manual Validation

```go
// Test configuration
config.TestConfiguration()

// Quick validation check
if config.QuickConfigCheck() {
    fmt.Println("Configuration is valid")
}
```

## Cloud Run Compatibility

### Required Environment Variables

```bash
# Set in Cloud Run
MONGOSTRING=your-mongodb-atlas-connection-string
PRIVATEKEY=your-private-key
PUBLICKEY=your-public-key
JWT_SECRET=secure-jwt-secret
GCP_PROJECT_ID=your-project-id
ENVIRONMENT=production
```

### Health Check Endpoint

Aplikasi menyediakan endpoint `/health` untuk Cloud Run health checks:

```bash
curl https://your-cloud-run-url/health
# Returns: {"status": "healthy", "environment": "production"}
```

### Port Configuration

```go
// PORT environment variable untuk Cloud Run compatibility
port := config.GetConfig().Server.Port
```

## Best Practices

### 1. Environment Separation

- **Development**: Local MongoDB, permissive CORS, debug enabled
- **Production**: MongoDB Atlas, strict CORS, debug disabled
- Use `ENVIRONMENT` variable untuk switch behavior

### 2. Security

- ✅ Never commit real keys to repository
- ✅ Use secure JWT secrets
- ✅ Validate all required variables
- ✅ Use proper CORS origins

### 3. Connection Management

- ✅ Connection pooling untuk performance
- ✅ Timeout configuration untuk reliability
- ✅ Graceful error handling
- ✅ Automatic reconnection

### 4. Monitoring

- ✅ Health check endpoint untuk Cloud Run
- ✅ Configuration validation
- ✅ Database connection testing
- ✅ Logging untuk debugging

## Migration Guide

### From Legacy Configuration

```go
// Old way
config.SetEnv()
mongoString := config.MongoString

// New way (backward compatible)
config.LoadConfig()
mongoString := config.GetMongoString() // Still works
```

### Controller Integration

Controllers tetap menggunakan fungsi legacy untuk backward compatibility:

```go
// In controllers
config.SetEnv() // Triggers new configuration loading
db, _ := config.GetMongoDB()
privateKey := config.GetPrivateKey()
```

## Deployment

### Cloud Run Deployment

1. Set environment variables di Cloud Run Console
2. Update Dockerfile untuk ekspose PORT 8080
3. Deploy dengan Cloud Run configuration

### Local Development

```bash
# Copy .env.example to .env
cp .env.example .env

# Edit .env with your values
nano .env

# Set environment variables
export $(cat .env | xargs)
```

## Troubleshooting

### Common Issues

1. **Missing MONGOSTRING**: Set MongoDB Atlas connection string
2. **Invalid PRIVATEKEY**: Generate RSA key pair
3. **JWT_SECRET validation**: Set secure secret value
4. **Connection timeout**: Check MongoDB Atlas whitelist

### Debug Mode

```bash
# Enable debug untuk development
ENVIRONMENT=development
LOG_LEVEL=debug

# Check configuration info
curl http://localhost:8080/config
```

## Summary

Implementasi GoCroot configuration ini memberikan:

- ✅ **Modern Configuration Patterns**: Typed structures, validation, environment management
- ✅ **Production Ready**: Security, monitoring, health checks
- ✅ **Cloud Run Compatible**: PORT configuration, health endpoints
- ✅ **Backward Compatible**: Legacy functions tetap bekerja
- ✅ **Developer Friendly**: Debug mode, validation tools, documentation

Konfigurasi ini siap untuk production deployment di Google Cloud Run dengan MongoDB Atlas dan Google Cloud services.