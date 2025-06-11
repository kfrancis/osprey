#include "http_shared.h"

// Create HTTP client - returns client_id or negative error
int64_t http_create_client(char* base_url, int64_t timeout) {
    if (!base_url) {
        return -1;
    }
    
    if (timeout < 0) {
        return -2;
    }
    
    int64_t id = get_next_id();
    HttpClient* client = malloc(sizeof(HttpClient));
    if (!client) {
        return -3;
    }
    
    client->id = id;
    client->base_url = strdup(base_url);
    client->timeout = (int)timeout;
    client->is_persistent = false;
    
    // Parse base URL
    char* path;
    if (parse_url(base_url, &client->host, &client->port, &path) != 0) {
        free(client->base_url);
        free(client);
        return -4;
    }
    free(path);  // We only need host and port for client
    
    pthread_mutex_lock(&runtime_mutex);
    clients[id] = client;
    pthread_mutex_unlock(&runtime_mutex);
    
    return id;
}

// Make HTTP request - returns HTTP status code or negative error
int64_t http_request(int64_t client_id, int64_t method, char* path, char* headers, char* body) {
    pthread_mutex_lock(&runtime_mutex);
    HttpClient* client = clients[client_id];
    pthread_mutex_unlock(&runtime_mutex);
    
    if (!client) {
        return -1;
    }
    
    if (!path) {
        return -2;
    }
    
    // Create socket
    int sock = socket(AF_INET, SOCK_STREAM, 0);
    if (sock < 0) {
        return -3;
    }
    
    // Set timeout
    struct timeval tv;
    tv.tv_sec = client->timeout / 1000;
    tv.tv_usec = (client->timeout % 1000) * 1000;
    setsockopt(sock, SOL_SOCKET, SO_RCVTIMEO, (const char*)&tv, sizeof tv);
    
    // Connect to server
    struct sockaddr_in server_addr;
    server_addr.sin_family = AF_INET;
    server_addr.sin_port = htons(client->port);
    
    struct hostent* server = gethostbyname(client->host);
    if (!server) {
        close(sock);
        return -4;
    }
    
    memcpy(&server_addr.sin_addr.s_addr, server->h_addr, server->h_length);
    
    if (connect(sock, (struct sockaddr*)&server_addr, sizeof(server_addr)) < 0) {
        close(sock);
        return -5;
    }
    
    // Build HTTP request
    char request[MAX_HTTP_BUFFER];
    const char* method_str = http_method_to_string((HttpMethod)method);
    
    int request_len = snprintf(request, sizeof(request),
        "%s %s HTTP/1.1\r\n"
        "Host: %s:%d\r\n"
        "Connection: close\r\n",
        method_str, path, client->host, client->port);
    
    // Add headers if provided
    if (headers && strlen(headers) > 0) {
        request_len += snprintf(request + request_len, sizeof(request) - request_len,
            "%s\r\n", headers);
    }
    
    // Add body if provided
    if (body && strlen(body) > 0) {
        request_len += snprintf(request + request_len, sizeof(request) - request_len,
            "Content-Length: %zu\r\n\r\n%s", strlen(body), body);
    } else {
        request_len += snprintf(request + request_len, sizeof(request) - request_len, "\r\n");
    }
    
    // Send request
    if (send(sock, request, request_len, 0) < 0) {
        close(sock);
        return -6;
    }
    
    // Read response (simplified - just get status code)
    char response[MAX_HTTP_BUFFER];
    int bytes_received = recv(sock, response, sizeof(response) - 1, 0);
    close(sock);
    
    if (bytes_received <= 0) {
        return -7;
    }
    
    response[bytes_received] = '\0';
    
    // Parse status code
    if (strncmp(response, "HTTP/1.1 ", 9) == 0 || strncmp(response, "HTTP/1.0 ", 9) == 0) {
        return atoi(response + 9);
    }
    
    return -8;  // Invalid response format
}

// Close HTTP client - returns 0 on success
int64_t http_close_client(int64_t client_id) {
    pthread_mutex_lock(&runtime_mutex);
    HttpClient* client = clients[client_id];
    if (client) {
        clients[client_id] = NULL;
        free(client->base_url);
        free(client->host);
        free(client);
    }
    pthread_mutex_unlock(&runtime_mutex);
    
    return 0;
}

// Convenience functions for specific HTTP methods
int64_t http_get(int64_t client_id, char* path, char* headers) {
    return http_request(client_id, HTTP_GET, path, headers, NULL);
}

int64_t http_post(int64_t client_id, char* path, char* body, char* headers) {
    return http_request(client_id, HTTP_POST, path, headers, body);
}

int64_t http_put(int64_t client_id, char* path, char* body, char* headers) {
    return http_request(client_id, HTTP_PUT, path, headers, body);
}

int64_t http_delete(int64_t client_id, char* path, char* headers) {
    return http_request(client_id, HTTP_DELETE, path, headers, NULL);
} 