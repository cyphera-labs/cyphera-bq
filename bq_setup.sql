-- BigQuery Remote UDF Setup
-- Replace <PROJECT_ID>, <REGION>, <DATASET>, and <SERVICE_URL> with your values.

-- 1. Create a Cloud Resource connection
-- bq mk --connection --connection_type=CLOUD_RESOURCE \
--   --project_id=<PROJECT_ID> --location=<REGION> cyphera-connection

-- 2. Grant the connection's service account permission to invoke the Cloud Run service
-- gcloud run services add-iam-policy-binding cyphera-bq-udf \
--   --region=<REGION> \
--   --member="serviceAccount:<CONNECTION_SA>" \
--   --role="roles/run.invoker"

-- 3. Create the remote functions
CREATE OR REPLACE FUNCTION `<PROJECT_ID>.<DATASET>`.cyphera_protect(policy_name STRING, value STRING)
RETURNS STRING
REMOTE WITH CONNECTION `<PROJECT_ID>.<REGION>.cyphera-connection`
OPTIONS (
  endpoint = '<SERVICE_URL>'
);

CREATE OR REPLACE FUNCTION `<PROJECT_ID>.<DATASET>`.cyphera_unprotect(policy_name STRING, value STRING)
RETURNS STRING
REMOTE WITH CONNECTION `<PROJECT_ID>.<REGION>.cyphera-connection`
OPTIONS (
  endpoint = '<SERVICE_URL>/decrypt'
);

-- 4. Use it!
-- SELECT cyphera_protect('ssn', ssn) FROM `dataset.patients`;
