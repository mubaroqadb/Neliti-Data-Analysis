# GitHub Actions Workflows

File-file workflow ini perlu dipindahkan ke folder `.github/workflows/` agar dapat aktif.

## Cara Mengaktifkan Workflows

### Langkah 1: Pindahkan file workflow

```bash
# Clone repository
git clone https://github.com/mubaroqadb/Neliti-Data-Analysis.git
cd Neliti-Data-Analysis

# Buat folder .github/workflows
mkdir -p .github/workflows

# Pindahkan file workflow
mv workflows/deploy-backend.yml .github/workflows/
mv workflows/deploy-frontend.yml .github/workflows/

# Hapus folder workflows yang sudah kosong
rm -rf workflows

# Commit dan push
git add .
git commit -m "Move workflows to .github/workflows/"
git push origin main
```

### Langkah 2: Setup GitHub Secrets dan Variables

Buka Settings > Secrets and variables > Actions di repository GitHub.

#### Secrets yang Diperlukan:
- `GOOGLE_CREDENTIALS`: JSON service account key dari GCP
- `MONGOSTRING`: Connection string MongoDB Atlas
- `PRIVATEKEY`: Ed25519 private key (hex format) untuk PASETO
- `PUBLICKEY`: Ed25519 public key (hex format) untuk PASETO
- `GCS_BUCKET`: Nama bucket Google Cloud Storage

#### Variables yang Diperlukan:
- `GCP_PROJECT_ID`: ID project Google Cloud
- `GCP_REGION`: Region GCP (default: asia-southeast1)
- `VERTEXAI_REGION`: Region Vertex AI (default: asia-southeast1)

### Langkah 3: Enable GitHub Pages

1. Buka Settings > Pages
2. Pilih Source: GitHub Actions
3. Workflow akan otomatis deploy ke GitHub Pages

### Langkah 4: Trigger Deployment

Workflows akan berjalan otomatis ketika:
- Ada push ke branch `main` dengan perubahan di folder `backend/` atau `frontend/`
- Manual trigger melalui Actions tab > Run workflow
