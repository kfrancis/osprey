// Playground JavaScript functionality
class OspreyPlayground {
  constructor() {
    this.examples = {};
    this.currentExample = null;
    this.autoRun = true;
    this.init();
  }

  init() {
    this.loadExamples();
    this.setupEventListeners();
    this.setupTabs();
    this.setupSettings();
    this.loadDefaultExample();
  }

  loadExamples() {
    const examplesScript = document.getElementById('playground-examples');
    if (examplesScript) {
      try {
        this.examples = JSON.parse(examplesScript.textContent);
      } catch (e) {
        console.error('Failed to load examples:', e);
      }
    }
  }

  setupEventListeners() {
    // Example buttons
    const exampleButtons = document.querySelectorAll('.example-btn');
    exampleButtons.forEach(btn => {
      btn.addEventListener('click', () => {
        const example = btn.dataset.example;
        this.loadExample(example);
        this.setActiveExample(btn);
      });
    });

    // Editor actions
    const runBtn = document.getElementById('run-code');
    const formatBtn = document.getElementById('format-code');
    const shareBtn = document.getElementById('share-code');
    const clearOutputBtn = document.getElementById('clear-output');

    runBtn?.addEventListener('click', () => this.runCode());
    formatBtn?.addEventListener('click', () => this.formatCode());
    shareBtn?.addEventListener('click', () => this.shareCode());
    clearOutputBtn?.addEventListener('click', () => this.clearOutput());

    // Code input changes
    const codeInput = document.getElementById('code-input');
    if (codeInput) {
      codeInput.addEventListener('input', () => {
        if (this.autoRun) {
          this.debounce(() => this.runCode(), 500)();
        }
      });

      // Tab key handling for code editor
      codeInput.addEventListener('keydown', (e) => {
        if (e.key === 'Tab') {
          e.preventDefault();
          const start = codeInput.selectionStart;
          const end = codeInput.selectionEnd;
          
          // Insert tab character
          codeInput.value = codeInput.value.substring(0, start) + 
                          '  ' + 
                          codeInput.value.substring(end);
          
          // Put cursor after tab
          codeInput.selectionStart = codeInput.selectionEnd = start + 2;
        }
      });
    }

    // URL parameter handling for shared code
    this.handleUrlParameters();
  }

  setupTabs() {
    const tabs = document.querySelectorAll('.editor-tab');
    const panels = document.querySelectorAll('.editor-panel');

    tabs.forEach(tab => {
      tab.addEventListener('click', () => {
        const targetPanel = tab.dataset.tab;
        
        // Update active tab
        tabs.forEach(t => t.classList.remove('active'));
        tab.classList.add('active');
        
        // Update active panel
        panels.forEach(p => p.classList.remove('active'));
        const panel = document.querySelector(`[data-panel="${targetPanel}"]`);
        if (panel) {
          panel.classList.add('active');
        }

        // Generate AST if AST tab is selected
        if (targetPanel === 'ast') {
          this.generateAST();
        }
      });
    });
  }

  setupSettings() {
    const autoRunCheckbox = document.getElementById('auto-run');
    const showTypesCheckbox = document.getElementById('show-types');
    const fontSizeSelect = document.getElementById('font-size');

    autoRunCheckbox?.addEventListener('change', (e) => {
      this.autoRun = e.target.checked;
    });

    showTypesCheckbox?.addEventListener('change', (e) => {
      this.showTypes = e.target.checked;
      // TODO: Implement type annotations in editor
    });

    fontSizeSelect?.addEventListener('change', (e) => {
      const fontSize = e.target.value;
      const codeInput = document.getElementById('code-input');
      if (codeInput) {
        codeInput.style.fontSize = fontSize + 'px';
      }
    });
  }

  loadExample(exampleKey) {
    if (this.examples[exampleKey]) {
      const codeInput = document.getElementById('code-input');
      if (codeInput) {
        codeInput.value = this.examples[exampleKey];
        this.currentExample = exampleKey;
        if (this.autoRun) {
          this.runCode();
        }
      }
    }
  }

  setActiveExample(button) {
    document.querySelectorAll('.example-btn').forEach(btn => {
      btn.classList.remove('active');
    });
    button.classList.add('active');
  }

  loadDefaultExample() {
    const firstExample = document.querySelector('.example-btn');
    if (firstExample) {
      firstExample.click();
    }
  }

  runCode() {
    const codeInput = document.getElementById('code-input');
    const programOutput = document.getElementById('program-output');
    const compilationOutput = document.getElementById('compilation-output');

    if (!codeInput || !programOutput || !compilationOutput) return;

    const code = codeInput.value;
    
    // Simulate compilation
    this.simulateCompilation(code, compilationOutput, programOutput);
  }

  simulateCompilation(code, compilationEl, outputEl) {
    // Clear previous output
    outputEl.innerHTML = '';
    
    // Show compilation in progress
    compilationEl.innerHTML = '<div class="compilation-status">⏳ Compiling...</div>';

    // Simulate compilation delay
    setTimeout(() => {
      try {
        const result = this.parseAndExecute(code);
        
        if (result.success) {
          compilationEl.innerHTML = '<div class="compilation-status success">✅ Code compiled successfully</div>';
          outputEl.innerHTML = this.formatOutput(result.output);
        } else {
          compilationEl.innerHTML = `<div class="compilation-status error">❌ Compilation failed<br><small>${result.error}</small></div>`;
          outputEl.innerHTML = '<div class="output-error">Compilation failed. See compilation messages above.</div>';
        }
      } catch (e) {
        compilationEl.innerHTML = `<div class="compilation-status error">❌ Compilation failed<br><small>${e.message}</small></div>`;
        outputEl.innerHTML = '<div class="output-error">Compilation failed. See compilation messages above.</div>';
      }
    }, 500);
  }

  parseAndExecute(code) {
    // Simple parser simulation for demo purposes
    // In a real implementation, this would call the Osprey compiler
    
    const output = [];
    
    // Look for print statements and extract their arguments
    const printRegex = /print\s*\(\s*"([^"]+)"\s*\)/g;
    const interpolationRegex = /print\s*\(\s*"([^"]*\${[^}]+}[^"]*)"\s*\)/g;
    
    let match;
    
    // Handle simple string prints
    while ((match = printRegex.exec(code)) !== null) {
      output.push(match[1]);
    }
    
    // Handle string interpolation (basic simulation)
    code = code.replace(interpolationRegex, (match, str) => {
      // Simple interpolation simulation
      const processed = str.replace(/\$\{([^}]+)\}/g, (m, expr) => {
        // Very basic expression evaluation
        if (expr.includes('name')) return 'World';
        if (expr.includes('value')) return '5';
        if (expr.includes('error')) return 'DivisionByZero';
        return expr;
      });
      output.push(processed);
      return match;
    });

    // Check for syntax errors (very basic)
    const hasUnclosedBraces = (code.match(/\{/g) || []).length !== (code.match(/\}/g) || []).length;
    const hasUnclosedParens = (code.match(/\(/g) || []).length !== (code.match(/\)/g) || []).length;
    
    if (hasUnclosedBraces) {
      return { success: false, error: 'Syntax error: Unclosed braces' };
    }
    
    if (hasUnclosedParens) {
      return { success: false, error: 'Syntax error: Unclosed parentheses' };
    }

    // Simulate some outputs for different examples
    if (code.includes('Hello World') || code.includes('greet')) {
      output.push('Hello World! Welcome to Osprey.');
    }
    
    if (code.includes('divide') && code.includes('10') && code.includes('2')) {
      output.push('10 / 2 = 5');
    }
    
    if (code.includes('divide') && code.includes('0')) {
      output.push('Cannot divide by zero!');
    }
    
    if (code.includes('Circle') && code.includes('area')) {
      output.push('Circle area: 75');
      output.push('Rectangle area: 24');
      output.push('Triangle area: 12');
    }

    if (code.includes('createUser')) {
      output.push('User: Alice, Age: 30, Email: alice@example.com, Active: true');
      output.push('User: Bob, Age: 25, Email: bob@example.com, Active: false');
    }

    if (code.includes('range') && code.includes('filter')) {
      output.push('Sum of squares of even numbers 1-19: 1140');
      output.push('Welcome to Osprey v1!');
    }

    if (code.includes('Fiber') && code.includes('await')) {
      output.push('Worker 1 completed');
      output.push('Worker 2 completed');
      output.push('Received: 1, 2, 3');
    }

    if (code.includes('UserAccount') && code.includes('validation')) {
      output.push('Valid account created for alice123');
      output.push('Account creation failed: Validation error');
    }

    if (output.length === 0) {
      output.push('Program executed successfully (no output)');
    }

    return { success: true, output };
  }

  formatOutput(outputArray) {
    return outputArray.map(line => `<div class="output-line">${this.escapeHtml(line)}</div>`).join('');
  }

  escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
  }

  formatCode() {
    const codeInput = document.getElementById('code-input');
    if (!codeInput) return;

    // Basic code formatting (in a real implementation, this would use the Osprey formatter)
    let code = codeInput.value;
    
    // Simple indentation fixes
    const lines = code.split('\n');
    let indentLevel = 0;
    const formattedLines = lines.map(line => {
      const trimmed = line.trim();
      
      if (trimmed === '') return '';
      
      // Decrease indent for closing braces
      if (trimmed.startsWith('}')) {
        indentLevel = Math.max(0, indentLevel - 1);
      }
      
      const formatted = '  '.repeat(indentLevel) + trimmed;
      
      // Increase indent after opening braces
      if (trimmed.endsWith('{')) {
        indentLevel++;
      }
      
      return formatted;
    });
    
    codeInput.value = formattedLines.join('\n');
    
    if (this.autoRun) {
      this.runCode();
    }
  }

  shareCode() {
    const codeInput = document.getElementById('code-input');
    if (!codeInput) return;

    const code = codeInput.value;
    const encodedCode = encodeURIComponent(code);
    const shareUrl = `${window.location.origin}${window.location.pathname}?code=${encodedCode}`;
    
    if (navigator.clipboard) {
      navigator.clipboard.writeText(shareUrl).then(() => {
        this.showToast('Share URL copied to clipboard!');
      });
    } else {
      // Fallback for older browsers
      const textArea = document.createElement('textarea');
      textArea.value = shareUrl;
      document.body.appendChild(textArea);
      textArea.select();
      document.execCommand('copy');
      document.body.removeChild(textArea);
      this.showToast('Share URL copied to clipboard!');
    }
  }

  clearOutput() {
    const programOutput = document.getElementById('program-output');
    if (programOutput) {
      programOutput.innerHTML = '<div class="output-placeholder">Click "Run" to see the output of your program.</div>';
    }
  }

  generateAST() {
    const codeInput = document.getElementById('code-input');
    const astOutput = document.getElementById('ast-output');
    
    if (!codeInput || !astOutput) return;

    const code = codeInput.value;
    
    // Simulate AST generation (in a real implementation, this would call the Osprey parser)
    const ast = this.simulateAST(code);
    astOutput.innerHTML = `<pre>${this.escapeHtml(JSON.stringify(ast, null, 2))}</pre>`;
  }

  simulateAST(code) {
    // Very basic AST simulation for demo
    return {
      type: "Program",
      statements: [
        {
          type: "FunctionDeclaration",
          name: "greet",
          parameters: [{ name: "name", type: "String" }],
          returnType: "String",
          body: {
            type: "StringLiteral",
            value: "Hello ${name}! Welcome to Osprey."
          }
        },
        {
          type: "LetDeclaration",
          name: "message",
          value: {
            type: "FunctionCall",
            name: "greet",
            arguments: [
              { name: "name", value: { type: "StringLiteral", value: "World" } }
            ]
          }
        },
        {
          type: "ExpressionStatement",
          expression: {
            type: "FunctionCall",
            name: "print",
            arguments: [{ type: "Identifier", name: "message" }]
          }
        }
      ]
    };
  }

  handleUrlParameters() {
    const urlParams = new URLSearchParams(window.location.search);
    const code = urlParams.get('code');
    
    if (code) {
      const codeInput = document.getElementById('code-input');
      if (codeInput) {
        codeInput.value = decodeURIComponent(code);
        if (this.autoRun) {
          this.runCode();
        }
      }
    }
  }

  showToast(message) {
    // Simple toast notification
    const toast = document.createElement('div');
    toast.className = 'toast';
    toast.textContent = message;
    toast.style.cssText = `
      position: fixed;
      top: 20px;
      right: 20px;
      background: var(--color-success);
      color: white;
      padding: 12px 24px;
      border-radius: 6px;
      z-index: 1000;
      font-size: 14px;
      box-shadow: 0 4px 12px rgba(0,0,0,0.15);
    `;
    
    document.body.appendChild(toast);
    
    setTimeout(() => {
      toast.remove();
    }, 3000);
  }

  debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
      const later = () => {
        clearTimeout(timeout);
        func(...args);
      };
      clearTimeout(timeout);
      timeout = setTimeout(later, wait);
    };
  }
}

// Initialize playground when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
  new OspreyPlayground();
}); 