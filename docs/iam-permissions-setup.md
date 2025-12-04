# IAM Permissions Setup for GitHub Actions

This document explains how to set up the required IAM permissions for the GitHub Actions workflow to deploy to Google Cloud Run.

## Required Permissions

The service account `github-actions-sa@neliti-480014.iam.gserviceaccount.com` needs the following permissions:

1. **Service Account Token Creator** (`iam.serviceAccounts.getAccessToken`)
   - Required for: Authentication and token refresh
   - Resource: Service account itself
   - Role: `roles/iam.serviceAccountTokenCreator`

2. **Artifact Registry Writer** (`artifactregistry.repositories.uploadArtifacts`)
   - Required for: Pushing Docker images to Artifact Registry
   - Resource: Artifact Registry repository
   - Role: `roles/artifactregistry.writer`

3. **Cloud Run Developer** (`run.services.create`, `run.services.update`, etc.)
   - Required for: Deploying to Cloud Run
   - Resource: Cloud Run service
   - Role: `roles/run.developer`

## How to Add Permissions

### Option 1: Using Google Cloud Console

1. Go to [IAM & Admin](https://console.cloud.google.com/iam-admin) in Google Cloud Console
2. Select the project `neliti-480014`
3. Go to "Service Accounts"
4. Find and click on `github-actions-sa@neliti-480014.iam.gserviceaccount.com`
5. Click on "Permissions" tab
6. Click "Grant Access"
7. Add the following roles:
   - `roles/iam.serviceAccountTokenCreator`
   - `roles/artifactregistry.writer`
   - `roles/run.developer`

### Option 2: Using gcloud CLI

```bash
# Set your project
gcloud config set project neliti-480014

# Add Service Account Token Creator role
gcloud iam service-accounts add-iam-policy-binding \
  github-actions-sa@neliti-480014.iam.gserviceaccount.com \
  --member="serviceAccount:github-actions-sa@neliti-480014.iam.gserviceaccount.com" \
  --role="roles/iam.serviceAccountTokenCreator"

# Add Artifact Registry Writer role
gcloud projects add-iam-policy-binding \
  neliti-480014 \
  --member="serviceAccount:github-actions-sa@neliti-480014.iam.gserviceaccount.com" \
  --role="roles/artifactregistry.writer"

# Add Cloud Run Developer role
gcloud projects add-iam-policy-binding \
  neliti-480014 \
  --member="serviceAccount:github-actions-sa@neliti-480014.iam.gserviceaccount.com" \
  --role="roles/run.developer"
```

### Option 3: Using Terraform

```hcl
resource "google_service_account_iam_binding" "token_creator" {
  service_account_id = "projects/neliti-480014/serviceAccounts/github-actions-sa@neliti-480014.iam.gserviceaccount.com"
  role              = "roles/iam.serviceAccountTokenCreator"
  members           = ["serviceAccount:github-actions-sa@neliti-480014.iam.gserviceaccount.com"]
}

resource "google_project_iam_binding" "artifact_registry_writer" {
  project = "neliti-480014"
  role    = "roles/artifactregistry.writer"
  members = ["serviceAccount:github-actions-sa@neliti-480014.iam.gserviceaccount.com"]
}

resource "google_project_iam_binding" "cloud_run_developer" {
  project = "neliti-480014"
  role    = "roles/run.developer"
  members = ["serviceAccount:github-actions-sa@neliti-480014.iam.gserviceaccount.com"]
}
```

## Verify Permissions

After adding the permissions, you can verify them with:

```bash
# Check service account permissions
gcloud iam service-accounts get-iam-policy \
  github-actions-sa@neliti-480014.iam.gserviceaccount.com

# Check project-level permissions
gcloud projects get-iam-policy neliti-480014 \
  --flatten="bindings[].members" \
  --format="table(bindings.role, bindings.members)"
```

## Troubleshooting

If you still encounter permission issues after adding these roles:

1. Ensure the permissions have propagated (can take a few minutes)
2. Check if there are any deny policies that might override these permissions
3. Verify the service account email is correct
4. Check organization-level policies that might restrict these permissions