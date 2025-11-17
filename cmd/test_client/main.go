package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
)

type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: test_client <tool_name> [args_json]")
		fmt.Println("Example: test_client argocd_list_clusters '{}'")
		os.Exit(1)
	}

	toolName := os.Args[1]
	argsJSON := "{}"
	if len(os.Args) > 2 {
		argsJSON = os.Args[2]
	}

	// Parse arguments
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		log.Fatalf("Invalid JSON arguments: %v", err)
	}

	fmt.Printf("==> Testing MCP Server\n")
	fmt.Printf("Tool: %s\n", toolName)
	fmt.Printf("Args: %s\n\n", argsJSON)

	// Start the MCP server
	cmd := exec.Command("go", "run", "./cmd/mcp_server")
	cmd.Env = os.Environ()

	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	// Read stderr in background
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			fmt.Fprintf(os.Stderr, "[SERVER STDERR] %s\n", scanner.Text())
		}
	}()

	// Read responses in background
	responseChan := make(chan *JSONRPCResponse, 10)
	go func() {
		decoder := json.NewDecoder(stdout)
		for {
			var resp JSONRPCResponse
			if err := decoder.Decode(&resp); err != nil {
				if err != io.EOF {
					fmt.Fprintf(os.Stderr, "[ERROR] Failed to decode response: %v\n", err)
				}
				close(responseChan)
				return
			}
			responseChan <- &resp
		}
	}()

	// Send initialize request
	initReq := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo": map[string]string{
				"name":    "test-client",
				"version": "1.0.0",
			},
		},
	}
	sendRequest(stdin, initReq)

	// Wait for initialize response
	resp := <-responseChan
	if resp != nil {
		fmt.Printf("[1] Initialize response:\n%s\n\n", prettyJSON(resp.Result))
	}

	// Send tools/list request
	listReq := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
		Params:  map[string]interface{}{},
	}
	sendRequest(stdin, listReq)

	// Wait for tools/list response
	resp = <-responseChan
	if resp != nil {
		fmt.Printf("[2] Tools list:\n%s\n\n", prettyJSON(resp.Result))
	}

	// Send tool call request
	callReq := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      3,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      toolName,
			"arguments": args,
		},
	}
	sendRequest(stdin, callReq)

	// Wait for tool call response
	resp = <-responseChan
	if resp != nil {
		if resp.Error != nil {
			fmt.Printf("[3] Tool call error: %s (code %d)\n\n", resp.Error.Message, resp.Error.Code)
		} else {
			fmt.Printf("[3] Tool call result:\n%s\n\n", prettyJSON(resp.Result))
		}
	}

	// Close stdin to signal end of input
	stdin.Close()

	// Wait for server to exit
	cmd.Wait()

	fmt.Println("==> Test complete")
}

func sendRequest(w io.Writer, req JSONRPCRequest) {
	data, err := json.Marshal(req)
	if err != nil {
		log.Fatalf("Failed to marshal request: %v", err)
	}
	fmt.Fprintf(w, "%s\n", data)
}

func prettyJSON(raw json.RawMessage) string {
	var v interface{}
	if err := json.Unmarshal(raw, &v); err != nil {
		return string(raw)
	}
	pretty, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return string(raw)
	}
	return string(pretty)
}

