#!/bin/bash

# Script to generate Go client from OpenAPI specification
# This script uses the OpenAPI Generator Docker image to generate a Go client
# from the Operaton REST API specification

set -e

# Configuration
SPEC_FILE="resources/operaton-rest-api.json"
OUTPUT_DIR="internal/operaton"
PACKAGE_NAME="operaton"
GIT_USER_ID="kthoms"
GIT_REPO_ID="o8n"

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}Starting OpenAPI client generation...${NC}"

# Check if spec file exists
if [ ! -f "$SPEC_FILE" ]; then
    echo -e "${RED}Error: Specification file not found at $SPEC_FILE${NC}"
    exit 1
fi

# Create output directory if it doesn't exist
mkdir -p "$OUTPUT_DIR"

echo -e "${YELLOW}Generating Go client using OpenAPI Generator...${NC}"

# Use OpenAPI Generator Docker image to generate the client
# We use docker to ensure consistent generation across environments
docker run --rm \
    -v "${PWD}:/local" \
    openapitools/openapi-generator-cli:v7.2.0 generate \
    -i "/local/${SPEC_FILE}" \
    -g go \
    -o "/local/${OUTPUT_DIR}" \
    --additional-properties=packageName="${PACKAGE_NAME}" \
    --additional-properties=isGoSubmodule=true \
    --additional-properties=withGoMod=false \
    --git-user-id="${GIT_USER_ID}" \
    --git-repo-id="${GIT_REPO_ID}"

echo -e "${GREEN}Client generation completed!${NC}"
echo -e "${YELLOW}Generated files are in: ${OUTPUT_DIR}${NC}"

# Run go mod tidy to update dependencies
echo -e "${YELLOW}Running go mod tidy...${NC}"
go mod tidy

echo -e "${GREEN}Done!${NC}"
