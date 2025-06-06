#!/bin/bash
# Build new image
docker build -t rust-demo .

# Clean up unused images
docker image prune -f