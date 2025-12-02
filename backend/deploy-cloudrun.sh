#!/bin/bash

# Research Data Analysis - Cloud Run Deployment Script
# ======================================================

set -e  # Exit on error

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== Research Data Analysis - Cloud Run Deployment ===${NC}\n"

# Configuration
PROJECT_ID="${GCP_PROJECT_ID:-neliti-480014}"
REGION="${GCP_REGION:-asia-southeast1}"
SERVICE_NAME="research-data-api"
BUCKET_NAME="${GCS_BUCKET:-research-data-uploads}"

# Check if running in Cloud Shell
if [ -n "$CLOUD_SHELL" ]; then
    echo -e "${GREEN}✓ Running in Cloud Shell${NC}"
else
    echo -e "${YELLOW}⚠ Not running in Cloud Shell. Make sure gcloud is configured.${NC}"
fi

# Set project
echo -e "\n${YELLOW}Setting GCP project...${NC}"
gcloud config set project $PROJECT_ID

# Enable required APIs (if not already enabled)
echo -e "\n${YELLOW}Ensuring required APIs are enabled...${NC}"
gcloud services enable run.googleapis.com \
    cloudbuild.googleapis.com \
    storage.googleapis.com \
    aiplatform.googleapis.com

# Create Cloud Storage bucket if it doesn't exist
echo -e "\n${YELLOW}Creating Cloud Storage bucket...${NC}"
if gsutil ls -b gs://$BUCKET_NAME &> /dev/null; then
    echo -e "${GREEN}✓ Bucket gs://$BUCKET_NAME already exists${NC}"
else
    gsutil mb -p $PROJECT_ID -c STANDARD -l $REGION gs://$BUCKET_NAME
    gsutil uniformbucketlevelaccess set on gs://$BUCKET_NAME
    gsutil iam ch allUsers:objectViewer gs://$BUCKET_NAME
    echo -e "${GREEN}✓ Bucket gs://$BUCKET_NAME created${NC}"
fi

# Build and deploy to Cloud Run
echo -e "\n${YELLOW}Building and deploying to Cloud Run...${NC}"
gcloud run deploy $SERVICE_NAME \
    --source . \
    --platform managed \
    --region $REGION \
    --allow-unauthenticated \
    --set-env-vars "MONGOSTRING=${MONGOSTRING},PRIVATEKEY=${PRIVATEKEY},PUBLICKEY=${PUBLICKEY},GCS_BUCKET=${BUCKET_NAME},GCP_PROJECT_ID=${PROJECT_ID},VERTEXAI_REGION=${REGION}" \
    --min-instances 0 \
    --max-instances 10 \
    --memory 512Mi \
    --cpu 1 \
    --timeout 300

# Get service URL
SERVICE_URL=$(gcloud run services describe $SERVICE_NAME --region $REGION --format 'value(status.url)')

echo -e "\n${GREEN}========================================${NC}"
echo -e "${GREEN}✓ Deployment completed successfully!${NC}"
echo -e "${GREEN}========================================${NC}"
echo -e "\nBackend API URL: ${GREEN}$SERVICE_URL${NC}"
echo -e "GCS Bucket: ${GREEN}gs://$BUCKET_NAME${NC}"
echo -e "\nNext steps:"
echo -e "1. Test API: curl $SERVICE_URL"
echo -e "2. Update frontend API_BASE_URL with this URL"
echo -e "3. Configure GitHub Actions secrets"
echo -e "\n"