package protocol

import "encoding/json"

// Standard JSON-RPC 2.0 error codes.
const (
	ParsedErrorCode       = -32700 // parse_error
	InvalidRequestCode    = -32600 // invalid_request
	MethodNotFoundCode    = -32601 // method_not_found
	InvalidParamsCode     = -32602 // invalid_params
	InternalErrorCode     = -32603 // internal_error
)

// Standard error messages.
const (
	ParseErrorMessage       = "Parse error"
	InvalidRequestMessage   = "Invalid Request"
	MethodNotFoundMessage   = "Method not found"
	InvalidParamsMessage    = "Invalid params"
	InternalErrorMessage    = "Internal error"
)

// Error codes and their messages.
var ErrorMessages = map[int]string{
	ParsedErrorCode:     ParseErrorMessage,
	InvalidRequestCode:  InvalidRequestMessage,
	MethodNotFoundCode:  MethodNotFoundMessage,
	InvalidParamsCode:   InvalidParamsMessage,
	InternalErrorCode:   InternalErrorMessage,
}

// NewError creates a standard JSON-RPC error.
func NewError(code int, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// NewParseError returns a parse_error.
func NewParseError() *Error {
	return NewError(ParsedErrorCode, ParseErrorMessage)
}

// NewInvalidRequestError returns an invalid_request.
func NewInvalidRequestError() *Error {
	return NewError(InvalidRequestCode, InvalidRequestMessage)
}

// NewMethodNotFoundError returns a method_not_found.
func NewMethodNotFoundError() *Error {
	return NewError(MethodNotFoundCode, MethodNotFoundMessage)
}

// NewInvalidParamsError returns an invalid_params.
func NewInvalidParamsError() *Error {
	return NewError(InvalidParamsCode, InvalidParamsMessage)
}

// NewInternalError returns an internal_error.
func NewInternalError() *Error {
	return NewError(InternalErrorCode, InternalErrorMessage)
}

// PromptListParams lists available prompts.
type PromptListParams struct {
	// Optional hint for listing prompts.
	Hint string `json:"hint,omitempty"`
}

// ResourceListParams lists available resources.
type ResourceListParams struct {
	// Optional hint for listing resources.
	Hint string `json:"hint,omitempty"`
}

// ToolListParams lists available tools.
type ToolListParams struct {
	// Optional hint for listing tools.
	Hint string `json:"hint,omitempty"`
}

// PromptListResult lists prompts.
type PromptListResult struct {
	Prompts []Prompt `json:"prompts,omitempty"`
}

// ResourceListResult lists resources.
type ResourceListResult struct {
	Resources []Resource `json:"resources,omitempty"`
}

// ToolListResult lists tools.
type ToolListResult struct {
	Tools []Tool `json:"tools,omitempty"`
}

// Prompt is a prompt definition.
type Prompt struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Arguments   []Argument `json:"arguments,omitempty"`
}

// Resource is a resource reference.
type Resource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
}

// Tool is a tool definition.
type Tool struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	InputSchema any      `json:"inputSchema,omitempty"`
	Arguments   []Argument `json:"arguments,omitempty"`
}

// Argument is a parameter/argument definition.
type Argument struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

// ResourceTemplate defines a URI template for resources.
type ResourceTemplate struct {
	URITemplate string `json:"uriTemplate"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
}

// Implementation describes the client/server implementation.
type Implementation struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ServerCapabilities declares server-supported features.
type ServerCapabilities struct {
	Prompts      *PromptsCapability      `json:"prompts,omitempty"`
	Resources    *ResourcesCapability    `json:"resources,omitempty"`
	Tools        *ToolsCapability        `json:"tools,omitempty"`
	Logging      *LoggingCapability      `json:"logging,omitempty"`
	Completions  *CompletionsCapability  `json:"completions,omitempty"`
}

// ClientCapabilities declares client-supported features.
type ClientCapabilities struct {
	Roots      *RootsCapability      `json:"roots,omitempty"`
	Sampling   *SamplingCapability   `json:"sampling,omitempty"`
	Completions *CompletionsCapability `json:"completions,omitempty"`
}

// PromptsCapability holds prompts-related capabilities.
type PromptsCapability struct {
	ListAvailable  bool `json:"listAvailable,omitempty"`
	Get            bool `json:"get,omitempty"`
}

// ResourcesCapability holds resources-related capabilities.
type ResourcesCapability struct {
	ListAvailable    bool `json:"listAvailable,omitempty"`
	Subscribe        bool `json:"subscribe,omitempty"`
	Get              bool `json:"get,omitempty"`
}

// ToolsCapability holds tools-related capabilities.
type ToolsCapability struct {
	ListAvailable bool `json:"listAvailable,omitempty"`
	Call          bool `json:"call,omitempty"`
}

// LoggingCapability holds logging-related capabilities.
type LoggingCapability struct {
	Log bool `json:"log,omitempty"`
}

// CompletionsCapability holds completions-related capabilities.
type CompletionsCapability struct {
	Complete bool `json:"complete,omitempty"`
}

// RootsCapability holds roots-related capabilities for client.
type RootsCapability struct {
	ListAvailable bool `json:"listAvailable,omitempty"`
}

// SamplingCapability holds sampling-related capabilities for client.
type SamplingCapability struct {
	CreateMessage bool `json:"createMessage,omitempty"`
}

// InitializeRequest is sent from client to server.
type InitializeRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *int64          `json:"id"`
	Method  string          `json:"method"`
	Params  InitializeParams `json:"params"`
}

// InitializeParams contains initialization parameters.
type InitializeParams struct {
	ProtocolVersion string            `json:"protocolVersion"`
	Capabilities    ClientCapabilities  `json:"capabilities"`
	ClientInfo      Implementation       `json:"clientInfo"`
	ProcessID       *int                 `json:"processId,omitempty"`
}

// InitializeResponse is sent from server to client.
type InitializeResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *int64          `json:"id"`
	Result  InitializeResult `json:"result"`
}

// InitializeResult contains initialization response data.
type InitializeResult struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ServerCapabilities   `json:"capabilities"`
	ServerInfo      Implementation       `json:"serverInfo"`
}

// JSONRPCMessage is a union type for JSON-RPC messages.
type JSONRPCMessage struct {
	*Request
	*Response
}

// MarshalJSON serializes a JSONRPCMessage.
func (m JSONRPCMessage) MarshalJSON() ([]byte, error) {
	if m.Response != nil {
		return json.Marshal(m.Response)
	}
	return json.Marshal(m.Request)
}

// UnmarshalJSON deserializes a JSONRPCMessage.
func (m *JSONRPCMessage) UnmarshalJSON(data []byte) error {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if e, ok := raw["error"]; ok && e != nil {
		// Response
		var resp Response
		if err := json.Unmarshal(data, &resp); err != nil {
			return err
		}
		m.Response = &resp
	} else {
		// Request or Response
		var req Request
		if err := json.Unmarshal(data, &req); err == nil && req.Method != "" {
			m.Request = &req
		} else {
			var resp Response
			if err := json.Unmarshal(data, &resp); err != nil {
				return err
			}
			m.Response = &resp
		}
	}
	return nil
}

// Notification represents a JSON-RPC notification (no ID).
type Notification struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}
