package lsp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync"

	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
	"go.uber.org/zap"
)

// rwCloser combines a Reader and Writer into a ReadWriteCloser
type rwCloser struct {
	io.Reader
	io.Writer
}

func (rwCloser) Close() error { return nil }

// Client represents a simplified LSP client
type Client struct {
	conn       jsonrpc2.Conn
	server     protocol.Server
	cmd        *exec.Cmd
	rootPath   string
	serverName string
	
	mu           sync.Mutex
	initialized  bool
	capabilities *protocol.ServerCapabilities
	diagnostics  map[string][]protocol.Diagnostic
	
	// Callbacks
	onDiagnostics func(string, []protocol.Diagnostic)
}

// NewClient creates a new simplified LSP client
func NewClient(serverName string, rootPath string) *Client {
	return &Client{
		serverName:  serverName,
		rootPath:    rootPath,
		diagnostics: make(map[string][]protocol.Diagnostic),
	}
}

// Start starts the language server process
func (c *Client) Start(ctx context.Context, cmd string, args ...string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if c.initialized {
		return fmt.Errorf("client already started")
	}
	
	// Start the language server process
	c.cmd = exec.CommandContext(ctx, cmd, args...)
	
	// Get stdin/stdout pipes
	stdin, err := c.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdin pipe: %w", err)
	}
	
	stdout, err := c.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}
	
	// Start the process
	if err := c.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start language server: %w", err)
	}
	
	// Create a ReadWriteCloser from the pipes
	rwc := &rwCloser{
		Reader: stdout,
		Writer: stdin,
	}
	
	// Create the connection
	stream := jsonrpc2.NewStream(rwc)
	conn := jsonrpc2.NewConn(stream)
	
	// Set up a simple handler for server notifications
	handler := &simpleHandler{client: c}
	conn.Go(ctx, handler.Handle)
	
	c.conn = conn
	
	// Create a no-op logger to avoid nil pointer issues
	logger := zap.NewNop()
	c.server = protocol.ServerDispatcher(conn, logger)
	
	// Initialize the server with minimal capabilities
	if err := c.initialize(ctx); err != nil {
		c.Stop()
		return fmt.Errorf("failed to initialize: %w", err)
	}
	
	c.initialized = true
	return nil
}

// Stop stops the language server
func (c *Client) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if !c.initialized {
		return nil
	}
	
	ctx := context.Background()
	
	// Send shutdown request
	if c.server != nil {
		c.server.Shutdown(ctx)
	}
	
	// Close connection
	if c.conn != nil {
		c.conn.Close()
	}
	
	// Kill process if still running
	if c.cmd != nil && c.cmd.Process != nil {
		c.cmd.Process.Kill()
		c.cmd.Wait()
	}
	
	c.initialized = false
	return nil
}

// initialize sends a minimal initialize request
func (c *Client) initialize(ctx context.Context) error {
	rootURI := protocol.DocumentURI(uri.File(c.rootPath))
	
	// Create minimal initialization params
	params := &protocol.InitializeParams{
		RootURI: rootURI,
		ClientInfo: &protocol.ClientInfo{
			Name:    "aied",
			Version: "0.1.0",
		},
		// Minimal capabilities - let server decide what to support
		Capabilities: protocol.ClientCapabilities{},
	}
	
	result, err := c.server.Initialize(ctx, params)
	if err != nil {
		return fmt.Errorf("initialize failed: %w", err)
	}
	
	c.capabilities = &result.Capabilities
	
	// Send initialized notification
	if err := c.server.Initialized(ctx, &protocol.InitializedParams{}); err != nil {
		return fmt.Errorf("initialized notification failed: %w", err)
	}
	
	return nil
}

// OpenFile opens a file in the language server
func (c *Client) OpenFile(ctx context.Context, filename string, content string, languageID string) error {
	if !c.initialized {
		return fmt.Errorf("client not initialized")
	}
	
	fileURI := protocol.DocumentURI(uri.File(filename))
	
	params := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        fileURI,
			LanguageID: protocol.LanguageIdentifier(languageID),
			Version:    1,
			Text:       content,
		},
	}
	
	return c.server.DidOpen(ctx, params)
}

// CloseFile closes a file in the language server
func (c *Client) CloseFile(ctx context.Context, filename string) error {
	if !c.initialized {
		return fmt.Errorf("client not initialized")
	}
	
	fileURI := protocol.DocumentURI(uri.File(filename))
	
	params := &protocol.DidCloseTextDocumentParams{
		TextDocument: protocol.TextDocumentIdentifier{
			URI: fileURI,
		},
	}
	
	return c.server.DidClose(ctx, params)
}

// UpdateFile sends file changes to the language server
func (c *Client) UpdateFile(ctx context.Context, filename string, content string, version int32) error {
	if !c.initialized {
		return fmt.Errorf("client not initialized")
	}
	
	fileURI := protocol.DocumentURI(uri.File(filename))
	
	params := &protocol.DidChangeTextDocumentParams{
		TextDocument: protocol.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: protocol.TextDocumentIdentifier{
				URI: fileURI,
			},
			Version: version,
		},
		ContentChanges: []protocol.TextDocumentContentChangeEvent{
			{
				Text: content,
			},
		},
	}
	
	return c.server.DidChange(ctx, params)
}

// GetHover requests hover information
func (c *Client) GetHover(ctx context.Context, filename string, line, character uint32) (*protocol.Hover, error) {
	if !c.initialized {
		return nil, fmt.Errorf("client not initialized")
	}
	
	fileURI := protocol.DocumentURI(uri.File(filename))
	
	params := &protocol.HoverParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: fileURI,
			},
			Position: protocol.Position{
				Line:      line,
				Character: character,
			},
		},
	}
	
	return c.server.Hover(ctx, params)
}

// GetCompletion requests code completions
func (c *Client) GetCompletion(ctx context.Context, filename string, line, character uint32) ([]protocol.CompletionItem, error) {
	if !c.initialized {
		return nil, fmt.Errorf("client not initialized")
	}
	
	fileURI := protocol.DocumentURI(uri.File(filename))
	
	params := &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: fileURI,
			},
			Position: protocol.Position{
				Line:      line,
				Character: character,
			},
		},
	}
	
	result, err := c.server.Completion(ctx, params)
	if err != nil {
		return nil, err
	}
	
	// result should be *protocol.CompletionList
	if result == nil {
		return nil, nil
	}
	
	return result.Items, nil
}

// GetDefinition requests the definition location
func (c *Client) GetDefinition(ctx context.Context, filename string, line, character uint32) ([]protocol.Location, error) {
	if !c.initialized {
		return nil, fmt.Errorf("client not initialized")
	}
	
	fileURI := protocol.DocumentURI(uri.File(filename))
	
	params := &protocol.DefinitionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: fileURI,
			},
			Position: protocol.Position{
				Line:      line,
				Character: character,
			},
		},
	}
	
	result, err := c.server.Definition(ctx, params)
	if err != nil {
		return nil, err
	}
	
	// The result should be []protocol.Location
	return result, nil
}

// GetDiagnostics returns diagnostics for a file
func (c *Client) GetDiagnostics(filename string) []protocol.Diagnostic {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	return c.diagnostics[filename]
}

// SetDiagnosticsHandler sets the callback for diagnostics
func (c *Client) SetDiagnosticsHandler(handler func(string, []protocol.Diagnostic)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onDiagnostics = handler
}

// IsInitialized returns whether the client is initialized
func (c *Client) IsInitialized() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.initialized
}

// simpleHandler handles incoming server notifications
type simpleHandler struct {
	client *Client
}

func (h *simpleHandler) Handle(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	switch req.Method() {
	case protocol.MethodTextDocumentPublishDiagnostics:
		var params protocol.PublishDiagnosticsParams
		if err := json.Unmarshal(req.Params(), &params); err != nil {
			return reply(ctx, nil, err)
		}
		
		// Convert URI to filename
		fileURI := uri.URI(params.URI)
		filename := fileURI.Filename()
		
		// Store diagnostics
		h.client.mu.Lock()
		h.client.diagnostics[filename] = params.Diagnostics
		handler := h.client.onDiagnostics
		h.client.mu.Unlock()
		
		// Call handler if set
		if handler != nil {
			handler(filename, params.Diagnostics)
		}
		
		return reply(ctx, nil, nil)
		
	case protocol.MethodWindowShowMessage:
		// Ignore window messages for now
		return reply(ctx, nil, nil)
		
	case protocol.MethodWindowLogMessage:
		// Ignore log messages for now
		return reply(ctx, nil, nil)
		
	default:
		// Unknown notification, ignore
		return reply(ctx, nil, nil)
	}
}