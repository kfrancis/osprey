#!/usr/bin/env node

console.log('Cross-platform Osprey compiler mock');
console.log('Platform: ' + process.platform);
console.log('Architecture: ' + process.arch);
console.log('Node version: ' + process.version);
console.log('');

const args = process.argv.slice(2);
const fs = require('fs');
const path = require('path');

// Parse command-line arguments
let runMode = false;
let inputFile = null;

args.forEach(arg => {
  if (arg === '--run') {
    runMode = true;
  } else if (!inputFile && !arg.startsWith('--')) {
    inputFile = arg;
  }
});

if (!inputFile) {
  console.error('Error: No input file specified');
  process.exit(1);
}

try {
  console.log(`Processing file: ${inputFile}`);
  const fileContent = fs.readFileSync(inputFile, 'utf8');
  
  // Simple check for syntax errors
  let hasErrors = false;
  
  if (runMode) {
    console.log('Running Osprey program in cross-platform mode...');
    
    // Extract main function content
    if (fileContent.includes('fn main()')) {
      // Find print statements and simulate their execution
      const printMatches = fileContent.match(/print\(["'](.*?)["']\)/g);
      if (printMatches) {
        console.log('Program output:');
        printMatches.forEach(match => {
          const content = match.match(/["'](.*?)["']/)[1];
          console.log(content);
        });
      }
      
      console.log('Program execution complete');
    }
  } else {
    console.log('Compilation successful');
  }
} catch (err) {
  console.error(`Error: ${err.message}`);
  process.exit(1);
}
