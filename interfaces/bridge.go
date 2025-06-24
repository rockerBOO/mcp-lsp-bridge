package interfaces

import (
	"rockerboo/mcp-lsp-bridge/lsp"

	"github.com/myleshyson/lsprotocol-go/protocol"
)

// BridgeInterface defines the interface that the bridge must implement
type BridgeInterface interface {
	GetClientForLanguageInterface(language string) (any, error)
	InferLanguage(filePath string) (string, error)
	CloseAllClients()
	GetConfig() *lsp.LSPServerConfig
	DetectProjectLanguages(projectPath string) ([]string, error)
	DetectPrimaryProjectLanguage(projectPath string) (string, error)
	// Enhanced project analysis methods
	FindSymbolReferences(language, uri string, line, character uint32, includeDeclaration bool) ([]protocol.Location, error)
	FindSymbolDefinitions(language, uri string, line, character uint32) ([]protocol.Or2[protocol.LocationLink, protocol.Location], error)
	SearchTextInWorkspace(language, query string) ([]protocol.WorkspaceSymbol, error)
	GetMultiLanguageClients(languages []string) (map[string]lsp.LanguageClientInterface, error)
	// Core information tools
	GetHoverInformation(uri string, line, character uint32) (*protocol.Hover, error)
	GetDiagnostics(uri string) ([]any, error)
	GetWorkspaceDiagnostics(workspaceUri string, identifier string) ([]protocol.WorkspaceDiagnosticReport, error)
	GetSignatureHelp(uri string, line, character uint32) (*protocol.SignatureHelp, error)
	// Code actions and formatting tools
	GetCodeActions(uri string, line, character, endLine, endCharacter uint32) ([]protocol.CodeAction, error)
	FormatDocument(uri string, tabSize uint32, insertSpaces bool) ([]protocol.TextEdit, error)
	ApplyTextEdits(uri string, edits []protocol.TextEdit) error
	// Advanced navigation tools
	RenameSymbol(uri string, line, character uint32, newName string, preview bool) (*protocol.WorkspaceEdit, error)
	ApplyWorkspaceEdit(edit *protocol.WorkspaceEdit) error
	FindImplementations(uri string, line, character uint32) ([]protocol.Location, error)
	PrepareCallHierarchy(uri string, line, character uint32) ([]protocol.CallHierarchyItem, error)
	GetIncomingCalls(item protocol.CallHierarchyItem) ([]protocol.CallHierarchyIncomingCall, error)
	GetOutgoingCalls(item protocol.CallHierarchyItem) ([]protocol.CallHierarchyOutgoingCall, error)

	// Document symbol operations
	GetDocumentSymbols(uri string) ([]protocol.DocumentSymbol, error)
}
