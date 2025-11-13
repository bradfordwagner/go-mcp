# Development Guide

## Fast Feedback Loop

This project includes several tools to speed up your development workflow when building MCP server tools.

### First Time Setup

Create your environment configuration:

```bash
# Copy the example file and edit with your credentials
cp .env.sh.example .env.sh
# Edit .env.sh to add your ARGOCD_BASE_URL and ARGOCD_API_TOKEN
```

The `.env.sh` file is automatically loaded by all task commands and is gitignored for security.

### Quick Start

```bash
# Test your changes immediately (fastest)
task test

# Test ArgoCD functionality
task test-argo

# Test any tool with custom arguments
task test-custom TOOL=greet ARGS='{"name":"Brad"}'

# Watch mode - auto-test on file changes
task watch
```

### Test Client

The project includes a custom Go test client (`cmd/test_client`) that provides instant feedback without needing to restart Cursor. This is the fastest way to test your changes during development.

### Development Workflow Options

#### Option 1: Quick Test (Fastest)
Best for: Testing individual tool changes without restarting Cursor

```bash
# Test the greet tool
task test
# or directly: go run ./cmd/test_client greet '{"name":"Developer"}'

# Test ArgoCD list clusters
task test-argo
# or directly: go run ./cmd/test_client argocd_list_clusters '{}'

# Test any tool with custom arguments
task test-custom TOOL=your_tool_name ARGS='{"param":"value"}'

# Test directly with the Go client
go run ./cmd/test_client greet '{"name":"Brad"}'
```

**Pros:** Instant feedback (2-3 seconds), no need to restart Cursor, see full JSON-RPC flow, colored output
**Cons:** Not testing the actual Cursor integration

#### Option 2: Watch Mode (Recommended for TDD)
Best for: Continuous development with auto-testing

```bash
task watch
```

This will:
- Watch all `.go` files (except test_client)
- Run tests automatically on any change
- Clear screen and show results in terminal
- Uses `watchexec` for efficient file watching

**Pros:** Continuous feedback, no manual intervention, see results immediately after saving
**Cons:** Requires `watchexec` installed (`brew install watchexec`)

**Tip:** Keep this running in a terminal while you code. Every time you save a file, you'll see the test results instantly.

#### Option 3: Test in Cursor
Best for: Testing the actual integration in Cursor

After your changes pass local tests, reload Cursor to test the real integration:

**Reload Cursor window:**
- Press `Cmd+Shift+P`
- Type: "Developer: Reload Window"
- Cursor will restart the MCP server with your changes
- Test using Cursor's MCP tools

**Pros:** Tests real integration, sees actual Cursor behavior
**Cons:** Slower than option 1, requires manual reload

### Adding New Tools

1. Create your tool handler in `internal/tools/`
2. Register it in `cmd/mcp_server/main.go`
3. Test it immediately:
   ```bash
   go run ./cmd/test_client your_tool_name '{"param":"value"}'
   # or use the task command
   task test-custom TOOL=your_tool_name ARGS='{"param":"value"}'
   ```
4. Iterate quickly in watch mode:
   ```bash
   task watch
   # Now edit your code and see results instantly
   ```
5. When ready, test in Cursor:
   - Reload Cursor: `Cmd+Shift+P` -> "Developer: Reload Window"

### Environment Variables

All environment variables are configured in `.env.sh` which is automatically loaded by task commands.

**Setup:**
```bash
# One-time setup
cp .env.sh.example .env.sh

# Edit .env.sh with your values:
# - ARGOCD_BASE_URL
# - ARGOCD_API_TOKEN
```

**Override for specific tests:**
```bash
# Override environment variables for a single run
ARGOCD_BASE_URL="https://other-server" task test-argo
```

**Note:** The `.env.sh` file is gitignored to keep your credentials secure.

### Debugging

#### Enable verbose logging in your tool
The server runs over stdio, so use stderr for debugging:

```go
fmt.Fprintf(os.Stderr, "Debug: %+v\n", someData)
```

The test client will show stderr output with a `[SERVER STDERR]` prefix.

#### View full JSON-RPC flow
The test client shows the complete JSON-RPC flow:
1. Initialize request/response
2. Tools list request/response
3. Tool call request/response

```bash
go run ./cmd/test_client greet '{"name":"Debug"}'
```

#### Inspect server output directly
Run the server directly and send manual JSON-RPC commands:

```bash
go run ./cmd/mcp_server
# Then paste JSON-RPC commands manually
```

### Tips & Best Practices

1. **Development Workflow**
   ```bash
   # Start watch mode in one terminal
   task watch
   
   # Edit code in your editor
   # Save file → see results instantly
   
   # When ready to test in Cursor
   # Cmd+Shift+P -> "Developer: Reload Window"
   ```

2. **Use task commands** - They're shorter and easier to remember
   ```bash
   task test                                    # Quick test with greet
   task test-argo                               # Test ArgoCD functionality
   task test-custom TOOL=x ARGS='{"y":"z"}'    # Test any tool
   task watch                                   # Continuous development
   ```

3. **Fast iteration cycle** - The test client runs in 2-3 seconds, much faster than reloading Cursor

4. **Check both environments** - Test with client first (fast), then verify in Cursor (real)

5. **Use watch mode for TDD** - Write code, save, see results automatically

6. **Debug with stderr** - Any stderr output will be shown with `[SERVER STDERR]` prefix

### Common Issues

**"command not found: watchexec"** (optional, for watch mode)
```bash
brew install watchexec
```
Watch mode will work without it, but won't auto-reload on file changes.

**Changes not appearing in Cursor**
- Reload the Cursor window: `Cmd+Shift+P` -> "Developer: Reload Window"
- Check that your `~/.cursor/mcp.json` points to the correct path
- Make sure no compile errors exist: `go build ./cmd/mcp_server`

**Changes not appearing**
- Make sure you saved the file
- Check for compilation errors: `go build ./cmd/mcp_server`
- Try running `task test` to see if the server compiles and runs

**Test client hangs**
- Check that your tool handler doesn't block indefinitely
- Verify environment variables are set (ARGOCD_BASE_URL, ARGOCD_API_TOKEN)
- Look for error messages in the `[SERVER STDERR]` output

**ArgoCD connection issues**
```bash
# Check environment variables
echo $ARGOCD_BASE_URL
echo $ARGOCD_API_TOKEN

# Test with explicit values
ARGOCD_BASE_URL=https://your-server task test-argo
```

