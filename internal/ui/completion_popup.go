package ui

import (
	"github.com/gdamore/tcell/v2"
)

// CompletionItem represents a single completion option
type CompletionItem struct {
	Label       string // The text to display
	Detail      string // Additional detail (e.g., type information)
	InsertText  string // The text to insert when selected
	Kind        int    // Type of completion (e.g., 1=Text, 3=Function, 6=Variable)
}

// CompletionPopup displays a list of code completion options
type CompletionPopup struct {
	items         []CompletionItem
	selectedIndex int
	visible       bool
	x, y          int // Position of the popup
	maxHeight     int // Maximum height of the popup
	maxWidth      int // Maximum width of the popup
}

// NewCompletionPopup creates a new completion popup
func NewCompletionPopup() *CompletionPopup {
	return &CompletionPopup{
		items:         []CompletionItem{},
		selectedIndex: 0,
		visible:       false,
		maxHeight:     10,
		maxWidth:      50,
	}
}

// SetItems sets the completion items to display
func (p *CompletionPopup) SetItems(items []CompletionItem) {
	p.items = items
	p.selectedIndex = 0
	if len(items) > 0 {
		p.visible = true
	}
}

// Show displays the popup at the given position
func (p *CompletionPopup) Show(x, y int) {
	p.x = x
	p.y = y
	if len(p.items) > 0 {
		p.visible = true
	}
}

// Hide hides the popup
func (p *CompletionPopup) Hide() {
	p.visible = false
}

// IsVisible returns whether the popup is visible
func (p *CompletionPopup) IsVisible() bool {
	return p.visible
}

// MoveUp moves the selection up
func (p *CompletionPopup) MoveUp() {
	if p.selectedIndex > 0 {
		p.selectedIndex--
	}
}

// MoveDown moves the selection down
func (p *CompletionPopup) MoveDown() {
	if p.selectedIndex < len(p.items)-1 {
		p.selectedIndex++
	}
}

// GetSelectedItem returns the currently selected item
func (p *CompletionPopup) GetSelectedItem() *CompletionItem {
	if p.selectedIndex >= 0 && p.selectedIndex < len(p.items) {
		return &p.items[p.selectedIndex]
	}
	return nil
}

// Render draws the popup on the screen
func (p *CompletionPopup) Render(screen *Screen, styles *StyleConfig) {
	if !p.visible || len(p.items) == 0 {
		return
	}

	// Calculate popup dimensions
	width := p.calculateWidth()
	height := p.calculateHeight()

	// Adjust position if popup would go off screen
	screenWidth, screenHeight := screen.Size()
	if p.x+width > screenWidth {
		p.x = screenWidth - width
	}
	if p.y+height > screenHeight {
		p.y = screenHeight - height
	}

	// Draw popup background with border
	popupStyle := tcell.StyleDefault.Background(tcell.ColorDarkBlue).Foreground(tcell.ColorWhite)
	selectedStyle := tcell.StyleDefault.Background(tcell.ColorBlue).Foreground(tcell.ColorWhite)

	// Draw border
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			ch := ' '
			if y == 0 || y == height-1 {
				// Top and bottom border
				if x == 0 && y == 0 {
					ch = '┌'
				} else if x == width-1 && y == 0 {
					ch = '┐'
				} else if x == 0 && y == height-1 {
					ch = '└'
				} else if x == width-1 && y == height-1 {
					ch = '┘'
				} else {
					ch = '─'
				}
			} else if x == 0 || x == width-1 {
				// Side borders
				ch = '│'
			}
			screen.SetCell(p.x+x, p.y+y, ch, popupStyle)
		}
	}

	// Draw items
	for i, item := range p.items {
		if i >= height-2 {
			break // Don't draw more items than can fit
		}

		style := popupStyle
		if i == p.selectedIndex {
			style = selectedStyle
		}

		// Draw item label
		label := item.Label
		kindStr := getKindString(item.Kind)
		if kindStr != "" {
			label = kindStr + ": " + label
		}

		// Truncate if too long
		maxLabelWidth := width - 4 // Account for borders and padding
		if len(label) > maxLabelWidth {
			label = label[:maxLabelWidth-3] + "..."
		}

		// Clear the line first
		for x := 1; x < width-1; x++ {
			screen.SetCell(p.x+x, p.y+i+1, ' ', style)
		}

		// Draw the label
		screen.SetText(p.x+2, p.y+i+1, label, style)
	}
}

// calculateWidth calculates the popup width based on content
func (p *CompletionPopup) calculateWidth() int {
	maxLen := 20 // Minimum width
	for _, item := range p.items {
		label := item.Label
		kindStr := getKindString(item.Kind)
		if kindStr != "" {
			label = kindStr + ": " + label
		}
		if len(label) > maxLen {
			maxLen = len(label)
		}
	}

	// Add padding and borders
	width := maxLen + 4
	if width > p.maxWidth {
		width = p.maxWidth
	}
	return width
}

// calculateHeight calculates the popup height based on content
func (p *CompletionPopup) calculateHeight() int {
	height := len(p.items) + 2 // Add 2 for borders
	if height > p.maxHeight {
		height = p.maxHeight
	}
	return height
}

// getKindString converts LSP completion kind to string
func getKindString(kind int) string {
	switch kind {
	case 1:
		return "text"
	case 2:
		return "method"
	case 3:
		return "func"
	case 4:
		return "ctor"
	case 5:
		return "field"
	case 6:
		return "var"
	case 7:
		return "class"
	case 8:
		return "iface"
	case 9:
		return "module"
	case 10:
		return "prop"
	case 11:
		return "unit"
	case 12:
		return "value"
	case 13:
		return "enum"
	case 14:
		return "keyword"
	case 15:
		return "snippet"
	case 16:
		return "color"
	case 17:
		return "file"
	case 18:
		return "ref"
	case 19:
		return "folder"
	case 20:
		return "enum"
	case 21:
		return "const"
	case 22:
		return "struct"
	case 23:
		return "event"
	case 24:
		return "op"
	case 25:
		return "type"
	default:
		return ""
	}
}