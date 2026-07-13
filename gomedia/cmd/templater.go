package cmd

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"strings"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Templater struct {
	tmpl *template.Template
}

type namedWriter struct {
	io.Writer
	name string
}

func (nw *namedWriter) Name() string {
	return nw.name
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewTemplater(t string) (*Templater, error) {
	// Compile the template
	tmpl, err := template.New("metadata").Funcs(template.FuncMap{
		"base":  filepath.Base,
		"ext":   filepath.Ext,
		"dir":   filepath.Dir,
		"lower": strings.ToLower,
		"upper": strings.ToUpper,
		"name":  func(s string) string { return strings.TrimSuffix(filepath.Base(s), filepath.Ext(s)) },
	}).Parse(t)
	if err != nil {
		return nil, err
	}
	return &Templater{tmpl: tmpl}, nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return the path generated from the template and the given data.
func (t *Templater) Path(data map[string]any) (string, error) {
	var sb strings.Builder
	if err := t.tmpl.Execute(&sb, data); err != nil {
		return "", err
	}
	return sb.String(), nil
}

// Create any directories needed for the given path, then create the file,
// and return the writer for the file.
func (t *Templater) Create(data map[string]any, fn func(w io.Writer) error) error {
	path, err := t.Path(data)
	if err != nil {
		return err
	}

	// Create any intermediate directories needed for the path
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	// Create the file
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write to the file, then either commit or rollback the file depending on whether the write was successful.
	if err := fn(&namedWriter{Writer: f, name: path}); err != nil {
		// Remove the file if it was created and an error occurred during writing.
		if _, statErr := os.Stat(path); statErr == nil {
			err = errors.Join(err, os.Remove(path))
		}
		return err
	}
	return nil
}

// OLD

func PathFromTemplate(template string, args ...any) string {
	// Create a map of template variables to their values
	vars := make(map[string]string)
	for i := 0; i+1 < len(args); i += 2 {
		key := strings.TrimSpace(fmt.Sprintf("%v", args[i]))
		if key == "" {
			continue
		}
		vars[key] = fmt.Sprintf("%v", args[i+1])
	}

	// Expand {key} placeholders via local substitution.
	return expandTemplate(template, vars)
}

func expandTemplate(template string, vars map[string]string) string {
	var b strings.Builder
	b.Grow(len(template))

	for i := 0; i < len(template); {
		if template[i] != '{' {
			b.WriteByte(template[i])
			i++
			continue
		}

		// Find matching '}' for a {key} token.
		j := strings.IndexByte(template[i+1:], '}')
		if j < 0 {
			// No closing brace; keep the rest as-is.
			b.WriteString(template[i:])
			break
		}
		j += i + 1

		key := template[i+1 : j]
		if value, ok := vars[key]; ok {
			b.WriteString(value)
		} else {
			// Unknown key: keep original token unchanged.
			b.WriteString(template[i : j+1])
		}
		i = j + 1
	}

	return b.String()
}
