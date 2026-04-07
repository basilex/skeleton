// Package template provides template rendering implementations for notifications.
// This package contains the Go template engine for rendering notification content.
package template

import (
	"bytes"
	"fmt"
	"text/template"
)

// Engine defines the interface for template rendering.
type Engine interface {
	Render(template string, variables map[string]string) (string, error)
	RenderHTML(template string, variables map[string]string) (string, error)
}

// GoTemplateEngine implements the Engine interface using Go's text/template package.
type GoTemplateEngine struct{}

// NewGoTemplateEngine creates a new Go template engine.
func NewGoTemplateEngine() *GoTemplateEngine {
	return &GoTemplateEngine{}
}

// Render renders a text template with the provided variables.
func (e *GoTemplateEngine) Render(templateStr string, variables map[string]string) (string, error) {
	tmpl, err := template.New("notification").Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, variables); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	return buf.String(), nil
}

// RenderHTML renders an HTML template with the provided variables.
// Currently delegates to Render as Go templates handle both equally.
func (e *GoTemplateEngine) RenderHTML(templateStr string, variables map[string]string) (string, error) {
	return e.Render(templateStr, variables)
}
