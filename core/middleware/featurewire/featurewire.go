// Package middleware/featurewire provides feature-flag gating middleware
// for the MCP router dispatch pipeline.
package featurewire

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/project/mcp-go-core/core/feature"
	"github.com/project/mcp-go-core/core/middleware"
	"github.com/project/mcp-go-core/core/mcperror"
)

// FlagMapper maps an MCP method name to a feature flag name.
// e.g. "tools/call:advanced" → "advanced_tools"
type FlagMapper func(method string) string

func DefaultFlagMapper(method string) string {
	const (
		toolsPrefix    = "tools/"
		resourcesPrefix = "resources/"
		promptsPrefix  = "prompts/"
	)
	switch {
	case strings.HasPrefix(method, toolsPrefix):
		return method[len(toolsPrefix):]
	case strings.HasPrefix(method, resourcesPrefix):
		return method[len(resourcesPrefix):]
	case strings.HasPrefix(method, promptsPrefix):
		return method[len(promptsPrefix):]
	}
	return ""
}

// Middleware returns a middleware that gates requests by feature flag.
// If the flag for the method is disabled, the request is rejected with a
// JSON-RPC error before reaching the handler.
func Middleware(flags *feature.Flags, mapper FlagMapper) middleware.Middleware {
	return func(next middleware.Handler) middleware.Handler {
		return middleware.HandlerFunc(func(method string, params []byte) ([]byte, error) {
			flagName := mapper(method)
			if flagName != "" && flags.IsDisabled(flagName) {
				return nil, mcperror.NewError(mcperror.CodeValidation,
					fmt.Sprintf("feature flag '%s' is disabled for method '%s'", flagName, method))
			}
			return next.HandleRequest(method, params)
		})
	}
}

// HealthStatus returns current flag statuses for health endpoint reporting.
func HealthStatus(flags *feature.Flags) map[string]bool {
	snap := flags.Snapshot()
	status := make(map[string]bool, len(snap))
	for name, flag := range snap {
		status[name] = flag.Enabled
	}
	return status
}

// FlagStatus is a JSON-serializable flag status for health endpoints.
type FlagStatus struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

// MarshalFlagStatus returns JSON bytes of flag status for a given method.
func MarshalFlagStatus(flags *feature.Flags, method, flagName string) ([]byte, error) {
	flag := flags.Get(flagName)
	return json.Marshal(FlagStatus{
		Name:    flagName,
		Enabled: flag.Enabled,
	})
}

// ParseFlagParams extracts method name from JSON-RPC request params for flag lookup.
func ParseFlagParams(params []byte) (method string) {
	if len(params) == 0 {
		return ""
	}
	var raw struct {
		Method string `json:"method"`
	}
	_ = json.Unmarshal(params, &raw)
	return raw.Method
}
