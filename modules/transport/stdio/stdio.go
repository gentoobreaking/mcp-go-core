// Package stdio provides MCP over stdio transport.
package stdio

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// Handler processes a JSON-RPC message.
type Handler func(ctx context.Context, msg json.RawMessage) (any, error)

// Transport implements MCP over stdio.
type Transport struct {
	in  io.Reader
	out io.Writer
}

// New creates a new stdio transport.
func New(in io.Reader, out io.Writer) *Transport {
	if in == nil {
		in = os.Stdin
	}
	if out == nil {
		out = os.Stdout
	}
	return &Transport{in: in, out: out}
}

// Serve reads from stdin and writes responses to stdout.
func (t *Transport) Serve(ctx context.Context, handler Handler) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	scanner := bufio.NewScanner(t.in)
	var wg sync.WaitGroup
	errCh := make(chan error, 1)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		wg.Add(1)
		go func(line []byte) {
			defer wg.Done()

			// Parse the request
			var req map[string]any
			if err := json.Unmarshal(line, &req); err != nil {
				resp := map[string]any{
					"jsonrpc": "2.0",
					"error": map[string]any{
						"code":    -32700,
						"message": "Parse error",
					},
				}
				if id, ok := req["id"]; ok {
					resp["id"] = id
				}
				if data, err := json.Marshal(resp); err == nil {
					fmt.Fprintln(t.out, string(data))
				}
				return
			}

			id, _ := req["id"]
			result, err := handler(ctx, line)
			if err != nil {
				resp := map[string]any{
					"jsonrpc": "2.0",
					"id":      id,
					"error": map[string]any{
						"code":    -32000,
						"message": err.Error(),
					},
				}
				if data, err := json.Marshal(resp); err == nil {
					fmt.Fprintln(t.out, string(data))
				}
				return
			}

			resp := map[string]any{
				"jsonrpc": "2.0",
				"id":      id,
				"result":  result,
			}
			data, err := json.Marshal(resp)
			if err != nil {
				return
			}
			fmt.Fprintln(t.out, string(data))
		}(line)
	}

	// Wait for all handlers
	wg.Wait()

	select {
	case err := <-errCh:
		return err
	default:
	}

	// Check context
	<-ctx.Done()
	return nil
}

// Send writes a message to stdout.
func (t *Transport) Send(msg any) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(t.out, string(data))
	return err
}
