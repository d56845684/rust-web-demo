#!/bin/bash
set -e

# Run PostgreSQL container
docker stop rust-demo rust_postgres 2>/dev/null || true
docker rm rust-demo rust_postgres 2>/dev/null || true
docker run -d --name rust_postgres \
    -e POSTGRES_DB=rust_demo \
    -e POSTGRES_USER=rust_user \
    -e POSTGRES_PASSWORD=rust_password \
    -p 5432:5432 \
    -v postgres_data:/var/lib/postgresql/data \
    postgres:15

# Run application container
APP_CONTAINER=$(docker run -d --rm -p 8080:8080 \
    -v "$(pwd)/static:/app/static" \
    --name rust-demo \
    --link rust_postgres:postgres \
    -e RUST_LOG=info \
    rust-demo)

echo "Application container started with ID $APP_CONTAINER"
