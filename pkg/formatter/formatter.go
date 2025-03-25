package formatter

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Format formats the given data based on the output format
func Format(data interface{}, format string) (string, error) {
	switch strings.ToLower(format) {
	case "json", "j":
		return formatJSON(data, false)
	case "pretty", "p":
		return formatJSON(data, true)
	default:
		return formatJSON(data, true)
	}
}

// formatJSON formats the data as JSON with optional pretty printing
func formatJSON(data interface{}, pretty bool) (string, error) {
	var output []byte
	var err error

	if pretty {
		output, err = json.MarshalIndent(data, "", "  ")
	} else {
		output, err = json.Marshal(data)
	}

	if err != nil {
		return "", fmt.Errorf("error formatting JSON: %w", err)
	}

	return string(output), nil
} 