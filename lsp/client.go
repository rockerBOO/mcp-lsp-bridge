package lsp

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync/atomic"
	"time"

	"rockerboo/mcp-lsp-bridge/logger"

	"github.com/myleshyson/lsprotocol-go/protocol"
	"github.com/sourcegraph/jsonrpc2"
)

// NewLanguageClient creates a new Language Server Protocol client
func NewLanguageClient(command string, args ...string) (*LanguageClient, error) {

	err := sanitizeArgs(args)
	if err != nil {
		return nil, err
	}

	// Create a background context that will be replaced when connecting
	ctx, cancel := context.WithCancel(context.Background())

	client := LanguageClient{
		command:            command,
		args:               args,
		clientCapabilities: protocol.ClientCapabilities{},
		serverCapabilities: protocol.ServerCapabilities{},
		ctx:                ctx,
		cancel:             cancel,

		// Default configuration
		maxConnectionAttempts: 3,
		connectionTimeout:     10 * time.Second,
		idleTimeout:           30 * time.Minute,
		restartDelay:          1 * time.Second,

		status: StatusConnecting,
	}

	return &client, nil
}

func (cs ClientStatus) String() string {
	switch cs {
	case StatusUninitialized:
		return "uninitialized"
	case StatusConnecting:
		return "connecting"
	case StatusConnected:
		return "connected"
	case StatusError:
		return "error"
	case StatusRestarting:
		return "restarting"
	case StatusDisconnected:
		return "disconnected"
	default:
		return "unknown"
	}
}

func sanitizeArgs(args []string) error {
	for _, arg := range args {
		// Block shell metacharacters that could be dangerous if LSP server processes them
		if strings.ContainsAny(arg, ";|&$`") {
			return fmt.Errorf("dangerous character in argument: %s", arg)
		}
		// Block command substitution patterns
		if strings.Contains(arg, "$(") || strings.Contains(arg, "`") {
			return fmt.Errorf("command substitution not allowed: %s", arg)
		}
	}
	return nil
}

func (lc *LanguageClient) Connect() (LanguageClientInterface, error) {
	// Log the LSP server connection attempt
	logger.Info(fmt.Sprintf("Connecting to LSP server: %s %v", lc.command, lc.args))

	var conn LSPConnectionInterface

	// Cancel the existing context and create a new one for the connection session
	if lc.cancel != nil {
		lc.cancel()
	}
	ctx, cancel := context.WithCancel(context.Background())

	// Start the external process
	cmd := exec.CommandContext(ctx, lc.command, lc.args...) // #nosec G204

	stdin, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		closeErr := stdin.Close()

		cancel()

		if closeErr != nil {
			return nil, fmt.Errorf("failed to close stdin pipe: %w", closeErr)
		}

		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		stdCloseErr := stdin.Close()
		if stdCloseErr != nil {
			cancel()
			return nil, fmt.Errorf("failed to close stdin pipe: %w", stdCloseErr)
		}

		stdOutClose := stdout.Close()
		if stdOutClose != nil {
			cancel()
			return nil, fmt.Errorf("failed to close stdout pipe: %w", stdOutClose)
		}

		cancel()

		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		// Close all pipes and cancel context on startup failure
		var closeErrors []error
		if stdinErr := stdin.Close(); stdinErr != nil {
			closeErrors = append(closeErrors, fmt.Errorf("failed to close stdin: %w", stdinErr))
		}
		if stdoutErr := stdout.Close(); stdoutErr != nil {
			closeErrors = append(closeErrors, fmt.Errorf("failed to close stdout: %w", stdoutErr))
		}
		if stderrErr := stderr.Close(); stderrErr != nil {
			closeErrors = append(closeErrors, fmt.Errorf("failed to close stderr: %w", stderrErr))
		}
		cancel()
		
		if len(closeErrors) > 0 {
			return nil, fmt.Errorf("failed to start command: %w, cleanup errors: %v", err, closeErrors)
		}
		return nil, fmt.Errorf("failed to start command: %w", err)
	}

	// Store the command for cleanup
	lc.cmd = cmd

	// Create a ReadWriteCloser that combines stdin and stdout for LSP
	readWriteCloser := &stdioReadWriteCloser{
		stdin:  stdin,
		stdout: stdout,
	}

	// Create handler
	handler := &ClientHandler{client: lc}

	// Create JSON-RPC connection using VSCode Object Codec for LSP headers
	stream := jsonrpc2.NewBufferedStream(readWriteCloser, jsonrpc2.VSCodeObjectCodec{})
	conn = jsonrpc2.NewConn(ctx, stream, handler)

	lc.conn = conn
	lc.status = StatusConnected
	lc.lastInitialized = time.Now()
	lc.ctx = ctx
	lc.cancel = cancel

	// Log successful LSP server connection
	logger.Info(fmt.Sprintf("Successfully connected to LSP server: %s %v", lc.command, lc.args))

	// Handle stderr in background
	go func() {
		buf := make([]byte, 1024)

		for {
			n, err := stderr.Read(buf)
			if err != nil {
				break
			}

			if n > 0 {
				logger.Debug(fmt.Sprintf("[%s SERVER STDERR]: %s", lc.command, buf[:n]))
			}
		}

		stderrCloseErr := stderr.Close()
		if stderrCloseErr != nil {
			logger.Warn(fmt.Errorf("failed to close stderr pipe: %w", stderrCloseErr))
		}
	}()

	return lc, nil
}

func (ls *LanguageClient) IsConnected() bool {
	return ls.status == StatusConnected
}

func (ls *LanguageClient) Status() ClientStatus {
	return ls.status
}

func (lc *LanguageClient) ProjectRoots() []string {
	return lc.workspacePaths
}
func (lc *LanguageClient) SetProjectRoots(paths []string) {
	lc.workspacePaths = paths
}

// Close closes the language client and cleans up resources
func (lc *LanguageClient) Close() error {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	var errors []error

	// Cancel context first to signal shutdown
	if lc.cancel != nil {
		lc.cancel()
	}

	// Close JSON-RPC connection (this will close the stdin/stdout pipes)
	if lc.conn != nil {
		err := lc.conn.Close()
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to close JSON-RPC connection: %w", err))
		}
		lc.conn = nil
	}

	// Attempt graceful shutdown of language server process
	if lc.cmd != nil && lc.cmd.Process != nil {
		// Wait for process to exit or kill it
		done := make(chan error, 1)
		go func() {
			done <- lc.cmd.Wait()
		}()

		select {
		case waitErr := <-done:
			// Process exited, check if there was an error
			if waitErr != nil {
				errors = append(errors, fmt.Errorf("process wait failed: %w", waitErr))
			}
		case <-time.After(2 * time.Second):
			// Force kill if it doesn't exit
			if lc.cmd.Process != nil {
				if killErr := lc.cmd.Process.Kill(); killErr != nil {
					errors = append(errors, fmt.Errorf("failed to kill process: %w", killErr))
				}
				// Wait for it to actually exit after kill
				waitErr := <-done
				if waitErr != nil {
					errors = append(errors, fmt.Errorf("process wait after kill failed: %w", waitErr))
				}
			}
		}
		lc.cmd = nil
	}

	// Reset connection state
	lc.status = StatusUninitialized
	lc.lastError = nil

	// Return combined errors if any
	if len(errors) > 0 {
		return fmt.Errorf("cleanup errors: %v", errors)
	}

	return nil
}

// ClientCapabilities returns the client's capabilities
func (lc *LanguageClient) ClientCapabilities() protocol.ClientCapabilities {
	return lc.clientCapabilities
}

// ServerCapabilities returns the server's capabilities
func (lc *LanguageClient) ServerCapabilities() protocol.ServerCapabilities {
	return lc.serverCapabilities
}

// SetServerCapabilities sets the server's capabilities
func (lc *LanguageClient) SetServerCapabilities(capabilities protocol.ServerCapabilities) {
	lc.serverCapabilities = capabilities
}

// SendRequest sends a request with timeout
func (lc *LanguageClient) SendRequest(method string, params any, result any, timeout time.Duration) error {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	// Increment total requests
	atomic.AddInt64(&lc.totalRequests, 1)

	// Ensure connection is still valid by checking context and connection
	if lc.ctx == nil || lc.ctx.Err() != nil || lc.conn == nil {
		return errors.New("language server connection is closed")
	}

	// Reset status to connected if we have a valid connection
	if lc.status == StatusError && lc.ctx != nil && lc.ctx.Err() == nil && lc.conn != nil {
		lc.status = StatusConnected

		logger.Info("LSP client status reset from error to connected")
	}

	reqCtx, cancel := context.WithTimeout(lc.ctx, timeout)
	defer cancel()

	logger.Debug(fmt.Sprintf("LSP Request: method=%s params=%v", method, params))

	err := lc.conn.Call(reqCtx, method, params, result)
	if err != nil {
		// Increment failed requests
		atomic.AddInt64(&lc.failedRequests, 1)
		lc.lastErrorTime = time.Now()
		lc.lastError = err
		lc.status = StatusError

		// Log the error
		logger.Error(fmt.Sprintf("LSP Request Error: method=%s, error=%v", method, err))
	} else {
		// Increment successful requests
		atomic.AddInt64(&lc.successfulRequests, 1)
	}

	return err
}

// SendRequestNoTimeout sends a request without timeout
func (lc *LanguageClient) SendRequestNoTimeout(method string, params any, result any) error {
	return lc.conn.Call(lc.ctx, method, params, result)
}

// SendNotification sends a notification
func (lc *LanguageClient) SendNotification(method string, params any) error {
	if len(method) == 0 {
		return errors.New("empty notification method")
	}

	return lc.conn.Notify(lc.ctx, method, params)
}

// Context returns the client's context
func (lc *LanguageClient) Context() context.Context {
	return lc.ctx
}

// GetMetrics returns the current metrics for the language client
func (lc *LanguageClient) GetMetrics() ClientMetrics {
	lc.mu.RLock()
	defer lc.mu.RUnlock()

	return ClientMetrics{
		Command:            lc.command,
		Status:             lc.status,
		TotalRequests:      atomic.LoadInt64(&lc.totalRequests),
		SuccessfulRequests: atomic.LoadInt64(&lc.successfulRequests),
		FailedRequests:     atomic.LoadInt64(&lc.failedRequests),
		LastInitialized:    lc.lastInitialized,
		LastErrorTime:      lc.lastErrorTime,
		LastError:          fmt.Sprintf("%v", lc.lastError),
		IsConnected:        lc.IsConnected(),
		ProcessID:          lc.processID,
	}
}

func (lc *LanguageClient) SetupSemanticTokens() error {
	tokenTypes, tokenModifiers, err := GetTokenTypeFromServerCapabilities(&lc.serverCapabilities)
	if err != nil {
		logger.Warn(fmt.Sprintf("SetupSemanticTokens: Failed to get token types from server capabilities: %v", err))
		return err
	}

	logger.Debug(fmt.Sprintf("SetupSemanticTokens: Token Types: %+v, Token Modifiers: %+v", tokenTypes, tokenModifiers))

	lc.tokenParser = NewSemanticTokenParser(tokenTypes, tokenModifiers)

	return nil
}

func (lc *LanguageClient) TokenParser() *SemanticTokenParser {
	return lc.tokenParser
}
