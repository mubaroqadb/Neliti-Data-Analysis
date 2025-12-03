# README.md

## Research Data Analysis Backend

Backend aplikasi Research Data Analysis menggunakan GoCroot framework.

### Requirements
- Go 1.22.5+
- MongoDB
- Google Cloud Platform

### Dependencies
- MongoDB Driver v1.16.0
- Google Cloud AI Platform v1.68.0
- Google Cloud Storage v1.42.0
- Go Functions Framework v1.8.1
- Go Paseto v1.5.1
- Go Crypot v0.25.0
- Google APIs v0.190.0

### Building
```bash
cd backend
go mod tidy
go build -o main main-cloudrun.go
```

### Running
```bash
cd backend
./main
```

### Environment Variables
- MONGOSTRING: MongoDB connection string
- GCP_PROJECT_ID: Google Cloud Project ID
- PRIVATEKEY: Authentication private key
- PUBLICKEY: Authentication public key
- JWT_SECRET: JWT secret key
- GCS_BUCKET: Google Cloud Storage bucket name
- VERTEXAI_REGION: Vertex AI region
- PORT: Server port (default: 8080)
- ENVIRONMENT: Environment (development/production)
