# Osprey VS Code Extension

VS Code language support for Osprey - a modern functional programming language with focus on clarity, safety, and performance.

![Osprey Language](https://via.placeholder.com/800x200?text=Osprey+Programming+Language)

## Features

- **Syntax Highlighting**: Full syntax highlighting for Osprey (.osp) files
- **Language Server**: Intelligent code features powered by the Language Server Protocol
- **Compiler Integration**: Compile and run Osprey programs directly within VS Code
- **Smart Editing**: Auto-indentation, bracket matching, and code folding
- **Code Navigation**: Jump to definitions, find references, and document outline

## Installation

You can install this extension directly from the VS Code marketplace:

1. Open VS Code
2. Go to the Extensions view (Ctrl+Shift+X)
3. Search for "Osprey Language Support"
4. Click Install

Alternatively, you can build and install the extension from source:

```bash
# Clone repository 
git clone https://github.com/osprey/osprey-vscode.git
cd osprey-vscode

# Install dependencies
npm install

# Compile the extension
npm run compile

# Package the extension
npm run package

# Install the VSIX file
code --install-extension osprey-language-support-0.1.0.vsix
```

## Getting Started

1. Create a new file with the `.osp` extension
2. Start writing Osprey code
3. Use Ctrl+Shift+B (or Cmd+Shift+B on Mac) to compile your code
4. Use F5 to compile and run your code

### Example Osprey Code

```osprey
// Basic Osprey example
fn greet(name: String) -> String = "Hello, " + name + "!"

fn main() -> Unit = {
  print(greet("World"))
}
```

## Features

### Syntax Highlighting

The extension provides detailed syntax highlighting for:
- Keywords and control structures
- Function and type declarations
- String literals (with support for interpolation)
- Comments and documentation blocks
- Operators and built-in types

### Language Server

The Language Server provides:
- Error diagnostics as you type
- Code completion suggestions
- Hover information for symbols
- Document symbols for outline view
- Go-to-definition and find references

### Compiler Integration

Compile and run Osprey code without leaving VS Code:
- Compile: Ctrl+Shift+B (Cmd+Shift+B on Mac)
- Run: F5
- View output in the integrated terminal

## Configuration

You can configure the extension in your VS Code settings:

```json
{
  "osprey.server.enabled": true,
  "osprey.server.compilerPath": "${workspaceFolder}/bin/osprey",
  "osprey.diagnostics.enabled": true,
  "osprey.diagnostics.mode": "onType"
}
```

### Available Settings

| Setting | Description | Default |
|---------|-------------|---------|
| `osprey.server.enabled` | Enable/disable the Osprey language server | `true` |
| `osprey.server.compilerPath` | Path to the Osprey compiler executable | `${workspaceFolder}/bin/osprey` |
| `osprey.diagnostics.enabled` | Enable/disable diagnostic messages | `true` |
| `osprey.diagnostics.mode` | When to run diagnostics: onSave, onType, manual | `onType` |

## Commands

The extension provides the following commands (accessible from the Command Palette with Ctrl+Shift+P):

- **Osprey: Compile File**: Compile the current Osprey file
- **Osprey: Compile and Run File**: Compile and run the current Osprey file
- **Osprey: Set Language to Osprey**: Force the current file to be treated as Osprey code

## Troubleshooting

### Language Server Not Starting
- Check the "Osprey Debug" output channel for error messages
- Ensure the compiler path is set correctly in settings
- Restart VS Code and try again

### Compiler Not Found
- The extension includes a mock compiler for development
- By default, it looks in the bin folder of the extension
- You can set a custom path in settings

## Development

This extension consists of:

1. **Client**: The VS Code extension part (TypeScript)
2. **Server**: The Language Server implementation (TypeScript)
3. **Compiler**: A bundled Osprey compiler/interpreter

To contribute to this extension:

```bash
# Install dependencies
npm install

# Watch mode for development
npm run watch

# Run the extension in a new window
F5 (to debug)
```

## License

This extension is licensed under the MIT License. See the LICENSE file for details.

```
vscode-extension/
├── package.json              # Extension manifest
├── client/src/extension.ts   # Main entry point
├── server/src/server.ts      # Language server
├── syntaxes/osprey.tmGrammar.json # Syntax rules
└── language-configuration/language-configuration.json
``` 