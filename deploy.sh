#!/bin/bash
# Deploy Cyphera BQ UDF to Google Cloud Run
set -euo pipefail

PROJECT_ID="${GCP_PROJECT_ID:?Set GCP_PROJECT_ID}"
REGION="${GCP_REGION:-us-central1}"
SERVICE_NAME="cyphera-bq-udf"

echo "Building and deploying to Cloud Run..."
gcloud run deploy "$SERVICE_NAME" \
  --source . \
  --region "$REGION" \
  --allow-unauthenticated \
  --set-env-vars "CYPHERA_POLICY_FILE=/etc/cyphera/cyphera.yaml"

SERVICE_URL=$(gcloud run services describe "$SERVICE_NAME" --region "$REGION" --format='value(status.url)')
echo "Service URL: $SERVICE_URL"
echo
echo "Now create the BQ remote function — see bq_setup.sql"
