# cyphera-bq

Format-preserving encryption for BigQuery via Remote UDFs.

A high-performance Go HTTP server that BigQuery calls as a Remote Function. Deployable to Cloud Run or as a local Docker container.

> **Note**: Currently uses a dummy (reversible shift) cipher as a placeholder.
> Real FF1 FPE will be wired in via `cyphera-go`.

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
SELECT cyphera_protect('ssn', '123-45-6789');
SELECT cyphera_unprotect('ssn', cyphera_protect('ssn', '123-45-6789'));
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
