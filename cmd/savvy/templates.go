package main

import (
	"embed"
	"html/template"
)

//go:embed templates
var templatesFS embed.FS

var templates = template.Must(template.ParseFS(templatesFS, "templates/*.tmpl"))
