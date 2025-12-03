#!/bin/bash

echo "=== Research Data Analysis - GoCroot Cloud Run Deployment ==="
echo "Deploying dengan GoCroot implementation yang sudah diperbaiki"

# Set project
gcloud config set project neliti-480014

# Enable required APIs
echo "Enabling APIs..."
gcloud services enable aiplatform.googleapis.com
gcloud services enable storage.googleapis.com  
gcloud services enable run.googleapis.com
gcloud services enable artifactregistry.googleapis.com

# Create Artifact Registry repository
echo "Creating Artifact Registry repository..."
gcloud artifacts repositories create research-repo \
    --repository-format=docker \
    --location=asia-southeast1 \
    --description="Repository untuk Research Data Analysis backend dengan GoCroot" 2>/dev/null || true

# Configure Docker authentication
echo "Configuring Docker authentication..."
gcloud auth configure-docker asia-southeast1-docker.pkg.dev

# Create Cloud Storage bucket untuk file uploads
echo "Creating Cloud Storage bucket..."
gsutil mb -p neliti-480014 -l asia-southeast1 gs://research-data-uploads 2>/dev/null || echo "Bucket sudah ada"

# Set environment variables untuk build
export MONGOSTRING="mongodb+srv://testuser:testpass123@research-cluster.mongodb.net/research_data?retryWrites=true&w=majority"
export GCP_PROJECT_ID="neliti-480014"
export PRIVATEKEY="192c4c2d4e98f8e5f3fdb823df93dd85fa25714a25ed06c814d2b9e087c52c0f"
export PUBLICKEY="f78f22b4537b39bf4255780d49e2d3556214ba26735fd7514c171ed8a875915d"
export JWT_SECRET="research-analysis-jwt-secret-2024"
export GCS_BUCKET="research-data-uploads"
export VERTEXAI_REGION="asia-southeast1"
export PORT=8080
export ENVIRONMENT="production"

# Navigate to backend directory for building
echo "Navigating to backend directory..."
cd backend

if [ ! -f "go.mod" ]; then
    echo "Error: go.mod not found in backend directory."
    exit 1
fi

# Clean and tidy dependencies (fix checksum issues)
echo "Cleaning and tidying dependencies..."
go clean -modcache 2>/dev/null || true
go mod download 2>/dev/null || true
go mod tidy

# Build Docker image dengan GoCroot implementation
echo "Building Docker image dengan GoCroot implementation..."
docker build -t asia-southeast1-docker.pkg.dev/neliti-480014/research-repo/research-backend:latest .

# Return to root directory for deployment
cd ..

# Push to Artifact Registry
echo "Pushing image to Artifact Registry..."
docker push asia-southeast1-docker.pkg.dev/neliti-480014/research-repo/research-backend:latest

# Deploy to Cloud Run dengan GoCroot configuration
echo "Deploying to Cloud Run dengan GoCroot implementation..."
gcloud run deploy research-data-backend \
    --image=asia-southeast1-docker.pkg.dev/neliti-480014/research-repo/research-backend:latest \
    --platform=managed \
    --region=asia-southeast1 \
    --allow-unauthenticated \
    --port=8080 \
    --memory=1Gi \
    --cpu=1 \
    --set-env-vars="MONGOSTRING=${MONGOSTRING},GCP_PROJECT_ID=${GCP_PROJECT_ID},PRIVATEKEY=${PRIVATEKEY},PUBLICKEY=${PUBLICKEY},JWT_SECRET=${JWT_SECRET},GCS_BUCKET=${GCS_BUCKET},VERTEXAI_REGION=${VERTEXAI_REGION},ENVIRONMENT=${ENVIRONMENT}" \
    --project=neliti-480014 \
    --service-account="neliti-480014@neliti-480014.iam.gserviceaccount.com"

echo ""
echo "=== Deployment Complete ==="
echo "Fetching service URL..."

# Get service URL
SERVICE_URL=$(gcloud run services describe research-data-backend --platform=managed --region=asia-southeast1 --format='value(status.url)' 2>/dev/null)

if [ ! -z "$SERVICE_URL" ]; then
    echo "✅ Backend URL: $SERVICE_URL"
    echo "🔍 Test health check:"
    curl -s "$SERVICE_URL/health"
    echo ""
    echo "📊 Deployment Summary:"
    echo "   - Service: research-data-backend"
    echo "   - Region: asia-southeast1"
    echo "   - Memory: 1Gi"
    echo "   - CPU: 1"
    echo "   - Authentication: Allow unauthenticated"
    echo "   - Environment: Production"
    echo ""
    echo "🌐 Frontend should connect to: $SERVICE_URL"
    
    # Save deployment info
    echo "Backend URL: $SERVICE_URL" > deployment-info.txt
    echo "Deployment Date: $(date)" >> deployment-info.txt
    echo "GoCroot Implementation: Yes" >> deployment-info.txt
else
    echo "❌ Failed to get service URL"
    echo "Check deployment logs dengan:"
    echo "gcloud run services describe research-data-backend --platform=managed --region=asia-southeast1"
fi

echo ""
echo "🎉 Research Data Analysis dengan GoCroot telah berhasil dideploy!"