#!/bin/bash

# Hardcoded configuration
AWS_REGION="us-east-1"
ECR_REPOSITORY="osprey-web-compiler"
IMAGE_TAG="latest"
ECS_CLUSTER="osprey-cluster"
ECS_SERVICE="osprey-web-compiler"
ECS_TASK_DEFINITION="osprey-web-compiler"

# Get AWS account ID dynamically
AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)

# Computed values
ECR_URI="${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/${ECR_REPOSITORY}"
FULL_IMAGE_URI="${ECR_URI}:${IMAGE_TAG}" 