#!/bin/bash
set -e

# Run PostgreSQL container
docker stop go-demo go_postgres 2>/dev/null || true
docker rm go-demo go_postgres 2>/dev/null || true
docker run -d --name go_postgres \
    -e POSTGRES_DB=go_demo \
    -e POSTGRES_USER=go_user \
    -e POSTGRES_PASSWORD=go_password \
    -p 5432:5432 \
    -v postgres_data:/var/lib/postgresql/data \
    postgres:15

# Run application container
APP_CONTAINER=$(docker run -d --rm -p 8080:8080 \
    -v "$(pwd)/static:/app/static" \
    --name go-demo \
    --link go_postgres:postgres \
    go-demo)

echo "Application container started with ID $APP_CONTAINER"
