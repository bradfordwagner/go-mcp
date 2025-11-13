# Environment Setup

## Quick Setup (Single Command)

```bash
cp .env.sh.example .env.sh && echo "✅ Created .env.sh - Now edit it with your credentials"
```

Then edit `.env.sh` with your ArgoCD credentials.

## What is .env.sh?

The `.env.sh` file contains environment variables needed by the MCP server:
- `ARGOCD_BASE_URL` - Your ArgoCD server URL
- `ARGOCD_API_TOKEN` - Your ArgoCD API token

This file is:
- ✅ Automatically loaded by all `task` commands
- ✅ Gitignored for security (won't be committed)
- ✅ Used by test scripts and the MCP server

## Getting Your ArgoCD Credentials

### ArgoCD Base URL

Your ArgoCD server URL, for example:
```bash
https://argocd-server.akp-gitops.svc.cluster.local
https://argocd.example.com
```

### ArgoCD API Token

Generate an API token from ArgoCD:

**Option 1: Using ArgoCD CLI**
```bash
argocd account generate-token --account admin
```

**Option 2: Using ArgoCD UI**
1. Login to ArgoCD UI
2. Go to Settings → Accounts → your account
3. Generate a new token

**Option 3: Using kubectl**
```bash
# Port-forward to ArgoCD server
kubectl port-forward svc/argocd-server -n argocd 8080:443

# Login and generate token
argocd login localhost:8080 --username admin
argocd account generate-token
```

## Example .env.sh

```bash
#!/bin/bash
# Environment variables for MCP server

# ArgoCD Configuration
export ARGOCD_BASE_URL="https://argocd.example.com"
export ARGOCD_API_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# Optional: Set to "1" to skip TLS verification
# export ARGOCD_INSECURE="0"
```

## Verifying Your Setup

Test that your credentials work:

```bash
# Test the ArgoCD connection
task test-argo

# You should see a list of clusters
```

## Troubleshooting

**Error: "argo CD server address not provided"**
- Make sure `.env.sh` exists
- Verify it has `export ARGOCD_BASE_URL="..."`
- Check that the file is sourced: `source .env.sh && echo $ARGOCD_BASE_URL`

**Error: "argo CD auth token not provided"**
- Make sure `.env.sh` has `export ARGOCD_API_TOKEN="..."`
- Verify the token is valid by testing with ArgoCD CLI

**Connection errors**
- Verify the URL is accessible
- Check if you need to set `ARGOCD_INSECURE="1"` for self-signed certs
- Ensure you can reach the server: `curl -k $ARGOCD_BASE_URL`

## Security Notes

- ⚠️ **Never commit `.env.sh`** - It's gitignored for security
- ✅ Share `.env.sh.example` - It has no real credentials
- 🔒 Keep your API tokens secure and rotate them regularly

