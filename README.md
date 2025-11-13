# MCP Server for ArgoCD

A Model Context Protocol (MCP) server providing tools to interact with ArgoCD.

## Quick Start

### First Time Setup

```bash
# Quick setup (single command)
cp .env.sh.example .env.sh

# Edit .env.sh with your ArgoCD credentials
# Required: ARGOCD_BASE_URL and ARGOCD_API_TOKEN
```

See [ENV_SETUP.md](ENV_SETUP.md) for detailed environment setup instructions, including how to get your ArgoCD credentials.

See [DEVELOPMENT.md](DEVELOPMENT.md) for detailed development workflow.

### Fast Feedback Loop

```bash
# Test your changes immediately
task test

# Watch for changes (auto-test)
task watch
```

## Development

See [DEVELOPMENT.md](DEVELOPMENT.md) for the complete development guide including:
- Fast feedback loop options
- Testing strategies
- Debugging tips
- Common workflows
