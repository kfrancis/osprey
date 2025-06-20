# Using the Osprey Extension with TaskMaster

This guide explains how to use the Osprey VS Code Extension with TaskMaster to manage your Osprey language development projects.

## What is TaskMaster?

[TaskMaster](https://github.com/christianfindlay/taskmaster) is a task management tool for software development that helps you organize and track development tasks. It integrates with VS Code to provide a structured way to manage your development workflow.

## Setting Up TaskMaster for an Osprey Project

1. **Initialize a TaskMaster Project**

```bash
cd your-osprey-project
taskmaster-ai init
```

Or use the TaskMaster extension in VS Code to initialize a project.

2. **Create Tasks for Osprey Development**

Create a Product Requirements Document (PRD) in `.taskmaster/docs/prd.txt` describing your Osprey project, then generate tasks:

```bash
taskmaster-ai parse-prd
```

## Sample TaskMaster Tasks for Osprey Development

Here are some example tasks you might create for an Osprey project:

1. **Set up Osprey development environment**
   - Install Osprey compiler
   - Configure VS Code extension
   - Create example project structure

2. **Implement core language features**
   - Define data types and structures
   - Implement functions and procedures
   - Create control flow structures

3. **Build standard library**
   - IO operations
   - String manipulation
   - Mathematical functions
   - Collection types

4. **Create test suite**
   - Unit tests for language features
   - Integration tests for larger programs
   - Benchmarks for performance evaluation

## VS Code Tasks for Osprey

The Osprey VS Code extension includes several pre-defined tasks:

1. **Install Dependencies**: Install all required dependencies
2. **Compile Extension**: Compile the extension code
3. **Watch and Compile Extension**: Continuously compile changes
4. **Package Extension**: Create a VSIX package
5. **Run Extension Tests**: Run the test suite
6. **Bundle Osprey Compiler**: Bundle the Osprey compiler with the extension
7. **Build and Run Extension**: Compile and launch the extension

You can run these tasks using the VS Code command palette (Ctrl+Shift+P) and typing "Tasks: Run Task".

## Creating Custom Tasks for Osprey Projects

You can create custom tasks for your Osprey projects by adding them to the `.vscode/tasks.json` file:

```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "Compile Osprey Project",
      "type": "shell",
      "command": "osprey compile ${workspaceFolder}/src/main.osp -o ${workspaceFolder}/bin/main",
      "group": {
        "kind": "build",
        "isDefault": true
      },
      "presentation": {
        "reveal": "always",
        "panel": "shared"
      }
    },
    {
      "label": "Run Osprey Project",
      "type": "shell",
      "command": "${workspaceFolder}/bin/main",
      "group": "test",
      "presentation": {
        "reveal": "always",
        "panel": "shared"
      },
      "dependsOn": ["Compile Osprey Project"]
    }
  ]
}
```

## Integrating with TaskMaster Workflow

When using TaskMaster with Osprey projects, follow this workflow:

1. **Plan**: Create tasks in TaskMaster for your Osprey development
2. **Implement**: Write your Osprey code and use the VS Code extension features
3. **Test**: Use the built-in compiler integration to test your code
4. **Track**: Update task status in TaskMaster as you complete work
5. **Iterate**: Expand tasks into subtasks as needed for complex features

## Tips for Productive Osprey Development

1. Use the VS Code extension's syntax highlighting and code completion
2. Set up snippets for common Osprey patterns
3. Use the built-in compiler integration to quickly test changes
4. Track your progress using TaskMaster's task management
5. Create custom tasks for your specific project needs

By combining the Osprey VS Code extension with TaskMaster, you can create a powerful development environment for Osprey programming language projects.
