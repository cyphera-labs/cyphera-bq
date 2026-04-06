#!/bin/bash
# Cyphera BQ UDF Demo — test the Go HTTP server locally
# Run: docker compose up -d && bash demo.sh

BASE_URL="${1:-http://localhost:8080}"

echo "=== Health Check ==="
curl -s "$BASE_URL/health" | python3 -m json.tool
echo

echo "=== Encrypt SSNs (policy-based) ==="
curl -s -X POST "$BASE_URL/" \
  -H "Content-Type: application/json" \
  -d '{"calls": [["ssn", "123-45-6789"], ["ssn", "987-65-4321"], ["ssn", "555-12-3456"]]}' \
  | python3 -m json.tool
echo

echo "=== Round-trip decrypt ==="
ENCRYPTED=$(curl -s -X POST "$BASE_URL/" \
  -H "Content-Type: application/json" \
  -d '{"calls": [["ssn", "123-45-6789"]]}' | python3 -c "import sys,json; print(json.load(sys.stdin)['replies'][0])")
echo "Encrypted: $ENCRYPTED"
curl -s -X POST "$BASE_URL/decrypt" \
  -H "Content-Type: application/json" \
  -d "{\"calls\": [[\"ssn\", \"$ENCRYPTED\"]]}" \
  | python3 -m json.tool
echo

echo "=== Direct FF1 encrypt (3-arg) ==="
curl -s -X POST "$BASE_URL/" \
  -H "Content-Type: application/json" \
  -d '{"calls": [["123456789", "2B7E151628AED2A6ABF7158809CF4F3C", "digits"]]}' \
  | python3 -m json.tool
echo

echo "=== Encrypt credit cards ==="
curl -s -X POST "$BASE_URL/" \
  -H "Content-Type: application/json" \
  -d '{"calls": [["credit_card", "4111-1111-1111-1111"], ["credit_card", "5500-0000-0000-0004"]]}' \
  | python3 -m json.tool
