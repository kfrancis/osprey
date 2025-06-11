#!/bin/bash

set -e

# Source configuration
source "$(dirname "$0")/config.sh"

echo "🐳 ECR IMAGE BUILD & PUSH"
echo "========================="

# Navigate to webcompiler directory
cd "$(dirname "$0")"

# Check required tools
if ! command -v aws &> /dev/null; then
    echo "❌ AWS CLI is required but not installed!"
    exit 1
fi

if ! command -v docker &> /dev/null; then
    echo "❌ Docker is required but not installed!"
    exit 1
fi

# Check AWS credentials
if ! aws sts get-caller-identity &> /dev/null; then
    echo "❌ AWS credentials not configured!"
    echo "Run: aws configure"
    exit 1
fi

echo "📋 Configuration:"
echo "  AWS Account: ${AWS_ACCOUNT_ID}"
echo "  Region: ${AWS_REGION}"
echo "  ECR Repository: ${ECR_REPOSITORY}"
echo "  Image URI: ${FULL_IMAGE_URI}"
echo ""

# Create ECR repository if it doesn't exist
echo "🏗️  Ensuring ECR repository exists..."
aws ecr describe-repositories --repository-names "${ECR_REPOSITORY}" --region "${AWS_REGION}" > /dev/null 2>&1 || {
    echo "Creating ECR repository: ${ECR_REPOSITORY}"
    aws ecr create-repository \
        --repository-name "${ECR_REPOSITORY}" \
        --region "${AWS_REGION}" \
        --image-scanning-configuration scanOnPush=true
}

# Login to ECR
echo "🔐 Logging in to ECR..."
aws ecr get-login-password --region "${AWS_REGION}" | docker login --username AWS --password-stdin "${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com"

# Force delete existing image if it exists
echo "🗑️  Deleting existing image..."
aws ecr batch-delete-image --repository-name "${ECR_REPOSITORY}" --image-ids imageTag="${IMAGE_TAG}" --region "${AWS_REGION}" 2>/dev/null || echo "No existing image to delete"

# Build and push the Docker image
echo "🔨 Building Docker image for linux/amd64..."
docker build --platform linux/amd64 --no-cache -t "${FULL_IMAGE_URI}" -f Dockerfile ..

echo "⬆️  Pushing to ECR (force replace)..."
docker push "${FULL_IMAGE_URI}"

echo ""
echo "✅ ECR deployment complete!"
echo "🌐 Image URI: ${FULL_IMAGE_URI}"
echo ""
echo "🔧 Next step: ./deploy-ecs.sh" 