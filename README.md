# cyphera-bq

[![CI](https://github.com/cyphera-labs/cyphera-bq/actions/workflows/ci.yml/badge.svg)](https://github.com/cyphera-labs/cyphera-bq/actions/workflows/ci.yml)
[![Security](https://github.com/cyphera-labs/cyphera-bq/actions/workflows/codeql.yml/badge.svg)](https://github.com/cyphera-labs/cyphera-bq/actions/workflows/codeql.yml)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue)](LICENSE)

Format-preserving encryption for BigQuery via Remote UDFs.

A high-performance Go HTTP server that BigQuery calls as a Remote Function. Deployable to Cloud Run or as a local Docker container.

Built on [`cyphera-go`](https://github.com/cyphera-labs/cyphera-go).

## Quick Start (Local)

```bash
docker compose up -d
bash demo.sh
```

## Deploy to GCP

```bash
export GCP_PROJECT_ID=your-project
bash deploy.sh
```

Then run the DDL in `bq_setup.sql` to create the BQ remote functions.

## Usage in BigQuery

```sql
-- Protect with a named policy
SELECT cyphera_protect('ssn', '123-45-6789');
-- → 'T01948-37-2150' (tagged, format preserved)

-- Access — tag tells Cyphera which policy to use, no policy name needed
SELECT cyphera_access(cyphera_protect('ssn', '123-45-6789'));
-- → '123-45-6789'
```

## HTTP API

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/` | POST | Encrypt (BQ Remote UDF protocol) |
| `/decrypt` | POST | Decrypt |
| `/health` | GET | Health check |

### Request/Response

```json
{"calls": [["ssn", "123-45-6789"], ["ssn", "987-65-4321"]]}
→ {"replies": ["456-78-9012", "210-98-7654"]}
```
