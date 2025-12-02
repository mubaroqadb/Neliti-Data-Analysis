# Research Data Analysis - Backend API

Backend API untuk aplikasi Research Data Analysis, dibangun dengan GoCroot framework dan di-deploy ke Google Cloud Functions.

## Teknologi yang Digunakan

- **Framework**: GoCroot (Go + Google Cloud Functions Framework)
- **Database**: MongoDB Atlas
- **Storage**: Google Cloud Storage
- **AI**: Vertex AI Gemini 2.0 Flash
- **Deployment**: Google Cloud Functions dengan GitHub Actions CI/CD

## Struktur Direktori

```
backend/
├── .github/
│   └── workflows/
│       └── deploy.yml       # GitHub Actions untuk CI/CD
├── config/
│   └── config.go            # Konfigurasi aplikasi dan environment
├── controller/
│   ├── auth.go              # Handler untuk autentikasi
│   ├── base.go              # Handler dasar (home, not found)
│   ├── project.go           # Handler untuk proyek penelitian
│   ├── upload.go            # Handler untuk upload data
│   ├── analysis.go          # Handler untuk analisis
│   └── export.go            # Handler untuk ekspor hasil
├── helper/
│   ├── at/
│   │   └── at.go            # Utility functions
│   ├── atdb/
│   │   └── atdb.go          # Database helper functions
│   ├── storage/
│   │   └── storage.go       # Google Cloud Storage helper
│   ├── vertexai/
│   │   └── vertexai.go      # Vertex AI integration
│   └── watoken/
│       └── watoken.go       # Token generation/validation
├── model/
│   └── model.go             # Data models/structs
├── route/
│   └── route.go             # URL routing
├── main.go                  # Entry point
├── go.mod                   # Go module definition
└── README.md
```

## API Endpoints

### Autentikasi
- `POST /auth/register` - Registrasi pengguna baru
- `POST /auth/login` - Login pengguna
- `GET /auth/profile` - Dapatkan profil pengguna (memerlukan token)

### Proyek
- `POST /api/project` - Buat proyek baru
- `GET /api/project` - Daftar semua proyek pengguna
- `GET /api/project/:id` - Dapatkan detail proyek
- `PUT /api/project` - Update proyek
- `DELETE /api/project` - Hapus proyek

### Upload Data
- `POST /api/upload/:projectId` - Upload file data (CSV, XLSX, JSON)
- `GET /api/preview/:uploadId` - Preview data yang diupload
- `GET /api/stats/:uploadId` - Statistik data

### Analisis
- `POST /api/recommend/:projectId` - Dapatkan rekomendasi metode analisis
- `POST /api/process` - Proses analisis dengan metode yang dipilih
- `GET /api/results/:analysisId` - Dapatkan hasil analisis
- `POST /api/refine/:analysisId` - Refine analisis berdasarkan feedback

### Ekspor
- `GET /api/export/:analysisId?format=pdf|csv|json` - Ekspor hasil analisis

## Environment Variables

Variabel lingkungan yang diperlukan:

```
MONGOSTRING=mongodb+srv://user:pass@cluster.mongodb.net/
PRIVATEKEY=your_ed25519_private_key_hex
PUBLICKEY=your_ed25519_public_key_hex
GCS_BUCKET=your-gcs-bucket-name
GCP_PROJECT_ID=your-gcp-project-id
VERTEXAI_REGION=asia-southeast1
```

## Setup GitHub Secrets

Untuk deployment, tambahkan secrets berikut di repository GitHub:

1. `GOOGLE_CREDENTIALS` - JSON key dari service account GCP
2. `MONGOSTRING` - Connection string MongoDB Atlas
3. `PRIVATEKEY` - Ed25519 private key untuk token
4. `PUBLICKEY` - Ed25519 public key untuk token
5. `GCS_BUCKET` - Nama bucket Google Cloud Storage

Dan variables:
1. `GCP_PROJECT_ID` - ID project GCP
2. `GCP_REGION` - Region GCP (default: asia-southeast1)
3. `VERTEXAI_REGION` - Region Vertex AI

## Deployment

Deployment otomatis dilakukan melalui GitHub Actions saat push ke branch `main`.

### Manual Deployment

```bash
# Login ke GCP
gcloud auth login

# Set project
gcloud config set project YOUR_PROJECT_ID

# Deploy
gcloud functions deploy ResearchDataAnalysis \
  --gen2 \
  --runtime=go122 \
  --region=asia-southeast1 \
  --source=. \
  --entry-point=ResearchDataAnalysis \
  --trigger-http \
  --allow-unauthenticated
```

## Development Lokal

```bash
# Install dependencies
go mod download

# Run locally (memerlukan environment variables)
export MONGOSTRING="your-mongo-string"
export PRIVATEKEY="your-private-key"
# ... set other env vars

go run main.go
```

## MongoDB Collections

- `users` - Data pengguna
- `projects` - Proyek penelitian
- `uploads` - Metadata file yang diupload
- `analyses` - Hasil analisis
- `audit_logs` - Log aktivitas

## Keamanan

- Autentikasi menggunakan PASETO v4 tokens
- Password di-hash menggunakan bcrypt
- CORS dikonfigurasi untuk origin yang diizinkan
- Semua aktivitas dicatat dalam audit log