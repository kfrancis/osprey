import * as path from 'path';
import * as vscode from 'vscode';
import { workspace, ExtensionContext, window, ConfigurationChangeEvent, commands, Uri, debug, languages } from 'vscode';
import { execFile } from 'child_process';
import * as fs from 'fs';
import {
  LanguageClient,
  LanguageClientOptions,
  ServerOptions,
  TransportKind
} from 'vscode-languageclient/node';

// Define custom event classes for better telemetry
class OspreyEvent {
  constructor(public type: string) {}
}

class ActivationEvent extends OspreyEvent {
  constructor(public success: boolean, public timeTaken: number, public error?: any) {
    super('activation');
  }
}

class CompilerEvent extends OspreyEvent {
  constructor(public action: 'check' | 'compile' | 'run', public success: boolean, public error?: any) {
    super('compiler');
  }
}

class ServerEvent extends OspreyEvent {
  constructor(public action: 'start' | 'stop' | 'restart' | 'error', public error?: any) {
    super('server');
  }
}

// Event stream for telemetry
class EventStream {
  private listeners: ((event: OspreyEvent) => void)[] = [];

  subscribe(listener: (event: OspreyEvent) => void): void {
    this.listeners.push(listener);
  }

  post(event: OspreyEvent): void {
    this.listeners.forEach(listener => listener(event));
  }
}

let client: LanguageClient;
let outputChannel: vscode.OutputChannel;
let statusBar: vscode.StatusBarItem;
const eventStream = new EventStream();

export function activate(context: ExtensionContext) {
  const startActivation = process.hrtime();
  console.log('Osprey extension is now active!');
  
  // Store context globally for use in other functions
  extensionContext = context;
  
  // Create output channel for diagnostics
  outputChannel = window.createOutputChannel('Osprey');
  outputChannel.appendLine('=== Osprey Extension Activation ===');
  
  // Setup event listeners for telemetry and logging
  setupEventListeners();
  
  // Create status bar with extended functionality
  setupStatusBar(context);
  updateStatusBar('starting');

  try {
    // Check if Osprey server is enabled
    const config = workspace.getConfiguration('osprey');
    if (!config.get('server.enabled', true)) {
      outputChannel.appendLine('Language server is disabled in configuration');
      updateStatusBar('disabled');
      return;
    }

    // Server options - use the TypeScript language server
    const serverModule = context.asAbsolutePath(path.join('server', 'out', 'src', 'server.js'));
    outputChannel.appendLine(`Server module path: ${serverModule}`);
    
    // Check if server file exists
    if (!fs.existsSync(serverModule)) {
      const errorMsg = `Server module not found at: ${serverModule}`;
      outputChannel.appendLine(`ERROR: ${errorMsg}`);
      
      // Try to diagnose the issue
      try {
        const serverOutDir = path.dirname(serverModule);
        const serverSrcDir = path.join(context.extensionPath, 'server', 'src');
        
        if (!fs.existsSync(path.dirname(serverOutDir))) {
          outputChannel.appendLine(`Server output directory does not exist: ${path.dirname(serverOutDir)}`);
          outputChannel.appendLine('This suggests the server was not compiled or installed correctly.');
        }
        
        if (fs.existsSync(serverSrcDir)) {
          const files = fs.readdirSync(serverSrcDir);
          outputChannel.appendLine(`Server source files exist: ${files.join(', ')}`);
          outputChannel.appendLine('Try running "npm run compile" to build the server.');
        }
      } catch (err) {
        outputChannel.appendLine(`Error diagnosing server issue: ${err}`);
      }
      
      // Create an error notification with actionable buttons
      window.showErrorMessage(errorMsg, 'Compile Server', 'Show Output', 'Report Issue')
        .then(selection => {
          if (selection === 'Compile Server') {
            compileServer(context);
          } else if (selection === 'Show Output') {
            outputChannel.show();
          } else if (selection === 'Report Issue') {
            vscode.env.openExternal(Uri.parse('https://github.com/yourorg/osprey/issues/new'));
          }
        });
      
      updateStatusBar('error');
      eventStream.post(new ActivationEvent(false, getDurationInMs(startActivation), new Error(errorMsg)));
      return;
    }
    
    outputChannel.appendLine('Server module exists, proceeding with setup...');
    
    const debugOptions = { execArgv: ['--nolazy', '--inspect=6009'] };
    
    const serverOptions: ServerOptions = {
      run: { module: serverModule, transport: TransportKind.ipc },
      debug: {
        module: serverModule,
        transport: TransportKind.ipc,
        options: debugOptions
      }
    };

    // Client options
    const clientOptions: LanguageClientOptions = {
      documentSelector: [{ scheme: 'file', language: 'osprey' }],
      synchronize: {
        fileEvents: workspace.createFileSystemWatcher('**/*.osp')
      },
      outputChannel: outputChannel,
      revealOutputChannelOn: 4, // Error
      initializationFailedHandler: (error: Error | any) => {
        outputChannel.appendLine(`Initialization failed: ${error}`);
        window.showErrorMessage(`Osprey language server initialization failed: ${error}`, 'Show Output')
          .then(selection => {
            if (selection === 'Show Output') {
              outputChannel.show();
            }
          });
        updateStatusBar('error');
        eventStream.post(new ServerEvent('error', error));
        return false;
      },
      errorHandler: {
        error: (error: Error, message: any, count: number | undefined) => {
          outputChannel.appendLine(`Language server error: ${error.message}, message: ${message}, count: ${count}`);
          if (count === 1) {
            window.showWarningMessage(`Osprey language server error: ${error.message}`, 'Show Output', 'Restart Server')
              .then(selection => {
                if (selection === 'Show Output') {
                  outputChannel.show();
                } else if (selection === 'Restart Server') {
                  commands.executeCommand('osprey.restartServer');
                }
              });
            eventStream.post(new ServerEvent('error', error));
          }
          return { action: count && count > 3 ? 2 : 1 }; // Continue for first 3 errors, then shutdown
        },
        closed: () => {
          outputChannel.appendLine('Language server connection closed');
          updateStatusBar('stopped');
          return { action: 1 }; // Restart
        }
      }
    };

    // Create and start the language client
    client = new LanguageClient(
      'ospreyLanguageServer',
      'Osprey Language Server',
      serverOptions,
      clientOptions
    );

    outputChannel.appendLine('Starting language client...');
    
    // Set up client state listeners
    client.onDidChangeState((event) => {
      if (event.newState === 1) { // Running state
        updateStatusBar('running');
        outputChannel.appendLine('✅ Osprey language server is running');
        eventStream.post(new ServerEvent('start'));
      } else if (event.newState === 3) { // Stopped state
        updateStatusBar('stopped');
        outputChannel.appendLine('⚠️ Osprey language server has stopped');
        eventStream.post(new ServerEvent('stop'));
      }
    });

    // Start the client and server
    client.start().then(() => {
      outputChannel.appendLine('SUCCESS: Osprey language server started successfully');
      console.log('Osprey language server started successfully');
      updateStatusBar('running');
      
      // After server starts, check if compiler is available
      checkCompilerAvailability();
      
      const timeTaken = getDurationInMs(startActivation);
      outputChannel.appendLine(`Extension activated in ${timeTaken.toFixed(3)}ms`);
      eventStream.post(new ActivationEvent(true, timeTaken));
    }).catch((error: any) => {
      const errorMsg = `Failed to start Osprey language server: ${error.message || error}`;
      outputChannel.appendLine(`ERROR: ${errorMsg}`);
      outputChannel.appendLine(`Error stack: ${error.stack || 'No stack trace'}`);
      console.error('Failed to start Osprey language server:', error);
      updateStatusBar('error');
      window.showErrorMessage(errorMsg, 'Show Output', 'Restart Server')
        .then(selected => {
          if (selected === 'Show Output') {
            outputChannel.show();
          } else if (selected === 'Restart Server') {
            commands.executeCommand('osprey.restartServer');
          }
        });
      eventStream.post(new ActivationEvent(false, getDurationInMs(startActivation), error));
    });

    // Import and set up the debug adapter
    const { OspreyDebugSession } = require('./debugAdapter');
    
    // Register debug adapter
    const provider = debug.registerDebugAdapterDescriptorFactory('osprey', {
      createDebugAdapterDescriptor(_session: any) {
        // Return an inline implementation of the debug adapter
        return new vscode.DebugAdapterInlineImplementation(new OspreyDebugSession());
      }
    });

    context.subscriptions.push(provider);

    // Register debug configuration provider
    context.subscriptions.push(debug.registerDebugConfigurationProvider('osprey', {
      resolveDebugConfiguration(folder: any, config: any, token: any) {
        // If no config is provided, create a default one
        if (!config.type && !config.request && !config.name) {
          const editor = window.activeTextEditor;
          if (editor && editor.document.languageId === 'osprey') {
            config.type = 'osprey';
            config.name = 'Run Osprey File';
            config.request = 'launch';
            config.program = editor.document.fileName;
          }
        }

        if (!config.program) {
          return window.showInformationMessage("Cannot find a program to run").then(_ => {
            return undefined;
          });
        }

        // Actually run the Osprey program instead of debugging
        compileAndRunCurrentFile();
        return undefined; // Cancel the debug session
      }
    }));

    // Auto-detect and force language association for .osp files
    handleLanguageAssociation(context);

    // Register commands
    registerCommands(context);

  } catch (error: any) {
    const errorMsg = `Error during extension activation: ${error.message || error}`;
    outputChannel.appendLine(`FATAL ERROR: ${errorMsg}`);
    outputChannel.appendLine(`Stack trace: ${error.stack || 'No stack trace available'}`);
    window.showErrorMessage(errorMsg, 'Show Output')
      .then(selection => {
        if (selection === 'Show Output') {
          outputChannel.show();
        }
      });
    updateStatusBar('error');
    eventStream.post(new ActivationEvent(false, getDurationInMs(startActivation), error));
  }
}

function setupEventListeners() {
  // Create event listeners for logging and telemetry
  eventStream.subscribe((event) => {
    // Log to output channel
    if (event instanceof ActivationEvent) {
      outputChannel.appendLine(`[Event] Activation ${event.success ? 'successful' : 'failed'} in ${event.timeTaken.toFixed(3)}ms`);
      if (!event.success && event.error) {
        outputChannel.appendLine(`[Event] Activation error: ${event.error.message || event.error}`);
      }
    } else if (event instanceof CompilerEvent) {
      outputChannel.appendLine(`[Event] Compiler ${event.action} ${event.success ? 'succeeded' : 'failed'}`);
      if (!event.success && event.error) {
        outputChannel.appendLine(`[Event] Compiler error: ${event.error.message || event.error}`);
      }
    } else if (event instanceof ServerEvent) {
      outputChannel.appendLine(`[Event] Server ${event.action}`);
      if (event.error) {
        outputChannel.appendLine(`[Event] Server error: ${event.error.message || event.error}`);
      }
    }
    
    // Could add telemetry reporting here if desired
  });
}

function setupStatusBar(context: ExtensionContext) {
  statusBar = window.createStatusBarItem(vscode.StatusBarAlignment.Right, 100);
  statusBar.text = "$(sync~spin) Osprey: Starting";
  statusBar.tooltip = "Osprey Language Server is initializing";
  statusBar.command = 'osprey.showStatusOptions';
  statusBar.show();
  context.subscriptions.push(statusBar);
  
  // Register status options command
  context.subscriptions.push(commands.registerCommand('osprey.showStatusOptions', () => {
    const items = [
      { label: "$(output) Show Output Channel", command: 'osprey.showOutputChannel' },
      { label: "$(refresh) Restart Language Server", command: 'osprey.restartServer' },
      { label: "$(tools) Check Compiler Status", command: 'osprey.checkCompilerStatus' },
      { label: "$(gear) Reset Settings", command: 'osprey.resetSettings' }
    ];
    
    window.showQuickPick(items, {
      placeHolder: 'Select an Osprey action'
    }).then(selection => {
      if (selection) {
        commands.executeCommand(selection.command);
      }
    });
  }));
}

function updateStatusBar(status: 'starting' | 'running' | 'error' | 'stopped' | 'disabled') {
  if (!statusBar) return;
  
  switch (status) {
    case 'starting':
      statusBar.text = "$(sync~spin) Osprey: Starting";
      statusBar.tooltip = "Osprey Language Server is starting";
      break;
    case 'running':
      statusBar.text = "$(check) Osprey";
      statusBar.tooltip = "Osprey Language Server is running. Click for options.";
      break;
    case 'error':
      statusBar.text = "$(error) Osprey";
      statusBar.tooltip = "Osprey Language Server encountered an error. Click for options.";
      break;
    case 'stopped':
      statusBar.text = "$(stop) Osprey: Stopped";
      statusBar.tooltip = "Osprey Language Server is stopped. Click to restart.";
      break;
    case 'disabled':
      statusBar.text = "$(circle-slash) Osprey: Disabled";
      statusBar.tooltip = "Osprey Language Server is disabled in settings. Click for options.";
      break;
  }
}

function compileServer(context: ExtensionContext) {
  outputChannel.appendLine('Attempting to compile the language server...');
  outputChannel.show();
  
  const serverDir = path.join(context.extensionPath, 'server');
  
  const terminal = window.createTerminal('Osprey Server Compile');
  terminal.show();
  
  if (fs.existsSync(path.join(serverDir, 'package.json'))) {
    terminal.sendText(`cd "${serverDir}" && npm install && npm run compile`);
    outputChannel.appendLine('Running npm install && npm run compile in server directory');
  } else {
    outputChannel.appendLine(`ERROR: Cannot find package.json in ${serverDir}`);
    window.showErrorMessage('Cannot find server package.json for compilation');
  }
}

function handleLanguageAssociation(context: ExtensionContext) {
  // Auto-detect and force language association for .osp files
  workspace.onDidOpenTextDocument((document) => {
    outputChannel.appendLine(`📁 Document opened: ${document.fileName}`);
    if (document.fileName.endsWith('.osp') && document.languageId !== 'osprey') {
      outputChannel.appendLine(`🔧 Forcing language association for ${document.fileName} (was: ${document.languageId})`);
      // Use the proper API to set language
      languages.setTextDocumentLanguage(document, 'osprey').then(() => {
        outputChannel.appendLine(`✅ Successfully set language to osprey for ${document.fileName}`);
      }, (error: any) => {
        outputChannel.appendLine(`❌ Failed to set language: ${error}`);
      });
    }
  });

  // Check already open documents
  workspace.textDocuments.forEach((document) => {
    if (document.fileName.endsWith('.osp') && document.languageId !== 'osprey') {
      outputChannel.appendLine(`🔧 Forcing language association for already open file: ${document.fileName}`);
      languages.setTextDocumentLanguage(document, 'osprey');
    }
  });
  
  context.subscriptions.push(workspace.onDidChangeConfiguration((event: ConfigurationChangeEvent) => {
    if (event.affectsConfiguration('osprey')) {
      window.showInformationMessage('Osprey configuration changed. Restart required.', 'Restart Now')
        .then(selection => {
          if (selection === 'Restart Now') {
            commands.executeCommand('osprey.restartServer');
          }
        });
    }
  }));
}

function registerCommands(context: ExtensionContext) {
  // Register commands
  context.subscriptions.push(
    commands.registerCommand('osprey.compile', () => {
      compileCurrentFile();
    }),
    commands.registerCommand('osprey.run', () => {
      compileAndRunCurrentFile();
    }),
    commands.registerCommand('osprey.setLanguage', () => {
      const activeEditor = window.activeTextEditor;
      if (activeEditor) {
        languages.setTextDocumentLanguage(activeEditor.document, 'osprey');
        window.showInformationMessage('Set language to Osprey');
      }
    }),
    commands.registerCommand('osprey.resetSettings', () => {
      // Reset the compiler path setting to use the bundled compiler
      const config = workspace.getConfiguration('osprey');
      config.update('server.compilerPath', '', true).then(() => {
        window.showInformationMessage('Osprey compiler settings reset to use bundled compiler', 'Check Compiler')
          .then(selection => {
            if (selection === 'Check Compiler') {
              commands.executeCommand('osprey.checkCompilerStatus');
            }
          });
      });
    }),
    commands.registerCommand('osprey.restartServer', async () => {
      if (client) {
        updateStatusBar('starting');
        outputChannel.appendLine('🔄 Restarting Osprey language server...');
        try {
          await client.stop();
          await client.start();
          updateStatusBar('running');
          window.showInformationMessage('Osprey language server restarted successfully');
          eventStream.post(new ServerEvent('restart'));
        } catch (error) {
          updateStatusBar('error');
          outputChannel.appendLine(`Error restarting server: ${error}`);
          window.showErrorMessage(`Failed to restart server: ${error}`, 'Show Output')
            .then(selection => {
              if (selection === 'Show Output') {
                outputChannel.show();
              }
            });
          eventStream.post(new ServerEvent('error', error));
        }
      } else {
        window.showErrorMessage('Osprey language server is not running', 'Start Server')
          .then(selection => {
            if (selection === 'Start Server') {
              // Try to initialize the server from scratch
              activate(extensionContext);
            }
          });
      }
    }),
    commands.registerCommand('osprey.showOutputChannel', () => {
      outputChannel.show();
    }),
    commands.registerCommand('osprey.checkCompilerStatus', () => {
      checkCompilerAvailability(true); // Force UI feedback
    })
  );
}

// Store extension context globally so compiler functions can access it
let extensionContext: ExtensionContext;

/**
 * Checks if the compiler is available and working
 * @param showUI Whether to show UI notifications about the result
 */
function checkCompilerAvailability(showUI: boolean = false) {
  const compilerPath = getCompilerPath();
  if (!compilerPath) {
    const message = 'Osprey compiler not found. Some features may not work correctly.';
    outputChannel.appendLine(`⚠️ WARNING: ${message}`);
    
    if (showUI) {
      window.showWarningMessage(message, 'Reset Settings', 'Show Output')
        .then(selected => {
          if (selected === 'Reset Settings') {
            commands.executeCommand('osprey.resetSettings');
          } else if (selected === 'Show Output') {
            outputChannel.show();
          }
        });
    }
    
    eventStream.post(new CompilerEvent('check', false, new Error('Compiler not found')));
    return false;
  }

  // Check if compiler works
  outputChannel.appendLine(`🔍 Checking compiler: ${compilerPath}`);
  if (showUI) {
    outputChannel.show();
  }

  return new Promise<boolean>((resolve) => {
    execFile(compilerPath, ['--version'], (error, stdout, stderr) => {
      if (error) {
        outputChannel.appendLine(`⚠️ Compiler check failed: ${error.message}`);
        
        if (showUI) {
          window.showErrorMessage(`Compiler check failed: ${error.message}`, 'Reset Settings', 'Show Output')
            .then(selected => {
              if (selected === 'Reset Settings') {
                commands.executeCommand('osprey.resetSettings');
              } else if (selected === 'Show Output') {
                outputChannel.show();
              }
            });
        }
        
        eventStream.post(new CompilerEvent('check', false, error));
        resolve(false);
      } else {
        outputChannel.appendLine(`✅ Compiler check successful: ${stdout.trim()}`);
        
        if (showUI) {
          window.showInformationMessage(`Osprey compiler is working correctly (${stdout.trim()})`);
        }
        
        eventStream.post(new CompilerEvent('check', true));
        resolve(true);
      }
    });
  });
}

/**
 * Gets the path to the Osprey compiler executable
 * This function handles platform differences and fallbacks
 */
function getCompilerPath(): string {
  // Get the compiler path from settings or use bundled compiler
  const config = workspace.getConfiguration('osprey');
  const compilerPathSetting = config.get<string>('server.compilerPath', '');
  
  outputChannel.appendLine(`Looking for compiler path (setting: ${compilerPathSetting || 'not set'})`);
  
  // This variable will hold our final compiler path
  let compilerPath = '';
  
  // Check if the user has specified a custom compiler path
  if (compilerPathSetting && compilerPathSetting.trim() !== '') {
    outputChannel.appendLine(`Custom compiler path specified: ${compilerPathSetting}`);
    
    // Replace ${workspaceFolder} with actual path if present
    compilerPath = compilerPathSetting;
    if (compilerPath.includes('${workspaceFolder}') && workspace.workspaceFolders && workspace.workspaceFolders.length > 0) {
      compilerPath = compilerPath.replace('${workspaceFolder}', workspace.workspaceFolders[0].uri.fsPath);
      outputChannel.appendLine(`Resolved workspace path: ${compilerPath}`);
    }
    
    // Check if the specified compiler exists
    if (!fs.existsSync(compilerPath)) {
      outputChannel.appendLine(`❌ Specified compiler not found at: ${compilerPath}. Falling back to bundled compiler.`);
      compilerPath = ''; // Reset to use bundled compiler
    } else {
      outputChannel.appendLine(`✅ Using specified compiler: ${compilerPath}`);
      return compilerPath; // Return early if we found a valid custom compiler
    }
  } else {
    outputChannel.appendLine('No custom compiler path specified, looking for compiler in standard locations...');
  }
  
  // Try to find compiler in PATH if no custom path was provided
  if (!compilerPath) {
    // Check PATH environment variable first
    const pathVar = process.env.PATH || '';
    const pathDirs = pathVar.split(path.delimiter);
    const exeExtension = process.platform === 'win32' ? '.exe' : '';
    
    outputChannel.appendLine('Checking PATH for osprey compiler...');
    
    for (const dir of pathDirs) {
      const possiblePath = path.join(dir, `osprey${exeExtension}`);
      if (fs.existsSync(possiblePath)) {
        compilerPath = possiblePath;
        outputChannel.appendLine(`✅ Found osprey compiler in PATH: ${compilerPath}`);
        return compilerPath;
      }
    }
    
    outputChannel.appendLine('Osprey compiler not found in PATH, checking bundled compiler...');
  }
  
  // If still no valid compiler path, use the bundled compiler
  if (!compilerPath) {
    // Always use the compiler bundled with the extension
    if (process.platform === 'win32') {
      // On Windows, prefer the cmd wrapper if available, then osprey.exe, then osprey.js
      const cmdPath = extensionContext.asAbsolutePath(path.join('bin', 'osprey.cmd'));
      const exePath = extensionContext.asAbsolutePath(path.join('bin', 'osprey.exe'));
      const jsPath = extensionContext.asAbsolutePath(path.join('bin', 'osprey.js'));
      
      outputChannel.appendLine(`Checking for Windows bundled compiler...`);
      
      if (fs.existsSync(cmdPath)) {
        compilerPath = cmdPath;
        outputChannel.appendLine(`✅ Using bundled Windows cmd wrapper: ${compilerPath}`);
      } else if (fs.existsSync(exePath)) {
        compilerPath = exePath;
        outputChannel.appendLine(`✅ Using bundled Windows exe: ${compilerPath}`);
      } else if (fs.existsSync(jsPath)) {
        compilerPath = jsPath;
        outputChannel.appendLine(`✅ Using bundled JS version: ${compilerPath}`);
      }
    } else {
      // On Unix-like systems, try native binary then js
      const nativePath = extensionContext.asAbsolutePath(path.join('bin', 'osprey'));
      const jsPath = extensionContext.asAbsolutePath(path.join('bin', 'osprey.js'));
      
      outputChannel.appendLine(`Checking for Unix bundled compiler...`);
      
      if (fs.existsSync(nativePath)) {
        compilerPath = nativePath;
        // Ensure the compiler is executable on Unix-like systems
        try {
          fs.chmodSync(compilerPath, 0o755); // rwxr-xr-x
          outputChannel.appendLine(`✅ Using bundled Unix native binary: ${compilerPath}`);
        } catch (error) {
          console.error('Failed to set executable permissions on compiler:', error);
          outputChannel.appendLine(`❌ Failed to set executable permissions: ${error}`);
        }
      } else if (fs.existsSync(jsPath)) {
        compilerPath = jsPath;
        try {
          fs.chmodSync(compilerPath, 0o755); // rwxr-xr-x
          outputChannel.appendLine(`✅ Using bundled JS version: ${compilerPath}`);
        } catch (error) {
          console.error('Failed to set executable permissions on compiler:', error);
          outputChannel.appendLine(`❌ Failed to set executable permissions: ${error}`);
        }
      }
    }
  }
  
  // Final check to ensure we have a valid compiler path
  if (!compilerPath || !fs.existsSync(compilerPath)) {
    // Last ditch effort - try to find the compiler in the extension directory
    const extensionPath = extensionContext.extensionPath;
    const possiblePaths = [
      path.join(extensionPath, 'bin', 'osprey.cmd'),
      path.join(extensionPath, 'bin', 'osprey.exe'),
      path.join(extensionPath, 'bin', 'osprey'),
      path.join(extensionPath, 'bin', 'osprey.js')
    ];
    
    outputChannel.appendLine(`🔎 Last attempt to find compiler in extension directory...`);
    
    for (const p of possiblePaths) {
      if (fs.existsSync(p)) {
        console.log(`Found compiler at: ${p}`);
        outputChannel.appendLine(`✅ Found compiler at: ${p}`);
        compilerPath = p;
        break;
      }
    }
    
    // If still not found, show error and return empty string
    if (!compilerPath || !fs.existsSync(compilerPath)) {
      outputChannel.appendLine('❌ CRITICAL ERROR: Osprey compiler not found after exhaustive search!');
      return ''; // Return empty string to indicate compiler not found
    }
  }
  
  console.log(`Using Osprey compiler at: ${compilerPath}`);
  outputChannel.appendLine(`✅ Final compiler path: ${compilerPath}`);
  return compilerPath;
}

function compileCurrentFile() {
  const activeEditor = window.activeTextEditor;
  if (!activeEditor) {
    window.showErrorMessage('No active Osprey file found');
    return;
  }

  const document = activeEditor.document;
  if (!document.fileName.endsWith('.osp')) {
    window.showErrorMessage('Please open a .osp file to compile');
    return;
  }

  // Save the file first
  document.save().then(() => {
    const compilerOutput = window.createOutputChannel('Osprey Compiler');
    compilerOutput.show();
    compilerOutput.clear();
    compilerOutput.appendLine(`Compiling ${document.fileName}...`);
    compilerOutput.appendLine('----------------------------------');

    // Get the directory containing the file (no workspace required)
    const fileDir = path.dirname(document.fileName);
    
    // Get the compiler path from settings or use bundled compiler
    const compilerPath = getCompilerPath();
    
    // Check if we have a valid compiler path
    if (!compilerPath) {
      compilerOutput.appendLine('⛔ ERROR: Osprey compiler not found.');
      compilerOutput.appendLine('');
      compilerOutput.appendLine('Try the following:');
      compilerOutput.appendLine('1. Run the command "Osprey: Reset Compiler Settings" from the command palette.');
      compilerOutput.appendLine('2. Close and reopen VS Code.');
      compilerOutput.appendLine('3. Check that the extension is properly installed.');
      compilerOutput.appendLine('=== END OUTPUT ===');
      
      window.showErrorMessage('Osprey compiler not found. Try running "Osprey: Reset Compiler Settings"');
      eventStream.post(new CompilerEvent('compile', false, new Error('Compiler not found')));
      return;
    }
    
    compilerOutput.appendLine(`Using Osprey compiler: ${compilerPath}`);
    
    // Use the configured osprey compiler
    execFile(compilerPath, [document.fileName], 
      { cwd: fileDir }, 
      (error: any, stdout: any, stderr: any) => {
        compilerOutput.appendLine(`=== COMPILATION OUTPUT ===`);
        
        if (stdout) {
          compilerOutput.appendLine(`STDOUT:`);
          compilerOutput.appendLine(stdout);
        }
        
        if (stderr) {
          compilerOutput.appendLine(`STDERR:`);
          compilerOutput.appendLine(stderr);
        }
        
        if (error) {
          compilerOutput.appendLine(`ERROR:`);
          compilerOutput.appendLine(`Exit code: ${error.code || 'unknown'}`);
          compilerOutput.appendLine(`Signal: ${error.signal || 'none'}`);
          compilerOutput.appendLine(`Error message: ${error.message}`);
          window.showErrorMessage('Compilation failed. Check output for details.');
          eventStream.post(new CompilerEvent('compile', false, error));
        } else {
          compilerOutput.appendLine('=== COMPILATION SUCCESS ===');
          window.showInformationMessage('Osprey file compiled successfully!');
          eventStream.post(new CompilerEvent('compile', true));
        }
        
        compilerOutput.appendLine(`=== END OUTPUT ===`);
      }
    );
  });
}

function compileAndRunCurrentFile() {
  const activeEditor = window.activeTextEditor;
  if (!activeEditor) {
    window.showErrorMessage('No active Osprey file found');
    return;
  }

  const document = activeEditor.document;
  if (!document.fileName.endsWith('.osp')) {
    window.showErrorMessage('Please open a .osp file to run');
    return;
  }
  
  // Save the file first
  document.save().then(() => {
    const runOutput = window.createOutputChannel('Osprey Runner');
    runOutput.show();
    runOutput.clear();
    runOutput.appendLine(`Compiling and running ${document.fileName}...`);

    // Get the directory containing the file (no workspace required)
    const fileDir = path.dirname(document.fileName);
    
    // Get the compiler path from settings or use bundled compiler
    const compilerPath = getCompilerPath();
    
    // Check if we have a valid compiler path
    if (!compilerPath) {
      runOutput.appendLine('ERROR: Osprey compiler not found. Please check extension installation.');
      runOutput.appendLine('=== END OUTPUT ===');
      window.showErrorMessage('Osprey compiler not found. Try running "Osprey: Reset Compiler Settings"');
      eventStream.post(new CompilerEvent('run', false, new Error('Compiler not found')));
      return;
    }
    
    runOutput.appendLine(`Using Osprey compiler: ${compilerPath}`);
    
    // Use the configured osprey compiler with --run flag
    execFile(compilerPath, [document.fileName, '--run'], 
      { cwd: fileDir }, 
      (error: any, stdout: any, stderr: any) => {
        runOutput.appendLine(`=== COMPILE AND RUN OUTPUT ===`);
        
        if (stdout) {
          runOutput.appendLine(`STDOUT:`);
          runOutput.appendLine(stdout);
        }
        
        if (stderr) {
          runOutput.appendLine(`STDERR:`);
          runOutput.appendLine(stderr);
        }
        
        if (error) {
          runOutput.appendLine(`ERROR:`);
          runOutput.appendLine(`Exit code: ${error.code || 'unknown'}`);
          runOutput.appendLine(`Signal: ${error.signal || 'none'}`);
          runOutput.appendLine(`Error message: ${error.message}`);
          window.showErrorMessage('Compilation or execution failed. Check output for details.');
          eventStream.post(new CompilerEvent('run', false, error));
        } else {
          runOutput.appendLine('=== SUCCESS ===');
          window.showInformationMessage('Osprey program executed successfully!');
          eventStream.post(new CompilerEvent('run', true));
        }
        
        runOutput.appendLine(`=== END OUTPUT ===`);
      }
    );
  });
}

function getDurationInMs(start: [number, number]): number {
  const diff = process.hrtime(start);
  return (diff[0] * 1000) + (diff[1] / 1000000);
}

export function deactivate(): Promise<void> | undefined {
  outputChannel.appendLine('Deactivating Osprey extension...');
  
  if (!client) {
    outputChannel.appendLine('Language client was never started, nothing to deactivate.');
    return undefined;
  }
  
  outputChannel.appendLine('Stopping language client...');
  updateStatusBar('stopped');
  
  return client.stop().then(() => {
    outputChannel.appendLine('Language client stopped successfully.');
  }).catch((error) => {
    outputChannel.appendLine(`Error stopping language client: ${error}`);
    console.error('Error stopping language client:', error);
  });
}