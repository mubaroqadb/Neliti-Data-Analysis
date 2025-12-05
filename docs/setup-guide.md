# Research Data Analysis - Panduan Setup dan Deployment

## Daftar Isi
1. [Persiapan MongoDB Atlas](#persiapan-mongodb-atlas)
2. [Persiapan Google Cloud Platform](#persiapan-google-cloud-platform)
3. [Deployment Backend](#deployment-backend)
4. [Deployment Frontend](#deployment-frontend)
5. [Konfigurasi GitHub Secrets](#konfigurasi-github-secrets)
6. [Testing](#testing)

---

## Persiapan MongoDB Atlas

### 1. Buat Akun dan Cluster

1. Daftar di [MongoDB Atlas](https://www.mongodb.com/cloud/atlas)
2. Buat cluster baru (pilih Free Tier untuk testing)
3. Pilih region terdekat (misalnya: Singapore)

### 2. Konfigurasi Network Access

1. Buka **Network Access** > **+ ADD IP ADDRESS**
2. Pilih **ALLOW ACCESS FROM ANYWHERE** atau tambahkan IP spesifik
3. Konfirmasi

### 3. Buat Database User

1. Buka **Database Access** > **+ ADD NEW DATABASE USER**
2. Pilih **Password** authentication
3. Masukkan username dan password
4. Berikan role **Read and write to any database**
5. Simpan

### 4. Dapatkan Connection String

1. Klik **Connect** pada cluster
2. Pilih **Connect your application**
3. Copy connection string
4. Ganti `<password>` dengan password user

Format: `mongodb+srv://username:password@cluster.xxxxx.mongodb.net/`

### 5. Collections yang Diperlukan

Collections akan dibuat otomatis saat pertama kali digunakan:
- `users` - Data pengguna
- `projects` - Proyek penelitian
- `uploads` - Metadata file upload
- `analyses` - Hasil analisis
- `audit_logs` - Log aktivitas

---

## Persiapan Google Cloud Platform

### 1. Buat Project

1. Buka [Google Cloud Console](https://console.cloud.google.com)
2. Buat project baru
3. Catat Project ID

### 2. Enable APIs

Enable API berikut:
- Cloud Functions API
- Cloud Build API
- Cloud Storage API
- Vertex AI API
- Artifact Registry API

```bash
gcloud services enable cloudfunctions.googleapis.com
gcloud services enable cloudbuild.googleapis.com
gcloud services enable storage.googleapis.com
gcloud services enable aiplatform.googleapis.com
gcloud services enable artifactregistry.googleapis.com
```

### 3. Buat Service Account

```bash
# Set project
PROJECT_ID=your-project-id

# Buat service account
gcloud iam service-accounts create research-data-api \
  --display-name="Research Data Analysis API"

# Berikan roles
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:research-data-api@$PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/cloudfunctions.developer"

gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:research-data-api@$PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/storage.admin"

gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:research-data-api@$PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/aiplatform.user"

# Buat key JSON
gcloud iam service-accounts keys create key.json \
  --iam-account="research-data-api@$PROJECT_ID.iam.gserviceaccount.com"
```

### 4. Buat Cloud Storage Bucket

```bash
gsutil mb -l asia-southeast1 gs://research-data-uploads-$PROJECT_ID

# Set CORS
cat > cors.json << EOF
[
  {
    "origin": ["*"],
    "method": ["GET", "PUT", "POST", "DELETE"],
    "responseHeader": ["Content-Type", "Authorization"],
    "maxAgeSeconds": 3600
  }
]
EOF

gsutil cors set cors.json gs://research-data-uploads-$PROJECT_ID
```

### 5. Generate Token Keys

Untuk autentikasi PASETO, generate key pair:

```bash
# Menggunakan Go
go run -tags generate ./helper/watoken/

# Atau gunakan OpenSSL
openssl genpkey -algorithm ed25519 -outform DER | xxd -p -c 64
```

Simpan private key dan public key dalam format hex.

---

## Deployment Backend

### Deploy Manual

```bash
cd backend

gcloud functions deploy ResearchDataAnalysis \
  --gen2 \
  --runtime=go122 \
  --region=asia-southeast1 \
  --source=. \
  --entry-point=ResearchDataAnalysis \
  --trigger-http \
  --allow-unauthenticated \
  --set-env-vars="MONGOSTRING=mongodb+srv://..." \
  --set-env-vars="PRIVATEKEY=your-private-key-hex" \
  --set-env-vars="PUBLICKEY=your-public-key-hex" \
  --set-env-vars="GCS_BUCKET=research-data-uploads-xxx" \
  --set-env-vars="GCP_PROJECT_ID=your-project-id" \
  --set-env-vars="VERTEXAI_REGION=asia-southeast1" \
  --memory=512MB \
  --timeout=300s
```

### Deploy via GitHub Actions

Push ke branch `main` akan otomatis trigger deployment.

---

## Deployment Frontend

### 1. Fork/Clone Repository

```bash
git clone https://github.com/your-username/research-data-analysis-frontend
cd research-data-analysis-frontend
```

### 2. Update API URL

Edit `js/main.js`:

```javascript
const API_BASE_URL = 'https://asia-southeast1-YOUR_PROJECT_ID.cloudfunctions.net/ResearchDataAnalysis';
```

### 3. Enable GitHub Pages

1. Buka repository Settings
2. Scroll ke **Pages**
3. Source: **GitHub Actions**
4. Push perubahan ke `main`

---

## Konfigurasi GitHub Secrets

### Backend Repository

Tambahkan secrets di **Settings > Secrets and variables > Actions**:

| Secret Name | Deskripsi |
|------------|-----------|
| `GOOGLE_CREDENTIALS` | Isi dari key.json |
| `MONGOSTRING` | MongoDB connection string |
| `PRIVATEKEY` | Ed25519 private key (hex) |
| `PUBLICKEY` | Ed25519 public key (hex) |
| `GCS_BUCKET` | Nama bucket GCS |

Tambahkan variables:

| Variable Name | Deskripsi |
|--------------|-----------|
| `GCP_PROJECT_ID` | ID project GCP |
| `GCP_REGION` | Region (default: asia-southeast1) |
| `VERTEXAI_REGION` | Region Vertex AI |

---

## Testing

### Test Backend API

```bash
# Test endpoint root
curl https://asia-southeast1-YOUR_PROJECT_ID.cloudfunctions.net/ResearchDataAnalysis/

# Test registrasi
curl -X POST \
  https://asia-southeast1-YOUR_PROJECT_ID.cloudfunctions.net/ResearchDataAnalysis/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "full_name": "Test User",
    "institution": "Test University",
    "research_field": "Computer Science"
  }'

# Test login
curl -X POST \
  https://asia-southeast1-YOUR_PROJECT_ID.cloudfunctions.net/ResearchDataAnalysis/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

### Test Frontend

1. Buka URL GitHub Pages
2. Test registrasi dan login
3. Buat proyek baru
4. Upload file data (CSV/Excel/JSON)
5. Dapatkan rekomendasi AI
6. Proses analisis
7. Ekspor hasil

---

## Troubleshooting

### MongoDB Connection Error
- Pastikan IP sudah di-whitelist
- Cek username dan password
- Pastikan connection string benar

### Vertex AI Error
- Pastikan API sudah di-enable
- Cek service account memiliki role `aiplatform.user`
- Pastikan region Vertex AI benar

### CORS Error
- Pastikan backend sudah mengatur CORS headers
- Cek origin yang diizinkan di config.go

### File Upload Error
- Pastikan file tidak melebihi 50MB
- Cek format file (CSV, XLSX, JSON)
- Pastikan GCS bucket sudah dibuat dan accessible
