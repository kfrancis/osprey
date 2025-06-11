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

ELASTIC_IP_FILE="$(dirname "$0")/.elastic-ip"

echo "🎯 ONE-TIME ELASTIC IP SETUP"
echo "============================"
echo ""

# Check if we already have an Elastic IP saved
if [[ -f "$ELASTIC_IP_FILE" ]]; then
    SAVED_IP=$(cat "$ELASTIC_IP_FILE")
    print_warning "Elastic IP already exists: $SAVED_IP"
    
    # Find the current running instance and reassociate
    CURRENT_INSTANCE=$(aws ec2 describe-instances \
        --region "${AWS_REGION}" \
        --query 'Reservations[*].Instances[?State.Name==`running`].InstanceId' \
        --output text | head -n1)
    
    if [[ -n "$CURRENT_INSTANCE" ]]; then
        ALLOCATION_ID=$(aws ec2 describe-addresses \
            --public-ips "$SAVED_IP" \
            --region "${AWS_REGION}" \
            --query 'Addresses[0].AllocationId' \
            --output text 2>/dev/null)
        
        if [[ -n "$ALLOCATION_ID" && "$ALLOCATION_ID" != "None" ]]; then
            print_step "Reassociating existing Elastic IP with new instance..."
            aws ec2 associate-address \
                --instance-id "$CURRENT_INSTANCE" \
                --allocation-id "$ALLOCATION_ID" \
                --region "${AWS_REGION}" > /dev/null
            
            echo ""
            echo "🎯 YOUR PERMANENT URL: http://$SAVED_IP:3001"
            echo ""
            echo "✅ Elastic IP reassociated!"
            exit 0
        fi
    fi
fi

# Check AWS credentials
if ! aws sts get-caller-identity &> /dev/null; then
    print_error "AWS credentials not configured!"
    print_error "Run: aws configure"
    exit 1
fi

# Get the most recent instance (assuming it's the one we want)
print_step "Finding your EC2 instance..."
INSTANCE_ID=$(aws ec2 describe-instances \
    --region "${AWS_REGION}" \
    --query 'Reservations[*].Instances[?State.Name==`running`].InstanceId' \
    --output text | head -n1)

if [[ -z "$INSTANCE_ID" ]]; then
    print_error "No running EC2 instances found!"
    print_error "Run ./deploy.sh first to create an instance"
    exit 1
fi

print_status "Found running instance: $INSTANCE_ID"

# Allocate Elastic IP
print_step "Allocating Elastic IP..."
ALLOCATION_ID=$(aws ec2 allocate-address \
    --domain vpc \
    --region "${AWS_REGION}" \
    --query 'AllocationId' \
    --output text)

print_status "Allocated Elastic IP: $ALLOCATION_ID"

# Associate with instance
print_step "Associating Elastic IP with instance..."
ELASTIC_IP=$(aws ec2 describe-addresses \
    --allocation-ids "$ALLOCATION_ID" \
    --region "${AWS_REGION}" \
    --query 'Addresses[0].PublicIp' \
    --output text)

aws ec2 associate-address \
    --instance-id "$INSTANCE_ID" \
    --allocation-id "$ALLOCATION_ID" \
    --region "${AWS_REGION}" > /dev/null

# Save the IP for future reference
echo "$ELASTIC_IP" > "$ELASTIC_IP_FILE"

echo ""
echo "🎉 SUCCESS!"
echo "==========="
echo ""
echo "🎯 YOUR PERMANENT URL: http://$ELASTIC_IP:3001"
echo ""
echo "✅ Elastic IP saved to: $ELASTIC_IP_FILE"
echo "✅ This IP will never change!"
echo "✅ You never need to run this script again!"
echo ""
echo "🚀 Next steps:"
echo "   1. Wait 2-3 minutes for the service to start"
echo "   2. Test: curl http://$ELASTIC_IP:3001/api"
echo "   3. Setup Cloudflare: cd cloudflare && ./set-backend.sh http://$ELASTIC_IP:3001"
echo "" 