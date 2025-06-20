import * as path from 'path';
import { execFile, spawn, ChildProcess } from 'child_process';
import * as fs from 'fs';
import {
    Logger, logger,
    LoggingDebugSession,
    InitializedEvent, TerminatedEvent, OutputEvent
} from 'vscode-debugadapter';
import { DebugProtocol } from 'vscode-debugprotocol';
import * as vscode from 'vscode';
import * as os from 'os';

/**
 * Enhanced debug adapter for Osprey
 * This implementation includes better error handling, cross-platform support,
 * and is based on best practices from the VS Code C# extension.
 */
export class OspreyDebugSession extends LoggingDebugSession {
    private static threadID = 1;
    private configPath: string | undefined;
    private childProcess: ChildProcess | undefined;
    private workspaceFolders: string[] = [];
    private extensionPath: string | undefined;
    private hasShownCompilerError: boolean = false;

    public constructor() {
        super("osprey-debug.txt");
        this.setDebuggerLinesStartAt1(true);
        this.setDebuggerColumnsStartAt1(true);
        
        // Try to get workspace folders from VS Code
        if (vscode.workspace.workspaceFolders) {
            this.workspaceFolders = vscode.workspace.workspaceFolders.map(folder => folder.uri.fsPath);
        }
        
        // Try to get extension path
        const extension = vscode.extensions.getExtension('ospreyLang.osprey');
        if (extension) {
            this.extensionPath = extension.extensionPath;
        }
        
        this.logEvent({
            type: OspreyDebugEventType.Debug,
            message: 'Osprey debug adapter initialized'
        });
    }

    /**
     * Initialize the debug adapter
     */
    protected initializeRequest(response: DebugProtocol.InitializeResponse, args: DebugProtocol.InitializeRequestArguments): void {
        // Log platform information
        this.logEvent({
            type: OspreyDebugEventType.Debug,
            message: `Debug session starting on ${os.platform()} (${os.type()} ${os.release()})`,
            data: { platform: os.platform(), type: os.type(), release: os.release() }
        });
        
        response.body = response.body || {};
        response.body.supportsConfigurationDoneRequest = false;
        response.body.supportsEvaluateForHovers = false;
        response.body.supportTerminateDebuggee = true;
        response.body.supportsRestartRequest = true;

        this.sendResponse(response);
        this.sendEvent(new InitializedEvent());
    }    /**
     * Log a debug event
     */
    private logEvent(event: OspreyDebugEvent): void {
        switch (event.type) {
            case OspreyDebugEventType.Debug:
                logger.verbose(`[Osprey Debug] ${event.message}`);
                break;
            case OspreyDebugEventType.Info:
                logger.log(`[Osprey Info] ${event.message}`);
                break;
            case OspreyDebugEventType.Warning:
                logger.warn(`[Osprey Warning] ${event.message}`);
                break;
            case OspreyDebugEventType.Error:
            case OspreyDebugEventType.LaunchError:
            case OspreyDebugEventType.CompilerNotFound:
            case OspreyDebugEventType.FileNotFound:
                logger.error(`[Osprey Error] ${event.message}`);
                break;
            default:
                logger.log(`[Osprey] ${event.message}`);
        }
    }

    /**
     * Handle launch request (debug session start)
     */
    protected async launchRequest(response: DebugProtocol.LaunchResponse, args: any): Promise<void> {
        this.logEvent({
            type: OspreyDebugEventType.LaunchStart,
            message: `Launch request received for ${args.program}`,
            data: args
        });
        
        // Get the program path from launch args
        const programPath = args.program;
        if (!programPath || !fs.existsSync(programPath)) {
            const errorMsg = `Cannot find file: ${programPath}`;
            
            this.logEvent({
                type: OspreyDebugEventType.FileNotFound,
                message: errorMsg
            });
            
            this.sendEvent(new OutputEvent(`Error: ${errorMsg}\n`, 'stderr'));
            this.sendErrorResponse(response, 1000, errorMsg);
            return;
        }

        // Get compiler path
        const compilerPath = await this.getCompilerPath();
        if (!compilerPath) {
            const errorMsg = 'Cannot find Osprey compiler. Please check extension installation.';
            
            this.logEvent({
                type: OspreyDebugEventType.CompilerNotFound,
                message: errorMsg
            });
            
            if (!this.hasShownCompilerError) {
                this.sendEvent(new OutputEvent(`${errorMsg}\n\nPossible solutions:\n`, 'stderr'));
                this.sendEvent(new OutputEvent('1. Run the "Osprey: Reset Compiler Settings" command\n', 'stderr'));
                this.sendEvent(new OutputEvent('2. Check that the extension is properly installed\n', 'stderr'));
                this.sendEvent(new OutputEvent('3. Set a custom compiler path in settings\n', 'stderr'));
                this.hasShownCompilerError = true;
            } else {
                this.sendEvent(new OutputEvent(`${errorMsg}\n`, 'stderr'));
            }
            
            this.sendErrorResponse(response, 1001, errorMsg);
            return;
        }

        // Run the program with the Osprey compiler
        const fileDir = path.dirname(programPath);
        
        // Send a message about what we're doing
        const runMsg = `Running ${programPath} with Osprey compiler: ${compilerPath}`;
        this.logEvent({
            type: OspreyDebugEventType.Info,
            message: runMsg
        });
        this.sendEvent(new OutputEvent(`${runMsg}\n`, 'console'));
        
        try {
            // Kill any existing process
            if (this.childProcess) {
                this.childProcess.kill();
            }
            
            // Use spawn instead of execFile to get streaming output
            this.childProcess = spawn(compilerPath, [programPath, '--run'], { cwd: fileDir });
            
            // Stream stdout
            this.childProcess.stdout?.on('data', (data: Buffer) => {
                const output = data.toString();
                this.sendEvent(new OutputEvent(output, 'stdout'));
            });
            
            // Stream stderr
            this.childProcess.stderr?.on('data', (data: Buffer) => {
                const output = data.toString();
                this.sendEvent(new OutputEvent(output, 'stderr'));
            });
            
            // Handle process completion
            this.childProcess.on('close', (code: number | null) => {
                if (code !== 0) {
                    this.logEvent({
                        type: OspreyDebugEventType.ProgramExit,
                        message: `Program exited with code ${code}`,
                        data: { exitCode: code }
                    });
                    this.sendEvent(new OutputEvent(`Program exited with code ${code}\n`, 'console'));
                } else {
                    this.logEvent({
                        type: OspreyDebugEventType.LaunchSuccess,
                        message: 'Program execution complete'
                    });
                    this.sendEvent(new OutputEvent('Program execution complete with exit code 0\n', 'console'));
                }
                this.sendEvent(new TerminatedEvent());
                this.childProcess = undefined;
            });
            
            // Handle process errors
            this.childProcess.on('error', (err: Error) => {
                const errorMsg = `Error executing process: ${err.message}`;
                this.logEvent({
                    type: OspreyDebugEventType.LaunchError,
                    message: errorMsg,
                    data: err
                });
                
                // Handle common error cases with helpful messages
                if (err.message.includes('ENOENT')) {
                    this.sendEvent(new OutputEvent('The Osprey compiler was not found or is not executable.\n', 'stderr'));
                } else if (err.message.includes('EACCES')) {
                    this.sendEvent(new OutputEvent('The Osprey compiler does not have execute permissions.\n', 'stderr'));
                    if (process.platform !== 'win32') {
                        this.sendEvent(new OutputEvent('Try running: chmod +x <compiler-path>\n', 'stderr'));
                    }
                } else {
                    this.sendEvent(new OutputEvent(`${errorMsg}\n`, 'stderr'));
                }
                
                this.sendEvent(new TerminatedEvent());
                this.childProcess = undefined;
            });
            
            // Send response that we've started launching
            this.sendResponse(response);
            
        } catch (err: any) {
            const errorMsg = `Failed to run Osprey program: ${err.message}`;
            
            this.logEvent({
                type: OspreyDebugEventType.LaunchError,
                message: errorMsg,
                data: err
            });
            
            this.sendErrorResponse(response, 1002, errorMsg);
        }
    }

    /**
     * Handle termination requests
     */
    protected terminateRequest(response: DebugProtocol.TerminateResponse, args: DebugProtocol.TerminateArguments, request?: DebugProtocol.Request): void {
        this.logEvent({
            type: OspreyDebugEventType.Debug,
            message: 'Terminate request received'
        });
        
        // Kill any running process
        if (this.childProcess) {
            this.logEvent({
                type: OspreyDebugEventType.Debug,
                message: 'Killing child process'
            });
            
            this.childProcess.kill();
            this.childProcess = undefined;
        }
        
        this.sendResponse(response);
        this.sendEvent(new TerminatedEvent());
    }
    
    /**
     * Support for restart requests
     */
    protected restartRequest(response: DebugProtocol.RestartResponse, args: DebugProtocol.RestartArguments): void {
        this.logEvent({
            type: OspreyDebugEventType.Debug,
            message: 'Restart request received'
        });
        
        this.sendEvent(new OutputEvent('Restarting Osprey program...\n', 'console'));
        
        // For now, just acknowledge the restart but don't actually implement it
        // The client will handle restarting the debug session
        this.sendResponse(response);
    }    /**
     * Get the path to the Osprey compiler
     * This uses multiple search strategies and provides detailed logging
     */
    private async getCompilerPath(): Promise<string | undefined> {
        // Return cached path if available
        if (this.configPath) {
            return this.configPath;
        }
        
        this.logEvent({
            type: OspreyDebugEventType.Debug,
            message: 'Looking for compiler path'
        });
        
        // Try to get path from VS Code settings
        try {
            if (vscode.workspace.getConfiguration) {
                const config = vscode.workspace.getConfiguration('osprey');
                const compilerPathSetting = config.get<string>('server.compilerPath', '');
                
                if (compilerPathSetting && compilerPathSetting.trim() !== '') {
                    let resolvedPath = compilerPathSetting;
                    
                    // Handle variable substitution
                    if (resolvedPath.includes('${workspaceFolder}') && this.workspaceFolders.length > 0) {
                        resolvedPath = resolvedPath.replace('${workspaceFolder}', this.workspaceFolders[0]);
                        this.logEvent({
                            type: OspreyDebugEventType.Debug,
                            message: `Resolved workspace path: ${resolvedPath}`
                        });
                    }
                    
                    // Check if the specified compiler exists
                    if (fs.existsSync(resolvedPath)) {
                        this.logEvent({
                            type: OspreyDebugEventType.Info,
                            message: `Found compiler from settings: ${resolvedPath}`
                        });
                        this.configPath = resolvedPath;
                        return resolvedPath;
                    } else {
                        this.logEvent({
                            type: OspreyDebugEventType.Warning,
                            message: `Configured compiler not found at: ${resolvedPath}`
                        });
                    }
                }
            }
        } catch (err: any) {
            this.logEvent({
                type: OspreyDebugEventType.Warning,
                message: `Error accessing VS Code settings: ${err.message}`
            });
        }
        
        // Try to find the compiler in PATH
        try {
            const pathVar = process.env.PATH || '';
            const pathDirs = pathVar.split(path.delimiter);
            const exeExtension = process.platform === 'win32' ? '.exe' : '';
            
            this.logEvent({
                type: OspreyDebugEventType.Debug,
                message: 'Checking PATH for osprey compiler'
            });
            
            for (const dir of pathDirs) {
                const possiblePath = path.join(dir, `osprey${exeExtension}`);
                if (fs.existsSync(possiblePath)) {
                    this.logEvent({
                        type: OspreyDebugEventType.Info,
                        message: `Found osprey compiler in PATH: ${possiblePath}`
                    });
                    this.configPath = possiblePath;
                    return possiblePath;
                }
            }
        } catch (err: any) {
            this.logEvent({
                type: OspreyDebugEventType.Warning,
                message: `Error searching PATH: ${err.message}`
            });
        }
        
        // Try to get path from the extension bundle
        try {
            if (!this.extensionPath) {
                this.logEvent({
                    type: OspreyDebugEventType.Warning,
                    message: 'Extension path not available'
                });
            } else {
                // Check platform-specific possibilities
                if (process.platform === 'win32') {
                    // Windows: prefer cmd wrapper, then exe, then js
                    const cmdPath = path.join(this.extensionPath, 'bin', 'osprey.cmd');
                    const exePath = path.join(this.extensionPath, 'bin', 'osprey.exe');
                    const jsPath = path.join(this.extensionPath, 'bin', 'osprey.js');
                    
                    if (fs.existsSync(cmdPath)) {
                        this.logEvent({
                            type: OspreyDebugEventType.Info,
                            message: `Using bundled Windows cmd wrapper: ${cmdPath}`
                        });
                        this.configPath = cmdPath;
                        return cmdPath;
                    } else if (fs.existsSync(exePath)) {
                        this.logEvent({
                            type: OspreyDebugEventType.Info,
                            message: `Using bundled Windows exe: ${exePath}`
                        });
                        this.configPath = exePath;
                        return exePath;
                    } else if (fs.existsSync(jsPath)) {
                        this.logEvent({
                            type: OspreyDebugEventType.Info,
                            message: `Using bundled JS version: ${jsPath}`
                        });
                        this.configPath = jsPath;
                        return jsPath;
                    }
                } else {
                    // Unix: try native binary then js
                    const nativePath = path.join(this.extensionPath, 'bin', 'osprey');
                    const jsPath = path.join(this.extensionPath, 'bin', 'osprey.js');
                    
                    if (fs.existsSync(nativePath)) {
                        // Ensure executable permissions
                        try {
                            fs.chmodSync(nativePath, 0o755); // rwxr-xr-x
                        } catch (error: any) {
                            this.logEvent({
                                type: OspreyDebugEventType.Warning,
                                message: `Failed to set executable permissions: ${error.message}`
                            });
                        }
                        
                        this.logEvent({
                            type: OspreyDebugEventType.Info,
                            message: `Using bundled Unix native binary: ${nativePath}`
                        });
                        this.configPath = nativePath;
                        return nativePath;
                    } else if (fs.existsSync(jsPath)) {
                        // Ensure executable permissions
                        try {
                            fs.chmodSync(jsPath, 0o755); // rwxr-xr-x
                        } catch (error: any) {
                            this.logEvent({
                                type: OspreyDebugEventType.Warning,
                                message: `Failed to set executable permissions: ${error.message}`
                            });
                        }
                        
                        this.logEvent({
                            type: OspreyDebugEventType.Info,
                            message: `Using bundled JS version: ${jsPath}`
                        });
                        this.configPath = jsPath;
                        return jsPath;
                    }
                }
            }
        } catch (err: any) {
            this.logEvent({
                type: OspreyDebugEventType.Error,
                message: `Error finding bundled compiler: ${err.message}`
            });
        }
        
        // Log failure
        this.logEvent({
            type: OspreyDebugEventType.CompilerNotFound,
            message: 'Failed to find Osprey compiler after exhaustive search'
        });        
        return undefined;
    }
}
}

// Don't register the debug adapter factory here, it should be done in the extension.ts activate function
// The OspreyDebugSession class is already exported at the beginning of the file
