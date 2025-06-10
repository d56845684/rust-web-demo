#!/bin/bash
set -e
# Build new image
docker build -t go-demo .

# Clean up unused images
docker image prune -f
