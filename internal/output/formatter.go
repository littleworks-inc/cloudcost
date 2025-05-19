package output

import (
	"io"

	"github.com/littleworks-inc/cloudcost/pkg/model"
)

// Formatter defines the interface for report formatters
type Formatter interface {
	// Format formats a report and writes it to the given writer
	Format(report *model.Report, writer io.Writer) error

	// GetName returns the name of the formatter (e.g., "text", "json")
	GetName() string
}

// FormatterRegistry is a registry of formatters
type FormatterRegistry struct {
	Formatters map[string]Formatter
}

// NewFormatterRegistry creates a new formatter registry
func NewFormatterRegistry() *FormatterRegistry {
	return &FormatterRegistry{
		Formatters: make(map[string]Formatter),
	}
}

// RegisterFormatter registers a formatter
func (r *FormatterRegistry) RegisterFormatter(formatter Formatter) {
	r.Formatters[formatter.GetName()] = formatter
}

// GetFormatter returns a formatter by name
func (r *FormatterRegistry) GetFormatter(name string) (Formatter, bool) {
	formatter, ok := r.Formatters[name]
	return formatter, ok
}
