# Osprey VS Code Extension

VSCode language support for the Osprey programming language including syntax highlighting, error diagnostics, and code completion.

## Features

- **Syntax Highlighting** for Osprey (.osp) files
- **Code Completion** for Osprey language keywords and built-ins
- **Error Diagnostics** with inline error messages
- **Compiler Integration** with commands to compile and run Osprey code
- **Cross-Platform Support** for Windows, macOS, and Linux

## Installation

### From VS Code Marketplace

1. Open VS Code
2. Go to Extensions view (Ctrl+Shift+X or Cmd+Shift+X)
3. Search for "Osprey Language Support"
4. Click Install

### Manual Installation

1. Download the latest `.vsix` file from Releases
2. Open VS Code
3. Go to Extensions view
4. Click the "..." menu and select "Install from VSIX..."
5. Select the downloaded file

## Usage

### Editing Osprey Files

1. Create a new file with `.osp` extension
2. VS Code will automatically recognize it as an Osprey file
3. Start coding with syntax highlighting and error checking

### Compiling and Running

- **Compile:** Press `Ctrl+Shift+B` (or `Cmd+Shift+B` on macOS)
- **Compile and Run:** Press `F5`
- **Via Command Palette:** Press `Ctrl+Shift+P` (or `Cmd+Shift+P` on macOS) and search for "Osprey: Compile" or "Osprey: Run"

## Requirements

- VS Code 1.96.0 or higher
- Node.js 20.19.2 (for development)

## Extension Settings

This extension contributes the following settings:

- `osprey.server.enabled`: Enable/disable the Osprey language server
- `osprey.server.compilerPath`: Path to the Osprey compiler executable (defaults to bundled compiler)
- `osprey.diagnostics.enabled`: Enable/disable diagnostic messages

## Development

### Prerequisites

- Node.js 20.19.2
- VS Code 1.96.0 or higher

### Setup

1. Clone the repository
2. Run `npm install` to install dependencies
3. Run `npm run compile` to compile the extension
4. Press `F5` to launch the extension in debug mode

### Building

- `npm run compile`: Compile the TypeScript code
- `npm run watch`: Watch for changes and recompile
- `npm run package`: Create a `.vsix` package
- `npm run bundle-compiler`: Bundle the Osprey compiler with the extension (automatically run during install)

### Cross-Platform Development

The extension is designed to work on all platforms supported by VS Code:
- Windows
- macOS
- Linux

When testing, please verify functionality on your platform.

## Troubleshooting

### Language Server Not Starting
- Check the Osprey output channels in VS Code for error messages
- Verify that the bundled compiler was installed correctly
- Check `osprey.server.compilerPath` setting if using a custom compiler

### Syntax Highlighting Not Working
- Ensure the file has a `.osp` extension
- Try running the "Set Language to Osprey" command from the command palette

### Compiler Errors
- If you see "Command not found" errors, the compiler may not be bundled correctly
- Try reinstalling the extension
- Check the extension logs for more details

## Contributing

We welcome contributions from the community! Please see our contributing guidelines for details.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
