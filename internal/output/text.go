package output

import (
	"io"
	"text/template"

	"github.com/littleworks-inc/cloudcost/pkg/model"
)

// TextFormatter formats reports as plain text
type TextFormatter struct {
	TemplatePath string
}

// NewTextFormatter creates a new text formatter
func NewTextFormatter(templatePath string) *TextFormatter {
	if templatePath == "" {
		templatePath = "configs/templates/text_report.tmpl"
	}
	return &TextFormatter{
		TemplatePath: templatePath,
	}
}

// Format formats the report as text
func (f *TextFormatter) Format(report *model.Report, writer io.Writer) error {
	tmpl, err := template.ParseFiles(f.TemplatePath)
	if err != nil {
		return err
	}

	return tmpl.Execute(writer, report)
}

// GetName returns the name of the formatter
func (f *TextFormatter) GetName() string {
	return "text"
}
