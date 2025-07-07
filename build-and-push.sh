#!/bin/bash
set -e

# Build and push Crossplane provider-plausible to Docker Hub

REGISTRY=${REGISTRY:-docker.io/rossigee}
IMAGE_NAME=provider-plausible
TAG=${TAG:-develop}

echo "Building provider-plausible..."

# Generate code first
echo "Generating code..."
make generate || true

# Build the binary
echo "Building binary..."
make build || true

# Check if binary exists
if [ ! -f "_output/bin/linux_amd64/provider" ]; then
    echo "Binary not found at _output/bin/linux_amd64/provider"
    exit 1
fi

# Build the Docker image manually
echo "Building Docker image..."
cd cluster/images/provider-plausible
mkdir -p bin/linux_amd64
cp ../../../_output/bin/linux_amd64/provider bin/linux_amd64/
docker build --build-arg TARGETOS=linux --build-arg TARGETARCH=amd64 -t ${REGISTRY}/${IMAGE_NAME}:${TAG} .
cd ../../../

# Push to Docker Hub
echo "Pushing to Docker Hub..."
docker push ${REGISTRY}/${IMAGE_NAME}:${TAG}

echo "Successfully pushed ${REGISTRY}/${IMAGE_NAME}:${TAG}"

# Clean up
rm -rf cluster/images/provider-plausible/bin/