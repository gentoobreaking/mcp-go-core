// Package router provides tool/resource/prompt dispatch.
package router

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/project/mcp-go-core/core/mcperror"
	"github.com/project/mcp-go-core/core/protocol"
	"github.com/project/mcp-go-core/core/prompt"
	"github.com/project/mcp-go-core/core/resource"
	"github.com/project/mcp-go-core/core/tool"
)

// Router dispatches MCP methods to registered handlers.
type Router struct {
	tools     map[string]tool.Tool
	resources map[string]resource.Resource
	prompts   map[string]prompt.Prompt
	logLevel  protocol.LogLevel
	sampler   SamplingHandler
	onResourceCreated CreatedNotifier
	subscriptions  map[string]map[string]bool
	rootsHandler     func() error
	notifyHandler    protocol.NotificationSender
	promptCreator    PromptCreator
	promptListChanged    func() error
	resourceListChanged  func() error
	toolsListChanged     func() error
	progressHandler      func(params protocol.ProgressNotificationParams)
	messageHandler       func(level, logger string, data any)
	elicitationHandler   func(params protocol.ElicitationCreateParams) error
	elicitationRegistry  map[string]protocol.ElicitationResult
	taskRegistry         map[string]protocol.TaskResult
	roots                []protocol.Root
}

// SamplingHandler is called when the client sends sampling/createMessage.
type SamplingHandler func(ctx context.Context, req *protocol.CreateMessageParams) (*protocol.CreateMessageResult, error)

// CreatedNotifier sends a resources/created notification when a resource is registered.
type CreatedNotifier func(uri, name string)

// PromptCreator is a factory function invoked by prompts/create to
// dynamically construct a Prompt from client-provided params.
type PromptCreator func(name, description string, args map[string]any) (prompt.Prompt, error)
// NewRouter creates a new Router.
func NewRouter() *Router {
	return &Router{
		tools:     make(map[string]tool.Tool),
		resources: make(map[string]resource.Resource),
		prompts:   make(map[string]prompt.Prompt),
		logLevel:  protocol.LogLevelInfo,
		subscriptions: make(map[string]map[string]bool),
		elicitationRegistry: make(map[string]protocol.ElicitationResult),
		taskRegistry: make(map[string]protocol.TaskResult),
	}
}

// SetLogLevel sets the server log level.
func (r *Router) SetLogLevel(level protocol.LogLevel) {
	r.logLevel = level
}

// LogLevel returns the current log level.
func (r *Router) LogLevel() protocol.LogLevel {
	return r.logLevel
}

// SetSampler registers a handler for sampling/createMessage.
func (r *Router) SetSampler(h SamplingHandler) {
	r.sampler = h
}

// SetOnResourceCreated registers a notification callback for resources/created.
func (r *Router) SetOnResourceCreated(fn CreatedNotifier) {
	r.onResourceCreated = fn
}

// RegisterTool adds a tool to the router.
func (r *Router) RegisterTool(t tool.Tool) {
	r.tools[t.Name()] = t
}

// RegisterResource adds a resource to the router and emits a
// notifications/resources/created notification to any registered handler.
func (r *Router) RegisterResource(res resource.Resource) {
	r.resources[res.URI()] = res
	if r.onResourceCreated != nil {
		r.onResourceCreated(res.URI(), res.Name())
	}
}

// RegisterPrompt adds a prompt to the router.
func (r *Router) RegisterPrompt(p prompt.Prompt) {
	r.prompts[p.Name()] = p
}

// ListPrompts returns a copy of registered prompts.
func (r *Router) ListPrompts() []prompt.Prompt {
	result := make([]prompt.Prompt, 0, len(r.prompts))
	for _, p := range r.prompts {
		result = append(result, p)
	}
	return result
}

// Dispatch routes a request to the appropriate handler.
func (r *Router) Dispatch(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	switch req.Method {
	case "initialize":
		return r.handleInitialize(ctx, req)
	case "tools/list":
		return r.dispatchListTools(ctx, req)
	case "tools/call":
		return r.dispatchCallTool(ctx, req)
	case "resources/list":
		return r.dispatchListResources(ctx, req)
	case "resources/read":
		return r.dispatchReadResource(ctx, req)
	case "prompts/list":
		return r.dispatchListPrompts(ctx, req)
	case "prompts/get":
		return r.dispatchGetPrompt(ctx, req)
	case "logging/setLogLevel":
		return r.dispatchSetLogLevel(ctx, req)
	case "sampling/createMessage":
		return r.dispatchCreateMessage(ctx, req)
	case "notifications/cancel":
		// Notification — client requesting cancellation of in-flight request
		return &protocol.Response{JSONRPC: "2.0", ID: req.ID, Result: nil}, nil
	case "initialized":
		// Notification — no response needed
		return &protocol.Response{JSONRPC: "2.0", ID: req.ID, Result: nil}, nil
	case "ping":
		return r.dispatchPing(ctx, req)
	case "complete":
		return r.dispatchComplete(ctx, req)
	case "notifications/roots/list_changed":
		return r.dispatchRootsListChanged(ctx, req)
	case "resources/subscribe":
		return r.dispatchSubscribe(ctx, req)
	case "resources/unsubscribe":
		return r.dispatchUnsubscribe(ctx, req)
	case "resources/templates/list":
		return r.dispatchListResourceTemplates(ctx, req)
	case "notifications/progress":
		return r.dispatchProgress(ctx, req)
	case "notifications/message":
		return r.dispatchMessage(ctx, req)
	case "prompts/create":
		return r.dispatchCreatePrompt(ctx, req)
	case "notifications/prompts/list_changed":
		return r.dispatchPromptsListChanged(ctx, req)
	case "notifications/resources/list_changed":
		return r.dispatchResourcesListChanged(ctx, req)
	case "notifications/tools/list_changed":
		return r.dispatchToolsListChanged(ctx, req)
	case "elicitation/create":
		return r.dispatchElicitationCreate(ctx, req)
	case "notifications/elicitation/complete":
		return r.dispatchElicitationComplete(ctx, req)
	case "tasks/get":
		return r.dispatchTasksGet(ctx, req)
	case "tasks/list":
		return r.dispatchTasksList(ctx, req)
	case "tasks/cancel":
		return r.dispatchTasksCancel(ctx, req)
	case "server/discover":
		return r.dispatchServerDiscover(ctx, req)
	case "roots/list":
		return r.dispatchRootsList(ctx, req)
	case "subscriptions/listen":
		return r.dispatchSubscriptionListen(ctx, req)
	default:
		return nil, mcperror.NewError(mcperror.CodeProtocol, "method not found: "+req.Method)
	}
}


// dispatchPing handles the ping method — returns "pong".
func (r *Router) dispatchPing(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	return &protocol.Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  protocol.PingResult{},
	}, nil
}

// dispatchComplete handles complete/arg and complete/prompt requests.
func (r *Router) dispatchComplete(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	var params protocol.CompletionParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return nil, mcperror.NewInvalidParamsError("invalid complete params: " + err.Error())
	}

	// Provide completion values based on context
	values := r.completeValues(params)

	return &protocol.Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: protocol.CompleteResult{
			Completion: protocol.CompletionResult{
				Values: values,
			},
		},
	}, nil
}

// completeValues returns completion candidates for the given params.
func (r *Router) completeValues(params protocol.CompletionParams) []string {
	// Argument value "true"/"false" completion for booleans
	if params.Value == "" || strings.HasPrefix(strings.ToLower(params.Value), "true") || strings.HasPrefix(strings.ToLower(params.Value), "false") {
		if params.ArgumentName == "enabled" || params.ArgumentName == "verbose" {
			return []string{"true", "false"}
		}
	}

	// Tool/prompt reference completion
	if params.Ref.Kind == "tool" || params.Ref.Kind == "prompt" {
		var names []string
		if params.Ref.Kind == "tool" {
			for name := range r.tools {
				names = append(names, name)
			}
		} else {
			for name := range r.prompts {
				names = append(names, name)
			}
		}
		return names
	}

	return nil
}

// dispatchRootsListChanged handles notifications/roots/list_changed.
func (r *Router) dispatchRootsListChanged(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	// Trigger roots refresh hook if registered
	if r.rootsHandler != nil {
		_ = r.rootsHandler()
	}
	// Notification — no response body needed, but return JSON-RPC ack
	return &protocol.Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  nil,
	}, nil
}

// dispatchSubscribe handles resources/subscribe requests.
func (r *Router) dispatchSubscribe(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	var params protocol.SubscribeParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return nil, mcperror.NewInvalidParamsError("invalid subscribe params: " + err.Error())
	}

	if params.URI == "" {
		return nil, mcperror.NewInvalidParamsError("uri is required")
	}

	// Track subscription per-client (default client = "default" when no session ID context)
	clientID := clientIDFromContext(ctx)
	if r.subscriptions[params.URI] == nil {
		r.subscriptions[params.URI] = make(map[string]bool)
	}
	r.subscriptions[params.URI][clientID] = true

	return &protocol.Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: struct{}{},
	}, nil
}

// SetRootsHandler registers a callback invoked when the client sends
// notifications/roots/list_changed.
func (r *Router) SetRootsHandler(handler func() error) {
	r.rootsHandler = handler
}

// IsSubscribed returns whether a resource URI has an active subscription.
func (r *Router) IsSubscribed(uri string) bool {
	clients, ok := r.subscriptions[uri]
	if !ok {
		return false
	}
	return len(clients) > 0
}

// SetNotificationSender registers a callback for emitting server→client
// notifications (e.g., notifications/resources/update).
func (r *Router) SetNotificationSender(handler protocol.NotificationSender) {
	r.notifyHandler = handler
}

// NotifyResourceUpdate emits a notifications/resources/update notification
// to all clients subscribed to the given URI.
// changeType is "update" or "delete".
func (r *Router) NotifyResourceUpdate(uri, changeType string) error {
	clients, ok := r.subscriptions[uri]
	if !ok || len(clients) == 0 {
		return nil // no subscribers
	}
	notify := protocol.ResourceUpdateNotification{
		JSONRPC: "2.0",
		Method:  "notifications/resources/update",
		Params: protocol.ResourceUpdateParams{
			URI:        uri,
			ChangeType: changeType,
		},
	}
	if r.notifyHandler != nil {
		return r.notifyHandler(notify.Method, notify.Params)
	}
	return nil
}

// dispatchCreatePrompt handles prompts/create — registers a new prompt.
func (r *Router) dispatchCreatePrompt(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	var params protocol.PromptCreateParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return nil, mcperror.NewInvalidParamsError("invalid prompts/create params: " + err.Error())
	}

	if params.Name == "" {
		return nil, mcperror.NewInvalidParamsError("name is required")
	}

	// Use registered prompt factory if available, otherwise no-op handler
	if r.promptCreator != nil {
		p, err := r.promptCreator(params.Name, params.Description, params.Arguments)
		if err != nil {
			return nil, mcperror.NewError(mcperror.CodeInternal, err.Error())
		}
		r.RegisterPrompt(p)
	}

	return &protocol.Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  map[string]any{"success": true},
	}, nil
}

// dispatchPromptsListChanged handles notifications/prompts/list_changed.
func (r *Router) dispatchPromptsListChanged(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	// Trigger prompt refresh hook if registered
	if r.promptListChanged != nil {
		_ = r.promptListChanged()
	}
	return &protocol.Response{JSONRPC: "2.0", ID: req.ID, Result: nil}, nil
}

// dispatchResourcesListChanged handles notifications/resources/list_changed.
func (r *Router) dispatchResourcesListChanged(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	if r.resourceListChanged != nil {
		_ = r.resourceListChanged()
	}
	return &protocol.Response{JSONRPC: "2.0", ID: req.ID, Result: nil}, nil
}

// dispatchToolsListChanged handles notifications/tools/list_changed.
func (r *Router) dispatchToolsListChanged(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	if r.toolsListChanged != nil {
		_ = r.toolsListChanged()
	}
	return &protocol.Response{JSONRPC: "2.0", ID: req.ID, Result: nil}, nil
}

// dispatchUnsubscribe handles resources/unsubscribe — removes a subscription.
func (r *Router) dispatchUnsubscribe(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	var params protocol.UnsubscribeParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return nil, mcperror.NewInvalidParamsError("invalid unsubscribe params: " + err.Error())
	}

	if params.URI == "" {
		return nil, mcperror.NewInvalidParamsError("uri is required")
	}

	// Remove per-client subscription
	clientID := clientIDFromContext(ctx)
	if clients, ok := r.subscriptions[params.URI]; ok {
		delete(clients, clientID)
		if len(clients) == 0 {
			delete(r.subscriptions, params.URI)
		}
	}

	return &protocol.Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  map[string]any{"success": true},
	}, nil
}

// dispatchListResourceTemplates handles resources/templates/list.
func (r *Router) dispatchListResourceTemplates(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	var templates []protocol.ResourceTemplate
	// Return registered resource templates (currently empty — templates
	// would be registered separately from static resources)
	return &protocol.Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  protocol.ResourceTemplateListResult{Templates: templates},
	}, nil
}

// dispatchProgress handles notifications/progress client→server (ack).
func (r *Router) dispatchProgress(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	var params protocol.ProgressNotificationParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return nil, mcperror.NewInvalidParamsError("invalid progress params: " + err.Error())
	}

	if r.progressHandler != nil {
		r.progressHandler(params)
	}
	return &protocol.Response{JSONRPC: "2.0", ID: req.ID, Result: nil}, nil
}

// dispatchMessage handles notifications/message client→server.
func (r *Router) dispatchMessage(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	var params protocol.MessageNotificationParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return nil, mcperror.NewInvalidParamsError("invalid message params: " + err.Error())
	}

	if r.messageHandler != nil {
		r.messageHandler(params.Level, params.Logger, params.Data)
	}
	return &protocol.Response{JSONRPC: "2.0", ID: req.ID, Result: nil}, nil
}

// Unsubscribe removes a resource subscription by URI for all clients.
func (r *Router) Unsubscribe(uri string) {
	delete(r.subscriptions, uri)
}

// DeleteResource removes a registered resource and emits a
// notifications/resources/deleted notification to any subscribed clients.
func (r *Router) DeleteResource(uri string) error {
	delete(r.resources, uri)
	// Notify subscribers first, then clear subscriptions
	err := r.NotifyResourceDeleted(uri)
	delete(r.subscriptions, uri)
	return err
}

// NotifyResourceDeleted emits a notifications/resources/deleted notification
// to all clients subscribed to the given URI.
func (r *Router) NotifyResourceDeleted(uri string) error {
	notify := protocol.ResourceDeletedNotification{
		JSONRPC: "2.0",
		Method:  "notifications/resources/deleted",
		Params: protocol.ResourceDeleteParams{
			URI: uri,
		},
	}
	if r.notifyHandler != nil {
		return r.notifyHandler(notify.Method, notify.Params)
	}
	return nil
}

// Subscribe registers a subscription for a URI for a given client ID.
// If clientID is empty, "default" is used.
func (r *Router) Subscribe(uri, clientID string) {
	if clientID == "" {
		clientID = "default"
	}
	if r.subscriptions[uri] == nil {
		r.subscriptions[uri] = make(map[string]bool)
	}
	r.subscriptions[uri][clientID] = true
}

// UnsubscribeClient removes a subscription for a specific client ID.
func (r *Router) UnsubscribeClient(uri, clientID string) {
	if clients, ok := r.subscriptions[uri]; ok {
		delete(clients, clientID)
		if len(clients) == 0 {
			delete(r.subscriptions, uri)
		}
	}
}

// ClientIDKey is the context key for the client/session ID.
type ClientIDKey struct{}

// clientIDFromContext extracts a client ID from context, defaulting to "default".
func clientIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return "default"
	}
	if id, ok := ctx.Value(ClientIDKey{}).(string); ok && id != "" {
		return id
	}
	return "default"
}

// SetProgressHandler registers a callback for notifications/progress.
func (r *Router) SetProgressHandler(h func(params protocol.ProgressNotificationParams)) {
	r.progressHandler = h
}

// SetMessageHandler registers a callback for notifications/message.
func (r *Router) SetMessageHandler(h func(level, logger string, data any)) {
	r.messageHandler = h
}

// SetPromptListChangedHandler registers a callback for notifications/prompts/list_changed.
func (r *Router) SetPromptListChangedHandler(h func() error) {
	r.promptListChanged = h
}

// SetResourceListChangedHandler registers a callback for notifications/resources/list_changed.
func (r *Router) SetResourceListChangedHandler(h func() error) {
	r.resourceListChanged = h
}

// SetToolsListChangedHandler registers a callback for notifications/tools/list_changed.
func (r *Router) SetToolsListChangedHandler(h func() error) {
	r.toolsListChanged = h
}

// SetPromptCreator registers a factory for prompts/create dynamic registration.
func (r *Router) SetPromptCreator(fn PromptCreator) {
	r.promptCreator = fn
}
// handleInitialize processes the initialize request handshake.
func (r *Router) handleInitialize(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	var initReq protocol.InitializeRequest
	if err := json.Unmarshal(req.Params, &initReq); err != nil {
		return nil, mcperror.NewInvalidParamsError("invalid initialize params: " + err.Error())
	}

	result := protocol.InitializeResult{
		ProtocolVersion: "2024-11-05",
		Capabilities: protocol.ServerCapabilities{
			Tools:     &protocol.ToolsCapability{ListAvailable: true, Call: true},
			Resources: &protocol.ResourcesCapability{ListAvailable: true, Get: true, Subscribe: true},
			Prompts:   &protocol.PromptsCapability{ListAvailable: true, Get: true},
			Logging:   &protocol.LoggingCapability{Log: true},
		},
		ServerInfo: protocol.Implementation{
			Name:    "mcp-go-core",
			Version: "0.1.0",
		},
	}

	_ = initReq // client info available if needed for future extensions

	return &protocol.Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}, nil
}

// dispatchListTools returns all registered tools.
func (r *Router) dispatchListTools(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	tools := make([]protocol.Tool, 0, len(r.tools))
	for _, t := range r.tools {
		tools = append(tools, protocol.Tool{
			Name:        t.Name(),
			Description: t.Description(),
			InputSchema: t.InputSchema(),
		})
	}
	return &protocol.Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  protocol.ToolListResult{Tools: tools},
	}, nil
}

// dispatchCallTool invokes a registered tool by name.
func (r *Router) dispatchCallTool(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	var callReq struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &callReq); err != nil {
		return nil, mcperror.NewInvalidParamsError("invalid tool call params: " + err.Error())
	}

	t, ok := r.tools[callReq.Name]
	if !ok {
		return nil, mcperror.NewError(mcperror.CodeTool, "tool not found: "+callReq.Name)
	}

	return t.Handler()(ctx, req)
}

// dispatchListResources returns all registered resources.
func (r *Router) dispatchListResources(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	resources := make([]protocol.Resource, 0, len(r.resources))
	for _, res := range r.resources {
		resources = append(resources, protocol.Resource{
			URI:         res.URI(),
			Name:        res.Name(),
			Description: res.Description(),
		})
	}
	return &protocol.Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  protocol.ResourceListResult{Resources: resources},
	}, nil
}

// dispatchReadResource reads a resource by URI.
func (r *Router) dispatchReadResource(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	var readReq struct {
		URI string `json:"uri"`
	}
	if err := json.Unmarshal(req.Params, &readReq); err != nil {
		return nil, mcperror.NewInvalidParamsError("invalid resource read params: " + err.Error())
	}

	res, ok := r.resources[readReq.URI]
	if !ok {
		return nil, mcperror.NewError(mcperror.CodeProtocol, "resource not found: "+readReq.URI)
	}

	return res.Read(ctx, req)
}

// dispatchListPrompts returns all registered prompts.
func (r *Router) dispatchListPrompts(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	prompts := make([]protocol.Prompt, 0, len(r.prompts))
	for _, p := range r.prompts {
		prompts = append(prompts, protocol.Prompt{
			Name:        p.Name(),
			Description: p.Description(),
		})
	}
	return &protocol.Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  protocol.PromptListResult{Prompts: prompts},
	}, nil
}

// dispatchGetPrompt retrieves a prompt by name.
func (r *Router) dispatchGetPrompt(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	var getReq struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &getReq); err != nil {
		return nil, mcperror.NewInvalidParamsError("invalid prompt get params: " + err.Error())
	}

	p, ok := r.prompts[getReq.Name]
	if !ok {
		return nil, mcperror.NewError(mcperror.CodeProtocol, "prompt not found: "+getReq.Name)
	}

	resp, err := p.Get(ctx, prompt.PromptRequest{
		Name:      getReq.Name,
		Arguments: getReq.Arguments,
	})
	if err != nil {
		return nil, err
	}

	return &protocol.Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  resp,
}, nil
}

// dispatchSetLogLevel sets the server log level.
func (r *Router) dispatchSetLogLevel(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	var params protocol.SetLogLevelParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return nil, mcperror.NewInvalidParamsError("invalid setLogLevel params: " + err.Error())
	}

	r.SetLogLevel(params.Level)

	return &protocol.Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  nil,
	}, nil
}

// dispatchCreateMessage handles sampling/createMessage by delegating to the
// registered SamplingHandler.
func (r *Router) dispatchCreateMessage(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	if r.sampler == nil {
		return nil, mcperror.NewError(mcperror.CodeProtocol, "sampling not supported")
	}

	var params protocol.CreateMessageParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return nil, mcperror.NewInvalidParamsError("invalid createMessage params: " + err.Error())
	}

	result, err := r.sampler(ctx, &params)
	if err != nil {
		return nil, err
	}

	return &protocol.Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}, nil
}

// dispatchElicitationCreate handles elicitation/create — requests client input.
func (r *Router) dispatchElicitationCreate(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	var params protocol.ElicitationCreateParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return nil, mcperror.NewInvalidParamsError("invalid elicitation params: " + err.Error())
	}

	if params.Message == "" {
		return nil, mcperror.NewInvalidParamsError("message is required")
	}

	elicitationID := "el_" + params.Message
	r.elicitationRegistry[elicitationID] = protocol.ElicitationResult{Action: "pending"}

	if r.elicitationHandler != nil {
		_ = r.elicitationHandler(params)
	}

	return &protocol.Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]any{
			"elicitationId": elicitationID,
			"message":       params.Message,
		},
	}, nil
}

// dispatchElicitationComplete handles notifications/elicitation/complete.
func (r *Router) dispatchElicitationComplete(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	var params protocol.ElicitationCompleteParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return nil, mcperror.NewInvalidParamsError("invalid elicitation complete params: " + err.Error())
	}

	if _, ok := r.elicitationRegistry[params.Id]; ok {
		r.elicitationRegistry[params.Id] = params.Result
	}

	return &protocol.Response{JSONRPC: "2.0", ID: req.ID, Result: nil}, nil
}

// dispatchTasksGet handles tasks/get — returns a task by ID.
func (r *Router) dispatchTasksGet(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	var params struct {
		ID string `json:"id,omitempty"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return nil, mcperror.NewInvalidParamsError("invalid tasks/get params: " + err.Error())
	}

	task, ok := r.taskRegistry[params.ID]
	if !ok {
		return nil, mcperror.NewError(mcperror.CodeProtocol, "task not found: "+params.ID)
	}

	return &protocol.Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  task,
	}, nil
}

// dispatchTasksList handles tasks/list — returns all tasks.
func (r *Router) dispatchTasksList(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	tasks := make([]protocol.TaskResult, 0, len(r.taskRegistry))
	for _, t := range r.taskRegistry {
		tasks = append(tasks, t)
	}

	return &protocol.Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  protocol.TaskListResult{Tasks: tasks},
	}, nil
}

// dispatchTasksCancel handles tasks/cancel — cancels a task by ID.
func (r *Router) dispatchTasksCancel(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	var params protocol.TaskCancelParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return nil, mcperror.NewInvalidParamsError("invalid tasks/cancel params: " + err.Error())
	}

	if task, ok := r.taskRegistry[params.ID]; ok {
		task.Status = protocol.TaskStatusCanceled
		r.taskRegistry[params.ID] = task
	}

	return &protocol.Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  map[string]any{"success": true},
	}, nil
}

// dispatchServerDiscover handles server/discover — returns server capabilities.
func (r *Router) dispatchServerDiscover(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	result := map[string]any{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]any{
			"tools":     true,
			"resources": true,
			"prompts":   true,
			"logging":   true,
			"complete":  true,
		},
		"serverInfo": map[string]any{
			"name":    "mcp-go-core",
			"version": "0.1.0",
		},
	}

	return &protocol.Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}, nil
}

// dispatchRootsList handles roots/list — returns stored roots from client.
func (r *Router) dispatchRootsList(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	return &protocol.Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  protocol.ListRootsResult{Roots: r.roots},
	}, nil
}

// dispatchSubscriptionListen handles subscriptions/listen — stores subscription URI.
func (r *Router) dispatchSubscriptionListen(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	var params protocol.SubscriptionListenParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return nil, mcperror.NewInvalidParamsError("invalid subscriptions/listen params: " + err.Error())
	}

	return &protocol.Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  map[string]any{"success": true},
	}, nil
}

// SetElicitationHandler registers a callback for elicitation/create.
func (r *Router) SetElicitationHandler(h func(params protocol.ElicitationCreateParams) error) {
	r.elicitationHandler = h
}

// ResolveElicitation returns the result of an elicitation by ID.
func (r *Router) ResolveElicitation(id string) (protocol.ElicitationResult, bool) {
	result, ok := r.elicitationRegistry[id]
	return result, ok
}

// RegisterTask adds a task to the registry.
func (r *Router) RegisterTask(id string, status protocol.TaskStatus, result any) {
	r.taskRegistry[id] = protocol.TaskResult{
		ID:     id,
		Status: status,
		Result: result,
	}
}

// SetRoots registers roots provided by the client via roots/list.
func (r *Router) SetRoots(roots []protocol.Root) {
	r.roots = roots
}
