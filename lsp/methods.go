package lsp

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/myleshyson/lsprotocol-go/protocol"
)

// LSP Protocol Method Implementations

// Initialize sends an initialize request to the language server
func (lc *LanguageClient) Initialize(params protocol.InitializeParams) (*protocol.InitializeResult, error) {
	var result protocol.InitializeResult
	err := lc.SendRequest("initialize", params, &result, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Initialized sends the initialized notification
func (lc *LanguageClient) Initialized() error {
	return lc.SendNotification("initialized", protocol.InitializedParams{})
}

// Shutdown sends a shutdown request
func (lc *LanguageClient) Shutdown() error {
	var result protocol.ShutdownResponse
	return lc.SendRequest("shutdown", nil, &result, 5*time.Second)
}

// Exit sends an exit notification
func (lc *LanguageClient) Exit() error {
	return lc.SendNotification("exit", nil)
}

// DidOpen sends a textDocument/didOpen notification
func (lc *LanguageClient) DidOpen(uri string, languageId protocol.LanguageKind, text string, version int32) error {
	params := protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			Uri:        protocol.DocumentUri(uri),
			LanguageId: languageId,
			Version:    version,
			Text:       text,
		},
	}
	return lc.SendNotification("textDocument/didOpen", params)
}

// DidChange sends a textDocument/didChange notification
func (lc *LanguageClient) DidChange(uri string, version int32, changes []protocol.TextDocumentContentChangeEvent) error {
	params := protocol.DidChangeTextDocumentParams{
		TextDocument: protocol.VersionedTextDocumentIdentifier{
			Uri:     protocol.DocumentUri(uri),
			Version: version,
		},
		ContentChanges: changes,
	}
	return lc.SendNotification("textDocument/didChange", params)
}

// DidSave sends a textDocument/didSave notification
func (lc *LanguageClient) DidSave(uri string, text *string) error {
	params := map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri": uri,
		},
	}
	if text != nil {
		params["text"] = *text
	}
	return lc.SendNotification("textDocument/didSave", params)
}

// DidClose sends a textDocument/didClose notification
func (lc *LanguageClient) DidClose(uri string) error {
	params := map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri": uri,
		},
	}
	return lc.SendNotification("textDocument/didClose", params)
}

func (lc *LanguageClient) WorkspaceSymbols(query string) ([]protocol.SymbolInformation, error) {
	var result []protocol.SymbolInformation
	err := lc.SendRequest("workspace/symbol", protocol.WorkspaceSymbolParams{
		Query: query,
	}, &result, 5*time.Second)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Definition requests definition locations for a symbol at a given position
// Returns LocationLink[] or converts Location[] to LocationLink[]
func (lc *LanguageClient) Definition(uri string, line, character int32) ([]protocol.LocationLink, error) {
	// Use raw JSON response to handle both Location[] and LocationLink[] formats
	var rawResult json.RawMessage
	err := lc.SendRequest("textDocument/definition", protocol.DefinitionParams{
		TextDocument: protocol.TextDocumentIdentifier{
			Uri: protocol.DocumentUri(uri),
		},
		Position: protocol.Position{
			Line:      uint32(line),
			Character: uint32(character),
		},
	}, &rawResult, 5*time.Second)
	if err != nil {
		return nil, err
	}

	// First try to unmarshal as LocationLink[]
	var resultLinks []protocol.LocationLink
	if err := json.Unmarshal(rawResult, &resultLinks); err == nil {
		// Check if we got valid LocationLink data
		if len(resultLinks) > 0 && resultLinks[0].TargetUri != "" {
			return resultLinks, nil
		}
	}

	// Try to unmarshal as Location[]
	var resultLocations []protocol.Location
	if err := json.Unmarshal(rawResult, &resultLocations); err != nil {
		return nil, fmt.Errorf("failed to unmarshal definition response: %w", err)
	}

	// Convert Location[] to LocationLink[]
	links := make([]protocol.LocationLink, len(resultLocations))
	for i, loc := range resultLocations {
		links[i] = protocol.LocationLink{
			TargetUri:         loc.Uri,
			TargetRange:       loc.Range,
			TargetSelectionRange: loc.Range,
		}
	}
	return links, nil
}

// References finds all references to a symbol at a given position
func (lc *LanguageClient) References(uri string, line, character int32, includeDeclaration bool) ([]protocol.Location, error) {
	var result []protocol.Location
	err := lc.SendRequest("textDocument/references", protocol.ReferenceParams{
		TextDocument: protocol.TextDocumentIdentifier{
			Uri: protocol.DocumentUri(uri),
		},
		Position: protocol.Position{
			Line:      uint32(line),
			Character: uint32(character),
		},
		Context: protocol.ReferenceContext{
			IncludeDeclaration: includeDeclaration,
		},
	}, &result, 5*time.Second)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Hover provides hover information at a given position
func (lc *LanguageClient) Hover(uri string, line, character int32) (*protocol.Hover, error) {
	var result protocol.Hover
	err := lc.SendRequest("textDocument/hover", protocol.HoverParams{
		TextDocument: protocol.TextDocumentIdentifier{
			Uri: protocol.DocumentUri(uri),
		},
		Position: protocol.Position{
			Line:      uint32(line),
			Character: uint32(character),
		},
	}, &result, 5*time.Second)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
