#!/bin/sh

# Cloudflare Worker Deployment Script for Osprey Web Compiler Proxy
# This script deploys the proxy worker to Cloudflare Workers

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# Load environment variables from .env file
if [ -f "$SCRIPT_DIR/.env" ]; then
    set -a
    . "$SCRIPT_DIR/.env"
    set +a
fi

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    printf "${GREEN}[INFO]${NC} %s\n" "$1"
}

print_warning() {
    printf "${YELLOW}[WARNING]${NC} %s\n" "$1"
}

print_error() {
    printf "${RED}[ERROR]${NC} %s\n" "$1"
}

# Check if wrangler is installed
check_wrangler() {
    if ! command -v wrangler > /dev/null 2>&1; then
        print_error "Wrangler CLI not found. Install it with: npm install -g wrangler"
        exit 1
    fi
}

# Check if Cloudflare API token is available
check_auth() {
    if [ -z "$CLOUDFLARE_API_TOKEN" ]; then
        print_error "CLOUDFLARE_API_TOKEN not found in .env file"
        print_error "Add CLOUDFLARE_API_TOKEN=your_token_here to .env file"
        exit 1
    fi
    print_status "Using API token for authentication"
}

# Check that required secrets exist
check_secrets() {
    print_status "Checking required secrets..."
    if ! wrangler secret list | grep -q "AWS_API_URL"; then
        print_error "AWS_API_URL secret not found!"
        print_error "Please set it first: echo 'http://your-api-url' | wrangler secret put AWS_API_URL"
        exit 1
    fi
    print_status "All required secrets found"
}

# Deploy the worker
deploy_worker() {
    print_status "Deploying Cloudflare Worker..."
    
    cd "$SCRIPT_DIR"
    
    # Secrets are managed manually in Cloudflare - nothing to set here
    
    # Deploy the worker
    print_status "Deploying worker code..."
    if ! wrangler deploy; then
        print_error "Failed to deploy worker"
        exit 1
    fi
    
    print_status "Worker deployed successfully!"
}

# Show deployment info
show_info() {
    WORKER_NAME=$(grep '^name = ' wrangler.toml | cut -d'"' -f2)
    
    print_status "Deployment completed!"
    print_status "Worker Name: $WORKER_NAME"
    print_status "Worker URL: https://$WORKER_NAME.your-account.workers.dev"
    print_status ""
    print_status "Next steps:"
    print_status "1. Test the gateway with: curl https://$WORKER_NAME.your-account.workers.dev/health"
    print_status "2. Test the API with: curl -X POST https://$WORKER_NAME.your-account.workers.dev/api/run -H 'Content-Type: application/json' -d '{\"code\":\"print \\\"Hello\\\"\"}'"
    print_status "3. Monitor logs with: wrangler tail"
}

# Main execution
main() {
    print_status "Starting Cloudflare Worker deployment for Osprey Web Compiler Proxy"
    
    check_wrangler
    check_auth
    check_secrets
    deploy_worker
    show_info
}

# Handle script arguments
case "${1:-}" in
    --help|-h)
        echo "Usage: $0 [OPTIONS]"
        echo "Deploy Osprey Web Compiler Proxy to Cloudflare Workers"
        echo ""
        echo "Options:"
        echo "  --help, -h    Show this help message"
        echo ""
        exit 0
        ;;
    *)
        main
        ;;
esac 