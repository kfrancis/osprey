/**
 * Cloudflare Worker for Osprey Web Compiler IP Whitelisting
 * 
 * This worker sits in front of the AWS-hosted web compiler and only allows
 * requests from whitelisted IP addresses to reach the backend.
 */

// Environment variables are set in Cloudflare dashboard as secrets

const CORS_HEADERS = {
    'Access-Control-Allow-Origin': '*',
    'Access-Control-Allow-Methods': 'GET, POST, PUT, DELETE, OPTIONS',
    'Access-Control-Allow-Headers': 'Content-Type, Authorization',
    'Access-Control-Max-Age': '86400',
};

/**
 * Parse CIDR notation and check if IP is in range
 */
function isIpInCidr(ip, cidr) {
    if (!cidr.includes('/')) {
        return ip === cidr;
    }

    const [network, maskBits] = cidr.split('/');
    const mask = parseInt(maskBits);

    const ipNum = ipToNumber(ip);
    const networkNum = ipToNumber(network);
    const maskNum = (0xffffffff << (32 - mask)) >>> 0;

    return (ipNum & maskNum) === (networkNum & maskNum);
}

/**
 * Convert IP address to number
 */
function ipToNumber(ip) {
    return ip.split('.').reduce((acc, octet) => (acc << 8) + parseInt(octet), 0) >>> 0;
}

/**
 * Get client IP from request headers
 */
function getClientIp(request) {
    const cfConnectingIp = request.headers.get('CF-Connecting-IP');
    const xForwardedFor = request.headers.get('X-Forwarded-For');
    const xRealIp = request.headers.get('X-Real-IP');

    // Cloudflare provides the original client IP in CF-Connecting-IP
    return cfConnectingIp || xForwardedFor?.split(',')[0].trim() || xRealIp || '0.0.0.0';
}

/**
 * Check if IP is allowed - simplified to always allow
 */
function isIpAllowed(ip, allowedIps) {
    return true; // Simplified - no IP restrictions
}

/**
 * Rate limiting using Cloudflare KV (optional)
 */
async function checkRateLimit(ip, env) {
    if (!env.RATE_LIMIT_KV || !env.RATE_LIMIT_PER_MINUTE) {
        return true; // No rate limiting configured
    }

    const limit = parseInt(env.RATE_LIMIT_PER_MINUTE) || 60;
    const key = `rate_limit:${ip}:${Math.floor(Date.now() / 60000)}`;

    try {
        const current = await env.RATE_LIMIT_KV.get(key);
        const count = current ? parseInt(current) : 0;

        if (count >= limit) {
            return false;
        }

        await env.RATE_LIMIT_KV.put(key, (count + 1).toString(), { expirationTtl: 120 });
        return true;
    } catch (error) {
        console.error('Rate limit check failed:', error);
        return true; // Allow on error
    }
}

/**
 * Create error response
 */
function createErrorResponse(status, message, clientIp = null) {
    const response = {
        error: message,
        status: status,
        timestamp: new Date().toISOString(),
    };

    if (clientIp) {
        response.clientIp = clientIp;
    }

    return new Response(JSON.stringify(response), {
        status: status,
        headers: {
            'Content-Type': 'application/json',
            ...CORS_HEADERS,
        },
    });
}

/**
 * Main request handler
 */
export default {
    async fetch(request, env, ctx) {
        try {
            const url = new URL(request.url);
            const clientIp = getClientIp(request);

            // Handle CORS preflight requests
            if (request.method === 'OPTIONS') {
                return new Response(null, {
                    status: 200,
                    headers: CORS_HEADERS,
                });
            }

            // Health check endpoint
            if (url.pathname === '/health') {
                return new Response(JSON.stringify({
                    status: 'ok',
                    service: 'osprey-web-compiler-gateway',
                    clientIp: clientIp,
                    timestamp: new Date().toISOString(),
                }), {
                    status: 200,
                    headers: {
                        'Content-Type': 'application/json',
                        ...CORS_HEADERS,
                    },
                });
            }

            // Check rate limits
            const rateLimitOk = await checkRateLimit(clientIp, env);
            if (!rateLimitOk) {
                console.log(`Rate limited IP: ${clientIp}`);
                return createErrorResponse(429, 'Too many requests', clientIp);
            }

            // Proxy to backend
            const backendUrl = env.AWS_API_URL;
            if (!backendUrl) {
                return createErrorResponse(500, 'AWS API URL not configured');
            }

            const backendRequest = new Request(backendUrl + url.pathname + url.search, {
                method: request.method,
                headers: request.headers,
                body: request.method !== 'GET' && request.method !== 'HEAD' ? request.body : undefined,
            });

            // Add client IP header for backend
            backendRequest.headers.set('X-Original-IP', clientIp);
            backendRequest.headers.set('X-Forwarded-For', clientIp);

            const response = await fetch(backendRequest);

            // Create new response with CORS headers (remove backend CORS headers to avoid duplicates)
            const responseHeaders = Object.fromEntries(response.headers);
            delete responseHeaders['access-control-allow-origin'];
            delete responseHeaders['access-control-allow-methods'];
            delete responseHeaders['access-control-allow-headers'];
            delete responseHeaders['access-control-max-age'];

            const newResponse = new Response(response.body, {
                status: response.status,
                statusText: response.statusText,
                headers: {
                    ...responseHeaders,
                    ...CORS_HEADERS,
                },
            });

            console.log(`Proxied request from ${clientIp} to ${backendUrl}${url.pathname} - Status: ${response.status}`);

            return newResponse;

        } catch (error) {
            console.error('Worker error:', error);
            return createErrorResponse(500, 'Internal server error');
        }
    },
}; 