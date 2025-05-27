package lsp

import (
	"fmt"
	"strings"

	"github.com/dshills/aied/internal/buffer"
	"go.lsp.dev/protocol"
)

// BufferToLSPPosition converts buffer position to LSP position
func BufferToLSPPosition(buf *buffer.Buffer, line, col int) protocol.Position {
	// LSP uses 0-based positions
	return protocol.Position{
		Line:      uint32(line),
		Character: uint32(col),
	}
}

// LSPToBufferPosition converts LSP position to buffer position
func LSPToBufferPosition(pos protocol.Position) (line, col int) {
	// Buffer also uses 0-based positions
	return int(pos.Line), int(pos.Character)
}

// BufferToLSPRange converts buffer range to LSP range
func BufferToLSPRange(startLine, startCol, endLine, endCol int) protocol.Range {
	return protocol.Range{
		Start: protocol.Position{
			Line:      uint32(startLine),
			Character: uint32(startCol),
		},
		End: protocol.Position{
			Line:      uint32(endLine),
			Character: uint32(endCol),
		},
	}
}

// LSPToBufferRange converts LSP range to buffer range
func LSPToBufferRange(r protocol.Range) (startLine, startCol, endLine, endCol int) {
	return int(r.Start.Line), int(r.Start.Character), 
	       int(r.End.Line), int(r.End.Character)
}

// GetBufferContent returns the entire buffer content as a string
func GetBufferContent(buf *buffer.Buffer) string {
	lines := buf.Lines()
	return strings.Join(lines, "\n")
}

// DiagnosticSeverityToString converts diagnostic severity to string
func DiagnosticSeverityToString(severity protocol.DiagnosticSeverity) string {
	switch severity {
	case protocol.DiagnosticSeverityError:
		return "Error"
	case protocol.DiagnosticSeverityWarning:
		return "Warning"
	case protocol.DiagnosticSeverityInformation:
		return "Info"
	case protocol.DiagnosticSeverityHint:
		return "Hint"
	default:
		return "Unknown"
	}
}

// CompletionKindToString converts completion kind to string
func CompletionKindToString(kind protocol.CompletionItemKind) string {
	switch kind {
	case protocol.CompletionItemKindText:
		return "Text"
	case protocol.CompletionItemKindMethod:
		return "Method"
	case protocol.CompletionItemKindFunction:
		return "Function"
	case protocol.CompletionItemKindConstructor:
		return "Constructor"
	case protocol.CompletionItemKindField:
		return "Field"
	case protocol.CompletionItemKindVariable:
		return "Variable"
	case protocol.CompletionItemKindClass:
		return "Class"
	case protocol.CompletionItemKindInterface:
		return "Interface"
	case protocol.CompletionItemKindModule:
		return "Module"
	case protocol.CompletionItemKindProperty:
		return "Property"
	case protocol.CompletionItemKindUnit:
		return "Unit"
	case protocol.CompletionItemKindValue:
		return "Value"
	case protocol.CompletionItemKindEnum:
		return "Enum"
	case protocol.CompletionItemKindKeyword:
		return "Keyword"
	case protocol.CompletionItemKindSnippet:
		return "Snippet"
	case protocol.CompletionItemKindColor:
		return "Color"
	case protocol.CompletionItemKindFile:
		return "File"
	case protocol.CompletionItemKindReference:
		return "Reference"
	case protocol.CompletionItemKindFolder:
		return "Folder"
	case protocol.CompletionItemKindEnumMember:
		return "EnumMember"
	case protocol.CompletionItemKindConstant:
		return "Constant"
	case protocol.CompletionItemKindStruct:
		return "Struct"
	case protocol.CompletionItemKindEvent:
		return "Event"
	case protocol.CompletionItemKindOperator:
		return "Operator"
	case protocol.CompletionItemKindTypeParameter:
		return "TypeParameter"
	default:
		return "Unknown"
	}
}

// GetCompletionDetail formats completion item for display
func GetCompletionDetail(item protocol.CompletionItem) string {
	var parts []string
	
	// Add kind
	if item.Kind != 0 {
		parts = append(parts, CompletionKindToString(item.Kind))
	}
	
	// Add detail if available
	if item.Detail != "" {
		parts = append(parts, item.Detail)
	}
	
	if len(parts) > 0 {
		return strings.Join(parts, " - ")
	}
	
	return ""
}

// ExtractHoverText extracts plain text from hover response
func ExtractHoverText(hover *protocol.Hover) string {
	if hover == nil || hover.Contents == (protocol.MarkupContent{}) {
		return ""
	}
	
	switch hover.Contents.Kind {
	case protocol.PlainText:
		return hover.Contents.Value
	case protocol.Markdown:
		// Simple markdown stripping - just remove code fence markers
		text := hover.Contents.Value
		lines := strings.Split(text, "\n")
		var result []string
		inCodeBlock := false
		
		for _, line := range lines {
			if strings.HasPrefix(line, "```") {
				inCodeBlock = !inCodeBlock
				continue
			}
			if !inCodeBlock || len(result) == 0 {
				result = append(result, line)
			}
		}
		
		return strings.Join(result, "\n")
	default:
		return hover.Contents.Value
	}
}

// FormatDiagnostic formats a diagnostic for display
func FormatDiagnostic(d protocol.Diagnostic) string {
	severity := DiagnosticSeverityToString(d.Severity)
	location := FormatRange(d.Range)
	
	message := d.Message
	if d.Source != "" {
		message = d.Source + ": " + message
	}
	
	if d.Code != nil {
		switch code := d.Code.(type) {
		case string:
			message = message + " [" + code + "]"
		case float64:
			codeStr := fmt.Sprintf("%.0f", code)
			message = message + " [" + codeStr + "]"
		case int:
			message = message + " [" + fmt.Sprintf("%d", code) + "]"
		}
	}
	
	return severity + " at " + location + ": " + message
}

// FormatRange formats a range for display
func FormatRange(r protocol.Range) string {
	if r.Start.Line == r.End.Line {
		return fmt.Sprintf("%d:%d", r.Start.Line+1, r.Start.Character+1)
	}
	return fmt.Sprintf("%d:%d-%d:%d", r.Start.Line+1, r.Start.Character+1, r.End.Line+1, r.End.Character+1)
}

// IsPositionInRange checks if a position is within a range
func IsPositionInRange(pos protocol.Position, r protocol.Range) bool {
	if pos.Line < r.Start.Line || pos.Line > r.End.Line {
		return false
	}
	
	if pos.Line == r.Start.Line && pos.Character < r.Start.Character {
		return false
	}
	
	if pos.Line == r.End.Line && pos.Character > r.End.Character {
		return false
	}
	
	return true
}