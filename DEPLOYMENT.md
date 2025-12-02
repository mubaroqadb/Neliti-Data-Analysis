# Research Data Analysis - Deployment Guide

## Prerequisites

- Google Cloud Project: `neliti-480014`
- Cloud Shell access: `sr_mubaroq@unfari.ac.id`
- MongoDB Atlas cluster configured
- GitHub repository: `mubaroqadb/Neliti-Data-Analysis`

## Deployment Steps

### 1. Backend Deployment ke Cloud Run

#### A. Setup Environment Variables

Buka Cloud Shell dan set environment variables:

```bash
export GCP_PROJECT_ID="neliti-480014"
export MONGOSTRING="mongodb+srv://testuser:testpass123@research-cluster.mongodb.net/research_data?retryWrites=true&w=majority"
export PRIVATEKEY="192c4c2d4e98f8e5f3fdb823df93dd85fa25714a25ed06c814d2b9e087c52c0f"
export PUBLICKEY="f78f22b4537b39bf4255780d49e2d3556214ba26735fd7514c171ed8a875915d"
export GCS_BUCKET="research-data-uploads"
export GCP_REGION="asia-southeast1"
```

#### B. Clone Repository

```bash
cd ~
git clone https://github.com/mubaroqadb/Neliti-Data-Analysis.git
cd Neliti-Data-Analysis/backend
```

#### C. Deploy ke Cloud Run

```bash
chmod +x deploy-cloudrun.sh
./deploy-cloudrun.sh
```

Script ini akan:
- Enable required APIs
- Create Cloud Storage bucket
- Build Docker container
- Deploy ke Cloud Run
- Configure environment variables

#### D. Get Backend URL

Setelah deployment berhasil, catat URL backend:

```bash
gcloud run services describe research-data-api --region asia-southeast1 --format 'value(status.url)'
```

Contoh output: `https://research-data-api-xxxxx-as.a.run.app`

### 2. Frontend Deployment ke GitHub Pages

#### A. Update API Base URL

Edit `frontend/js/main.js`, ubah `API_BASE_URL`:

```javascript
const API_BASE_URL = 'https://research-data-api-xxxxx-as.a.run.app';
```

#### B. Enable GitHub Pages

1. Go to repository settings
2. Navigate to **Pages** section
3. Set source: **Deploy from a branch**
4. Select branch: `main`
5. Select folder: `/frontend`
6. Click **Save**

#### C. Access Frontend

Frontend akan tersedia di:
```
https://mubaroqadb.github.io/Neliti-Data-Analysis/
```

### 3. Setup GitHub Actions CI/CD

#### A. Move Workflows ke .github/workflows/

Workflows sudah dipindahkan ke lokasi yang benar.

#### B. Configure GitHub Secrets

Go to repository **Settings > Secrets and variables > Actions**, tambahkan:

**Secrets:**
- `GOOGLE_CREDENTIALS`: Service account JSON key dengan permissions:
  - Cloud Run Admin
  - Cloud Build Editor
  - Service Account User
  - Storage Admin
  
- `MONGOSTRING`: 
  ```
  mongodb+srv://testuser:testpass123@research-cluster.mongodb.net/research_data?retryWrites=true&w=majority
  ```

- `PRIVATEKEY`:
  ```
  192c4c2d4e98f8e5f3fdb823df93dd85fa25714a25ed06c814d2b9e087c52c0f
  ```

- `PUBLICKEY`:
  ```
  f78f22b4537b39bf4255780d49e2d3556214ba26735fd7514c171ed8a875915d
  ```

**Variables:**
- `GCP_PROJECT_ID`: `neliti-480014`
- `GCP_REGION`: `asia-southeast1`
- `GCS_BUCKET`: `research-data-uploads`

#### C. Create Service Account

Di Cloud Shell:

```bash
# Create service account
gcloud iam service-accounts create github-actions \
    --display-name "GitHub Actions Deployer"

# Grant permissions
gcloud projects add-iam-policy-binding neliti-480014 \
    --member="serviceAccount:github-actions@neliti-480014.iam.gserviceaccount.com" \
    --role="roles/run.admin"

gcloud projects add-iam-policy-binding neliti-480014 \
    --member="serviceAccount:github-actions@neliti-480014.iam.gserviceaccount.com" \
    --role="roles/cloudbuild.builds.editor"

gcloud projects add-iam-policy-binding neliti-480014 \
    --member="serviceAccount:github-actions@neliti-480014.iam.gserviceaccount.com" \
    --role="roles/iam.serviceAccountUser"

gcloud projects add-iam-policy-binding neliti-480014 \
    --member="serviceAccount:github-actions@neliti-480014.iam.gserviceaccount.com" \
    --role="roles/storage.admin"

# Create and download key
gcloud iam service-accounts keys create ~/github-actions-key.json \
    --iam-account=github-actions@neliti-480014.iam.gserviceaccount.com

# Display key content
cat ~/github-actions-key.json
```

Copy seluruh isi JSON dan paste ke GitHub Secret `GOOGLE_CREDENTIALS`.

### 4. Testing

#### A. Test Backend API

```bash
BACKEND_URL=$(gcloud run services describe research-data-api --region asia-southeast1 --format 'value(status.url)')

# Test root endpoint
curl $BACKEND_URL

# Test register
curl -X POST $BACKEND_URL/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test User",
    "email": "test@example.com",
    "password": "password123",
    "institution": "Test University"
  }'
```

#### B. Test Frontend

Buka browser dan akses:
```
https://mubaroqadb.github.io/Neliti-Data-Analysis/
```

Test flows:
1. Register new account
2. Login
3. Create new project
4. Upload data file (CSV/XLSX/JSON)
5. View data preview
6. Get AI recommendations
7. Process analysis
8. View results with charts
9. Export results

### 5. Monitoring & Logs

#### View Cloud Run Logs

```bash
gcloud run services logs read research-data-api \
    --region asia-southeast1 \
    --limit 50
```

#### View GitHub Actions Logs

Go to repository **Actions** tab untuk melihat workflow runs.

## Environment Variables Reference

### Backend (Cloud Run)

| Variable | Description | Example |
|----------|-------------|---------|
| MONGOSTRING | MongoDB Atlas connection string | mongodb+srv://user:pass@cluster.mongodb.net/db |
| PRIVATEKEY | PASETO Ed25519 private key (hex) | 192c4c2d4e98f8e5... |
| PUBLICKEY | PASETO Ed25519 public key (hex) | f78f22b4537b39bf... |
| GCS_BUCKET | Cloud Storage bucket name | research-data-uploads |
| GCP_PROJECT_ID | Google Cloud project ID | neliti-480014 |
| VERTEXAI_REGION | Vertex AI region | asia-southeast1 |
| PORT | Server port (auto-set by Cloud Run) | 8080 |

### Frontend (GitHub Pages)

| Variable | Location | Value |
|----------|----------|-------|
| API_BASE_URL | js/main.js | https://research-data-api-xxxxx-as.a.run.app |

## Troubleshooting

### Backend tidak bisa deploy

1. Check Cloud Run quota
2. Verify environment variables
3. Check logs: `gcloud run services logs read research-data-api --region asia-southeast1`

### Frontend tidak bisa akses backend

1. Verify CORS headers di backend
2. Check API_BASE_URL di frontend
3. Verify Cloud Run service allows unauthenticated access

### MongoDB connection failed

1. Verify connection string
2. Check MongoDB Atlas network access (whitelist 0.0.0.0/0)
3. Verify database user credentials

### Vertex AI errors

1. Verify aiplatform.googleapis.com API enabled
2. Check ADC configured correctly
3. Verify project ID dan region

## Support

Untuk issues atau pertanyaan, buka issue di GitHub repository.