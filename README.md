# Neliti Data Analysis

Platform analisis data riset interaktif dengan bantuan AI untuk membantu peneliti mengolah dan menganalisis data penelitian secara efisien.

## Teknologi

### Backend
- **Framework**: GoCroot (Go + Google Cloud Functions)
- **Database**: MongoDB Atlas
- **Storage**: Google Cloud Storage
- **AI**: Vertex AI Gemini 2.0 Flash
- **Auth**: PASETO v4 (Ed25519)

### Frontend
- **Framework**: JSCroot (ES6+ Vanilla JavaScript)
- **Styling**: Custom CSS dengan CSS Variables
- **Charts**: Chart.js
- **Deployment**: GitHub Pages

## Fitur

- **Upload Data**: Upload file penelitian (CSV, Excel, JSON)
- **Rekomendasi AI**: Dapatkan rekomendasi metode analisis statistik dari AI
- **Analisis Interaktif**: Pilih dan proses analisis dengan metode yang direkomendasikan
- **Visualisasi**: Lihat hasil dalam bentuk grafik dan chart
- **Ekspor**: Ekspor hasil analisis dalam format PDF, CSV, atau JSON

## Struktur Proyek

```
.
├── backend/                    # Backend API (GoCroot)
│   ├── config/                 # Konfigurasi aplikasi
│   ├── controller/             # Request handlers
│   ├── helper/                 # Helper functions
│   │   ├── at/                 # Utility functions
│   │   ├── atdb/               # Database operations
│   │   ├── storage/            # GCS operations
│   │   ├── vertexai/           # AI integration
│   │   └── watoken/            # Token management
│   ├── model/                  # Data models
│   ├── route/                  # URL routing
│   ├── main.go                 # Entry point
│   └── go.mod                  # Go modules
│
├── frontend/                   # Frontend SPA (JSCroot)
│   ├── css/                    # Stylesheets
│   ├── js/                     # JavaScript modules
│   └── index.html              # Main HTML
│
└── docs/                       # Documentation
    └── setup-guide.md          # Setup instructions
```

## API Endpoints

### Authentication
- `POST /auth/register` - Registrasi pengguna baru
- `POST /auth/login` - Login pengguna
- `GET /auth/profile` - Dapatkan profil pengguna

### Projects
- `POST /api/project` - Buat proyek baru
- `GET /api/project` - Daftar semua proyek
- `GET /api/project/:id` - Detail proyek
- `PUT /api/project` - Update proyek
- `DELETE /api/project` - Hapus proyek

### Data Upload
- `POST /api/upload/:projectId` - Upload file data
- `GET /api/preview/:uploadId` - Preview data
- `GET /api/stats/:uploadId` - Statistik data

### Analysis
- `POST /api/recommend/:projectId` - Rekomendasi metode analisis
- `POST /api/process` - Proses analisis
- `GET /api/results/:analysisId` - Hasil analisis
- `POST /api/refine/:analysisId` - Refine analisis

### Export
- `GET /api/export/:analysisId?format=pdf|csv|json` - Ekspor hasil

## Environment Variables

```bash
MONGOSTRING=mongodb+srv://user:pass@cluster.mongodb.net/
PRIVATEKEY=your_ed25519_private_key_hex
PUBLICKEY=your_ed25519_public_key_hex
GCS_BUCKET=your-gcs-bucket-name
GCP_PROJECT_ID=your-gcp-project-id
VERTEXAI_REGION=asia-southeast1
```

## Deployment Status

- **Last Updated**: 2025-12-03 20:29:51
- **Backend**: Deploying to Cloud Run with Workload Identity Federation
- **GitHub Actions**: Using WIF authentication for secure deployment

### Backend (Cloud Run)
```bash
gcloud functions deploy ResearchDataAnalysis \
  --gen2 \
  --runtime=go122 \
  --region=asia-southeast1 \
  --source=./backend \
  --entry-point=ResearchDataAnalysis \
  --trigger-http \
  --allow-unauthenticated
```

### Frontend (GitHub Pages)
1. Enable GitHub Pages di repository settings
2. Pilih source: GitHub Actions
3. Push ke branch main

## Dokumentasi

Lihat [docs/setup-guide.md](docs/setup-guide.md) untuk panduan lengkap setup dan deployment.

## Lisensi

MIT License
# Trigger GitHub Actions workflow
# Trigger GitHub Actions workflow again
