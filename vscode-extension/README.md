# Osprey VSCode Extension

VSCode language support for Osprey including syntax highlighting, error diagnostics, and code completion.

## Architecture

**CRITICAL**: TypeScript for VSCode integration, Go for heavy computation.

- **TypeScript**: VSCode API, Language Server Protocol, UI interactions
- **Go**: Parsing, compilation, analysis, performance-critical operations

## Components

### 1. Syntax Highlighting (`syntaxes/osprey.tmGrammar.json`)
TextMate grammar supporting:
- Keywords, literals, operators
- String interpolation (`"Hello ${name}!"`)
- Function/type declarations
- Comments

### 2. Language Configuration (`language-configuration/language-configuration.json`)
- Bracket matching and auto-closing
- Comment toggling (`//`)
- Indentation and word patterns

### 3. Language Server (`server/src/server.ts`)
- Real-time error diagnostics via Go compiler
- Code completion for keywords/built-ins
- Configuration management

### 4. Client Extension (`client/src/extension.ts`)
- Extension entry point and lifecycle
- Status bar integration

## Error Diagnostics Flow

```
1. User edits .osp file
2. TypeScript server receives change
3. Server writes temp file
4. Server calls: go run cmd/osprey/main.go temp.osp
5. Server parses stderr for errors
6. Server sends diagnostics to VSCode
```

## Development

### Setup
```bash
cd vscode-extension
npm install
npm run compile
```

### Build & Install
```bash
npm run package         # Creates .vsix file
npm run install-extension
```

## Configuration

```json
{
  "osprey.server.enabled": true,
  "osprey.server.path": "/path/to/osprey/compiler",
  "osprey.diagnostics.enabled": true
}
```

## Performance Notes

- Temp files created/cleaned immediately
- Compilation on document changes (debounced)
- TypeScript server lightweight, Go handles heavy lifting
- Async compilation doesn't block UI

## Future Enhancements

**Go-based** (implement in Go, call from TypeScript):
- Semantic highlighting, go-to-definition, find references
- Advanced type checking, refactoring tools

**TypeScript-based** (lightweight UI/UX):
- Snippet expansion, formatting preferences
- Theme integration, status indicators

## Troubleshooting

### Language Server Not Starting
- Check `osprey.server.path` configuration
- Ensure Go compiler accessible
- Check VSCode Output → "Osprey Language Server"

### No Error Diagnostics
- Verify `osprey.diagnostics.enabled` is true
- Test: `go run cmd/osprey/main.go test.osp`
- Check temp directory permissions

## File Structure

```
vscode-extension/
├── package.json              # Extension manifest
├── client/src/extension.ts   # Main entry point
├── server/src/server.ts      # Language server
├── syntaxes/osprey.tmGrammar.json # Syntax rules
└── language-configuration/language-configuration.json
``` 