#!/bin/bash
set -euo pipefail

# Build image and run containers
./auto_build.sh
./auto_run.sh

# Ensure containers are stopped on exit
trap 'docker stop rust-demo rust_postgres >/dev/null 2>&1 || true' EXIT

# Wait for server to be ready
for i in {1..30}; do
    if curl -sf http://localhost:8080/test-db > /dev/null; then
        break
    fi
    sleep 1
done

# Login and obtain JWT
TOKEN=$(curl -s -X POST http://localhost:8080/api/login \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"password"}' | jq -r '.token')

if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
    echo "Failed to retrieve auth token" >&2
    exit 1
fi

# Test GET /todos
curl -f -H "Authorization: Bearer $TOKEN" http://localhost:8080/todos

# Test POST /todos
ID=$(curl -s -X POST http://localhost:8080/todos \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"title":"CI Todo","done":false}' | jq -r '.id')

# Test PUT /todos/{id}
curl -f -X PUT http://localhost:8080/todos/$ID \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"title":"Updated CI Todo"}'

# Test POST /todos/{id}/toggle
curl -f -X POST http://localhost:8080/todos/$ID/toggle \
    -H "Authorization: Bearer $TOKEN"

# Test DELETE /todos/{id}
curl -f -X DELETE http://localhost:8080/todos/$ID \
    -H "Authorization: Bearer $TOKEN"

echo "All API endpoints tested successfully."

