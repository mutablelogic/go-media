package cmd

import (
	"fmt"
	"strings"
)

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
