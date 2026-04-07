package template

import (
	"bytes"
	"fmt"
	"text/template"
)

type Engine interface {
	Render(template string, variables map[string]string) (string, error)
	RenderHTML(template string, variables map[string]string) (string, error)
}

type GoTemplateEngine struct{}

func NewGoTemplateEngine() *GoTemplateEngine {
	return &GoTemplateEngine{}
}

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

func (e *GoTemplateEngine) RenderHTML(templateStr string, variables map[string]string) (string, error) {
	return e.Render(templateStr, variables)
}
