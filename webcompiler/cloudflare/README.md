# Osprey Web Compiler Cloudflare Proxy

Simple Cloudflare Worker that proxies requests to your AWS-hosted Osprey Web Compiler.

## Setup

1. **Set your AWS API URL:**
   ```bash
   echo 'http://your-aws-api-url:3001' | wrangler secret put AWS_API_URL
   ```

2. **Deploy:**
   ```bash
   ./deploy.sh
   ```

That's it. Your worker will proxy all requests to your AWS backend.

## Files

- `worker.js` - The proxy worker code
- `deploy.sh` - Deployment script
- `wrangler.toml` - Wrangler configuration 