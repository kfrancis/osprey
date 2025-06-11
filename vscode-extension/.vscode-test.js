const { defineConfig } = require('@vscode/test-cli');

module.exports = defineConfig({
  files: 'out/test/suite/**/*.test.js',
  version: 'stable',
  mocha: {
    ui: 'tdd',
    timeout: 10000,
    color: true
  },
  launchArgs: [
    '--disable-extensions',
    '--disable-workspace-trust'
  ]
}); 