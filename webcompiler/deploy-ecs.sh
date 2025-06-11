#!/bin/bash

set -e

# Source configuration
source "$(dirname "$0")/config.sh"

echo "🚢 ECS FARGATE DEPLOYMENT"
echo "========================="

# Navigate to webcompiler directory
cd "$(dirname "$0")"

# Check required tools
if ! command -v aws &> /dev/null; then
    echo "❌ AWS CLI is required but not installed!"
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
echo "  ECR Image: ${FULL_IMAGE_URI}"
echo "  ECS Cluster: ${ECS_CLUSTER}"
echo "  ECS Service: ${ECS_SERVICE}"
echo "  Task Definition: ${ECS_TASK_DEFINITION}"
echo ""

# Verify ECR image exists
echo "🔍 Checking if ECR image exists..."
if ! aws ecr describe-images --repository-name "${ECR_REPOSITORY}" --image-ids imageTag="${IMAGE_TAG}" --region "${AWS_REGION}" > /dev/null 2>&1; then
    echo "❌ ECR image ${FULL_IMAGE_URI} not found!"
    echo "Run ./deploy-ecr.sh first to build and push the image"
    exit 1
fi
echo "✅ ECR image found: ${FULL_IMAGE_URI}"

# Ensure ECS cluster exists and is active
echo "🏗️  Ensuring ECS cluster exists..."
CLUSTER_STATUS=$(aws ecs describe-clusters --clusters "${ECS_CLUSTER}" --region "${AWS_REGION}" --query 'clusters[0].status' --output text 2>/dev/null || echo "NOTFOUND")
if [ "$CLUSTER_STATUS" != "ACTIVE" ]; then
    echo "Creating ECS cluster: ${ECS_CLUSTER}"
    aws ecs create-cluster --cluster-name "${ECS_CLUSTER}" --region "${AWS_REGION}"
fi
echo "✅ ECS cluster ready"

# Get default VPC and subnets
echo "🔍 Getting VPC configuration..."
DEFAULT_VPC_ID=$(aws ec2 describe-vpcs --filters "Name=is-default,Values=true" --query 'Vpcs[0].VpcId' --output text --region "${AWS_REGION}")
SUBNET_IDS=$(aws ec2 describe-subnets --filters "Name=vpc-id,Values=${DEFAULT_VPC_ID}" --query 'Subnets[*].SubnetId' --output text --region "${AWS_REGION}")
echo "VPC: ${DEFAULT_VPC_ID}"
echo "Subnets: ${SUBNET_IDS}"

# Create security group for the service (or reuse existing)
echo "🔐 Creating security group..."
SECURITY_GROUP_ID=$(aws ec2 describe-security-groups \
    --filters "Name=group-name,Values=osprey-web-compiler-sg" "Name=vpc-id,Values=${DEFAULT_VPC_ID}" \
    --query 'SecurityGroups[0].GroupId' \
    --output text \
    --region "${AWS_REGION}" 2>/dev/null)

if [ "$SECURITY_GROUP_ID" = "None" ] || [ -z "$SECURITY_GROUP_ID" ]; then
    echo "Creating new security group..."
    SECURITY_GROUP_ID=$(aws ec2 create-security-group \
        --group-name "osprey-web-compiler-sg" \
        --description "Security group for Osprey Web Compiler - Cloudflare IPs only" \
        --vpc-id "${DEFAULT_VPC_ID}" \
        --region "${AWS_REGION}" \
        --query 'GroupId' \
        --output text)
else
    echo "Using existing security group..."
fi

echo "Security Group: ${SECURITY_GROUP_ID}"

# Add HTTP access rule (restrict to Cloudflare IPs for security)
echo "🔐 Adding HTTP access rules for Cloudflare IPs only..."

# Cloudflare IPv4 ranges
CLOUDFLARE_IPV4_RANGES=(
    "173.245.48.0/20"
    "103.21.244.0/22"
    "103.22.200.0/22"
    "103.31.4.0/22"
    "141.101.64.0/18"
    "108.162.192.0/18"
    "190.93.240.0/20"
    "188.114.96.0/20"
    "197.234.240.0/22"
    "198.41.128.0/17"
    "162.158.0.0/15"
    "104.16.0.0/13"
    "104.24.0.0/14"
    "172.64.0.0/13"
    "131.0.72.0/22"
)

# Add rules for each Cloudflare IPv4 range
for cidr in "${CLOUDFLARE_IPV4_RANGES[@]}"; do
    aws ec2 authorize-security-group-ingress \
        --group-id "${SECURITY_GROUP_ID}" \
        --protocol tcp \
        --port 3001 \
        --cidr "${cidr}" \
        --region "${AWS_REGION}" 2>/dev/null || echo "Rule for ${cidr} already exists"
done

# Create CloudWatch log group
echo "📊 Creating CloudWatch log group..."
aws logs create-log-group --log-group-name "/ecs/${ECS_TASK_DEFINITION}" --region "${AWS_REGION}" 2>/dev/null || true

# Create task definition for Fargate
echo "📝 Creating Fargate task definition..."
TASK_DEFINITION_JSON=$(cat <<EOF
{
  "family": "${ECS_TASK_DEFINITION}",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "256",
  "memory": "512",
  "executionRoleArn": "arn:aws:iam::${AWS_ACCOUNT_ID}:role/ecsTaskExecutionRole",
  "containerDefinitions": [
    {
      "name": "${ECS_SERVICE}",
      "image": "${FULL_IMAGE_URI}",
      "portMappings": [
        {
          "containerPort": 3001,
          "protocol": "tcp"
        }
      ],
      "essential": true,
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/${ECS_TASK_DEFINITION}",
          "awslogs-region": "${AWS_REGION}",
          "awslogs-stream-prefix": "ecs"
        }
      },
      "environment": [
        {
          "name": "NODE_ENV",
          "value": "production"
        },
        {
          "name": "AWS_REGION",
          "value": "${AWS_REGION}"
        }
      ]
    }
  ]
}
EOF
)

# Register task definition
echo "📋 Registering task definition..."
TEMP_JSON_FILE=$(mktemp)
echo "$TASK_DEFINITION_JSON" > "$TEMP_JSON_FILE"
TASK_DEFINITION_ARN=$(aws ecs register-task-definition --region "${AWS_REGION}" --cli-input-json "file://${TEMP_JSON_FILE}" --query 'taskDefinition.taskDefinitionArn' --output text)
rm "$TEMP_JSON_FILE"
echo "Task definition registered: ${TASK_DEFINITION_ARN}"

# Create Application Load Balancer
echo "🎯 Creating Application Load Balancer..."
LOAD_BALANCER_NAME="osprey-alb"
TARGET_GROUP_NAME="osprey-targets"

# Create target group
echo "🎯 Creating target group..."
TARGET_GROUP_ARN=$(aws elbv2 create-target-group \
    --name "${TARGET_GROUP_NAME}" \
    --protocol HTTP \
    --port 3001 \
    --vpc-id "${DEFAULT_VPC_ID}" \
    --target-type ip \
    --health-check-path "/health" \
    --health-check-protocol HTTP \
    --health-check-port 3001 \
    --health-check-interval-seconds 30 \
    --health-check-timeout-seconds 5 \
    --healthy-threshold-count 2 \
    --unhealthy-threshold-count 5 \
    --region "${AWS_REGION}" \
    --query 'TargetGroups[0].TargetGroupArn' \
    --output text 2>/dev/null || \
    aws elbv2 describe-target-groups \
        --names "${TARGET_GROUP_NAME}" \
        --region "${AWS_REGION}" \
        --query 'TargetGroups[0].TargetGroupArn' \
        --output text)

echo "Target Group ARN: ${TARGET_GROUP_ARN}"

# Create load balancer
echo "🎯 Creating load balancer..."
LOAD_BALANCER_ARN=$(aws elbv2 create-load-balancer \
    --name "${LOAD_BALANCER_NAME}" \
    --subnets ${SUBNET_IDS} \
    --security-groups "${SECURITY_GROUP_ID}" \
    --region "${AWS_REGION}" \
    --query 'LoadBalancers[0].LoadBalancerArn' \
    --output text 2>/dev/null || \
    aws elbv2 describe-load-balancers \
        --names "${LOAD_BALANCER_NAME}" \
        --region "${AWS_REGION}" \
        --query 'LoadBalancers[0].LoadBalancerArn' \
        --output text)

echo "Load Balancer ARN: ${LOAD_BALANCER_ARN}"

# Get load balancer DNS name
LOAD_BALANCER_DNS=$(aws elbv2 describe-load-balancers \
    --load-balancer-arns "${LOAD_BALANCER_ARN}" \
    --region "${AWS_REGION}" \
    --query 'LoadBalancers[0].DNSName' \
    --output text)

echo "Load Balancer DNS: ${LOAD_BALANCER_DNS}"

# Create listener
echo "🎯 Creating load balancer listener..."
aws elbv2 create-listener \
    --load-balancer-arn "${LOAD_BALANCER_ARN}" \
    --protocol HTTP \
    --port 80 \
    --default-actions Type=forward,TargetGroupArn="${TARGET_GROUP_ARN}" \
    --region "${AWS_REGION}" 2>/dev/null || echo "Listener already exists"

# Convert subnet IDs to comma-separated format
SUBNET_LIST=$(echo $SUBNET_IDS | tr ' ' ',' | sed 's/,$//')

# Create or update the Fargate service
echo "🚢 Deploying Fargate service..."
SERVICE_STATUS=$(aws ecs describe-services --cluster "${ECS_CLUSTER}" --services "${ECS_SERVICE}" --region "${AWS_REGION}" --query 'services[0].status' --output text 2>/dev/null || echo "NOTFOUND")
if [ "$SERVICE_STATUS" = "ACTIVE" ]; then
    echo "Updating existing service with new task definition..."
    aws ecs update-service \
        --cluster "${ECS_CLUSTER}" \
        --service "${ECS_SERVICE}" \
        --task-definition "${TASK_DEFINITION_ARN}" \
        --region "${AWS_REGION}"
else
    echo "Creating new service..."
    aws ecs create-service \
        --cluster "${ECS_CLUSTER}" \
        --service-name "${ECS_SERVICE}" \
        --task-definition "${TASK_DEFINITION_ARN}" \
        --desired-count 1 \
        --launch-type FARGATE \
        --network-configuration "awsvpcConfiguration={subnets=[${SUBNET_LIST}],securityGroups=[\"${SECURITY_GROUP_ID}\"],assignPublicIp=ENABLED}" \
        --load-balancers targetGroupArn="${TARGET_GROUP_ARN}",containerName="${ECS_SERVICE}",containerPort=3001 \
        --region "${AWS_REGION}"
fi

# Clean up old task definition revisions (keep only latest)
echo "🧹 Cleaning up old task definition revisions..."
OLD_TASK_DEFS=$(aws ecs list-task-definitions --family-prefix "${ECS_TASK_DEFINITION}" --region "${AWS_REGION}" --status ACTIVE --output json | jq -r '.taskDefinitionArns[0:-1][]' 2>/dev/null || echo "")
if [ -n "$OLD_TASK_DEFS" ]; then
    echo "$OLD_TASK_DEFS" | xargs -I {} aws ecs deregister-task-definition --task-definition {} --region "${AWS_REGION}" 2>/dev/null || true
    echo "✅ Old task definitions cleaned up"
else
    echo "✅ No old task definitions to clean up"
fi

echo ""
echo "🎯 STABLE BACKEND URL: http://${LOAD_BALANCER_DNS}"
echo ""
echo "✅ Fargate service deployed successfully!"
echo ""
echo "📋 Next steps:"
echo "  1. Your service will be available at: http://${LOAD_BALANCER_DNS}"
echo "  2. Use this URL for the Cloudflare gateway backend_url"
echo "  3. Save this URL - it won't change!"
echo "" 