// Package resource provides the Resource interface for MCP resources.
package resource

import (
	"context"

	"github.com/project/mcp-go-core/core/protocol"
)

// Resource is the interface for an MCP resource.
type Resource interface {
	URI() string
	Name() string
	Description() string
	Read(ctx context.Context, req *protocol.Request) (*protocol.Response, error)
}

// BaseResource provides a partial implementation of Resource.
type BaseResource struct {
	uri         string
	name        string
	description string
	readFunc    func(ctx context.Context, req *protocol.Request) (*protocol.Response, error)
}

// NewResource creates a new Resource.
func NewResource(uri, name, description string, readFn func(ctx context.Context, req *protocol.Request) (*protocol.Response, error)) Resource {
	return &BaseResource{
		uri:         uri,
		name:        name,
		description: description,
		readFunc:    readFn,
	}
}

func (r *BaseResource) URI() string         { return r.uri }
func (r *BaseResource) Name() string        { return r.name }
func (r *BaseResource) Description() string { return r.description }
func (r *BaseResource) Read(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	return r.readFunc(ctx, req)
}
