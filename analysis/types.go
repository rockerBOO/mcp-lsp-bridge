package analysis

import (
	"time"
	"rockerboo/mcp-lsp-bridge/types"
	"github.com/myleshyson/lsprotocol-go/protocol"
)

// Analysis types for project-wide analysis
type AnalysisType string

const (
	WorkspaceAnalysis   AnalysisType = "workspace_analysis"
	SymbolRelationships AnalysisType = "symbol_relationships"
	FileAnalysis        AnalysisType = "file_analysis"
	PatternAnalysis     AnalysisType = "pattern_analysis"
)

// AnalysisRequest represents a request for project analysis
type AnalysisRequest struct {
	Type    AnalysisType
	Target  string
	Scope   string
	Depth   string
	Options map[string]interface{}
}

// AnalysisResult contains the results of a project analysis
type AnalysisResult struct {
	Type     AnalysisType
	Target   string
	Data     interface{}
	Metadata AnalysisMetadata
}

// AnalysisMetadata provides context and statistics about the analysis
type AnalysisMetadata struct {
	Duration      time.Duration
	FilesScanned  int
	SymbolsFound  int
	LanguagesUsed []types.Language
	CacheHits     int
	CacheMisses   int
	Errors        []AnalysisError
}

// AnalysisError represents an error encountered during analysis
type AnalysisError struct {
	Language types.Language
	Message  string
	Type     string // "warning" | "error"
}

// LanguageStats provides statistics about a specific language in the project
type LanguageStats struct {
	FileCount     int
	SymbolCount   int
	Percentage    float64
	ComplexityAvg float64
}

// WorkspaceAnalysisData provides comprehensive analysis of the project's workspace
type WorkspaceAnalysisData struct {
	LanguageDistribution map[types.Language]LanguageStats
	TotalSymbols         int
	TotalFiles           int
	DependencyPatterns   []DependencyPattern
	ArchitecturalHealth  ArchitecturalHealthMetrics
}

// DependencyPattern represents a dependency relationship in the project
type DependencyPattern struct {
	Type        string
	Source      string
	Target      string
	Frequency   int
	IsCircular  bool
	Depth       int
}

// ArchitecturalHealthMetrics provides an overview of the project's architectural health
type ArchitecturalHealthMetrics struct {
	CodeOrganization   HealthScore
	NamingConsistency  HealthScore
	ErrorHandling      HealthScore
	TestCoverage       HealthScore
	Documentation      HealthScore
	OverallScore       HealthScore
}

// HealthScore represents the quality of a specific aspect of the project
type HealthScore struct {
	Score       float64
	Level       string
	Issues      []string
	Suggestions []string
}

// SymbolRelationshipsData provides details about symbol interactions and dependencies
type SymbolRelationshipsData struct {
	Symbol             protocol.WorkspaceSymbol
	Language          types.Language
	References        []protocol.Location
	Definitions       []protocol.Location
	CallHierarchy     []protocol.CallHierarchyItem
	Implementations   []protocol.Location
	TypeHierarchy     []protocol.Location
	UsagePatterns     UsagePatternAnalysis
	RelatedSymbols    []RelatedSymbol
	ImpactAnalysis    ImpactAnalysisData
}

// RelatedSymbol represents a symbol with potential relationships
type RelatedSymbol struct {
	Symbol         protocol.WorkspaceSymbol
	Relationship   string // "multi_location", "inheritance", etc.
	Strength       float64 // 0-1 coupling strength
}

// UsagePatternAnalysis provides insights into how a symbol is used
type UsagePatternAnalysis struct {
	PrimaryUsage    string
	SecondaryUsage  string
	UsageFrequency  int
	CallerPatterns  []CallerPattern
	FileUsageMap    map[string]int
	UsageContexts   map[string]int
}

// CallerPattern describes how a symbol is called
type CallerPattern struct {
	CallerType     string
	CallFrequency  int
	CallContexts   []string
}

// ImpactAnalysisData provides details about potential impacts of changing a symbol
type ImpactAnalysisData struct {
	FilesAffected           int
	AffectedFiles           []string
	CriticalPaths          []string
	BreakingChanges        []BreakingChange
	Dependencies           []string
	Dependents             []string
	RefactoringComplexity  string
}

// BreakingChange represents a potential breaking change when modifying a symbol
type BreakingChange struct {
	Type        string
	Description string
	AffectedFiles []string
	Severity    string
}

// FileAnalysisData provides detailed analysis of a specific file
type FileAnalysisData struct {
	Uri                 string
	Language            types.Language
	Symbols             []protocol.DocumentSymbol
	Complexity          ComplexityMetrics
	ImportExport        ImportExportAnalysis
	CrossFileRelations  []CrossFileRelation
	CodeQuality         CodeQualityMetrics
	Recommendations     []Recommendation
}

// ComplexityMetrics provides metrics about code complexity
type ComplexityMetrics struct {
	TotalLines       int
	FunctionCount    int
	ClassCount       int
	VariableCount    int
	ComplexityScore  float64
	ComplexityLevel  string
}

// ImportExportAnalysis provides details about import and export dependencies
type ImportExportAnalysis struct {
	Imports          []ImportInfo
	Exports          []ExportInfo
	ExternalDeps     []ExternalDependency
	InternalDeps     []InternalDependency
	CircularDeps     []CircularDependency
	UnusedImports    []string
}

// ImportInfo details information about an import
type ImportInfo struct {
	Module      string
	ImportType  string // "default" | "named" | "namespace"
	Usage       []UsageLocation
	IsExternal  bool
}

// ExportInfo details information about an export
type ExportInfo struct {
	Name        string
	ExportType  string // "default" | "named"
	UsedBy      []string
	IsPublic    bool
}

// ExternalDependency represents an external package dependency
type ExternalDependency struct {
	Package     string
	Version     string
	Usage       []UsageLocation
	UpdateAvailable bool
}

// InternalDependency represents a dependency on another file in the project
type InternalDependency struct {
	File        string
	Symbols     []string
	Relationship string // "imports" | "extends" | "uses"
}

// CircularDependency represents a circular import or dependency
type CircularDependency struct {
	Files       []string
	Cycle       []string
	Severity    string // "warning" | "error"
}

// CrossFileRelation describes how files interact
type CrossFileRelation struct {
	TargetFile   string
	RelationType string // "imports" | "references" | "calls"
	Symbols      []string
	Strength     float64 // 0-1 coupling strength
}

// UsageLocation tracks where a symbol is used
type UsageLocation struct {
	File      string
	Line      uint32
	Character uint32
	Context   string
}

// PatternAnalysisData provides insights into code patterns and consistency
type PatternAnalysisData struct {
	PatternType       string
	Scope             string
	ConsistencyScore  float64
	PatternInstances []PatternInstance
	Violations       []PatternViolation
	TrendAnalysis    TrendAnalysis
}

// PatternInstance represents a specific occurrence of a code pattern
type PatternInstance struct {
	Pattern     string
	Location    protocol.Location
	Confidence  float64
	Variations  []string
	Quality     string
}

// PatternViolation represents a deviation from expected code patterns
type PatternViolation struct {
	Expected   string
	Actual     string
	Location   protocol.Location
	Severity   string
	Rule       string
	Suggestion string
}

// TrendAnalysis provides insights into pattern evolution over time
type TrendAnalysis struct {
	Direction   string // "improving" | "stable" | "declining"
	Confidence  float64
	Factors     []string
	Predictions []string
}

// CodeQualityMetrics provides comprehensive code quality assessment
type CodeQualityMetrics struct {
	DuplicationScore     float64
	CohesionScore        float64
	CouplingScore        float64
	MaintainabilityIndex float64
	TestCoverage         float64
	DocumentationScore   float64
}

// Recommendation represents a suggested improvement for code
type Recommendation struct {
	Type        string // "refactor" | "optimize" | "test" | "document"
	Priority    string // "low" | "medium" | "high"
	Description string
	Location    *protocol.Location
	Effort      string // "low" | "medium" | "high"
}