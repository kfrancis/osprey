# Bundle Osprey Compiler for VSCode Extension
# This script downloads or builds the Osprey compiler and packages it with the extension

# Script parameters
param(
    [string]$OspreyVersion = "0.1.0",
    [string]$DownloadUrl = "",
    [switch]$BuildFromSource = $false
)

Write-Host "🦅 Bundling Osprey compiler for VSCode extension..."

# Create the bin directory if it doesn't exist
$binDir = Join-Path -Path $PSScriptRoot -ChildPath ".." -AdditionalChildPath "bin"
if (-not (Test-Path -Path $binDir)) {
    New-Item -ItemType Directory -Path $binDir | Out-Null
    Write-Host "Created bin directory at: $binDir"
}

# Function to download the Osprey compiler
function Download-OspreyCompiler {
    $tempFile = Join-Path -Path $env:TEMP -ChildPath "osprey-$OspreyVersion.zip"
    
    if ([string]::IsNullOrEmpty($DownloadUrl)) {
        # If no URL is provided, construct a default one (this is hypothetical)
        $DownloadUrl = "https://github.com/osprey/osprey/releases/download/v$OspreyVersion/osprey-$OspreyVersion-win-x64.zip"
    }
    
    Write-Host "Downloading Osprey compiler from: $DownloadUrl"
    
    try {
        Invoke-WebRequest -Uri $DownloadUrl -OutFile $tempFile
        
        Write-Host "Extracting compiler to bin directory..."
        Expand-Archive -Path $tempFile -DestinationPath $binDir -Force
        
        # Rename the compiler executable if needed
        $compilerPath = Join-Path -Path $binDir -ChildPath "osprey.exe"
        if (-not (Test-Path -Path $compilerPath)) {
            # Try to find the executable in extracted folders
            $exeFiles = Get-ChildItem -Path $binDir -Recurse -Filter "*.exe"
            if ($exeFiles.Count -gt 0) {
                Copy-Item -Path $exeFiles[0].FullName -Destination $compilerPath
                Write-Host "Copied compiler executable to: $compilerPath"
            } else {
                Write-Host "Warning: Could not find Osprey compiler executable in extracted files" -ForegroundColor Yellow
            }
        }
        
        # Clean up
        Remove-Item -Path $tempFile -Force
    }
    catch {
        Write-Host "Error downloading or extracting Osprey compiler: $_" -ForegroundColor Red
        return $false
    }
    
    return $true
}

# Function to create a mock compiler for testing
function Create-MockCompiler {
    Write-Host "Creating mock Osprey compiler for development..."
    
    $compilerPath = Join-Path -Path $binDir -ChildPath "osprey.js"
    $compilerContent = @"
#!/usr/bin/env node

/**
 * Mock Osprey compiler for VSCode extension development
 * This is a simple JavaScript implementation to simulate the Osprey compiler
 */

const fs = require('fs');
const path = require('path');

// Parse command line arguments
const args = process.argv.slice(2);
let runMode = false;
let inputFile = null;

// Process arguments
args.forEach(arg => {
  if (arg === '--run') {
    runMode = true;
  } else if (!inputFile && !arg.startsWith('--')) {
    inputFile = arg;
  }
});

// Check if we have an input file
if (!inputFile) {
  console.error('Error: No input file specified');
  process.exit(1);
}

// Read the file content
try {
  const fileContent = fs.readFileSync(inputFile, 'utf8');
  
  // Simple syntax check - detect common errors
  let hasErrors = false;
  const lines = fileContent.split('\n');
  
  lines.forEach((line, index) => {
    // Example checks:
    
    // Check for missing equals in function definition
    if (line.trim().startsWith('fn ') && !line.includes('=')) {
      console.error(`${inputFile}:${index + 1}:${line.indexOf('fn') + 3}: Error: Function declaration missing '='`);
      hasErrors = true;
    }
    
    // Check for unclosed brackets    if ((line.includes('{') -and !fileContent.includes('}')) -or 
        (line.includes('[') -and !fileContent.includes(']')) -or
        (line.includes('(') -and !fileContent.includes(')'))) {
      console.error("$inputFile`:$(index + 1)`:$(line.indexOf('{') -or line.indexOf('[') -or line.indexOf('(')): Error: Unclosed bracket");
      hasErrors = true;
    }
  });
  
  // If in run mode and no errors, "execute" the code
  if (runMode && !hasErrors) {
    // Extract and run the main function if it exists
    if (fileContent.includes('fn main()')) {
      console.log('Executing Osprey program...');
      
      // Find print statements and simulate their output
      const printMatches = fileContent.match(/print\(["'](.*?)["']\)/g);
      if (printMatches) {
        printMatches.forEach(match => {
          // Extract the content inside the quotes
          const content = match.match(/["'](.*?)["']/)[1];
          console.log(content);
        });
      }
      
      console.log('Program execution complete');
    } else {
      console.log('No main function found, nothing to execute');
    }
  }
  
  // Exit with error code if errors were found
  if (hasErrors) {
    process.exit(1);
  }
} catch (err) {
  console.error(`Error reading or processing file: ${err.message}`);
  process.exit(1);
}
"@

    # Write the mock compiler
    Set-Content -Path $compilerPath -Value $compilerContent
    
    # Create a batch file wrapper for Windows
    $batchPath = Join-Path -Path $binDir -ChildPath "osprey.cmd"
    $batchContent = @"
@echo off
node "%~dp0osprey.js" %*
"@

    Set-Content -Path $batchPath -Value $batchContent
    
    Write-Host "Created mock compiler at: $compilerPath and wrapper at: $batchPath"
    return $true
}

# Main script execution
if ($BuildFromSource) {
    Write-Host "Build from source option not implemented yet. Using mock compiler instead." -ForegroundColor Yellow
    $success = Create-MockCompiler
} else {
    # Try downloading first, fall back to mock compiler
    $success = Download-OspreyCompiler
    
    if (-not $success) {
        Write-Host "Failed to download compiler. Creating mock compiler instead..." -ForegroundColor Yellow
        $success = Create-MockCompiler
    }
}

# Update package.json to include the compiler path
Write-Host "Updating package.json with compiler path..."
$packageJsonPath = Join-Path -Path $PSScriptRoot -ChildPath ".." -AdditionalChildPath "package.json"
$packageJson = Get-Content -Path $packageJsonPath | ConvertFrom-Json

# Ensure the configuration section exists
if (-not $packageJson.contributes.configuration) {
    $packageJson.contributes | Add-Member -Type NoteProperty -Name "configuration" -Value @{}
}

# Add or update the compiler path setting
if (-not $packageJson.contributes.configuration.properties) {
    $packageJson.contributes.configuration | Add-Member -Type NoteProperty -Name "properties" -Value @{}
}

$packageJson.contributes.configuration.properties | Add-Member -Type NoteProperty -Name "osprey.server.compilerPath" -Value @{
    "type" = "string"
    "default" = "${workspaceFolder}/bin/osprey"
    "description" = "Path to the Osprey compiler executable"
} -Force

$packageJson | ConvertTo-Json -Depth 10 | Set-Content -Path $packageJsonPath

Write-Host "✅ Osprey compiler bundling complete!"
