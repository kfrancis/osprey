#!/bin/bash

set -e

# Source configuration
source "$(dirname "$0")/config.sh"

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

show_help() {
    echo "🚀 OSPREY WEB COMPILER DEPLOYMENT"
    echo "================================"
    echo ""
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --help, -h      Show this help"
    echo ""
    echo "Deploys to EC2 (~$3.50/month) by default"
    echo ""
}

deploy_to_ecr() {
    print_step "Building and pushing Docker image to ECR..."
    ./deploy-ecr.sh
}

deploy_to_ec2() {
    print_step "Deploying to EC2 instance..."
    
    # Navigate to webcompiler directory
    cd "$(dirname "$0")"

    # Check required tools
    if ! command -v aws &> /dev/null; then
        print_error "AWS CLI is required but not installed!"
        exit 1
    fi

    # Check AWS credentials
    if ! aws sts get-caller-identity &> /dev/null; then
        print_error "AWS credentials not configured!"
        print_error "Run: aws configure"
        exit 1
    fi

    print_status "Configuration:"
    print_status "  AWS Account: ${AWS_ACCOUNT_ID}"
    print_status "  Region: ${AWS_REGION}"
    print_status "  ECR Image: ${FULL_IMAGE_URI}"

    # Verify ECR image exists
    print_status "Checking if ECR image exists..."
    if ! aws ecr describe-images --repository-name "${ECR_REPOSITORY}" --image-ids imageTag="${IMAGE_TAG}" --region "${AWS_REGION}" > /dev/null 2>&1; then
        print_error "ECR image ${FULL_IMAGE_URI} not found!"
        print_error "Building image first..."
        deploy_to_ecr
    fi
    print_status "✅ ECR image found: ${FULL_IMAGE_URI}"

    # Get default VPC
    print_status "Getting VPC configuration..."
    DEFAULT_VPC_ID=$(aws ec2 describe-vpcs --filters "Name=is-default,Values=true" --query 'Vpcs[0].VpcId' --output text --region "${AWS_REGION}")
    SUBNET_ID=$(aws ec2 describe-subnets --filters "Name=vpc-id,Values=${DEFAULT_VPC_ID}" --query 'Subnets[0].SubnetId' --output text --region "${AWS_REGION}")
    print_status "VPC: ${DEFAULT_VPC_ID}"
    print_status "Subnet: ${SUBNET_ID}"

    # Create security group
    print_status "Creating security group..."
    SECURITY_GROUP_ID=$(aws ec2 create-security-group \
        --group-name "osprey-sg" \
        --description "Security group for Osprey Web Compiler" \
        --vpc-id "${DEFAULT_VPC_ID}" \
        --region "${AWS_REGION}" \
        --query 'GroupId' \
        --output text 2>/dev/null || \
        aws ec2 describe-security-groups \
            --filters "Name=group-name,Values=osprey-sg" "Name=vpc-id,Values=${DEFAULT_VPC_ID}" \
            --query 'SecurityGroups[0].GroupId' \
            --output text \
            --region "${AWS_REGION}")

    print_status "Security Group: ${SECURITY_GROUP_ID}"

    # Add HTTP and SSH access rules
    print_status "Adding access rules..."
    aws ec2 authorize-security-group-ingress \
        --group-id "${SECURITY_GROUP_ID}" \
        --protocol tcp \
        --port 3001 \
        --cidr 0.0.0.0/0 \
        --region "${AWS_REGION}" 2>/dev/null || echo "Port 3001 rule already exists"

    aws ec2 authorize-security-group-ingress \
        --group-id "${SECURITY_GROUP_ID}" \
        --protocol tcp \
        --port 22 \
        --cidr 0.0.0.0/0 \
        --region "${AWS_REGION}" 2>/dev/null || echo "SSH rule already exists"

    # Get latest Amazon Linux 2 AMI
    AMI_ID=$(aws ec2 describe-images \
        --owners amazon \
        --filters "Name=name,Values=amzn2-ami-hvm-*-x86_64-gp2" "Name=state,Values=available" \
        --query 'Images | sort_by(@, &CreationDate) | [-1].ImageId' \
        --output text \
        --region "${AWS_REGION}")

    # Create SIMPLE user data script for Docker setup
    USER_DATA=$(cat <<EOF
#!/bin/bash
yum update -y
yum install -y docker
systemctl start docker
systemctl enable docker
usermod -a -G docker ec2-user

# Install AWS CLI v2
curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip awscliv2.zip
./aws/install

# Login to ECR and run container
aws ecr get-login-password --region ${AWS_REGION} | docker login --username AWS --password-stdin ${ECR_URI}
docker run -d --restart unless-stopped -p 3001:3001 --name osprey-compiler ${FULL_IMAGE_URI}
EOF
)

    # Launch EC2 instance (t2.nano - cheapest!)
    print_status "Launching t2.nano EC2 instance..."
    INSTANCE_ID=$(aws ec2 run-instances \
        --image-id "${AMI_ID}" \
        --count 1 \
        --instance-type t2.nano \
        --security-group-ids "${SECURITY_GROUP_ID}" \
        --subnet-id "${SUBNET_ID}" \
        --user-data "${USER_DATA}" \
        --region "${AWS_REGION}" \
        --query 'Instances[0].InstanceId' \
        --output text)

    print_status "Instance ID: ${INSTANCE_ID}"

    # Wait for instance to be running
    print_status "Waiting for instance to start..."
    aws ec2 wait instance-running --instance-ids "${INSTANCE_ID}" --region "${AWS_REGION}"

    # Get temporary IP for initial deployment
    INSTANCE_IP=$(aws ec2 describe-instances \
        --instance-ids "${INSTANCE_ID}" \
        --region "${AWS_REGION}" \
        --query 'Reservations[0].Instances[0].PublicIpAddress' \
        --output text)

    echo ""
    echo "🎯 TEMPORARY URL: http://${INSTANCE_IP}:3001"
    echo ""
    echo "⚠️  This IP will change if you reboot the instance!"
    echo "    Run ./setup-elastic-ip.sh for a permanent URL"
    echo ""
    
    echo ""
    echo "✅ EC2 deployment completed!"
    echo ""
    echo "💰 Monthly cost: ~$3.50"
    echo ""
    echo "🚀 Next steps:"
    echo "  1. Setup permanent URL: ./setup-elastic-ip.sh"
    echo "  2. Deploy Cloudflare: cd cloudflare && ./deploy.sh"
    echo ""
    echo "🔧 Management:"
    echo "  - Instance ID: ${INSTANCE_ID}"
    echo "  - Temp IP: ${INSTANCE_IP}"
    echo "  - SSH: ssh ec2-user@${INSTANCE_IP}"
}

# Main execution
main() {
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --help|-h)
                show_help
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    echo "💰 OSPREY WEB COMPILER DEPLOYMENT"
    echo "================================="
    echo "Target cost: ~$3.50/month"
    echo ""
    deploy_to_ec2
}

# Navigate to webcompiler directory
cd "$(dirname "$0")"

main "$@" 