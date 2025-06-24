package mocks

import (
	"context"
	"rockerboo/mcp-lsp-bridge/lsp"
	"time"

	"github.com/myleshyson/lsprotocol-go/protocol"
	"github.com/stretchr/testify/mock"
)

// MockLanguageClient implements LanguageClientInterface for testing
type MockLanguageClient struct {
	mock.Mock
}

// Core client methods
func (m *MockLanguageClient) Connect() (lsp.LanguageClientInterface, error) {
	args := m.Called()
	return args.Get(0).(lsp.LanguageClientInterface), args.Error(1)
}

func (m *MockLanguageClient) SendRequest(method string, params any, result any, timeout time.Duration) error {
	args := m.Called(method, params, result, timeout)
	return args.Error(0)
}

func (m *MockLanguageClient) SendNotification(method string, params any) error {
	args := m.Called(method, params)
	return args.Error(0)
}

func (m *MockLanguageClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockLanguageClient) Context() context.Context {
	args := m.Called()
	return args.Get(0).(context.Context)
}

func (m *MockLanguageClient) GetMetrics() lsp.ClientMetrics {
	args := m.Called()
	return args.Get(0).(lsp.ClientMetrics)
}

func (m *MockLanguageClient) IsConnected() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockLanguageClient) Status() lsp.ClientStatus {
	args := m.Called()
	return args.Get(0).(lsp.ClientStatus)
}

// Lifecycle methods
func (m *MockLanguageClient) Initialize(params protocol.InitializeParams) (*protocol.InitializeResult, error) {
	args := m.Called(params)
	return args.Get(0).(*protocol.InitializeResult), args.Error(1)
}

func (m *MockLanguageClient) Initialized() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockLanguageClient) Shutdown() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockLanguageClient) Exit() error {
	args := m.Called()
	return args.Error(0)
}

// Capabilities
func (m *MockLanguageClient) ClientCapabilities() protocol.ClientCapabilities {
	args := m.Called()
	return args.Get(0).(protocol.ClientCapabilities)
}

func (m *MockLanguageClient) ServerCapabilities() protocol.ServerCapabilities {
	args := m.Called()
	return args.Get(0).(protocol.ServerCapabilities)
}

func (m *MockLanguageClient) SetServerCapabilities(capabilities protocol.ServerCapabilities) {
	m.Called(capabilities)
}

// Text document synchronization
func (m *MockLanguageClient) DidOpen(uri string, languageId protocol.LanguageKind, text string, version int32) error {
	args := m.Called(uri, languageId, text, version)
	return args.Error(0)
}

func (m *MockLanguageClient) DidChange(uri string, version int32, changes []protocol.TextDocumentContentChangeEvent) error {
	args := m.Called(uri, version, changes)
	return args.Error(0)
}

func (m *MockLanguageClient) DidSave(uri string, text *string) error {
	args := m.Called(uri, text)
	return args.Error(0)
}

func (m *MockLanguageClient) DidClose(uri string) error {
	args := m.Called(uri)
	return args.Error(0)
}

// Language features
func (m *MockLanguageClient) WorkspaceSymbols(query string) ([]protocol.WorkspaceSymbol, error) {
	args := m.Called(query)
	return args.Get(0).([]protocol.WorkspaceSymbol), args.Error(1)
}

func (m *MockLanguageClient) Definition(uri string, line, character int32) ([]protocol.Or2[protocol.LocationLink, protocol.Location], error) {
	args := m.Called(uri, line, character)
	return args.Get(0).([]protocol.Or2[protocol.LocationLink, protocol.Location]), args.Error(1)
}

func (m *MockLanguageClient) References(uri string, line, character int32, includeDeclaration bool) ([]protocol.Location, error) {
	args := m.Called(uri, line, character, includeDeclaration)
	return args.Get(0).([]protocol.Location), args.Error(1)
}

func (m *MockLanguageClient) Hover(uri string, line, character int32) (*protocol.Hover, error) {
	args := m.Called(uri, line, character)
	return args.Get(0).(*protocol.Hover), args.Error(1)
}

func (m *MockLanguageClient) DocumentSymbols(uri string) ([]protocol.DocumentSymbol, error) {
	args := m.Called(uri)
	return args.Get(0).([]protocol.DocumentSymbol), args.Error(1)
}

func (m *MockLanguageClient) Implementation(uri string, line, character int32) ([]protocol.Location, error) {
	args := m.Called(uri, line, character)
	return args.Get(0).([]protocol.Location), args.Error(1)
}

func (m *MockLanguageClient) SignatureHelp(uri string, line, character int32) (*protocol.SignatureHelp, error) {
	args := m.Called(uri, line, character)
	return args.Get(0).(*protocol.SignatureHelp), args.Error(1)
}
