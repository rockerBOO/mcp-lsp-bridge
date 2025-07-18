package mcpserver

import (
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

// LSPBridgeSession implements the ClientSession interface for MCP
type LSPBridgeSession struct {
	id            string
	notifChannel  chan mcp.JSONRPCNotification
	isInitialized bool
	createdAt     time.Time
	lastAccessed  time.Time
}

// NewLSPBridgeSession creates a new session instance
func NewLSPBridgeSession(sessionID string) *LSPBridgeSession {
	return &LSPBridgeSession{
		id:           sessionID,
		notifChannel: make(chan mcp.JSONRPCNotification, 10),
		createdAt:    time.Now(),
		lastAccessed: time.Now(),
	}
}

// SessionID returns the session identifier
func (s *LSPBridgeSession) SessionID() string {
	return s.id
}

// NotificationChannel returns the notification channel for this session
func (s *LSPBridgeSession) NotificationChannel() chan<- mcp.JSONRPCNotification {
	return s.notifChannel
}

// Initialize marks the session as initialized
func (s *LSPBridgeSession) Initialize() {
	s.isInitialized = true
	s.lastAccessed = time.Now()
}

// Initialized returns whether the session has been initialized
func (s *LSPBridgeSession) Initialized() bool {
	return s.isInitialized
}

// GetLastAccessed returns when the session was last accessed
func (s *LSPBridgeSession) GetLastAccessed() time.Time {
	return s.lastAccessed
}

// GetCreatedAt returns when the session was created
func (s *LSPBridgeSession) GetCreatedAt() time.Time {
	return s.createdAt
}
