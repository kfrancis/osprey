# Osprey Web Playground

Web-based Osprey compiler playground with Monaco editor and LSP support.

## Architecture

**Node.js API server (port 3001)** - Handles compilation, execution, and LSP bridge

```mermaid
graph TB
    subgraph "Browser (Port 8000)"
        A["Monaco Editor<br/>index.html"]
    end
    
    subgraph "Node.js Server (Port 3001)"
        B["Express API<br/>/api/compile<br/>/api/run"]
        C["WebSocket Bridge<br/>/lsp"]
    end
    
    subgraph "External Processes"
        D["Osprey Compiler<br/>osprey compile"]
        E["Osprey LSP<br/>stdio interface"]
    end
    
    A -->|"HTTP Requests"| B
    A -->|"WebSocket"| C
    B -->|"Spawn Process"| D
    C -->|"stdio"| E
```

## Quick Start

```bash
./start.sh
```

Open http://localhost:8000

## Features

- Monaco editor with Osprey syntax highlighting
- Compile/run buttons
- WebSocket LSP connection
- Real-time error feedback

## Requirements

- Node.js
- Python 3
- Osprey compiler in PATH

## Files

- `index.html` - Single-page playground app
- `src/server.js` - Node.js API and WebSocket server
- `start.sh` - Startup script 