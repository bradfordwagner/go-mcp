# Quick Start Guide

## Setup (One-Time)

Create your environment configuration:

```bash
# Single command setup
cp .env.sh.example .env.sh

# Then edit .env.sh with your ArgoCD credentials
```

The `.env.sh` file is gitignored and will be automatically loaded by all task commands.

**Need help getting credentials?** See [ENV_SETUP.md](ENV_SETUP.md) for detailed instructions.

## Test Your MCP Server in Seconds

```bash
# Test the greet tool (2-3 seconds)
task test

# Test ArgoCD list clusters
task test-argo

# Test any tool with custom arguments
task test-custom TOOL=greet ARGS='{"name":"Brad"}'
```

## Watch Mode (Best for Development)

```bash
# Auto-test on every file save
task watch

# Now edit any .go file and save
# Results appear automatically in the terminal!
```

## Test in Cursor

When ready to test in Cursor, reload the Cursor window to pick up your changes:

```
Cmd+Shift+P -> "Developer: Reload Window"
```

## Fast Feedback Loop Workflow

**Step 1:** Start watch mode in a terminal
```bash
task watch
```

**Step 2:** Edit your code in Cursor
- Make changes to any `.go` file
- Save the file
- See test results instantly in the watch terminal

**Step 3:** When ready, test in Cursor
- Reload Cursor window: `Cmd+Shift+P` -> "Developer: Reload Window"
- Test using MCP tools in Cursor

## What Just Happened?

You now have:

✅ **Instant Testing** - Test any MCP tool in 2-3 seconds with `task test`  
✅ **Auto-Testing** - Watch mode runs tests every time you save  
✅ **Full Visibility** - See the complete JSON-RPC flow (init, tools list, call)  
✅ **Debugging** - stderr output is captured and displayed  
✅ **Real Integration** - Test in Cursor with window reload

## Example Workflow

```bash
# Terminal 1: Start watch mode
task watch

# Cursor: Edit internal/tools/argo/list_clusters.go
# Save the file
# Watch terminal shows test results automatically

# When tests pass, test in Cursor
# Reload: Cmd+Shift+P -> "Developer: Reload Window"
# Then use the MCP tools in Cursor
```

## Tips

- Keep watch mode running while developing
- Test with `task test` for one-off checks
- Use `task test-custom` to test specific tools with specific args
- Check `DEVELOPMENT.md` for detailed documentation

